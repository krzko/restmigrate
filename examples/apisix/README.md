# APISIX Example

This example demonstrates how to use `restmigrate` with APISIX.

## Up

To apply your REST schema, run the following command:

```bash
restmigrate up --base-url http://localhost:9180 --api-key "xxxxxxxxxx" --type apisix
```

## Down

To remove your REST schema, run the following command:

```bash
restmigrate down --base-url http://localhost:9180 --api-key "xxxxxxxxxx" --type apisix
```

You can remove all the schema changes by passing the `--all` flag:

```bash
restmigrate down --base-url http://localhost:9180 --api-key "xxxxxxxxxx" --type apisix --all
```