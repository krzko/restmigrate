migrations: [
    {
        timestamp: 1719795943
        name:      "create_upstream"
        up: {
            "/apisix/admin/upstreams/1": {
                method: "PUT"
                body: {
                    type: "roundrobin"
                    nodes: "httpbin.org:80": 1
                }
            }
        }
        down: {
            "/apisix/admin/upstreams/1": {
                method: "DELETE"
            }
        }
    }
]
