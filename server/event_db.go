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

	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type EventDB struct {
	*ResourceDB
}

func NewEventDB(es *elasticsearch.ElasticsearchClient, index string) (*EventDB, error) {

	esi := elasticsearch.NewElasticsearchIndex(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	erdb := EventDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id Ident, alertDB *AlertDB) (*[]Ident, error) {

	resp, err := db.Esi.AddPercolationDocument(eventType, data)
	if err != nil {
		return nil, err
	}

	db.Flush()

	// add the triggers to the alert queue
	ids := make([]Ident, len(resp.Matches))
	for i, v := range resp.Matches {
		ids[i] = Ident(v.Id)
		alert := Alert{ID: NewIdent(), EventId: id, TriggerId: Ident(v.Id)}
		_, err = alertDB.PostData("Alert", &alert, alert.ID)
		if err != nil {
			return nil, err
		}
	}

	alertDB.Flush()

	return &ids, nil
}

func (db *EventDB) GetByMapping(mapping string) ([]Event, error) {

	searchResult, err := db.Esi.FilterByMatchAll(mapping)
	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, nil
	}

	ary := make([]Event, searchResult.TotalHits())

	for i, hit := range searchResult.Hits.Hits {
		var tmp Event
		err = json.Unmarshal([]byte(*hit.Source), tmp)
		if err != nil {
			return nil, err
		}
		ary[i] = tmp
	}
	return ary, nil
}
