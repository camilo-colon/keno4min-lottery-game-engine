package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/handler"
)

func lambdaHandler(ctx context.Context, input handler.CloseBetsInput) error {
	// Inicializar handler
	h, cleanup, err := handler.Setup(ctx)
	if err != nil {
		log.Printf("Error setting up handler: %v", err)
		return err
	}
	defer cleanup()

	// Ejecutar lógica
	return h.Handle(ctx, input)
}

func main() {
	lambda.Start(lambdaHandler)
}
