package rest

import (
	"net/http"

	"github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/process/cordinator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type routeRestHandler struct {
	cp  *cordinator.CordinatorProcess
	log *zap.Logger
}

func CreateRouteRestHandler(cp *cordinator.CordinatorProcess, log *zap.Logger) *routeRestHandler {
	return &routeRestHandler{
		cp:  cp,
		log: log,
	}
}

func (rrh *routeRestHandler) GetRoute(c *gin.Context) {
	id := c.Param("id")

	route, err := rrh.cp.GetRoute(id)
	if route == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

func (rrh *routeRestHandler) SetRoute(c *gin.Context) {
	var route routemodels.Route
	c.BindJSON(&route)

	if err := rrh.cp.SetRoute(&route); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (jrh *routeRestHandler) DeleteRoute(c *gin.Context) {
	id := c.Param("id")

	if err := jrh.cp.DeleteRoute(id); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
