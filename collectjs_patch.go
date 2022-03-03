package tdtl

import (
	"fmt"
	"github.com/tkeel-io/tdtl/pkg/json/gjson"
	"github.com/tkeel-io/tdtl/pkg/json/jsonparser"
	"github.com/tkeel-io/tdtl/pkg/json/sjson"
	"strings"
)

func _gjson2JsonNode(ret gjson.Result) Node {
	switch ret.Type {
	case gjson.True:
		return BoolNode(true)
	case gjson.False:
		return BoolNode(false)
	case gjson.Number: // return  Float\Int
		r := StringNode(ret.Raw)
		if strings.Index(r.String(), ".") == -1 {
			return r.To(Int)
		}
		return r.To(Float)
	case gjson.String:
		return StringNode(ret.Str)
	case gjson.JSON:
		return JSONNode{
			value:    []byte(ret.Raw),
			datatype: JSON,
		}
	}
	return NULL_RESULT
}

func get(raw []byte, path string) *Collect {
	path = path2GJSON(path)
	ret := gjson.GetBytes(raw, path)
	return New(ret)
}

//
//func Get(raw []byte, path string) []byte {
//	//keys := path2JSONPARSER(path)
//	//
//	//if value, dataType, _, err := jsonparser.Get(raw, keys...); err == nil {
//	//	return warpValue(dataType, value)
//	//} else {
//	//
//	//}
//	path = path2GJSON(path)
//	ret := gjson.GetBytes(raw, path)
//	ee := gjson.Get(ret.String(), "")
//	fmt.Println(ee, ee.Type)
//	return []byte(ret.String())
//}

func set(raw []byte, path string, value []byte) ([]byte, error) {
	//keys := path2JSONPARSER(path)
	//return	jsonparser.Set(raw, value, keys...)
	return sjson.SetRawBytes(raw, path, value)
}
func add(raw []byte, path string, value []byte) ([]byte, error) {
	keys := path2JSONPARSER(path)
	return jsonparser.Append(raw, value, keys...)
}

func del(raw []byte, path ...string) []byte {
	for _, v := range path {
		keys := path2JSONPARSER(v)
		raw = jsonparser.Delete(raw, keys...)
	}
	return raw
}

func forEach(raw []byte, datatype Type, fn func(key []byte, value *Collect)) []byte {
	// dispose object.
	if datatype == Object {
		jsonparser.ObjectEach(raw, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			fn(key, newCollectFromJsonparserResult(dataType, value))
			return nil
		})
	}

	// dispose array.
	if datatype == Array {
		idx := 0
		jsonparser.ArrayEach(raw, func(value []byte, dataType jsonparser.ValueType, offset int) error {
			fn(Byte(fmt.Sprintf("[%d]", idx)), newCollectFromJsonparserResult(dataType, value))
			idx++
			return nil
		})
	}
	return raw
}
