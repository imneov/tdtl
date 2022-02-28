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
	"math"
	"strings"

	"github.com/tkeel-io/tdtl/parser"
)

func EvalRuleQL(ctx Context, expr Expr) Node {
	return evalRuleQL(
		MutilContext{
			DefaultValue,
			ctx,
		},
		expr,
	)
}

func EvalFilter(ctx Context, expr Expr) bool {
	return evalFilter(
		MutilContext{
			DefaultValue,
			ctx,
		},
		expr,
	)
}

func EvalSelect(ctx Context, expr Expr) Node {
	return evalSelect(
		MutilContext{
			DefaultValue,
			ctx,
		},
		expr,
	)
}

func HasDimensions(expr Expr) bool {
	if expr == nil {
		return false
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		return expr.dimensions != nil
	case *DimensionsExpr:
		return true
	}
	return false
}

func evalRuleQL(ctx Context, expr Expr) Node {
	if expr == nil {
		return UNDEFINED_RESULT
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		return evalRuleQL(ctx, expr.fields)
	case FieldsExpr:
		return evalFieldListExpr(ctx, expr)
	case *FilterExpr:
		return eval(ctx, expr.exp)
	case *BinaryExpr:
		return evalBinaryExpr(ctx, expr)
	case *CallExpr:
		return evalCallExpr(ctx, expr)
	case *JSONPathExpr:
		return evalJSONExpr(ctx, expr)
	case BoolNode:
		return expr
	case IntNode:
		return expr
	case FloatNode:
		return expr
	case StringNode:
		return expr
	case JSONNode:
		return expr
	}
	return UNDEFINED_RESULT
}

func evalFilter(ctx Context, expr Expr) bool {
	if expr == nil {
		return true
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		if expr.filter == nil || expr.filter.exp == nil {
			return true
		}
		node := eval(ctx, expr.filter)
		switch node := node.(type) {
		case BoolNode:
			return bool(node)
		}
	case FieldsExpr:
		return false
	default:
		node := eval(ctx, expr)
		switch node := node.(type) {
		case BoolNode:
			return bool(node)
		}
	}
	return false
}

func eval(ctx Context, expr Expr) Node {
	if expr == nil {
		return UNDEFINED_RESULT
	}
	switch expr := expr.(type) {
	case *FilterExpr:
		return eval(ctx, expr.exp)
	case *BinaryExpr:
		return evalBinaryExpr(ctx, expr)
	case *JSONPathExpr:
		return evalJSONExpr(ctx, expr)
	case *SwitchExpr:
		return evalSwitchExpr(ctx, expr)
	case BoolNode:
		return expr
	case IntNode:
		return expr
	case FloatNode:
		return expr
	case StringNode:
		return expr
	case JSONNode:
		return expr
	case *CallExpr:
		return evalCallExpr(ctx, expr)
	}
	return UNDEFINED_RESULT
}

func evalSelect(ctx Context, expr Expr) Node {
	if expr == nil {
		return UNDEFINED_RESULT
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		return evalSelect(ctx, expr.fields)
	case FieldsExpr:
		return evalFieldListExpr(ctx, expr)
	}
	return UNDEFINED_RESULT
}

func evalFieldListExpr(ctx Context, list FieldsExpr) Node {
	v := New("{}")
	for _, expr := range list {
		ret := eval(ctx, expr.exp)
		if expr.alias != "" {
			v.Set(expr.alias, ret)
			if v.Error() != nil {
				//fmt.Println("error in %v", v.Error())
				continue
			}
		}
	}
	return v
}

func evalBinaryExpr(ctx Context, expr *BinaryExpr) Node {
	lhs := eval(ctx, expr.LHS)
	rhs := eval(ctx, expr.RHS)
	if ret := evalBinaryOverload(expr.Op, lhs, rhs); ret != nil {
		return ret
	}

	return evalBinary(expr.Op, lhs, rhs)
}

