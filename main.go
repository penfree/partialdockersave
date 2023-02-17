package main

import (
	"context"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApp(ctx)
	app.Run(os.Args)
}
