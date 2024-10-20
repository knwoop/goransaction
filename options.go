package transaction

var _ TransactionOption = (*funcOption)(nil)

// TransactionOption is a type that defines functional options for transaction settings.
type TransactionOption interface {
	apply(*transactionOption)
}

// TransactionOptions holds the configurable options for a transaction.
type transactionOption struct {
	usePrimary bool
	readOnly   bool
}

type funcOption struct {
	f func(*transactionOption)
}

func (fo *funcOption) apply(o *transactionOption) {
	fo.f(o)
}

func newFuncOption(f func(*transactionOption)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func newOption(opts ...TransactionOption) *transactionOption {
	o := &transactionOption{}
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

// WithReadOnly returns a TransactionOption that sets the transaction to be read-only.
func WithReadOnly() TransactionOption {
	return newFuncOption(func(o *transactionOption) {
		o.readOnly = true
	})
}

// WithUsePrimary returns a TransactionOption that forces the transaction to use the primary database.
func WithUsePrimary() TransactionOption {
	return newFuncOption(func(o *transactionOption) {
		o.usePrimary = true
	})
}
