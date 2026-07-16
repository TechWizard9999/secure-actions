package main

import (
	"log"

	"github.com/kotakarthik/secure-actions/internal/app"
)

func main() {

	application := app.New()

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}