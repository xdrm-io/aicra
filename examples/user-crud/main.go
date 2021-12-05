package main

import (
	"log"
	"os"

	"github.com/xdrm-io/aicra/examples/user-crud/storage"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalf("missing argument: configuration file")
	}

	// load api definition
	config, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("cannot open config: %s", err)
	}

	// create dummy database
	db := &storage.DB{}

	// build app
	app, err := NewApp(config, db)
	config.Close()
	if err != nil {
		log.Fatalf("cannot create app: %s", err)
	}

	// launch
	log.Printf("server up at ':8080'")
	log.Fatal(app.Listen(":8080"))
}
