package main

import (
	"context"
	"log"
	"os"

	"github.com/kriuchkov/jiraforge/internal/app"
)

func main() {
	ctx := context.Background()
	root := newRootCommand(os.Stdout, os.Stderr)
	if err := root.ExecuteContext(ctx); err != nil && !app.IsContextCanceled(err) {
		log.Fatalf("run failed: %v", err)
	}
}
