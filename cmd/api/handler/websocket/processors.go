package websocket

import (
	"github.com/NobleScope/noble-indexer/cmd/api/handler/responses"
	"github.com/NobleScope/noble-indexer/internal/storage"
)

func blockProcessor(block storage.Block) Notification[*responses.Block] {
	response := responses.NewBlock(block)
	return NewBlockNotification(response)
}

func headProcessor(state storage.State) Notification[*responses.State] {
	response := responses.NewState(state)
	return NewStateNotification(response)
}
