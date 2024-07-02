migrations: [
    {
        timestamp: 1719885656
        name:      "add_example_route"
        up: {
            "/services/example_service/routes": {
                method: "POST"
                body: {
                    paths: [
                        "/mock",
                    ]
                    name: "example_route"
                }
            }
        }
        down: {
            "/services/example_service/routes/example_route": {
                method: "DELETE"
            }
        }
    }
]
