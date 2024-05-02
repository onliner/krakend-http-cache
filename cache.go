package main

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	rd "github.com/redis/go-redis/v9"

	"github.com/faabiosr/cachego/redis"
	"github.com/faabiosr/cachego/sync"
)

type (
	Driver string
	pool   map[string]Cache

	Cache interface {
		Fetch(key string) (string, error)
		Save(key string, value string, lifeTime time.Duration) error
		Flush() error
	}

	CacheCnf struct {
		Driver
		Opts map[string]interface{} `mapstructure:"options"`
	}

	RedisCnf struct {
		Addr     string
		User     string
		Pass     string
		DB       int
		PoolSize int
	}
)

var cachePool = make(pool)

func RegisterCache(name string, cnf *CacheCnf) error {
	switch cnf.Driver {
	case "redis":
		opt, err := buildRedisCnf(cnf.Opts)
		if err != nil {
			return err
		}

		cachePool[name] = redis.New(rd.NewClient(opt))
	case "memory":
		cachePool[name] = sync.New()
	default:
		return fmt.Errorf("cannot create connection %s for %s", name, cnf.Driver)
	}

	return nil
}

func GetCache(name string) Cache {
	return cachePool[name]
}

func buildRedisCnf(input map[string]interface{}) (*rd.Options, error) {
	var config RedisCnf

	if err := mapstructure.WeakDecode(input, &config); err != nil {
		return nil, err
	}

	return &rd.Options{
		Addr:     config.Addr,
		Username: config.User,
		Password: config.Pass,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	}, nil
}
