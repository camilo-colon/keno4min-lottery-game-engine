package ports

import "context"

// TransactionManager ejecuta una función dentro de una transacción del datastore.
// Si la función devuelve error, se hace rollback de todas las escrituras.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
