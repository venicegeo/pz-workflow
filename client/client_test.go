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

package client

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-workflow/common"
	_server "github.com/venicegeo/pz-workflow/server"
	"log"
	"testing"
	"time"
)

type ClientTester struct {
	suite.Suite
	logger     loggerPkg.ILoggerService
	uuidgenner uuidgenPkg.IUuidGenService
	workflow   *PzWorkflowService
	sys        *piazza.System
}

func (suite *ClientTester) SetupSuite() {
	t := suite.T()
	assert := assert.New(t)

	config, err := piazza.NewConfig(piazza.PzWorkflow, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	suite.logger, err = loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	suite.uuidgenner, err = uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	routes, err := _server.CreateHandlers(sys, suite.logger, suite.uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	suite.workflow, err = NewPzWorkflowService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	assert.Len(sys.Services, 5)

	suite.assertNoData()
}

func (suite *ClientTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func (suite *ClientTester) assertNoData() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	var err error

	es, err := workflow.GetAllEvents()
	assert.NoError(err)
	assert.Len(*es, 0)

	ts, err := workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*ts, 0)

	as, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*as, 0)

	xs, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*xs, 0)

}

func TestRunSuite(t *testing.T) {
	s := new(ClientTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *ClientTester) TestOne() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error
	var eventTypeName = "EventTypeA"

	var etId common.Ident
	{
		mapping := map[string]piazza.MappingElementTypeName{
			"num": piazza.MappingElementTypeInteger,
			"str": piazza.MappingElementTypeString,
		}

		eventType := &common.EventType{Name: eventTypeName, Mapping: mapping}

		etId, err = workflow.PostEventType(eventType)
		assert.NoError(err)
	}

	var tId common.Ident
	{
		x1 := &common.Trigger{
			Title: "the x1 trigger",
			Condition: common.Condition{
				EventType: etId,
				Query: `{
					"query": {
						"match": {
							"num": 17
						}
					}
				}`,
			},
			Job: common.Job{
				Task: "the x1 task",
			},
		}

		tId, err = workflow.PostTrigger(x1)
		assert.NoError(err)
	}

	var e1Id common.Ident
	{
		// will cause trigger t1Id
		e1 := &common.Event{
			EventType: etId,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}

		e1Id, err = workflow.PostEvent(eventTypeName, e1)
		assert.NoError(err)
	}

	var e2Id common.Ident
	{
		// will cause no triggers
		e2 := &common.Event{
			EventType: etId,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		e2Id, err = workflow.PostEvent(eventTypeName, e2)
		assert.NoError(err)
	}

	var aId common.Ident
	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)
		var alert0 common.Alert = (*alerts)[0]
		assert.EqualValues(e1Id, alert0.EventId)
		assert.EqualValues(tId, alert0.TriggerId)

		aId = alert0.ID
	}

	{
		err = workflow.DeleteEventType(etId)
		err = workflow.DeleteTrigger(tId)
		err = workflow.DeleteEvent(eventTypeName, e1Id)
		err = workflow.DeleteEvent(eventTypeName, e2Id)
		err = workflow.DeleteAlert(aId)
		suite.assertNoData()
	}
}

func (suite *ClientTester) TestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow

	settings, err := workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != false {
		t.Error("settings not false")
	}

	settings.Debug = true
	err = workflow.PostToAdminSettings(settings)
	assert.NoError(err)

	settings, err = workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != true {
		t.Error("settings not true")
	}
}

func (suite *ClientTester) TestAlertResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error

	a1 := common.Alert{TriggerId: "dummyT1", EventId: "dummyE1"}
	id, err := workflow.PostAlert(&a1)
	assert.NoError(err)

	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].ID)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerId)
	assert.EqualValues("dummyE1", (*alerts)[0].EventId)

	alert, err := workflow.GetAlert(id)
	assert.NoError(err)
	assert.EqualValues(id, alert.ID)

	alert, err = workflow.GetAlert("nosuchalert1")
	assert.Error(err)

	err = workflow.DeleteAlert(id)
	assert.NoError(err)

	err = workflow.DeleteAlert("nosuchalert2")
	assert.Error(err)

	alert, err = workflow.GetAlert(id)
	assert.Error(err)
	assert.Nil(alert)

	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)

	suite.assertNoData()
}

