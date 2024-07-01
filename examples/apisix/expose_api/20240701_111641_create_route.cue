migrations: [
    {
        timestamp: 1719796601
        name:      "create_route"
        up: {
            "/apisix/admin/routes/1": {
                method: "PUT"
                body: {
                    methods: ["GET"]
                    host:        "example.com"
                    uri:         "/anything/*"
                    upstream_id: "1"
                }
            }
        }
        down: {
            "/apisix/admin/routes/1": {
                method: "DELETE"
            }
        }
    }
]
