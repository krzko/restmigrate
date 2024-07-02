migrations: [
    {
        timestamp: 1719885375
        name:      "update_retries_example_service"
        up: {
            "/services/example_service": {
                method: "PATCH"
                body: {
                    retries:         10
                }
            }
        }
        down: {
            "/services/example_service": {
                method: "PATCH"
                body: {
                    retries:         5
                }
            }
        }
    }
]
