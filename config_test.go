package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientConfig(t *testing.T) {
	raw := []byte(`{
		"onliner/krakend-http-cache": {
			"ttl": 1,
			"connection": "memory"
		}
	}`)
	var input map[string]interface{}
	json.Unmarshal(raw, &input)
	cnf, err := NewClientConfig(input)

	assert.Nil(t, err)
	assert.EqualValues(t, 1, cnf.Ttl)
	assert.Equal(t, "memory", cnf.Conn)
	assert.Nil(t, cnf.Headers)
}

func TestClientConfigHeadersNormalize(t *testing.T) {
	raw := []byte(`{
		"onliner/krakend-http-cache": {
			"ttl": 1,
			"connection": "memory",
			"headers": ["x-custom-header", "Content-type", "X-custom-header"]
		}
	}`)
	var input map[string]interface{}
	json.Unmarshal(raw, &input)
	cnf, err := NewClientConfig(input)

	assert.Nil(t, err)
	assert.EqualValues(t, 1, cnf.Ttl)
	assert.Equal(t, "memory", cnf.Conn)
	assert.Equal(t, []string{"Content-Type", "X-Custom-Header"}, cnf.Headers)
}

func TestNewSrvConfig(t *testing.T) {
	raw := []byte(`{
		"onliner/krakend-http-cache": {
			"connections": {
				"inmemory": {
					"driver": "memory"
				},
				"redis": {
					"driver": "redis",
					"options": {
						"addr": "127.0.0.1:6379",
						"pass": "123qwe",
						"db": 1
					}
				}
			}
		}
	}`)

	var input map[string]interface{}
	json.Unmarshal(raw, &input)
	cnf, err := NewSrvConfig(input)

	assert.Nil(t, err)

	expected := map[string]CacheCnf{
		"inmemory": CacheCnf{
			Driver("memory"),
			nil,
		},
		"redis": CacheCnf{
			Driver("redis"),
			map[string]interface{}{
				"addr": "127.0.0.1:6379",
				"pass": "123qwe",
				"db":   float64(1),
			},
		},
	}

	assert.Equal(t, expected, cnf.Conns)
}
