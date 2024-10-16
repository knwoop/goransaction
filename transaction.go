package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var errAlreadyInTransaction = errors.New("already in a transaction")

// Transaction is a wrapper around the standard sql.Tx, representing a database transaction.
type Transaction struct {
	*sql.Tx
}

// RunTransaction executes the given function `f` within a transaction context.
// If the context already contains an active transaction, it returns an error indicating
// that a transaction is already in progress. Otherwise, it starts a new transaction
// using the provided options and executes the function `f` within that transaction.
func RunTrunsaction(ctx context.Context, c *Client, f func(ctx context.Context, txn *Transaction) error, opts ...TransactionOption) (err error) {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return errAlreadyInTransaction
	}

	if err := c.runInTrunsaction(ctx, func(ctx context.Context, txn *Transaction) error {
		return f(setTransactionToContext(ctx, txn), txn)
	}, opts...); err != nil {
		return fmt.Errorf("failed to execute a transaction: %w", err)
	}
	return nil
}

// TransactionOption is a type that defines functional options for transaction settings.
type TransactionOption func(options *TransactionOptions)

// TransactionOptions holds the configurable options for a transaction.
type TransactionOptions struct {
	usePrimary bool
	readOnly   bool
}

// WithReadOnly returns a TransactionOption that sets the transaction to be read-only.
func WithReadOnly() TransactionOption {
	return func(options *TransactionOptions) {
		options.readOnly = true
	}
}

// WithUsePrimary returns a TransactionOption that forces the transaction to use the primary database.
func WithUsePrimary() TransactionOption {
	return func(options *TransactionOptions) {
		options.usePrimary = true
	}
}
