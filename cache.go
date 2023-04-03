package main

import (
	"errors"
	"fmt"

	rd "github.com/go-redis/redis/v8"

	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/redis"
	"github.com/faabiosr/cachego/sync"
)

type pool map[string]cachego.Cache
type Driver string

type ConnCnf struct {
	Driver
	Opts map[string]interface{} `mapstructure:"options"`
}

var cachePool = make(pool)

func RegisterCacheConn(name string, cnf *ConnCnf) error {
	switch cnf.Driver {
	case "redis":
		opt, err := buildRedisCnf(cnf.Opts)
		if err != nil {
			return err
		}

		cachePool[name] = redis.New(rd.NewClient(opt))
	case "inmemory":
		cachePool[name] = sync.New()
	default:
		return fmt.Errorf("Cannot create connection %s for %s", name, cnf.Driver)
	}

	return nil
}

func GetCacheConn(name string) cachego.Cache {
	return cachePool[name]
}

func buildRedisCnf(input map[string]interface{}) (*rd.Options, error) {
	var result rd.Options

	addr, ok := input["addr"].(string)
	if !ok || addr == "" {
		return nil, errors.New("missing or empty address")
	}

	result.Addr = addr

	if user, ok := input["user"].(string); ok {
		result.Username = user
	}

	if pass, ok := input["pass"].(string); ok {
		result.Password = pass
	}

	if db, ok := input["db"].(float64); ok {
		result.DB = int(db)
	}

	if poolSize, ok := input["pool_size"].(float64); ok {
		result.PoolSize = int(poolSize)
	}

	return &result, nil
}
