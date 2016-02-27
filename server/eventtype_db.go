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
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-workflow/common"
)

var eventTypeID = 1

func NewEventTypeID() common.Ident {
	id := common.NewIdentFromInt(eventTypeID)
	eventTypeID++
	return common.Ident("ET" + string(id))
}

//---------------------------------------------------------------------------


type EventTypeRDB struct {
	*ResourceDB
}

func NewEventTypeDB(es *piazza.EsClient, index string) (*EventTypeRDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	etrdb := EventTypeRDB{ResourceDB: rdb}
	return &etrdb, nil
}
