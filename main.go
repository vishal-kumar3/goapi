package main

import "log"

func main() {
	store, err := NewPostgresStorage()
	if err != nil {
		log.Fatal("Error while db connection: ", err)
	}

	if err := store.Init(); err != nil {
		log.Fatal("Error while db init: ", err)
	}

	server := NewAPIServer(":8000", store)
	server.Run()
}
