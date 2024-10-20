package transaction

import (
	"context"
	"errors"
	"fmt"
)

var errAlreadyInTransaction = errors.New("already in a transaction")

// RunTransaction executes the given function `f` within a transaction context.
// If the context already contains an active transaction, it returns an error indicating
// that a transaction is already in progress. Otherwise, it starts a new transaction
// using the provided options and executes the function `f` within that transaction.
func RunTrunsaction(ctx context.Context, c *Client, f func(ctx context.Context, conn Conn) error, opts ...TransactionOption) (err error) {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return errAlreadyInTransaction
	}

	if err := c.runInTrunsaction(ctx, func(ctx context.Context, conn Conn) error {
		return f(setTransactionToContext(ctx, conn), conn)
	}, opts...); err != nil {
		return fmt.Errorf("failed to execute a transaction: %w", err)
	}
	return nil
}
