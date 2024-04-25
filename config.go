package main

import (
	"errors"
	"net/textproto"
	"slices"

	"github.com/mitchellh/mapstructure"
)

var Namespace = "onliner/krakend-http-cache"

type ClientConfig struct {
	Ttl     uint64
	Conn    string `mapstructure:"connection"`
	Headers []string
}

type SrvConfig struct {
	Conns map[string]CacheCnf `mapstructure:"connections"`
}

func NewClientConfig(input map[string]interface{}) (*ClientConfig, error) {
	var config ClientConfig
	err := parseConfig(input, &config)
	if err != nil {
		return nil, err
	}

	config.Headers = normalizeHeaders(config.Headers)

	return &config, nil
}

func NewSrvConfig(input map[string]interface{}) (*SrvConfig, error) {
	var config SrvConfig
	err := parseConfig(input, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func parseConfig(input map[string]interface{}, output interface{}) error {
	cnf, ok := input[Namespace].(map[string]interface{})
	if !ok {
		return errors.New("configuration not found")
	}

	err := mapstructure.WeakDecode(cnf, &output)
	if err != nil {
		return err
	}

	return nil
}

func normalizeHeaders(headers []string) []string {
	res := make(map[string]bool)
	for _, h := range headers {
		res[textproto.CanonicalMIMEHeaderKey(h)] = true
	}

	var values []string
	for h := range res {
		values = append(values, h)
	}

	slices.Sort(values)

	return values
}
