package main

import (
	"log"
	"os"
)

func helper() {
	os.Exit(1) 
	log.Fatal()
}

func main() {
	os.Exit(0)
	log.Fatal("error")
}
