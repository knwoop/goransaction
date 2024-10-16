package transaction

import (
	"context"
)

type transactionKey struct{}

func getTransactionFromContext(ctx context.Context) *Transaction {
	value := ctx.Value(transactionKey{})

	if value == nil {
		return nil
	}

	txn, ok := value.(*Transaction)
	if !ok {
		return nil
	}

	return txn
}

func setTransactionToContext(ctx context.Context, txn *Transaction) context.Context {
	if txn := getTransactionFromContext(ctx); txn != nil {
		return ctx
	}

	return context.WithValue(ctx, transactionKey{}, txn)
}
