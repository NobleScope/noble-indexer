package bus

import (
	"context"
	"sync"

	"github.com/NobleScope/noble-indexer/internal/storage"
	"github.com/dipdup-io/workerpool"
	"github.com/goccy/go-json"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	listener storage.Listener

	mx        *sync.RWMutex
	observers []*Observer

	g workerpool.Group
}

func NewDispatcher(factory storage.ListenerFactory) (*Dispatcher, error) {
	if factory == nil {
		return nil, errors.New("nil listener factory")
	}
	listener := factory.CreateListener()
	return &Dispatcher{
		listener:  listener,
		observers: make([]*Observer, 0),
		mx:        new(sync.RWMutex),
		g:         workerpool.NewGroup(),
	}, nil
}

func (d *Dispatcher) Observe(channels ...string) *Observer {
	if observer := NewObserver(channels...); observer != nil {
		d.mx.Lock()
		d.observers = append(d.observers, observer)
		d.mx.Unlock()
		return observer
	}

	return nil
}

func (d *Dispatcher) Start(ctx context.Context) {
	if err := d.listener.Subscribe(ctx, storage.ChannelHead, storage.ChannelBlock); err != nil {
		log.Err(err).Msg("subscribe on postgres notifications")
		return
	}
	d.g.GoCtx(ctx, d.listen)
}

func (d *Dispatcher) Close() error {
	d.g.Wait()

	d.mx.Lock()
	for i := range d.observers {
		if err := d.observers[i].Close(); err != nil {
			return err
		}
	}
	d.mx.Unlock()

	return nil
}

func (d *Dispatcher) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case notification, ok := <-d.listener.Listen():
			if !ok {
				return
			}
			if notification == nil {
				log.Warn().Msg("nil notification")
				continue
			}
			if err := d.handleNotification(notification); err != nil {
				log.Err(err).Str("channel", notification.Channel).Msg("handle notification")
			}
		}
	}
}

func (d *Dispatcher) handleNotification(notification *pq.Notification) error {
	switch notification.Channel {
	case storage.ChannelHead:
		return d.handleState(notification.Extra)
	case storage.ChannelBlock:
		return d.handleBlock(notification.Extra)
	default:
		return errors.Errorf("unknown channel name: %s", notification.Channel)
	}
}

func (d *Dispatcher) handleBlock(payload string) error {
	block := new(storage.Block)
	if err := json.Unmarshal([]byte(payload), block); err != nil {
		return err
	}

	d.mx.RLock()
	for i := range d.observers {
		d.observers[i].notifyBlocks(block)
	}
	d.mx.RUnlock()
	return nil
}

func (d *Dispatcher) handleState(payload string) error {
	var state storage.State
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return err
	}

	d.mx.RLock()
	for i := range d.observers {
		d.observers[i].notifyState(&state)
	}
	d.mx.RUnlock()
	return nil
}
