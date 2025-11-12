package main

import "go.uber.org/zap"

func main() {
	InitLogger()
	defer SyncLogger()

	store, err := NewPostgresStorage()
	if err != nil {
		Log.Fatal("Error while db connection: ", zap.Error(err))
	}

	if err := store.Init(); err != nil {
		Log.Fatal("Error while db init: ", zap.Error(err))
	}

	server := NewAPIServer(":8000", store)
	server.Run()
}
