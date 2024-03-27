package main

import (
	"fmt"
	"net/http"

	"github.com/aarthikrao/timeMachine/components/consensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/handlers/rest"
	"github.com/aarthikrao/timeMachine/process/cordinator"
	"github.com/aarthikrao/timeMachine/process/nodemanager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitTimeMachineHttpServer(
	cp *cordinator.CordinatorProcess,
	appDht dht.DHT,
	con consensus.Consensus,
	nodeMgr *nodemanager.NodeManager,
	log *zap.Logger,
	port int,
) *http.Server {
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gin.Recovery())
	// gin.SetMode(gin.ReleaseMode)

	// Health handler
	r.GET("/health", func(c *gin.Context) {
		// Return status ok
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Cluster handlers
	crh := rest.CreateClusterRestHandler(con, appDht, nodeMgr, log)
	cluster := r.Group("/cluster")
	{
		cluster.GET("", crh.GetStats)
		cluster.POST("/join", crh.Join)
		cluster.POST("/remove", crh.Remove)
		cluster.POST("/configure", crh.Configure)
	}

	// Job handlers
	jrh := rest.CreateJobRestHandler(cp, log)
	job := r.Group("/job")
	{
		job.GET("/:collection/:jobID", jrh.GetJob)
		job.POST("/:collection", jrh.SetJob)
		job.DELETE("/:collection/:jobID", jrh.DeleteJob)
	}

	// Route Handlers
	rrh := rest.CreateRouteRestHandler(cp, log)
	route := r.Group("/route")
	{
		route.GET("/:id", rrh.GetRoute)
		route.POST("/", rrh.SetRoute)
		route.DELETE("/:id", rrh.DeleteRoute)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	return srv
}
