// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"errors"
	"fmt"
	"log"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

//-TRIGGER----------------------------------------------------------------------

// TriggerDBMapping is the name of the Elasticsearch type to which Triggers are added
const TriggerDBMapping string = "Trigger"

type JobRequest struct {
	CreatedBy string  `json:"createdBy"`
	JobType   JobType `json:"jobType" binding:"required"`
}

type JobType struct {
	Data map[string]interface{} `json:"data" binding:"required"`
	Type string                 `json:"type" binding:"required"`
}

// Trigger does something when the and'ed set of Conditions all are true
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
type Trigger struct {
	TriggerID     piazza.Ident           `json:"triggerId"`
	Name          string                 `json:"name" binding:"required"`
	EventTypeID   piazza.Ident           `json:"eventTypeId" binding:"required"`
	Condition     map[string]interface{} `json:"condition" binding:"required"`
	Job           JobRequest             `json:"job" binding:"required"`
	PercolationID piazza.Ident           `json:"percolationId"`
	CreatedBy     string                 `json:"createdBy"`
	CreatedOn     piazza.TimeStamp       `json:"createdOn"`
	Enabled       bool                   `json:"enabled"`
}
type TriggerUpdate struct {
	Enabled bool `json:"enabled"`
}

// TriggerList is a list of triggers
type TriggerList []Trigger

//-EVENT------------------------------------------------------------------------

const EventDBMapping string = "_default_"

// An Event is posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
type Event struct {
	EventID      piazza.Ident           `json:"eventId"`
	EventTypeID  piazza.Ident           `json:"eventTypeId" binding:"required"`
	Data         map[string]interface{} `json:"data"`
	CreatedBy    string                 `json:"createdBy"`
	CreatedOn    piazza.TimeStamp       `json:"createdOn"`
	CronSchedule string                 `json:"cronSchedule"`
}

// EventList is a list of events
type EventList []Event

//-EVENTTYPE--------------------------------------------------------------------

// EventTypeDBMapping is the name of the Elasticsearch type to which Events are added
const EventTypeDBMapping string = "EventType"

// EventType describes an Event that is to be sent to workflow by a client or service
type EventType struct {
	EventTypeID piazza.Ident           `json:"eventTypeId"`
	Name        string                 `json:"name" binding:"required"`
	Mapping     map[string]interface{} `json:"mapping" binding:"required"`
	CreatedBy   string                 `json:"createdBy"`
	CreatedOn   piazza.TimeStamp       `json:"createdOn"`
}

// EventTypeList is a list of EventTypes
type EventTypeList []EventType

//-ALERT------------------------------------------------------------------------

// AlertDBMapping is the name of the Elasticsearch type to which Alerts are added
const AlertDBMapping string = "Alert"

// Alert is a notification, automatically created when a Trigger happens
type Alert struct {
	AlertID   piazza.Ident     `json:"alertId"`
	TriggerID piazza.Ident     `json:"triggerId"`
	EventID   piazza.Ident     `json:"eventId"`
	JobID     piazza.Ident     `json:"jobId"`
	CreatedBy string           `json:"createdBy"`
	CreatedOn piazza.TimeStamp `json:"createdOn"`
}

type AlertExt struct {
	AlertID   piazza.Ident     `json:"alertId"`
	Trigger   Trigger          `json:"trigger" binding:"required"`
	Event     Event            `json:"event" binding:"required"`
	JobID     piazza.Ident     `json:"jobId"`
	CreatedBy string           `json:"createdBy"`
	CreatedOn piazza.TimeStamp `json:"createdOn"`
}

//-CRON-------------------------------------------------------------------------

const CronDBMapping = "Cron"

//-- Stats ------------------------------------------------------------

type Stats struct {
	CreatedOn        piazza.TimeStamp `json:"createdOn"`
	NumEventTypes    int              `json:"numEventTypes"`
	NumEvents        int              `json:"numEvents"`
	NumTriggers      int              `json:"numTriggers"`
	NumAlerts        int              `json:"numAlerts"`
	NumTriggeredJobs int              `json:"numTriggeredJobs"`
}

func (stats *Stats) incrCounter(counter *int) {
	*counter++
}

func (stats *Stats) IncrEventTypes() {
	stats.incrCounter(&stats.NumEventTypes)
}

func (stats *Stats) IncrEvents() {
	stats.incrCounter(&stats.NumEvents)
}

func (stats *Stats) IncrTriggers() {
	stats.incrCounter(&stats.NumTriggers)
}

func (stats *Stats) IncrAlerts() {
	stats.incrCounter(&stats.NumAlerts)
}

func (stats *Stats) IncrTriggerJobs() {
	stats.incrCounter(&stats.NumTriggeredJobs)
}

//-UTILITY----------------------------------------------------------------------

// LoggedError logs the error's message and creates an error
func LoggedError(mssg string, args ...interface{}) error {
	str := fmt.Sprintf(mssg, args...)
	log.Print(str)
	return errors.New(str)
}

//-INIT-------------------------------------------------------------------------

func init() {
	piazza.JsonResponseDataTypes["*workflow.EventType"] = "eventtype"
	piazza.JsonResponseDataTypes["[]workflow.EventType"] = "eventtype-list"
	piazza.JsonResponseDataTypes["*workflow.Event"] = "event"
	piazza.JsonResponseDataTypes["[]workflow.Event"] = "event-list"
	piazza.JsonResponseDataTypes["*workflow.Trigger"] = "trigger"
	piazza.JsonResponseDataTypes["[]workflow.Trigger"] = "trigger-list"
	piazza.JsonResponseDataTypes["*workflow.Alert"] = "alert"
	piazza.JsonResponseDataTypes["[]workflow.Alert"] = "alert-list"
	piazza.JsonResponseDataTypes["[]workflow.AlertExt"] = "alertext-list"
	piazza.JsonResponseDataTypes["workflow.Stats"] = "workflowstats"
	piazza.JsonResponseDataTypes["*workflow.TestElasticsearchBody"] = "testelasticsearch"
	piazza.JsonResponseDataTypes["[]workflow.TestElasticsearchBody"] = "testelasticsearch-list"
}
