package main

import (
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
	defer app.Shutdown()

	if err != nil {
		return err
	}

	if err := app.Migrate(); err != nil {
		return err
	}

	return app.Run()
}
