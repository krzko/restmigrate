// package migrations  // Optional: uncomment if you want to use packages

migrations: [
    {
        timestamp: 1719748890
        name:      "add_admins"
        up: {
            // You can define multiple actions here
            "/api/v1/endpoint1": {
                method: "POST"
                body: {
                    key1: "value1"
                }
            }
            "/api/v1/endpoint2": {
                method: "PUT"
                body: {
                    key2: "value2"
                }
            }
        }
        down: {
            // Corresponding down actions
            "/api/v1/endpoint2": {
                method: "DELETE"
            }
            "/api/v1/endpoint1": {
                method: "DELETE"
            }
        }
    }
]
