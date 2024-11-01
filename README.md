# go-eventsourcing-pgx

This project is a Go-based event sourcing integration for [hallgren/eventsourcing](https://github.com/hallgren/eventsourcing) using PostgreSQL with the `pgx` driver.  
It provides an event store and a snapshot store implementation for the `eventsourcing` package. 
Also included is a migration tool to create the necessary tables in the database and a logging package for logging pgx and migration errors.



## Table of Contents

- [Installation](#installation)
- [Usage](#usage)

## Installation

To install the dependencies, run:

```sh
go get github.com/CentralConcept/go-eventsourcing-pgx/eventstore/pgx
go get github.com/CentralConcept/go-eventsourcing-pgx/snapshotstore/pgx
```

## Usage
Initializing the Database
To initialize the database, create a pgxpool.Pool and pass it to the NewPostgresPgx function:

```go
import (
    eventstore "github.com/CentralConcept/go-eventsourcing-pgx/eventstore/pgx"
    snapshotstore "github.com/CentralConcept/go-eventsourcing-pgx/snapshotstore/pgx"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
   dba, dbaErr := pgxpool.New(ctx, "user=postgres dbname=eventstore sslmode=disable password=mysecretpassword host=localhost")
   if dbaErr != nil {
        logger.Fatal(err)
   }
   sqlEventstore := pgx.NewPostgresPgx(dba)
   sqlSnapshotstore := snapshotstore.NewPostgresPgx(dba)
   migrateErr := sqlEventstore.Migrate(dba)
   if migrateErr != nil {
        logger.Fatal(migrateErr)
   }
   eventRepo := eventsourcing.NewEventRepository(sqlEventstore)
   snapshotRepo := eventsourcing.NewSnapshotRepository(sqlSnapshotstore, eventRepo)

// Use the REPO...
}
```
