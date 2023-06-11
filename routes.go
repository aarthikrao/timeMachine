package main

import (
	"fmt"
	"net/http"

	"github.com/aarthikrao/timeMachine/components/client"
	"github.com/aarthikrao/timeMachine/components/concensus"
	"github.com/aarthikrao/timeMachine/components/dht"
	"github.com/aarthikrao/timeMachine/handlers/rest"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitTimeMachineHttpServer(
	cp *client.ClientProcess,
	appDht dht.DHT,
	con concensus.Concensus,
	onClusterFormHandler func(),
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
	crh := rest.CreateClusterRestHandler(con, appDht, onClusterFormHandler, log)
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

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	return srv
}