// evalBinary eval simple types.
func evalBinaryOverload(op int, lhs, rhs Node) Node {
	// overload '+'
	switch op {
	case parser.TDTLParserADD:
		switch lhs := lhs.(type) {
		case StringNode:
			switch rhs := rhs.(type) {
			case IntNode:
				return evalBinaryOverload(op, lhs, rhs.To(String))
			case FloatNode:
				return evalBinaryOverload(op, lhs, rhs.To(String))
			case StringNode:
				return lhs + rhs
			}
		}
		switch rhs := rhs.(type) {
		case StringNode:
			switch lhs := lhs.(type) {
			case IntNode:
				return evalBinaryOverload(op, lhs.To(String), rhs)
			case FloatNode:
				return evalBinaryOverload(op, lhs.To(String), rhs)
			case StringNode:
				return lhs + rhs
			}
		}
	}
	return nil
}

// evalBinary eval simple types.
func evalBinary(op int, lhs, rhs Node) Node {
	switch lhs := lhs.(type) {
	case StringNode:
		switch rhs := rhs.(type) {
		case FloatNode, IntNode:
			return evalBinary(op, lhs.To(Number), rhs)
		case BoolNode:
			return evalBinary(op, lhs.To(Bool), rhs)
		case StringNode:
			return evalBinaryString(op, lhs, rhs)
		}
		return UNDEFINED_RESULT
	case FloatNode:
		switch rhs := rhs.(type) {
		case FloatNode:
			return evalBinaryFloat(op, lhs, rhs)
		case IntNode:
			return evalBinaryFloat(op, lhs, FloatNode(rhs))
		case StringNode:
			switch rhs := rhs.To(Float).(type) {
			case FloatNode:
				return evalBinaryFloat(op, lhs, rhs)
			}
		}
		return UNDEFINED_RESULT
	case IntNode:
		switch rhs := rhs.(type) {
		case IntNode:
			return evalBinaryInt(op, lhs, rhs)
		case FloatNode:
			return evalBinaryFloat(op, FloatNode(lhs), rhs)
		case StringNode:
			switch rhs := rhs.To(Number).(type) {
			case FloatNode:
				return evalBinaryFloat(op, FloatNode(lhs), rhs)
			case IntNode:
				return evalBinaryInt(op, lhs, rhs)
			}
		}
		return UNDEFINED_RESULT
	case BoolNode:
		switch rhs := rhs.(type) {
		case BoolNode:
			return evalBinaryBool(op, lhs, rhs)
		case StringNode:
			switch rhs := rhs.To(Bool).(type) {
			case BoolNode:
				return evalBinaryBool(op, lhs, rhs)
			}
		case *JSONNode:
			if isBooleanOP(op) {
				return evalBinary(op, BoolNode(false), rhs)
			}
		}
		return UNDEFINED_RESULT
	case *JSONNode:
		if isBooleanOP(op) {
			return BoolNode(false)
		}
		if isLogicOP(op) {
			return evalBinary(op, BoolNode(false), rhs)
		}
	default:
		return UNDEFINED_RESULT
	}
	return UNDEFINED_RESULT
}

func evalBinaryString(op int, lhs, rhs StringNode) Node {
	if !isBooleanOP(op) {
		return evalBinary(op, lhs.To(Number), rhs.To(Number))
	}
	// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
	ret := strings.Compare(string(lhs), string(rhs))
	switch op {
	case parser.TDTLLexerEQ:
		return BoolNode(ret == 0)
	case parser.TDTLLexerLT:
		return BoolNode(ret < 0)
	case parser.TDTLLexerLTE:
		return BoolNode(ret <= 0)
	case parser.TDTLLexerGT:
		return BoolNode(ret > 0)
	case parser.TDTLLexerGTE:
		return BoolNode(ret >= 0)
	case parser.TDTLLexerNE:
		return BoolNode(ret != 0)
	}
	return UNDEFINED_RESULT
}

func evalBinaryInt(op int, lhs, rhs IntNode) Node {
	switch op {
	case parser.TDTLParserADD:
		return lhs + rhs
	case parser.TDTLParserSUB:
		return lhs - rhs
	case parser.TDTLParserMUL:
		return lhs * rhs
	case parser.TDTLParserDIV:
		if rhs == 0 {
			return UNDEFINED_RESULT
		}
		return lhs / rhs
	case parser.TDTLParserMOD:
		return lhs % rhs
	case parser.TDTLLexerEQ:
		return BoolNode(lhs == rhs)
	case parser.TDTLLexerNE:
		return BoolNode(lhs != rhs)
	case parser.TDTLLexerLT:
		return BoolNode(lhs < rhs)
	case parser.TDTLLexerLTE:
		return BoolNode(lhs <= rhs)
	case parser.TDTLLexerGT:
		return BoolNode(lhs > rhs)
	case parser.TDTLLexerGTE:
		return BoolNode(lhs >= rhs)
	}
	return UNDEFINED_RESULT
}

