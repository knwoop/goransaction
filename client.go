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
func (c *Client) RunInTrunsaction(ctx context.Context, f func(ctx context.Context, txn *Transaction) error, opts ...TransactionOption) error {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return f(ctx, txn)
	}

	if err := c.runInTrunsaction(ctx, func(ctx context.Context, txn *Transaction) error {
		return f(setTransactionToContext(ctx, txn), txn)
	}, opts...); err != nil {
		return fmt.Errorf("failed to execute a transaction: %w", err)
	}

	return nil
}

func (c *Client) runInTrunsaction(ctx context.Context, f func(context.Context, *Transaction) error, opts ...TransactionOption) (rerr error) {
	var txOpts TransactionOptions
	for _, opt := range opts {
		opt(&txOpts)
	}

	conn := c.primary
	if !txOpts.usePrimary && txOpts.readOnly {
		conn = c.replicas[0] // TODO: use replica in rotation
	}
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{
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

	txn := &Transaction{Tx: tx}
	if err := f(setTransactionToContext(ctx, txn), txn); err != nil {
		return fmt.Errorf("failed to run function in a transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit TX: %w", err)
	}

	succeed = true
	return nil
}
