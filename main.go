package main

import (
	"context"
	"fmt"
	"net/http"
)

type registerer string

var ClientRegisterer = registerer(Namespace)
var HandlerRegisterer = registerer(Namespace)
var logger Logger = nil
var cacheHandler *CacheHandler

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", Namespace))
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(Namespace, r.registerClients)
}

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(Namespace, r.registerHandlers)
}

func (r registerer) registerClients(_ context.Context, extra map[string]interface{}) (http.Handler, error) {
	config, err := NewClientConfig(extra)
	if err != nil {
		return nil, err
	}

	if cacheHandler == nil {
		cacheHandler = NewCacheHandler(http.DefaultClient, logger)
	}

	return cacheHandler.Handle(config), nil
}

func (r registerer) registerHandlers(_ context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	config, err := NewSrvConfig(extra)
	if err != nil {
		return h, err
	}

	for name, cacheCnf := range config.Conns {
		if err = RegisterCache(name, &cacheCnf); err != nil {
			return h, err
		}
	}

	return h, nil
}

func main() {}
