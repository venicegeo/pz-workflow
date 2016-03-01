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

package server

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-workflow/common"
)

var eventID = 1

func NewEventID() common.Ident {
	id := common.NewIdentFromInt(eventID)
	eventID++
	return common.Ident("E" + string(id))
}

//---------------------------------------------------------------------------

type EventRDB struct {
	*ResourceDB
}

func NewEventDB(es *piazza.EsClient, index string) (*EventRDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	erdb := EventRDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventRDB) PercolateEventData(eventType string, data map[string]interface{}, id common.Ident, alertDB *AlertRDB) (*[]common.Ident, error) {

	resp, err := db.Esi.AddPercolationDocument(eventType, data)
	if err != nil {
		return nil, err
	}

	// add the triggers to the alert queue
	ids := make([]common.Ident, len(resp.Matches))
	for i,v := range resp.Matches {
		ids[i] = common.Ident(v.Id)
		alert := common.Alert{ID: NewAlertIdent(), EventId: id, TriggerId: common.Ident(v.Id)}
		_, err = alertDB.PostData("Alert", &alert, alert.ID)
		if err != nil {
			return nil, err
		}
	}

	return &ids, nil
}

func ConvertRawsToEvents(raws []*json.RawMessage) ([]common.Event, error) {
	objs := make([]common.Event, len(raws))
	for i, _ := range raws {
		err := json.Unmarshal(*raws[i], &objs[i])
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}