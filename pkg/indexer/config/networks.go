package config

import "github.com/cockroachdb/errors"

type Network struct {
	PrecompiledContracts []string `validate:"omitempty,dive,eth_addr" yaml:"precompiled_contracts,omitempty"`
}

type NetworksConfig map[string]Network

func (nc NetworksConfig) Get(network string) (Network, error) {
	if netCfg, ok := nc[network]; ok {
		return netCfg, nil
	}
	return Network{}, errors.Errorf("network %s config not found", network)
}

func (nc NetworksConfig) Substitute() error {
	return nil
}
