package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ds "github.com/aarthikrao/timeMachine/components/datastore"
	"github.com/aarthikrao/timeMachine/process/client"
	"go.uber.org/zap"
)

func main() {
	// Initialise datastore
	datastore, err := ds.CreateBoltDataStore("test", "/data/")
	if err != nil {
		panic(err)
	}
	defer datastore.Close()

	// Initialise process
	clientProcess := client.CreateClientProcess(datastore)

	log, _ := zap.NewDevelopment()

	srv := InitTimeMachineHttpServer(clientProcess, log)
	go srv.ListenAndServe()

	log.Info("Started time machine DB 🐓")

	// ---------------------################-------------------------
	// Wait for the shut down signal. Add all the teardown code below
	// ---------------------################-------------------------

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	srv.Shutdown(context.Background())
	log.Info("shutdown completed")
	log.Sync()
}
