package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aarthikrao/timeMachine/components/client"
	"github.com/aarthikrao/timeMachine/components/concensus"
	ds "github.com/aarthikrao/timeMachine/components/datastore"
	"go.uber.org/zap"
)

func main() {
	serverID := flag.String("serverID", "", "Raft serverID of this node. Must be unique across cluster")
	dataDir := flag.String("datadir", "data", "Provide the data directory without trailing '/'")
	raftPort := flag.Int("raftPort", 8101, "raft listening port")
	httpPort := flag.Int("httpPort", 8001, "http listening port")
	bootstrap := flag.Bool("bootstrap", false, "Should be `true` for the first node of the cluster")
	flag.Parse()

	if *serverID == "" || *dataDir == "" || *raftPort == 0 {
		flag.PrintDefaults()
		fmt.Println("try: ./timeMachine --serverID=node1 --raftPort=8101 --httpPort=8001 --bootstrap=true")
		panic("Invalid flags.")
	}

	// Prepare data and raft folder
	baseDir := *dataDir + "/" + *serverID
	boltDataDir := baseDir + "/data"
	raftDataDir := baseDir + "/raft"

	// Initialise datastore
	datastore, err := ds.CreateBoltDataStore(boltDataDir + "/" + "test")
	if err != nil {
		panic(err)
	}
	defer datastore.Close()

	// Initialise process
	clientProcess := client.CreateClientProcess(datastore)

	log, _ := zap.NewDevelopment()
	raft, err := concensus.NewRaftConcensus(*serverID, *raftPort, raftDataDir, log, *bootstrap)
	if err != nil {
		log.Fatal("Unable to start raft", zap.Error(err))
	}

	srv := InitTimeMachineHttpServer(clientProcess, raft, log, *httpPort)
	go srv.ListenAndServe()

	// Just for testing
	// go func() {
	// 	for {
	// 		if raft.IsLeader() {
	// 			log.Info("Is leader")
	// 			val := fsm.NodeConfig{
	// 				LastContactTime: timeUtils.GetCurrentMillis(),
	// 			}
	// 			by, err := json.Marshal(val)
	// 			if err != nil {
	// 				log.Error("Unable to marshal", zap.Error(err))
	// 			}
	// 			raft.Apply(by)
	// 		}
	// 		log.Info("sleep")
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

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
