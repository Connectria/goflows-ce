// stats.go	- functions for GoFlows statistics

package goflows

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// return statistics based on time range
func GoFlowStats(metric string, startTime, endTime time.Time) interface{} {
	var pipeline []bson.D

	// time range
	switch metric {
	case "HumanRange":
		return fmt.Sprintf("%v through %v", startTime.Format("2006-01-02 15:04"), endTime.Format("2006-01-02 15:04"))

	case "EventsReceived":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countEventsMongo(pipeline))
		return retVal

	case "EventsReaderRejected":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", bson.D{
						{"$regex", "^Rejecting MessageId"},
					}},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countReaderHistoryMongo(pipeline))
		return retVal

	case "EventsProcessorRejected":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", bson.D{
						{"$regex", "^empty OpsGenie event"},
					}},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countFlowHistoryMongo(pipeline))
		return retVal

	case "EventsEvaluated":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.function", "processOpsGenieEvent()"},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", "assigned jobID"},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countFlowHistoryMongo(pipeline))
		return retVal

	case "EventsMatching":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.function", "processOpsGenieEvent()"},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", "all event rules (criteria) met; initiating trigger"},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countFlowHistoryMongo(pipeline))
		return retVal

	case "JobsCreated":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"jobCreateTime", bson.D{
						{"$gte", startTime.Unix()},
						{"$lte", endTime.Unix()},
					}},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countStatusMongo(pipeline))
		return retVal

	case "JobsRunning":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"jobCreateTime", bson.D{
						{"$gte", startTime.Unix()},
						{"$lte", endTime.Unix()},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"jobControlExit", -1},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countStatusMongo(pipeline))
		return retVal

	case "JobsFailed":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"jobCreateTime", bson.D{
						{"$gte", startTime.Unix()},
						{"$lte", endTime.Unix()},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"jobControlExit", 1},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countStatusMongo(pipeline))
		return retVal

	case "JobsSuccessful":
		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"jobCreateTime", bson.D{
						{"$gte", startTime.Unix()},
						{"$lte", endTime.Unix()},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"jobControlExit", 0},
				}},
			},
			bson.D{
				{"$count", "count"},
			},
		}
		retVal := fmt.Sprintf("%v", countStatusMongo(pipeline))
		return retVal

	case "EventTriggerList":
		var returnList = make([]StatsTrigger, 0)

		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.function", "processOpsGenieEvent()"},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", "all event rules (criteria) met; initiating trigger"},
				}},
			},
			bson.D{
				{"$group", bson.D{
					{"_id", "$LogData.triggerId"},
					{"trigger", bson.D{
						{"$first", "$LogData.triggerId"},
					}},
					{"count", bson.D{
						{"$sum", 1},
					}},
				}},
			},
		}

		triggersOpsGenie := getFlowHistoryTriggerInfo(pipeline)
		for _, v := range triggersOpsGenie {
			name := "[ name not set ]"

			t, err := GetTrigger(v.Trigger)
			if err == nil {
				name = t.Name
			}

			returnList = append(returnList, StatsTrigger{
				Name:  name,
				ID:    v.Trigger,
				Count: strconv.Itoa(v.Count),
			})
		}

		// per SSP-789, sort by count, name
		sortByCount := func(c1, c2 *StatsTrigger) bool {
			if c1.Count < c2.Count {
				return true
			}
			if c1.Count > c2.Count {
				return false
			}
			if c1.Count == c2.Count {
				if c1.Name < c2.Name {
					return true
				}
				if c1.Name > c2.Name {
					return false
				}
			}

			return false
		}
		StatsOrderedBy(sortByCount).Sort(returnList)
		return returnList

	case "SchedulerTriggerList":
		var returnList = make([]StatsTrigger, 0)

		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.function", "processSchedulerTask()"},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.message", "Initiating TriggerID."},
				}},
			},
			bson.D{
				{"$group", bson.D{
					{"_id", "$LogData.triggerId"},
					{"trigger", bson.D{
						{"$first", "$LogData.triggerId"},
					}},
					{"count", bson.D{
						{"$sum", 1},
					}},
				}},
			},
		}

		triggersScheduler := getFlowHistoryTriggerInfo(pipeline)
		for _, v := range triggersScheduler {
			name := "[ name not set ]"

			t, err := GetTrigger(v.Trigger)
			if err == nil {
				name = t.Name
			}

			returnList = append(returnList, StatsTrigger{
				Name:  name,
				ID:    v.Trigger,
				Count: strconv.Itoa(v.Count),
			})
		}

		// per SSP-789, sort by count, then name
		sortByCount := func(c1, c2 *StatsTrigger) bool {
			if c1.Count < c2.Count {
				return true
			}
			if c1.Count > c2.Count {
				return false
			}
			if c1.Count == c2.Count {
				if c1.Name < c2.Name {
					return true
				}
				if c1.Name > c2.Name {
					return false
				}
			}

			return false
		}
		StatsOrderedBy(sortByCount).Sort(returnList)
		return returnList

	case "JobsWithErrors":
		type StatsJobErrors struct {
			ID      string
			Message string
		}

		var returnList = make([]StatsJobErrors, 0)

		pipeline = []bson.D{
			bson.D{
				{"$match", bson.D{
					{"createdAt", bson.D{
						{"$gte", startTime},
						{"$lte", endTime},
					}},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.level", "error"},
				}},
			},
			bson.D{
				{"$match", bson.D{
					{"LogData.jobID", bson.D{
						{"$exists", true},
					}},
				}},
			},
		}

		jobsWithErrors := getFlowHistoryTriggerInfo(pipeline)
		for _, v := range jobsWithErrors {

			returnList = append(returnList, StatsJobErrors{
				ID:      v.LogData.JobID,
				Message: v.LogData.Message,
			})
		}

		return returnList

	default:
		return ("Unknown")
	}
}

// per SSP-789, sort by count, name, trigger`
type StatsTrigger struct {
	Name  string
	ID    string
	Count string
}

type statsTriggerLessFunc func(p1, p2 *StatsTrigger) bool

type statsTriggerSorter struct {
	triggers []StatsTrigger
	less     []statsTriggerLessFunc
}

func (st *statsTriggerSorter) Sort(triggers []StatsTrigger) {
	st.triggers = triggers
	sort.Sort(st)
}

func StatsOrderedBy(less ...statsTriggerLessFunc) *statsTriggerSorter {
	return &statsTriggerSorter{
		less: less,
	}
}

func (st *statsTriggerSorter) Len() int {
	return len(st.triggers)
}

func (st *statsTriggerSorter) Swap(i, j int) {
	st.triggers[i], st.triggers[j] = st.triggers[j], st.triggers[i]
}

func (st *statsTriggerSorter) Less(i, j int) bool {
	p, q := &st.triggers[i], &st.triggers[j]

	var k int
	for k = 0; k < len(st.less)-1; k++ {
		less := st.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}

	return st.less[k](p, q)
}
