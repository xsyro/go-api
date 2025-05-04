# Go API

Simple Go backend service with Chi and Postgres with SQLc.

## Getting started

Here's how to run the service on your local machine for development and testing purposes.

### Project structure

- **cmd** - contains main files for casha core.
- **config** - hold main app configuration.
- **docs** - swagger documentation.
- **internal** - internal modules
- **internal/db** - migration's sql files, initialization, config, types and models
- **internal/db** - generated mocks

### Stack:

- **golang go1.22**
- **postgresql 15**
- **pgx v5**

### Docker compose stack:

To start up project services, and apply all migrations:

```shell
$ make up
```

To start up project services, and apply all migrations with log:

```shell
$ make up_log
```

Shut down all services:

```shell
$ make down
```

### Running the tests

```shell
$ make test
```

To run unit tests for all subdirectories.

#### Mocking

Project uses [gomock](https://github.com/uber-go/mock) and [testify](https://github.com/stretchr/testify)

You can regenerate mocking structures by using:

```shell
$ go generate ./...
```

## Built With

- [chi framework](https://github.com/go-chi/chi)
- [sqlc](https://github.com/sqlc-dev/sqlc)
- [pgx](https://github.com/jackc/pgx)
- [docker](https://www.docker.com)
- [docker-compose](https://docs.docker.com/compose)
- [make](https://www.gnu.org/s/make/manual/make.html)

# Other helpful resource

- Go - https://go.dev/learn/
- Sqlc - https://conroy.org/introducing-sqlc
- Check `NOTE.md` running docker-compose in development mode.
- Pgx - https://www.youtube.com/watch?v=sXMSWhcHCf8&t=2s
