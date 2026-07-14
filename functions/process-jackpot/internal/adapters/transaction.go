package adapters

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoTransactionManager ejecuta trabajo dentro de una transacción de MongoDB.
// Requiere que el despliegue sea un replica set (Atlas lo es).
type MongoTransactionManager struct {
	client *mongo.Client
}

func NewMongoTransactionManager(client *mongo.Client) *MongoTransactionManager {
	return &MongoTransactionManager{client: client}
}

// WithinTransaction abre una sesión y corre fn dentro de una transacción. El
// contexto de sesión se pasa a fn: toda operación de repositorio que lo use se
// enrola automáticamente en la transacción. WithTransaction reintenta la callback
// ante errores transitorios y hace commit/abort según el resultado.
func (tm *MongoTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, fn(sessCtx)
	})
	return err
}
