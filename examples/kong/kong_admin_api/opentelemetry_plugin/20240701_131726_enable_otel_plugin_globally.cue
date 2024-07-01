migrations: [
    {
        timestamp: 1719803846
        name:      "enable_otel_plugin_globally"
        up: {
            "/plugins/": {
                method: "POST"
                body: {
                    name: "opentelemetry"
                    config: {
                        endpoint: "http://opentelemetry.collector:4318/v1/traces"
                        headers: "X-Auth-Token": "secret-token"
                    }
                }
            }
        }
        down: {
            "/plugins/9175fd5b-e364-4759-bbe2-e7d00da5edd2": {
                method: "DELETE"
            }
        }
    }
]
