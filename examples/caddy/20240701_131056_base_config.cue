migrations: [
    {
        timestamp: 1719803456
        name:      "base_config"
        up: {
            "/load": {
                method: "PUT"
                body: {
                    apps: layer4: servers: socks: {
                        listen: [
                            "0.0.0.0:3000",
                        ]
                        routes: [{
                            match: [
                                {
                                    socks5: {}
                                }]
                            handle: [{
                                handler: "proxy"
                                upstreams: [{
                                    dial: ["localhost:9050"]
                                }]
                            }]
                        }, {
                            match: [
                                {
                                    http: []
                                },
                            ]
                            handle: [{
                                handler: "proxy"
                                upstreams: [{
                                    dial: ["localhost:8080"]
                                }]
                            }]
                        }]
                    }
                }
            }
        }
        down: {
            "/config": {
                method: "DELETE"
            }
        }
    }
]
