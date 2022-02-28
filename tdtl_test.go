/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tdtl

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQL(t *testing.T) {
	tqlString := `
insert into entity3 select entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3, entity2.property7
`
	tql, _ := NewTDTL(tqlString, nil)
	fmt.Println(tql.Target())
}

func TestExec3(t *testing.T) {
	tqlString := `insert into entity3 
	select entity4.*,entity1.property1, entity1.property1 as target1.uuu, 
			entity2.property2.name as target2, entity1.property1 + '/' + entity2.property3 as target3`

	tqlInst, err := NewTDTL(tqlString, nil)
	assert.Nil(t, err)

	t.Log(err)
	t.Log(tqlString)
	t.Log("target: ", tqlInst.Target())
	t.Log("sources: ", tqlInst.Entities())
	for k, tentacle := range tqlInst.Fields() {
		t.Log("tentacle: ", k, tentacle)
	}

	result, err := tqlInst.Exec(map[string]Node{
		"0074c68f-679c-4290-a2be-3878c8fb75f6.sysField._spacePath": StringNode("test"),
		"entity1.property1":      StringNode("123"),
		"entity2.property2.name": StringNode("g123"),
		"entity2.property3":      StringNode("d123"),
	})

	bytes, _ := json.Marshal(result)
	t.Log(string(bytes))

	t.Log(err)
	t.Log(result)
	//map[property1: property2: property3:]
}

//func Example_TQL() {
//	tqlString := `insert into entity3 select entity1.*,entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3`
//	expr, _ := Parse(tqlString)
//	Dump(expr)
//	//OUTPUT:
//	//
//	//0  Root {
//	//     1  .  Select {
//	//     2  .  .  Field (property1) {
//	//     3  .  .  .  "ref:&{entity1.property1}"
//	//     4  .  .  }
//	//     5  .  .  Field (property2) {
//	//     6  .  .  .  "ref:&{entity2.property2.name}"
//	//     7  .  .  }
//	//     8  .  .  Field (property3) {
//	//     9  .  .  .  Op [ADD] {
//	//    10  .  .  .  .  "ref:&{entity1.property1}"
//	//    11  .  .  .  .  "ref:&{entity2.property3}"
//	//    12  .  .  .  }
//	//    13  .  .  }
//	//    14  .  }
//	//    15  .  Topic [] {}
//	//    16  .  Where {
//	//    17  .  .  <nil>
//	//    18  .  }
//	//    19  }
//}

//func TestExec3(t *testing.T) {
//	tqlString := `insert into entity3 select entity1.property1 as property1, entity2.property2.name as property2, entity1.property1 + entity2.property3 as property3`
//
//	mInstance, err := NewMapper("mapper123", tqlString)
//
//	t.Log(err)
//
//	t.Log("target: ", mInstance.TargetEntity())
//	t.Log("sources: ", mInstance.SourceEntities())
//	for _, tentacle := range mInstance.Tentacles() {
//		t.Log("tentacle: ", tentacle)
//	}
//
//	result, err := mInstance.Exec(map[string]constraint.Node{
//		"0074c68f-679c-4290-a2be-3878c8fb75f6.sysField._spacePath": constraint.NewNode("test"),
//		"entity.property2.name": constraint.NewNode("123"),
//		"entity.property3":      constraint.NewNode("g123"),
//	})
//
//	t.Log(err)
//	t.Log(result)
//}

func TestString(t *testing.T) {
	tqlText := "insert into SS4c1e33a1-6899-4643-a6b3-46cf37950b7f select 54cf69fc-78c3-4f79-9f6b-5d5e5bd8d3c0.sysField._spacePath  + '/4c1e33a1-6899-4643-a6b3-46cf37950b7f' as sysField._spacePath"
	tqlIns, err := NewTDTL(tqlText, nil)
	assert.Nil(t, err)
	t.Log(tqlIns.Target())

}
