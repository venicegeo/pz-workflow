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

//---------------------------------------------------------------------------

type EventTypeDB struct {
	*ResourceDB
	mapping string
}

func NewEventTypeDB(server *Server, esi elasticsearch.IIndex) (*EventTypeDB, error) {

	rdb, err := NewResourceDB(server, esi)
	if err != nil {
		return nil, err
	}
	etrdb := EventTypeDB{ResourceDB: rdb, mapping: "EventType"}
	return &etrdb, nil
}

func (db *EventTypeDB) PostData(obj interface{}, id Ident) (Ident, error) {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return NoIdent, LoggedError("EventTypeDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return NoIdent, LoggedError("EventTypeDB.PostData failed: not created")
	}

	err = db.Esi.Flush()
	if err != nil {
		return NoIdent, err
	}

	return id, nil
}

func (db *EventTypeDB) GetAll() (*[]EventType, error) {
	var eventTypes []EventType

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &eventTypes, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, "")
	if err != nil {
		return nil, LoggedError("EventTypeDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, LoggedError("EventTypeDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var eventType EventType
			err := json.Unmarshal(*hit.Source, &eventType)
			if err != nil {
				return nil, err
			}
			eventTypes = append(eventTypes, eventType)
		}
	}

	return &eventTypes, nil
}

func (db *EventTypeDB) GetOne(id Ident) (*EventType, error) {

	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, LoggedError("EventTypeDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, LoggedError("EventTypeDB.GetOne failed: no getResult")
	}

	if !getResult.Found {
		return nil, nil
	}

	src := getResult.Source
	var eventType EventType
	err = json.Unmarshal(*src, &eventType)
	if err != nil {
		return nil, err
	}

	return &eventType, nil
}

func (db *EventTypeDB) DeleteByID(id Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return false, LoggedError("EventTypeDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventTypeDB.DeleteById failed: no deleteResult")
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}

	return deleteResult.Found, nil
}
