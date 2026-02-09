package main

import (
	"fmt"
	"log"

	"github.com/newmersedez/urlshort/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	app, err := app.NewApp()

	if err != nil {
		return fmt.Errorf("failed to initiailze the application object: %w", err)
	}

	return app.Run()
}
