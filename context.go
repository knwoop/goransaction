package transaction

import (
	"context"
	"database/sql"
)

type transactionKey struct{}

// Transaction is a wrapper around the standard sql.Tx, representing a database transaction.
type transaction struct {
	*sql.Tx
}

func getTransactionFromContext(ctx context.Context) Conn {
	value := ctx.Value(transactionKey{})

	if value == nil {
		return nil
	}

	txn, ok := value.(Conn)
	if !ok {
		return nil
	}

	return txn
}

func setTransactionToContext(ctx context.Context, txn Conn) context.Context {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return ctx
	}

	return context.WithValue(ctx, transactionKey{}, txn)
}
