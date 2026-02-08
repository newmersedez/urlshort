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
	application, err := app.NewApp()
	if err != nil {
		return err
	}

	return application.Run()
}
