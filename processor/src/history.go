// history.go - database filter and handler for processor logs  "/api/history"

package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// handleHistory
func handleApiHistory(c *gin.Context) {
	var start int64 = 0
	var end int64 = 0
	var err error

	// validate time range
	if len(c.Query("startTime")) > 0 {
		start, err = strconv.ParseInt(c.Query("startTime"), 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "invalid startTime",
			})
			return
		}
	}

	if len(c.Query("endTime")) > 0 {
		end, err = strconv.ParseInt(c.Query("endTime"), 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "invalid endTime",
			})
			return
		}
	}

	if (end > 0) && (start == 0) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "must use startTime if using endTime",
		})
		return
	}

	// history in range
	if (start > 0) && (end > 0) {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", time.Unix(start, 0)},
						{"$lte", time.Unix(end, 0)},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"startTime":         c.Query("startTime"),
			"endTime":           c.Query("endTime"),
			"processor-history": history,
		})
		return
	}

	// history since
	if len(c.Query("startTime")) > 0 {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", time.Unix(start, 0)},
					}},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"startTime":         c.Query("startTime"),
			"processor-history": history,
		})
		return
	}

	// specific messageID
	if len(c.Query("messageID")) > 0 {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"LogData.messageID", c.Query("messageID")},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"messageID":         c.Query("messageID"),
			"processor-history": history,
		})
		return
	}

	// all logs specific to eventID (i.e. OpsGenie alert)
	if len(c.Query("eventAlertID")) > 0 {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"LogData.eventAlertID", c.Query("eventAlertID")},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"eventAlertID":      c.Query("eventAlertID"),
			"processor-history": history,
		})
		return
	}

	// all logs specific to jobID (from GoFlow processing)
	if len(c.Query("jobID")) > 0 {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"LogData.jobID", c.Query("jobID")},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"jobID":             c.Query("jobID"),
			"processor-history": history,
		})
		return
	}

	// all logs specific to triggerId
	if len(c.Query("triggerId")) > 0 || len(c.Query("triggerID")) > 0 {
		pipeline := []bson.D{
			bson.D{
				{"$match", bson.D{
					{"LogData.triggerId", c.Query("triggerId")},
				}},
			},
			bson.D{
				{"$sort", bson.D{
					{"LogData.time", 1},
				}},
			},
		}

		history, err := getFlowHistory(pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          err.Error(),
				"error-function": "handleApiHistory()",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"triggerId":         c.Query("triggerId"),
			"processor-history": history,
		})
		return
	}

	// incomplete
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "garbled or incomplete request",
	})
}

// HistoryType is the structure of the processor logs
type HistoryType struct {
	Level        string  `json:"level"`
	Time         int     `json:"time"`
	Action       string  `json:"action,omitempty"`
	AlertActions string  `json:"alertActions,omitempty"`
	Duration     float64 `json:"duration,omitempty"`
	EventAlertID string  `json:"eventAlertID,omitempty"`
	Function     string  `json:"function,omitempty"`
	FuncName     string  `json:"funcName,omitempty"`
	JobID        string  `json:"jobID,omitempty"`
	JobStepID    int     `json:"jobStepID,omitempty"`
	Latency      float64 `json:"latency,omitempty"`
	Message      string  `json:"message,omitempty"`
	MessageID    string  `json:"messageID,omitempty"`
	Method       string  `json:"method,omitempty"`
	Path         string  `json:"path,omitempty"`
	Status       int     `json:"status,omitempty"`
	TriggerID    string  `json:"triggerId,omitempty"`
	UserAgent    string  `json:"user_agent,omitempty"`
}

type HistoryMongoType struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt primitive.DateTime `bson:"createdAt,omitempty"`
	LogData   HistoryType        `bson:"LogData"`
}

// return flow histoy
func getFlowHistory(p []primitive.D) ([]HistoryType, error) {
	opts := options.Aggregate().SetMaxTime(cfg.MongoClientTimeout).SetAllowDiskUse(true)
	loadedStructCursor, err := logCollection.Aggregate(context.TODO(), p, opts)
	if err != nil {
		logger.Error().
			Str("function", "getFlowHistory()").
			Msgf("logCollection.Aggregate() returns: '%v'", err.Error())
		return []HistoryType{}, err
	}

	var history []HistoryType
	for loadedStructCursor.Next(context.TODO()) {
		var h HistoryMongoType
		err = loadedStructCursor.Decode(&h)
		if err != nil {
			logger.Error().
				Str("function", "getFlowHistory()").
				Msgf("loadedStructCursor.Decocde() returns: '%v'", err.Error())
			return []HistoryType{}, err
		}

		history = append(history, h.LogData)
	}

	loadedStructCursor.Close(context.TODO())
	return history, nil
}
