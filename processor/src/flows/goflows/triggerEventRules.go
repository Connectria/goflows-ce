// triggerEventRules.go	- functions for trigger event rules cache

package goflows

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// TriggerEventRulesCache contains the rules for events
type TriggerEventRulesCacheType struct {
	Triggers   []TriggerType `json:"triggers"`
	LastUpdate int64         `json:"lastUpdate"` // Unix time
	TTL        int64         `json:"TTL"`
	mutex      sync.Mutex
}

// UpdateEventTriggersCache
func (rules *TriggerEventRulesCacheType) UpdateEventTriggersCache() error {
	if time.Now().Unix() > (rules.LastUpdate + rules.TTL) {
		body, err := frontEndAPI("GET", "/api/triggers?triggerLogic.triggerType=Event", nil)
		if err != nil {
			return err
		}

		rules.mutex.Lock()
		rules.Triggers = []TriggerType{}

		err = json.Unmarshal(body, &rules)
		if err != nil {
			os.Stderr.WriteString(err.Error())
		}

		/*
		   SSP-847  - Implement "weighted" (prioritized) event rules

		   RULE: weight	- a way for event rules (triggers) to be prioritized

		   	1. The higher the number, gets evaluated "sooner"
		   	2. No number, or 0, will be the default (natural order)
		   	3. Numbers less than 0 will be evaluated last last, lower the number the "more last it becomes"

		   EXAMPLE:
		   "weight": 0,	// natural order (default)
		   "weight": -1,	// after 0
		   "weight": 10,	// before 0 but after 12
		   "weight": 12,	// first

		*/

		for i, t := range rules.Triggers {
			rules.Triggers[i].Weight = 0
			for _, rule := range t.Triggerlogic.EventRules {
				ruleWeight := gjson.GetBytes(rule, "weight")
				if ruleWeight.Index > 0 {
					ruleWeight.ForEach(
						func(k, v gjson.Result) bool {
							rules.Triggers[i].Weight = v.Int()
							return true
						})
					continue
				}
			}
		}

		sortByWeight := func(c1, c2 *TriggerType) bool {
			return c1.Weight > c2.Weight
		}

		WeightedOrderedBy(sortByWeight).Sort(rules.Triggers)
		rules.LastUpdate = time.Now().Unix()
		rules.mutex.Unlock()
	}

	return nil
}

// the below funcs enforoce the desired sort of the struct
type weightedTriggerLessFunc func(p1, p2 *TriggerType) bool

type weightedTriggerSorter struct {
	triggers []TriggerType
	less     []weightedTriggerLessFunc
}

func (wt *weightedTriggerSorter) Sort(triggers []TriggerType) {
	wt.triggers = triggers
	sort.Sort(wt)
}

func WeightedOrderedBy(less ...weightedTriggerLessFunc) *weightedTriggerSorter {
	return &weightedTriggerSorter{
		less: less,
	}
}

func (wt *weightedTriggerSorter) Len() int {
	return len(wt.triggers)
}

func (wt *weightedTriggerSorter) Swap(i, j int) {
	wt.triggers[i], wt.triggers[j] = wt.triggers[j], wt.triggers[i]
}

func (wt *weightedTriggerSorter) Less(i, j int) bool {
	p, q := &wt.triggers[i], &wt.triggers[j]

	var k int
	for k = 0; k < len(wt.less)-1; k++ {
		less := wt.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}

	return wt.less[k](p, q)
}
