// router.go

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	ginlogger "github.com/gin-contrib/logger"
)

func router() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	r := gin.New()

	r.Use(ginlogger.SetLogger(ginlogger.Config{
		Logger: &logger,
		UTC:    true,
	}))

	// ping to verify API is up
	r.GET("/ping", handlePing)

	// API group operates on scheduler
	api := r.Group("/api")
	{
		// add task to scheduler
		api.POST("/add", handleAdd)

		// immediately execute GoFlows task/function/flow
		api.POST("/runNow", handleRunNow)

		// remove scheduled task based on internal id
		api.DELETE("/remove/schedulerJobID/:schedulerJobID", handleRemove)

		// remove scheduled task based on triggerId
		api.DELETE("/remove/triggerId/:triggerId", handleRemove)

		// list the current scheduled tasks (optional parameters under handleList)
		api.GET("/list", handleList)

		// scheduler log history (optional parameters under handleList)
		api.GET("/history", handleHistory)

		// validate a cron expression
		api.POST("/validate", handleValidate)
	}

	return r
}

// handlePing
func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "pong"})
}
