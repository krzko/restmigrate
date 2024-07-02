migrations: [
    {
        timestamp: 1719880244
        name:      "create_example_service"
        up: {
            "/services": {
                method: "POST"
                body: {
                    name:            "example_service"
                    retries:         5
                    protocol:        "http"
                    host:            "example.com"
                    port:            80
                    path:            "/some_api"
                    connect_timeout: 6000
                    write_timeout:   6000
                    read_timeout:    6000
                    tags: [
                        "user-level",
                    ]
                    enabled: true
                }
            }
        }
        down: {
            "/services/example_service": {
                method: "DELETE"
            }
        }
    }
]
