package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/handler"
)

func lambdaHandler(ctx context.Context, event events.EventBridgeEvent) (*domain.Game, error) {
	// Inicializar handler
	h, cleanup, err := handler.Setup(ctx)
	if err != nil {
		log.Printf("Error setting up handler: %v", err)
		return nil, err
	}
	defer cleanup()

	// Ejecutar lógica
	return h.Handle(ctx)
}

func main() {
	lambda.Start(lambdaHandler)
}