func (suite *ClientTester) TestEventTypeResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error

	mapping := map[string]piazza.MappingElementTypeName{
		"myint": piazza.MappingElementTypeString,
		"mystr": piazza.MappingElementTypeString,
	}
	eventType := &common.EventType{Name: "typnam", Mapping: mapping}
	id, err := workflow.PostEventType(eventType)

	eventTypes, err := workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*eventTypes, 1)
	assert.EqualValues(id, (*eventTypes)[0].ID)

	tmp, err := workflow.GetEventType(id)
	assert.NoError(err)
	assert.EqualValues(id, tmp.ID)

	err = workflow.DeleteEventType(id)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestEventResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error

	mapping := map[string]piazza.MappingElementTypeName{
		"myint": piazza.MappingElementTypeString,
		"mystr": piazza.MappingElementTypeString,
	}
	eventTypeName := "mytype"
	eventType := &common.EventType{Name: eventTypeName, Mapping: mapping}
	etId, err := workflow.PostEventType(eventType)

	event := &common.Event{
		EventType: etId,
		Date:      time.Now(),
		Data: map[string]interface{}{
			"myint": 17,
			"mystr": "quick",
		},
	}
	eId, err := workflow.PostEvent(eventTypeName, event)
	assert.NoError(err)

	events, err := workflow.GetAllEvents()
	assert.NoError(err)
	assert.Len(*events, 1)
	assert.EqualValues(eId, (*events)[0].ID)

	tmp, err := workflow.GetEvent(eventTypeName, eId)
	assert.NoError(err)
	assert.EqualValues(eId, tmp.ID)

	err = workflow.DeleteEvent(eventTypeName, eId)
	assert.NoError(err)

	err = workflow.DeleteEventType(etId)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestTriggerResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error

	mapping := map[string]piazza.MappingElementTypeName{
		"myint": piazza.MappingElementTypeString,
		"mystr": piazza.MappingElementTypeString,
	}
	eventType := &common.EventType{Name: "typnam", Mapping: mapping}
	etId, err := workflow.PostEventType(eventType)

	t1 := common.Trigger{
		Title: "the x1 trigger",
		Condition: common.Condition{
			EventType: etId,
			Query: `{
				"query": {
					"match": {
						"myint": 17
					}
				}
			}`,
		},
		Job: common.Job{
			Task: "the x1 task",
		},
	}
	t1Id, err := workflow.PostTrigger(&t1)
	assert.NoError(err)

	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 1)
	assert.EqualValues(t1Id, (*triggers)[0].ID)

	tmp, err := workflow.GetTrigger(t1Id)
	assert.NoError(err)
	assert.EqualValues(t1Id, tmp.ID)

	err = workflow.DeleteTrigger(t1Id)
	assert.NoError(err)
	err = workflow.DeleteEventType(etId)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestTriggering() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()

	var err error

	//-----------------------------------------------------

	var etHId, etIId, etJId common.Ident
	{
		mapping := map[string]piazza.MappingElementTypeName{
			"id":  piazza.MappingElementTypeString,
			"num": piazza.MappingElementTypeInteger,
			"str": piazza.MappingElementTypeString,
		}
		eventTypeH := &common.EventType{Name: "EventTypeH", Mapping: mapping}
		eventTypeI := &common.EventType{Name: "EventTypeI", Mapping: mapping}
		eventTypeJ := &common.EventType{Name: "EventTypeJ", Mapping: mapping}
		etHId, err = workflow.PostEventType(eventTypeH)
		assert.NoError(err)
		etIId, err = workflow.PostEventType(eventTypeI)
		assert.NoError(err)
		etJId, err = workflow.PostEventType(eventTypeJ)
		assert.NoError(err)

		eventTypes, err := workflow.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*eventTypes, 3)
	}

	////////////////

	var t1Id, t2Id common.Ident
	{
		t1 := &common.Trigger{
			Title: "the x1 trigger",
			Condition: common.Condition{
				EventType: etHId,
				Query: `{
				"query": {
					"match": {
						"str":  "quick"
					}
				}
			}`,
			},
			Job: common.Job{
				Task: "the x1 task",
			},
		}
		t1Id, err = workflow.PostTrigger(t1)
		assert.NoError(err)

		t2 := &common.Trigger{
			Title: "the x2 trigger",
			Condition: common.Condition{
				EventType: etIId,
				Query: `{
				"query": {
					"match": {
						"num": {
							"query": 18
						}
					}
				}
			}`,
			},
			Job: common.Job{
				Task: "the x2 task",
			},
		}
		t2Id, err = workflow.PostTrigger(t2)
		assert.NoError(err)

		triggers, err := workflow.GetAllTriggers()
		assert.NoError(err)
		assert.Len(*triggers, 2)
	}

	var e1Id, e2Id, e3Id common.Ident
	{
		// will cause trigger X1
		e1 := common.Event{
			EventType: etHId,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}
		e1Id, err = workflow.PostEvent("EventTypeH", &e1)
		assert.NoError(err)

		// will cause trigger X2
		e2 := common.Event{
			EventType: etIId,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}
		e2Id, err = workflow.PostEvent("EventTypeI", &e2)
		assert.NoError(err)

		// will cause no triggers
		e3 := common.Event{
			EventType: etJId,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 19,
				"str": "fox",
			},
		}
		e3Id, err = workflow.PostEvent("EventTypeJ", &e3)
		assert.NoError(err)
	}

	var a1Id, a2Id common.Ident
	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 2)

		var alertList common.AlertList
		alertList = *alerts
		alerts2 := alertList.ToSortedArray()
		assert.Len(alerts2, 2)

		a1Id = alerts2[0].ID
		a2Id = alerts2[1].ID

		assert.EqualValues(alerts2[0].TriggerId, t1Id)
		assert.EqualValues(alerts2[1].TriggerId, t2Id)
	}

	{
		workflow.DeleteTrigger(t1Id)
		workflow.DeleteTrigger(t2Id)
		workflow.DeleteEvent("EventTypeH", e1Id)
		workflow.DeleteEvent("EventTypeI", e2Id)
		workflow.DeleteEvent("EventTypeJ", e3Id)
		workflow.DeleteAlert(a1Id)
		workflow.DeleteAlert(a2Id)

		err = workflow.DeleteEventType(etHId)
		assert.NoError(err)
		err = workflow.DeleteEventType(etIId)
		assert.NoError(err)
		err = workflow.DeleteEventType(etJId)
		assert.NoError(err)
	}

	suite.assertNoData()
}