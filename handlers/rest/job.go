package rest

import (
	"net/http"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"github.com/aarthikrao/timeMachine/process/cordinator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type jobRestHandler struct {
	cp  *cordinator.CordinatorProcess
	log *zap.Logger
}

func CreateJobRestHandler(cp *cordinator.CordinatorProcess, log *zap.Logger) *jobRestHandler {
	return &jobRestHandler{
		cp:  cp,
		log: log,
	}
}

func (jrh *jobRestHandler) GetJob(c *gin.Context) {
	collection := c.Param("collection")
	jobID := c.Param("jobID")

	job, err := jrh.cp.GetJob(collection, jobID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (jrh *jobRestHandler) SetJob(c *gin.Context) {
	collection := c.Param("collection")

	var job jobmodels.Job
	c.BindJSON(&job)

	offset, err := jrh.cp.SetJob(collection, &job)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jrh.log.Debug("Job set", zap.String("collection", collection), zap.Any("job", job))

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"offset": offset,
	})
}

func (jrh *jobRestHandler) DeleteJob(c *gin.Context) {
	collection := c.Param("collection")
	jobID := c.Param("jobID")

	offset, err := jrh.cp.DeleteJob(collection, jobID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"offset": offset,
	})
}
