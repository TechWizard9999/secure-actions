package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kotakarthik/secure-actions/internal/app"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("secure-actions %s (%s)\n", version, commit)
		return
	}

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}