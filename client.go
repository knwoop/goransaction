package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

// Client represents a database client that holds a primary connection for write operations
// and a set of replica connections for read operations.
type Client struct {
	primary  *sql.DB
	replicas []*sql.DB
}

type Conn interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// NewClient creates a new Client with a primary and an optional set of replicas.
// If no replicas are provided, the primary connection will be added to the replica list.
// This ensures there is always at least one replica (even if it's the primary itself).
func NewClient(ctx context.Context, primary *sql.DB, replicas []*sql.DB) (*Client, error) {
	if len(replicas) == 0 {
		replicas = append(replicas, primary)
	}
	return &Client{primary: primary, replicas: replicas}, nil
}

// Close closes all database connections held by the Client.
// This includes the primary connection and all replica connections.
func (c *Client) Close() {
	c.primary.Close()
	for _, v := range c.replicas {
		v.Close()
	}
}

// RunInTransaction executes the provided function `f` within a transaction context.
// If the context already contains an active transaction, it will use the existing transaction.
// Otherwise, it starts a new transaction, setting the transaction to the context, and executes the function within it.
// If starting or executing the transaction fails, an error is returned.
func (c *Client) RunInTrunsaction(ctx context.Context, f func(ctx context.Context, conn Conn) error, opts ...TransactionOption) error {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return f(ctx, txn)
	}

	if err := c.runInTrunsaction(ctx, func(ctx context.Context, conn Conn) error {
		return f(setTransactionToContext(ctx, conn), conn)
	}, opts...); err != nil {
		return fmt.Errorf("failed to execute a transaction: %w", err)
	}

	return nil
}

// RunInSingleTransaction executes the provided function `f` within the context of an existing transaction.
// If a transaction is already active in the context, it reuses the existing transaction by setting it in the context.
// If no transaction is present, it uses the provided options to determine which database connection to use
// (primary or replica) and then executes the function without starting a new transaction.
func (c *Client) RunInSingleTransaction(ctx context.Context, f func(ctx context.Context, conn Conn) error, opts ...TransactionOption) error {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return f(setTransactionToContext(ctx, txn), txn)
	}

	txOpts := newOption(opts...)

	conn := c.primary
	if !txOpts.usePrimary && txOpts.readOnly {
		conn = c.replicas[0] // TODO: use replica in rotation
	}

	return f(ctx, conn)
}

func (c *Client) runInTrunsaction(ctx context.Context, f func(context.Context, Conn) error, opts ...TransactionOption) (rerr error) {
	txOpts := newOption(opts...)

	db := c.primary
	if !txOpts.usePrimary && txOpts.readOnly {
		db = c.replicas[0] // TODO: use replica in rotation
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: txOpts.readOnly,
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return fmt.Errorf("failed to begin TX: %w", err)
	}

	succeed := false
	defer func() {
		if succeed {
			return
		}

		if err := tx.Rollback(); err != nil {
			rerr = fmt.Errorf("failed to rollback TX: %w", err)
		}
	}()

	txn := &transaction{Tx: tx}
	if err := f(setTransactionToContext(ctx, txn), txn); err != nil {
		return fmt.Errorf("failed to run function in a transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit TX: %w", err)
	}

	succeed = true
	return nil
}
