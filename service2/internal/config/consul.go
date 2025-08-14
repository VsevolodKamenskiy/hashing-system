package config

import (
	"context"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

type AppConfig struct {
	HasherPort string
	DBDSN      string
	HTTPPort   string
}

func Load(ctx context.Context, consulAddr string) (*AppConfig, error) {

	cfg := &AppConfig{}

	if consulAddr == "" {
		return nil, errors.New("consulAddr is empty")
	}

	// клиент Consul
	consulConf := consulapi.DefaultConfig()
	consulConf.Address = consulAddr
	client, err := consulapi.NewClient(consulConf)

	if err != nil {
		return nil, errors.Wrap(err, "consul client failed")
	}

	kv := client.KV()

	// helper: get string value by key with fallback to current value
	getKV := func(key, current string) string {
		opts := (&consulapi.QueryOptions{}).WithContext(ctx)
		pair, _, err := kv.Get(key, opts)
		if err != nil || pair == nil || len(pair.Value) == 0 {
			return current
		}
		return string(pair.Value)
	}

	cfg.HTTPPort = getKV("config/service2/http_port", cfg.HTTPPort)
	cfg.DBDSN = getKV("config/service2/db_dsn", cfg.DBDSN)
	cfg.HasherPort = getKV("config/grpc_port", cfg.HasherPort)

	return cfg, nil
}
