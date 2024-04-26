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

	if err := parseConfig(input, &config); err != nil {
		return nil, err
	}

	config.Headers = normalizeHeaders(config.Headers)

	return &config, nil
}

func NewSrvConfig(input map[string]interface{}) (*SrvConfig, error) {
	var config SrvConfig

	if err := parseConfig(input, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func parseConfig(input map[string]interface{}, output interface{}) error {
	cnf, ok := input[Namespace].(map[string]interface{})
	if !ok {
		return errors.New("configuration not found")
	}

	return mapstructure.WeakDecode(cnf, &output)
}

func normalizeHeaders(headers []string) []string {
	seen := make(map[string]bool)
	var values []string

	for _, h := range headers {
		h = textproto.CanonicalMIMEHeaderKey(h)
		if _, ok := seen[h]; !ok {
			seen[h] = true
			values = append(values, h)
		}
		seen[textproto.CanonicalMIMEHeaderKey(h)] = true
	}

	slices.Sort(values)

	return values
}
