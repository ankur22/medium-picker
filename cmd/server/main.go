package main

import (
	"context"

	"github.com/ankur22/medium-picker/internal/logging"
)

func main() {
	ctx := context.Background()
	ctx, sync := logging.NewContext(ctx)
	defer sync()

	logging.Info(ctx, "Starting medium-picker")

	logging.Info(ctx, "Shutting down medium-picker")
}
