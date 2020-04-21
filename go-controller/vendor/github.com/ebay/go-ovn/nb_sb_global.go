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
	return odbi.globalSetOptionsImp(options, tableNBGlobal)
}

func (odbi *ovndb) nbGlobalGetOptionsImp() (map[string]string, error) {
	return odbi.globalGetOptionsImp(tableNBGlobal)
}

func (odbi *ovndb) sbGlobalSetOptionsImp(options map[string]string) (*OvnCommand, error) {
	return odbi.globalSetOptionsImp(options, tableSBGlobal)
}

func (odbi *ovndb) sbGlobalGetOptionsImp() (map[string]string, error) {
	return odbi.globalGetOptionsImp(tableSBGlobal)
}

func (odbi *ovndb) globalSetOptionsImp(options map[string]string, table string) (*OvnCommand, error) {
	if options == nil || table == "" {
		return nil, fmt.Errorf("Invalid arguments passed to set options: table: %s, options:  %v", table, options)
	}
	mutatemap, err := libovsdb.NewOvsMap(options)
	if err != nil {
		return nil, err
	}

	uuid, err := func() (string, error) {
		odbi.cachemutex.RLock()
		defer odbi.cachemutex.RUnlock()
		cacheGlobal, ok := odbi.cache[table]
		if !ok {
			return "", ErrorSchema
		}
		for uuid, _ := range cacheGlobal {
			return uuid, nil
		}
		return "", fmt.Errorf("No row found in %s table", table)
	}()
	if err != nil {
		return nil, err
	}
	row := make(OVNRow)
	row["options"] = mutatemap
	condition := libovsdb.NewCondition("_uuid", "==", stringToGoUUID(uuid))

	// simple mutate operation
	mutateOp := libovsdb.Operation{
		Op:    opUpdate,
		Table: table,
		Row:   row,
		Where: []interface{}{condition},
	}
	operations := []libovsdb.Operation{mutateOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) globalGetOptionsImp(table string) (map[string]string, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()
	cacheGlobal, ok := odbi.cache[table]
	if !ok {
		return nil, ErrorSchema
	}
	for _, drows := range cacheGlobal {
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
				return nil, fmt.Errorf("Error getting options field of the %s table - unsupported type", table)
			}
		}
	}
	return nil, fmt.Errorf("No row found in %s table", table)
}
