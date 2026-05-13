package main

import (
	"log"
	"os"
)

func helper() {
	os.Exit(1)       // want `os\.Exit must not be called outside of main function in main package`
	log.Fatal("err") // want `log\.Fatal must not be called outside of main function in main package`
}

func main() {
	os.Exit(0)
	log.Fatal("error")
}
