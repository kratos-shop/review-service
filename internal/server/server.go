package server

import (
	"review-service/internal/conf"

	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRegistry, NewGRPCServer, NewHTTPServer)

func NewRegistry(conf *conf.Registry) *consul.Registry {
	// 配置 consul
	cfg := api.DefaultConfig()
	cfg.Address = conf.Consul.Address
	cfg.Scheme = conf.Consul.Scheme
	// new consul client
	log.Infof("new consul client: %v", cfg)
	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatalf("new consul client failed: %v", err)
		panic(err)
	}
	// new reg with consul client
	reg := consul.New(client)
	log.Infof("new consul registry: %v", reg)
	return reg
}
