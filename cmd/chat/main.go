package main

import (
	"log"

	"github.com/Rosa-Devs/core/core"
	"github.com/Rosa-Devs/core/store"
)

func main() {
	// Cereate app maneger instance
	Store, err := store.NewStore("/home/mihalic2040/rosa_test")
	if err != nil {
		log.Panic("Fail to create store:", err)
	}
	app := &core.Core{
		Store: *Store,
	}

	localadrr := app.Start(":8080")

	log.Println("Local API addr is:", localadrr)

	select {}
}
