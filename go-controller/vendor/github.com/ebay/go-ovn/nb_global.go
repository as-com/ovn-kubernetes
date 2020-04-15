/**
 * Copyright (c) 2017 eBay Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package goovn

import (
	"fmt"

	"github.com/ebay/libovsdb"
)

func (odbi *ovndb) nbGlobalSetOptionsImp(options map[string]string) (*OvnCommand, error) {
	mutatemap, _ := libovsdb.NewOvsMap(options)
	mutation := libovsdb.NewMutation("options", opInsert, mutatemap)
	condition := libovsdb.NewCondition("_uuid", "==", ".")

	// simple mutate operation
	mutateOp := libovsdb.Operation{
		Op:        opMutate,
		Table:     tableNBGlobal,
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}
	operations := []libovsdb.Operation{mutateOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) nbGlobalGetOptionsImp() (map[string]string, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()
	cacheNBGlobal, ok := odbi.cache[tableNBGlobal]
	if !ok {
		return nil, ErrorSchema
	}
	for _, drows := range cacheNBGlobal {
		if options, ok := drows.Fields["options"]; ok {
			switch options.(type) {
			case libovsdb.OvsMap:
				optionsGoMap := options.(libovsdb.OvsMap).GoMap
				optionsMap := make(map[string]string)
				for k, v := range optionsGoMap {
					key, keyOk := k.(string)
					value, valueOk := v.(string)
					if !keyOk || !valueOk {
						continue
					}
					optionsMap[key] = value
				}
				return optionsMap, nil
			default:
				return nil, fmt.Errorf("Error getting options field of the NB_Global table - unsupported type")
			}
		}
	}
	return nil, fmt.Errorf("No row found in NB_Global table")
}
