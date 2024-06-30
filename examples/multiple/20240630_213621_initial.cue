// package migrations  // Optional

migrations: [
    {
        timestamp: 1682434822
        name:      "add_user_and_profile"
        up: {
            "/api/v1/users": {
                method: "POST"
                body: {
                    name: "John Doe"
                    email: "john@example.com"
                }
            }
            "/api/v1/profiles": {
                method: "POST"
                body: {
                    user_id: 1
                    bio: "Software Developer"
                }
            }
        }
        down: {
            "/api/v1/profiles": {
                method: "DELETE"
                body: {
                    user_id: 1
                }
            }
            "/api/v1/users": {
                method: "DELETE"
                body: {
                    id: 1
                }
            }
        }
    }
]