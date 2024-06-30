# restmigrate

`restmigrate` is a Go-based tool for managing and applying configuration changes to REST APIs in a systematic, version-controlled manner. It uses a migration-like approach similar to database schema migrations, allowing you to define, apply, and revert changes to your REST API configurations.

## Features

- Create, apply, and revert REST API configuration changes
- Use CUE language for flexible and type-safe configuration definitions
- Track applied migrations with a local state file
- CLI interface for easy usage

## Installation

To install RestMigrate, make sure you have Go installed on your system, then run:

```bash
go install github.com/krzo/restmigrate/cmd/restmigrate@latest
```

## Usage

### Creating a new migration

To create a new migration file:

```bash
restmigrate create <migration_name>
```

This will create a new CUE file in the `migrations` directory with a timestamp prefix.

### Applying migrations

To apply all pending migrations:

```bash
restmigrate up --url <api_base_url> --token <api_token>
```

### Reverting the last migration

To revert the most recently applied migration:

```bash
restmigrate down --url <api_base_url> --token <api_token>
```

## Migration File Format

Migration files are written in CUE and should follow this structure:

```cue
package migrations

migration: {
    timestamp: 1625097600  // Unix timestamp
    name:      "add_new_endpoint"
    up: {
        "/api/v1/new_endpoint": {
            method: "POST"
            body: {
                // Define the request body here
            }
        }
    }
    down: {
        "/api/v1/new_endpoint": {
            method: "DELETE"
        }
    }
}
```

## Development

To set up the development environment:

1. Clone the repository:

```bash
git clone https://github.com/krzo/restmigrate.git
```

2. Change to the project directory:

```bash
cd restmigrate
```

3. Install dependencies:

```bash
go mod tidy
```

4. Build the project:

```bash
make build
```
