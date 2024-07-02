# restmigrate

`restmigrate` is a Go-based tool for managing and applying configuration changes to REST APIs in a systematic, version-controlled manner. It uses a migration-like approach similar to database schema migrations, allowing you to define, apply, and revert changes to your REST API configurations.

## Features

- Create, apply, and revert REST API configuration changes
- Support for multiple API gateways or generic API endpoints (e.g., Kong, APISIX, Generic)
- [CUE](https://cuelang.org/) language for defining migrations
- [OpenTelemetry](https://opentelemetry.io/) traces integration for observability

### OpenTelemetry

`restmigrate` supports OpenTelemetry for distributed tracing, follow the [configuration](#configuration) section to enable it.

<img
  src="/assets/images/trace.png"
  alt="Distributed trace"
  title="Distributed trace"
  style="display: inline-block; margin: 0 auto; max-width: 300px">

## Installation

### brew

Install [brew](https://brew.sh/) and then run:

```sh
brew install krzko/tap/restmigrate
```

### Download Binary

Download the latest version from the [Releases](https://github.com/krzko/restmigrate/releases) page.

## Commands

* `create`: Create a new migration file
* `up`: Apply pending migrations
* `down`: Revert the last applied migration (use `--all` to revert all)
* `list`: Display applied migrations

## Configuration

Set these environment variables to configure `restmigrate`:

* `OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry exporter endpoint
* `OTEL_EXPORTER_OTLP_INSECURE`: Set to "true" for insecure connection
* `OTEL_SDK_ENABLED`: Set to "true" to enable OpenTelemetry (disabled by default)

## Usage

### Creating a new migration

To create a new migration file:

```bash
restmigrate create <migration_name>
```

This will create a new CUE file in the `migrations` directory with a timestamp prefix.

### Applying migrations

To apply all pending migrations, `--token` and `--type` are optional if the API does not require authentication:

```bash
restmigrate up --url <api_base_url> --token <api_token> --type <type>
```

### Reverting the last migration

To revert the most recently applied migration:

```bash
restmigrate down --url <api_base_url> --token <api_token> --type <type>
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
