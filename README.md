# go-eventsourcing-pgx

This project is a Go-based event sourcing library using PostgreSQL with the `pgx` driver.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)

## Installation

To install the dependencies, run:

```sh
go get github.com/CentralConcept/go-eventsourcing-pgx/eventstore/pgx
```

## Usage
Initializing the Database
To initialize the database, create a pgxpool.Pool and pass it to the NewPostgresPgx function:

```go
import (
    "github.com/CentralConcept/go-eventsourcing-pgx/eventstore/pgx"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
   dba, dbaErr := pgxpool.New(ctx, "user=postgres dbname=eventstore sslmode=disable password=mysecretpassword host=localhost")
   if dbaErr != nil {
        logger.Fatal(err)
   }
   sqlEventstore := pgx.NewPostgresPgx(dba)
   migrateErr := sqlEventstore.Migrate(dba)
   if migrateErr != nil {
        logger.Fatal(migrateErr)
   }
   repo := eventsourcing.NewEventRepository(sqlEventstore)
      // Use the REPO...
}
```
