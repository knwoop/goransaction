package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type mockConn struct {
	id string
}

func (m *mockConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockConn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, nil
}

func (m *mockConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockConn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

func newMockConn(id string) Conn {
	return &mockConn{id: id}
}

func TestGetTransactionFromContext(t *testing.T) {
	// Iterate over the test cases directly in the for loop
	for name, tt := range map[string]struct {
		ctx  context.Context
		want Conn
	}{
		"get existing transaction": {
			ctx:  context.WithValue(context.Background(), transactionKey{}, newMockConn("txn1")),
			want: newMockConn("txn1"),
		},
		"no transaction in context": {
			ctx:  context.Background(),
			want: nil,
		},
		"wrong type in context": {
			ctx:  context.WithValue(context.Background(), transactionKey{}, "invalid type"),
			want: nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := getTransactionFromContext(tt.ctx)

			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(mockConn{})); diff != "" {
				t.Errorf("\n(-got +want)\n%s", diff)
			}
		})
	}
}

func TestSetTransactionToContext(t *testing.T) {
	for name, tt := range map[string]struct {
		initialCtx context.Context
		setTxn     Conn
		want       Conn
	}{
		"set new transaction in context": {
			initialCtx: context.Background(),
			setTxn:     newMockConn("txn2"),
			want:       newMockConn("txn2"),
		},
		"transaction already in context, should not overwrite": {
			initialCtx: context.WithValue(context.Background(), transactionKey{}, newMockConn("txn3")),
			setTxn:     newMockConn("txn4"),
			want:       newMockConn("txn3"),
		},
		"set nil transaction": {
			initialCtx: context.Background(),
			setTxn:     nil,
			want:       nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := setTransactionToContext(tt.initialCtx, tt.setTxn)
			got := getTransactionFromContext(ctx)

			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(mockConn{})); diff != "" {
				t.Errorf("\n(-got +want)\n%s", diff)
			}
		})
	}
}
