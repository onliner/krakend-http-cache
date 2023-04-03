# krakend-http-cache

## Krakend configuration example

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
                                "connection": "redis"
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
                    "driver": "inmemory"
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
