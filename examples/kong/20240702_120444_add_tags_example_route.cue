migrations: [
    {
        timestamp: 1719885884
        name:      "add_tags_example_route"
        up: {
            "/services/example_service/routes/example_route": {
                method: "PATCH"
                body: {
                    tags: [
                        "tutorial",
                    ]
                }
            }
        }
        down: {
            "/services/example_service/routes/example_route": {
                method: "PATCH"
                body: {
                    tags: []
                }
            }
        }
    }
]
