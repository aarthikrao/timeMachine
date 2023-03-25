package rest

import (
	"net/http"

	"github.com/aarthikrao/timeMachine/components/concensus"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type clusterMessage struct {
	NodeID      string `json:"node_id,omitempty" bson:"node_id,omitempty"`
	RaftAddress string `json:"raft_address,omitempty" bson:"raft_address,omitempty"`
}

type clusterRestHandler struct {
	cp  concensus.Concensus
	log *zap.Logger
}

func CreateClusterRestHandler(cp concensus.Concensus, log *zap.Logger) *clusterRestHandler {
	return &clusterRestHandler{
		cp:  cp,
		log: log,
	}
}

func (crh *clusterRestHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, crh.cp.Stats())
}

func (crh *clusterRestHandler) Join(c *gin.Context) {
	var cm clusterMessage
	c.BindJSON(&cm)

	if cm.NodeID == "" || cm.RaftAddress == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid nodeID or address"})
		return
	}

	if !crh.cp.IsLeader() {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error":  "not leader",
				"leader": crh.cp.GetLeaderAddress(),
			},
		)
		return
	}

	if err := crh.cp.Join(cm.NodeID, cm.RaftAddress); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (crh *clusterRestHandler) Remove(c *gin.Context) {
	var cm clusterMessage
	c.BindJSON(&cm)

	if cm.NodeID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid nodeID"})
		return
	}

	if !crh.cp.IsLeader() {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error":  "not leader",
				"leader": crh.cp.GetLeaderAddress(),
			},
		)
		return
	}

	if err := crh.cp.Remove(cm.NodeID); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (crh *clusterRestHandler) Redistribute(c *gin.Context) {

	if !crh.cp.IsLeader() {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"error":  "not leader",
				"leader": crh.cp.GetLeaderAddress(),
			},
		)
		return
	}

	// TODO: Yet to implement

}
