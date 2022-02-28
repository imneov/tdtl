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
	"fmt"
	"github.com/tkeel-io/tdtl/pkg/json/gjson"
	"github.com/tkeel-io/tdtl/pkg/json/jsonparser"
	"strconv"
	"strings"
)

type MapHandle func(key []byte, value *Collect) Node
type ForeachHandle func(key []byte, value *Collect)
type SortHandle func(p1 *Collect, p2 *Collect) bool

var (
	UNDEFINED_RESULT = &JSONNode{datatype: Undefined}
	NULL_RESULT      = &JSONNode{datatype: Undefined}
)

// result represents a json value that is returned from Get().
type Result = gjson.Result

// Type node type
type Type int

const (
	// Undefine is Not a value
	// This isn't explicitly representable in JSON except by omitting the value.
	Undefined Type = iota
	// Null is a null json value
	Null
	// Bool is a json boolean
	Bool
	// Number is json number, include Int and Float
	Number
	// Int is json number, a discrete Int
	Int
	// Float is json number
	Float
	// String is a json string
	String
	// JSON is a raw block of JSON
	JSON
	// Object is a type of JSON
	Object
	// Array is a type of JSON
	Array
)

// String returns a string representation of the type.
func (t Type) String() string {
	switch t {
	default:
		return "Undefined"
	case Null:
		return "Null"
	case Bool:
		return "Bool"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case JSON:
		return "JSON"
	}
}

var jsonparserDatetype = map[jsonparser.ValueType]Type{
	jsonparser.NotExist: Null,
	jsonparser.String:   String,
	jsonparser.Number:   Number,
	jsonparser.Object:   Object,
	jsonparser.Array:    Array,
	jsonparser.Boolean:  Bool,
	jsonparser.Null:     Null,
	jsonparser.Unknown:  Null,
}

var gjsonDatetype = map[gjson.Type]Type{
	gjson.Null:   Null,
	gjson.Number: Number,
	gjson.String: String,
	gjson.True:   Bool,
	gjson.False:  Bool,
	gjson.JSON:   JSON,
}

func datetype(data interface{}) Type {
	switch data := data.(type) {
	case jsonparser.ValueType:
		return jsonparserDatetype[data]
	case gjson.Result:
		typ := gjsonDatetype[data.Type]
		if typ == JSON {
			if data.IsArray() {
				return Array
			}
			if data.IsObject() {
				return Object
			}
		}
		return typ
	}
	return Null
}

// True is a json true boolean

// JSON is a raw block of JSON

//Node interface
type Node interface {
	Type() Type
	To(Type) Node
	Raw() []byte
	String() string
	Error() error
}

type BoolNode bool

func (r BoolNode) Type() Type   { return Bool }
func (r BoolNode) Error() error { return nil }
func (r BoolNode) To(typ Type) Node {
	switch typ {
	case Bool:
		return r
	case String:
		return StringNode(fmt.Sprintf("%t", r))
	}
	return UNDEFINED_RESULT
}
func (r BoolNode) Raw() []byte {
	return []byte(r.String())
}
func (r BoolNode) String() string {
	return fmt.Sprintf("%t", r)
}

type IntNode int64

func (r IntNode) Type() Type   { return Int }
func (r IntNode) Error() error { return nil }
func (r IntNode) To(typ Type) Node {
	switch typ {
	case Number, Int:
		return r
	case Float:
		return FloatNode(r)
	case String:
		return StringNode(fmt.Sprintf("%d", r))
	}
	return UNDEFINED_RESULT
}
func (r IntNode) Raw() []byte {
	return []byte(r.String())
}
func (r IntNode) String() string {
	return fmt.Sprintf("%d", r)
}

type FloatNode float64

func (r FloatNode) Type() Type   { return Float }
func (r FloatNode) Error() error { return nil }
func (r FloatNode) To(typ Type) Node {
	switch typ {
	case Number, Float:
		return r
	case Int:
		return IntNode(r)
	case String:
		return StringNode(fmt.Sprintf("%f", r))
	}
	return UNDEFINED_RESULT
}
func (r FloatNode) Raw() []byte {
	return []byte(r.String())
}
func (r FloatNode) String() string {
	return fmt.Sprintf("%.6f", r)
}

type StringNode string

func (r StringNode) Type() Type   { return String }
func (r StringNode) Error() error { return nil }
func (r StringNode) To(typ Type) Node {
	switch typ {
	case String:
		return r
	case Bool:
		b, err := strconv.ParseBool(string(r))
		if err != nil {
			return UNDEFINED_RESULT
		}
		return BoolNode(b)
	case Number:
		if strings.Index(string(r), ".") == -1 {
			return r.To(Int)
		}
		return r.To(Float)
	case Int:
		b, err := strconv.ParseInt(string(r), 10, 64)
		if err != nil {
			return UNDEFINED_RESULT
		}
		return IntNode(b)
	case Float:
		b, err := strconv.ParseFloat(string(r), 64)
		if err != nil {
			return UNDEFINED_RESULT
		}
		return FloatNode(b)
	}
	return UNDEFINED_RESULT
}
func (r StringNode) Raw() []byte {
	return []byte(fmt.Sprintf("\"%s\"", r))
}
func (r StringNode) String() string {
	return string(r)
}

// JSONNode maybe Object or Array
type JSONNode struct {
	value    []byte
	path     string
	datatype Type
	offset   int
	err      error
}

func (r JSONNode) Type() Type   { return r.datatype }
func (r JSONNode) Error() error { return r.err }
func (cc JSONNode) To(typ Type) Node {
	switch typ {
	case JSON, Object, Array:
		return cc
	case Bool:
		return cc.To(String).To(Bool)
	case Number:
		return cc.To(String).To(Number)
	case Int:
		return cc.To(String).To(Int)
	case Float:
		return cc.To(String).To(Float)
	case String:
		return StringNode(cc.String())
	case Null:
		return UNDEFINED_RESULT
	case Undefined:
		return UNDEFINED_RESULT
	default:
		return UNDEFINED_RESULT
	}
	return UNDEFINED_RESULT
}
func (r JSONNode) Raw() []byte {
	switch r.datatype {
	case String:
		ret := append([]byte{byte('"')}, r.value...)
		ret = append(ret, byte('"'))
		return ret
	default:
		return []byte(r.String())
	}
}
func (r JSONNode) String() string {
	return string(r.value)
}
