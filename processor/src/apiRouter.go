// apiRouter.go - http server for API

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	ginlogger "github.com/gin-contrib/logger"
)

// apiRouter is the server used for API calls
func apiRouter() *gin.Engine {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	apiRouter := gin.New()
	apiRouter.Use(
		ginlogger.SetLogger(ginlogger.Config{
			Logger: &logger,
			UTC:    true,
		}),
	)

	// ping to verify API is up
	apiRouter.GET("/ping", handlePing)
	apiRouter.HEAD("/ping", handlePing)

	// external systems post to call backs using the callBackID (Kong is proxy)
	/** Disable callbacks
		apiRouter.POST("/callback/:callBackID", handleApiCallBack)
	**/
	// API group
	api := apiRouter.Group("/api")
	{
		// authorization history (optional parameters under handleList)
		// Commented out by AKB for CE release
		//api.GET("/authHistory", handleApiAuthHistory)

		// processor log history (optional parameters under handleList)
		api.GET("/history", handleApiHistory)

		// list available (compiled) functions for for events/tasks
		api.GET("/listFuncs", handleApiListFuncs)

		// list current call backs and if available, POST'd data
		/** Disable Callbacks
				api.GET("/listCallBacks", handleApiListCallBacks)
		**/

		// list trigger event rules in cache
		api.GET("/listTriggerEventRules", handleApiListTriggerEventRules)

		// set trigger rules ttl
		api.PUT("/setTriggerEventTTL", handleApiSetTriggerEventTTL)

		// validate event against rules
		api.GET("/validateRules", handleApiValidateRulesGet)
	}

	return apiRouter
}
