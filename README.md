# Krakend Http Cache

Krakend plugin for caching backend responses

## Client configuration

```json
...
"plugin/http-client": {
    "name": "onliner/krakend-http-cache",
    "onliner/krakend-http-cache": {
        "ttl": 180,
        "connection": "redis",
        "headers": []
    }
}
...
```

`ttl` - cache ttl in seconds
`connection` - name of cache connection
`headers` - headers used for cache key

## Cache connections

```json
...
"plugin/http-server": {
    "name": ["onliner/krakend-http-cache"],
    "onliner/krakend-http-cache": {
        "connections": {
            "<connection_name>": {
                "driver": "<connection_driver>",
                "options": {}
            }
        }
    }
}
...
```

`connections` - list of named cache connections

### Supported cache drivers

- memory
- redis

### Redis connection options

- `addr` - host:port address (default: localhost:6379)
- `user` - username to authenticate the current connection (default: "")
- `pass` - password (default: "")
- `db` - redis db (default: 0)
- `pool_size` - maximum number of socket connections (default: 10)

## Сonfiguration example

```json
{
    "version": 3,
    "name": "KrakenD API Gateway",
    "plugin": {
        "pattern": ".so",
        "folder": "/etc/krakend/plugins"
    },
    "endpoints": [
        {
            "endpoint": "/hello",
            "backend": [
                {
                    "host": ["http://api:8080"],
                    "url_pattern": "/hello",
                    "extra_config": {
                        "plugin/http-client": {
                            "name": "onliner/krakend-http-cache",
                            "onliner/krakend-http-cache": {
                                "ttl": 180,
                                "connection": "redis",
                                "headers": ["X-Custom-Headers"]
                            }
                        }
                    }
                }
            ]
        }
    ],
    "extra_config": {
        "plugin/http-server": {
            "name": ["onliner/krakend-http-cache"],
            "onliner/krakend-http-cache": {
                "connections": {
                    "inmemory": {
                        "driver": "memory"
                    },
                    "redis": {
                        "driver": "redis",
                        "options": {
                            "addr": "127.0.0.1:6379",
                            "user": "root",
                            "pass": "123qwe",
                            "db": 1,
                            "pool_size": 5
                        }
                    }
                }
            }
        }
    }
}
```
