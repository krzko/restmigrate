# Kong Example

This example demonstrates how to use `restmigrate` with Kong. Ensure you run the database version of Kong on your local machine before running the commands below.

## Up

To apply your REST schema, run the following command:

```bash
restmigrate up --base-url http://localhost:8001
```

## Down

To remove your REST schema, run the following command:

```bash
restmigrate down --base-url http://localhost:8001
```

You can remove all the schema changes by passing the `--all` flag:

```bash
restmigrate up --base-url http://localhost:8001 --all
```