func evalBinaryFloat(op int, lhs, rhs FloatNode) Node {
	switch op {
	case parser.TDTLParserADD:
		return lhs + rhs
	case parser.TDTLParserSUB:
		return lhs - rhs
	case parser.TDTLParserMUL:
		return lhs * rhs
	case parser.TDTLParserDIV:
		if rhs == 0 {
			return UNDEFINED_RESULT
		}
		return lhs / rhs
	case parser.TDTLParserMOD:
		return FloatNode(math.Mod(float64(lhs), float64(rhs)))
	case parser.TDTLLexerEQ:
		return BoolNode(lhs == rhs)
	case parser.TDTLLexerNE:
		return BoolNode(lhs != rhs)
	case parser.TDTLLexerLT:
		return BoolNode(lhs < rhs)
	case parser.TDTLLexerLTE:
		return BoolNode(lhs <= rhs)
	case parser.TDTLLexerGT:
		return BoolNode(lhs > rhs)
	case parser.TDTLLexerGTE:
		return BoolNode(lhs >= rhs)
	}
	return UNDEFINED_RESULT
}

func evalBinaryBool(op int, lhs, rhs BoolNode) Node {
	switch op {
	case parser.TDTLParserAND:
		return BoolNode(lhs && rhs)
	case parser.TDTLParserOR:
		return BoolNode(lhs || rhs)
	case parser.TDTLParserEQ:
		return BoolNode(lhs == rhs)
	case parser.TDTLParserNE:
		return BoolNode(lhs != rhs)
	case parser.TDTLParserNOT:
		return !rhs
	}
	return UNDEFINED_RESULT
}

func evalCallExpr(ctx Context, expr *CallExpr) Node {
	n := len(expr.args)
	if n == 0 {
		return ctx.Call(expr, []Node{})
	}
	values := make([]Node, 0, n)
	for _, expr := range expr.args {
		values = append(values, eval(ctx, expr))
	}
	ret := ctx.Call(expr, values)
	if ret.Type() != Undefined {
		return ret
	}
	return EvalCallExpr(ctx, expr)
}

var EvalCallExpr = func(ctx Context, expr *CallExpr) Node {
	return UNDEFINED_RESULT
}

func evalSwitchExpr(ctx Context, expr *SwitchExpr) Node {
	value := eval(ctx, expr.exp)
	for _, e := range expr.list {
		if value == eval(ctx, e.when) {
			return eval(ctx, e.then)
		}
	}
	if expr.last != nil {
		return eval(ctx, expr.last)
	}
	return UNDEFINED_RESULT
}

func evalJSONExpr(ctx Context, expr *JSONPathExpr) Node {
	return ctx.Value(expr.val)
}

func isBooleanOP(op int) bool {
	switch op {
	case parser.TDTLParserEQ,
		parser.TDTLParserGT,
		parser.TDTLParserLT,
		parser.TDTLParserGTE,
		parser.TDTLParserLTE:
		return true
	}
	return false
}

func isLogicOP(op int) bool {
	switch op {
	case parser.TDTLParserAND,
		parser.TDTLParserOR,
		parser.TDTLParserNOT:
		return true
	}
	return false
}

func GetTopic(expr Expr) (string, bool) {
	if expr == nil {
		return "", false
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		if expr.topic == nil {
			return "", false
		}
		s := strings.Join(expr.topic, "/")
		return s, true
	}
	return "", false
}

func GetWindow(expr Expr) *WindowExpr {
	if expr == nil {
		return nil
	}
	switch expr := expr.(type) {
	case *SelectStatementExpr:
		if expr.dimensions == nil {
			return nil
		}
		return expr.dimensions.window
	}
	return nil
}

func evalDimensions(ctx Context, exprs []*JSONPathExpr) (dimension Node) {
	if exprs == nil {
		return UNDEFINED_RESULT
	}
	keys := make([]string, 0, len(exprs))
	for _, expr := range exprs {
		keys = append(keys, eval(ctx, expr).String())
	}

	return StringNode(strings.Join(keys, "-"))
}
