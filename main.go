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

	var (
		// appDht will store the distributed hash table of this node
		appDht  dht.DHT                              = dht.Create()
		dsmgr   *dsm.DataStoreManager                = dsm.CreateDataStore(boltDataDir, log)
		connMgr *connectionmanager.ConnectionManager = connectionmanager.CreateConnectionManager(log)
	)

	// Initialise the FSM store
	fsmStore := fsm.NewConfigFSM(
		appDht,
		log,
	)

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
		log.Fatal("Unable to start raft")
		panic(err)
	}

	// Initialise node manager
	nodeMgr := nodemanager.CreateNodeManager(
		*nodeID,
		dsmgr,
		connMgr,
		appDht,
		raft,
		log,
	)

	// This method will be called by the FSM store if there are any changes.
	// We will initialise the connections in the nodeMgr with the latest cluster configuration
	fsmStore.SetChangeHandler(nodeMgr.CreateConnections)

	// Initialise process
	clientProcess := client.CreateClientProcess(nodeMgr)

	// Initialises the data store and connections to other nodes in the cluster.
	// This method is called after the cluster is formed and slots are computed.
	// Here we are just defining the method. It will be called once the node vs slots
	// values are ready.
	var initialiseDatastoreAndConn = func() {
		thisNodeID := dht.NodeID(*nodeID)
		slots := appDht.GetSlotsForNode(thisNodeID)
		if len(slots) <= 0 {
			panic("There are no slots for this node. Did you mean to start this node in bootstrap mode")
		}

		if err := dsmgr.InitialiseDataStores(slots); err != nil {
			panic(err)
		}

		if err := nodeMgr.CreateConnections(); err != nil {
			panic(err)
		}
	}

	if !*bootstrap {
		initialiseDatastoreAndConn()

	} else {
		// This means the node has started in bootstrap mode.
		// If the cluster is being started for the first time, we will have to
		// 	1. Form a raft group and elect a leader.
		//  2. Ask the leader to create the inital node vs slot map with leader and follower details.
		// 		this can be done by calling `Initialise(slotCountperNode int, nodes []string)``
		// 	3. Communicate with all the nodes in the raft group and apply the DHT in all the nodes.
		log.Warn("NodeVsSlot and datastores not yet initialised. Consider rebalancing the cluster once started")
	}

	srv := InitTimeMachineHttpServer(
		clientProcess,
		appDht,
		raft,
		initialiseDatastoreAndConn,
		nodeMgr,
		log,
		*httpPort,
	)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	// TODO: Implement adding this config from REST API after bootstrap.
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
