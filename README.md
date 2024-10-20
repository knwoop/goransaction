# goransaction

This Go package provides a simple transactional database client that supports both primary and replica databases. It allows for flexible transaction management with configurable options like read-only transactions and the ability to specify whether to use a primary or replica database.

## Features

- Support for both primary and replica databases.
- Transactions with options for read-only or write operations.
- Easy integration with Go’s `context` for managing transactions.
- Functional options for configuring transaction behavior.
- Reuse existing transactions or create new ones based on context.

## Usage

### Creating a Client
To create a new `Client` instance, pass in the primary and replica database connections. If no replicas are provided, the primary connection will be used as the replica.

```go
primaryDB, _ := sql.Open("driver", "primary-dsn")
replicaDB1, _ := sql.Open("driver", "replica1-dsn")
replicaDB2, _ := sql.Open("driver", "replica2-dsn")

client, err := NewClient(context.Background(), primaryDB, []*sql.DB{replicaDB1, replicaDB2})
if err != nil {
    // error handling
}
```

### Running Transactions
You can execute a function within a transaction using the `RunTransaction` function. This ensures that if no active transaction exists, a new one is started.
It also supports options for read-only transactions or enforcing the use of the primary database.

``` go
if err := RunTransaction(context.Background(), client, func(ctx context.Context, conn Conn) error {
    // Perform database operations within the transaction.
    _, err := conn.ExecContext(ctx, "INSERT INTO example (col) VALUES (?)", "value")
    return err
}); err != nil {
    // error handling
}
```

### Reusing an Existing Transaction
The `RunInTransaction` function allows you to execute a function inside an existing transaction if one is already available in the context. If no transaction is present in the context, it creates a new transaction.

``` go
if err := client.RunInTransaction(ctx, func(ctx context.Context, conn Conn) error {
    _, err := txn.ExecContext(ctx, "UPDATE example SET col = ? WHERE id = ?", "new_value", 1)
    if err != nil {
        // error handling
    }
    return nil
}); err != nil {
    // error handling
}
```

### Running Operations Within a Single Transaction Context
The RunInSingleTransaction function executes the provided function f using an existing transaction if one is already active in the context. If a transaction is present, it reuses that transaction. If no transaction is available, it uses the provided options to decide whether to use a primary or a replica database connection and executes the function without starting a new transaction.
```go
if err := client.RunInSingleTransaction(ctx, func(ctx context.Context, conn Conn) error {
    // Perform operations with the chosen database connection.
    _, err := conn.ExecContext(ctx, "DELETE FROM example WHERE id = ?", 123)
    if err != nil {
        // error handling
    }
    return nil
}); err != nil {
    // error handling
}
```

### Transaction Options
- WithReadOnly(): Sets the transaction as read-only, ensuring that no writes are allowed.
- WithUsePrimary(): Forces the transaction to use the primary database even for read operations.
