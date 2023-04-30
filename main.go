package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/aarthikrao/timeMachine/components/client"
	"github.com/aarthikrao/timeMachine/components/concensus"
	"github.com/aarthikrao/timeMachine/components/concensus/fsm"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"go.uber.org/zap"
)

// Input flags
var (
	nodeID    = flag.String("nodeID", "", "Raft nodeID of this node. Must be unique across cluster")
	dataDir   = flag.String("datadir", "data", "Provide the data directory without trailing '/'")
	raftPort  = flag.Int("raftPort", 8101, "raft listening port")
	httpPort  = flag.Int("httpPort", 8001, "http listening port")
	bootstrap = flag.Bool("bootstrap", false, "Should be `true` for the first node of the cluster")
)

func main() {
	flag.Parse()
	if *nodeID == "" || *dataDir == "" || *raftPort == 0 {
		flag.PrintDefaults()
		panic("Invalid flags. try: ./timeMachine --nodeID=node1 --raftPort=8101 --httpPort=8001 --bootstrap=true")
	}

	// Prepare data and raft folder
	baseDir := *dataDir + "/" + *nodeID
	boltDataDir := baseDir + "/data"
	raftDataDir := baseDir + "/raft"

	log, _ := zap.NewDevelopment()

	// Initialise the FSM store
	fsmStore := fsm.NewConfigFSM(log)

	// Initialise raft
	raft, err := concensus.NewRaftConcensus(
		*nodeID,
		*raftPort,
		raftDataDir,
		fsmStore,
		log,
		*bootstrap,
	)
	if err != nil {
		log.Fatal("Unable to start raft", zap.Error(err))
	}

	var (
		// appDht will store the distributed hash table of this node
		appDht  dht.DHT                              = dht.Create()
		dsmgr   *dsm.DataStoreManager                = dsm.CreateDataStore(boltDataDir, log)
		connMgr *connectionmanager.ConnectionManager = connectionmanager.CreateConnectionManager(log)
	)

	if !*bootstrap {
		nodeVsSlot := fsmStore.GetNodeVsStruct()
		if len(nodeVsSlot) <= 0 {
			panic("There are no slots for this node. Did you mean to start this node in bootstrap mode")
		}

		// Load the hash table information // TODO: Initialise properly
		// if err := appDht.Load(nodeVsSlot); err != nil {
		// 	panic(err)
		// }

		slots := nodeVsSlot[dht.NodeID(*nodeID)]
		if err := dsmgr.InitialiseDataStores(slots); err != nil {
			panic(err)
		}

		// Initialse the connection manager
		addrMap := fsmStore.GetNodeAddressMap()

		// Create connections to other nodes
		for nodeID, address := range addrMap {
			connMgr.AddNewConnection(nodeID, address)
		}

	} else {
		// This means the node has started in bootstrap mode.
		// We will need to join the raft group first, and then ask the master to rebalance
		// The master will then re distribute the slots in a way that causes very minumum
		// data transafer accross the nodes
		// We then update the dht so that the traffic is sent to the right data node
		log.Warn("NodeVsSlot and datastores not yet initialised. Consider rebalancing the cluster once started")
	}

	// Initialise node manager
	nodeMgr := nodemanager.CreateNodeManager(
		*nodeID,
		dsmgr,
		connMgr,
		appDht,
	)

	// Initialise process
	clientProcess := client.CreateClientProcess(nodeMgr)

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
