package interpreter

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func Interpret(n *Node, env *Environment) (*Node, error) {
	// fmt.Printf("Interpret: \n%s\n\n", n.ToString())
	switch n.Type {
	case StmtNT:
		return interpretStmt(n, env)
	// binary operations
	case AddNT, SubtNT, DivNT, MultNT, ModuloNT:
		return interpretMathOp(n, env)
	case PowerNT:
		return interpretPower(n, env)
	case LogicAndNT, LogicOrNT, FallbackNT:
		return interpretLogicOp(n, env)
	case EqualNT, NotEqualNT, LessEqualNT, GreaterEqualNT, LessNT, GreaterNT:
		return interpretComparison(n, env)
	case InNT:
		return interpretIn(n, env)

	// unary operations
	case LogicNotNT, MaybeNT, CardinalityNT, UnaryNegNT:
		return interpretUnOp(n, env)
	// identifiers
	case IdentifierNT, UnderscoreNT, IndexNT:
		return resolveIdentifier(n, env)
	// literals
	case IntNT, FloatNT, BoolNT, StringNT, FailNT, SuccessNT, NullNT, SetNT:
		return &Node{
			Type:  n.Type,
			Val:   n.Val,
			Scope: n.Scope,
			L:     n.L,
			R:     n.R,
		}, nil
	case LambdaNT:
		scope := n.Scope
		if scope == nil {
			scope = &Environment{
				Parent: env,
				Consts: map[string]*Node{},
				Vars:   map[string]*Node{},
			}
		}
		return &Node{
			Type:  n.Type,
			Val:   n.Val,
			Scope: scope,
			L:     n.L,
			R:     n.R,
		}, nil
	case ObjectNT:
		if n.Scope == nil {
			return &Node{
				Type: ObjectNT,
				Scope: &Environment{
					Vars: map[string]*Node{},
				},
			}, nil
		}
		return n, nil
	case ModuleNT:
		return n, nil
	case ListNT:
		list := []*Node{}
		for _, m := range n.Val.([]*Node) {
			val, err := Interpret(m, env)
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		}
		return &Node{
			Type: ListNT,
			Val:  list,
		}, nil
	case KVPairNT:
		return interpretKVPair(n, env)
	case SetItemNT:
		return interpretSetItem(n, env)
	case ConstDeclNT, VarDeclNT:
		return declareVar(n, env)
	case AssignmentNT:
		return assignVar(n, env)
	case IfNT:
		return interpretIf(n, env)
	case CallNT:
		return interpretCall(n, env)
	case ReturnStmtNT:
		returnVal, err := Interpret(n.R, env)
		return &Node{Type: ReturnStmtNT, R: returnVal}, err
	case MapNT:
		return interpretMap(n, env)
	case WhereNT:
		return interpretWhere(n, env)
	case PipeNT:
		return interpretPipe(n, env)
	case BracketAccessNT:
		return interpretBracketAccess(n, env)
	case FieldAccessNT:
		return interpretFieldAccess(n, env)
	case ListSliceNT:
		return interpretListSlice(n, env)
	case BreakNT, ContinueNT:
		return n, nil
	case WhileStmtNT:
		return interpretWhile(n, env)
	case ForStmtNT:
		return interpretFor(n, env)
	case RangeNT:
		return interpretRange(n, env)
	case ImportNT:
		return importModule(n, env)
	}

	return nil, fmt.Errorf("Unknown node type")
}

func resolveIdentifier(n *Node, env *Environment) (res *Node, err error) {
	ident := n.Val.(string)
	for e := env; e != nil; e = e.Parent {
		if val, ok := e.Consts[ident]; ok {
			return val, nil
		}
		if val, ok := e.Vars[ident]; ok {
			return val, nil
		}
	}

	return nil, fmt.Errorf("\"%s\" is undefined", ident)
}

func declareVar(n *Node, env *Environment) (res *Node, err error) {
	val, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	// assign, err := getAssignmentTarget(n.L, env)
	// if err != nil {
	// 	return nil, err
	// }
	ident := n.L.Val.(string)
	if _, exists := env.Consts[ident]; exists {
		return nil, fmt.Errorf("\"%s\" is already defined", ident)
	}
	if _, exists := env.Vars[ident]; exists {
		return nil, fmt.Errorf("\"%s\" is already defined", ident)
	}

	// assign(val)
	if n.Type == VarDeclNT {
		env.Vars[ident] = val
	} else {
		env.Consts[ident] = val
	}

	return &Node{Type: SuccessNT}, nil
}

func assignVar(n *Node, env *Environment) (res *Node, err error) {
	assign, err := getAssignmentTarget(n.L, env, false)
	if err != nil {
		return nil, err
	}

	val, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	err = assign(val)
	if err != nil {
		return nil, err
	}

	return &Node{Type: SuccessNT}, nil
}

func getAssignmentTarget(lhs *Node, env *Environment, constant bool) (assignFunc func(*Node) error, err error) {
	// basic identifiers
	if lhs.L == nil && lhs.Type == IdentifierNT {
		ident := lhs.Val.(string)
		for e := env; e != nil; e = e.Parent {
			if _, exists := e.Consts[ident]; exists {
				return nil, fmt.Errorf("Cannot assign to constant variable \"%s\"", ident)
			}
			if _, exists := e.Vars[ident]; exists {
				return func(n *Node) error {
					if constant {
						e.Consts[ident] = n
					} else {
						e.Vars[ident] = n
					}
					return nil
				}, nil
			}
		}
		return nil, fmt.Errorf("Cannot assign to undefined variable \"%s\"", ident)
	}

	// destructured assigns
	if lhs.Type == KVPairNT || lhs.Type == ListNT {
		return getDestructuredAssign(lhs, env)
	}

	// nested assigns, e.g. list index or object field
	if lhs.L != nil && lhs.L.Type == ListNT || lhs.L.Type == ObjectNT {
		return getNestedAssign(lhs, env)
	}

	return func(_ *Node) error {
		return nil
	}, fmt.Errorf("Invalid assignemnt target")
}

func getDestructuredAssign(assignee *Node, env *Environment) (assignFunc func(*Node) error, err error) {
	switch assignee.Type {
	case ListNT:
		return func(n *Node) error {
			if n.Type != ListNT {
				for _, m := range assignee.Val.([]*Node) {
					env.Consts[m.Val.(string)] = &Node{Type: FailNT}
				}
				return nil
			}

			for i, m := range assignee.Val.([]*Node) {
				if i < len(n.Val.([]*Node)) {
					env.Consts[m.Val.(string)] = n.Val.([]*Node)[i]
				} else {
					env.Consts[m.Val.(string)] = &Node{Type: FailNT}
				}
			}

			return nil
		}, nil

	}

	return nil, fmt.Errorf("Invalid assignemnt target")
}

func getNestedAssign(assignee *Node, env *Environment) (assignFunc func(*Node) error, err error) {
	// assignments to list indexes and object fields
	container, err := Interpret(assignee.L, env)
	if err != nil {
		return nil, err
	}

	switch container.Type {
	case ListNT:
		{
			idxNode, err := Interpret(assignee.R, env)
			length := len(container.Val.([]*Node))
			if err != nil {
				return nil, err
			}
			switch idxNode.Type {
			case IntNT:
				idx := int(idxNode.Val.(int64))
				if idx < 0 {
					idx = length + idx
				}
				if idx >= length || (idx < 0 && -idx >= length+1) {
					return nil, fmt.Errorf("Cannot assign to list. Index out of range.")
				}
				return func(n *Node) error {
					container.Val.([]*Node)[idx] = n
					return nil
				}, nil
			case FloatNT:
				idx := int(idxNode.Val.(float64))
				if idx < 0 {
					idx = length + idx
				}
				if idx >= length || (idx < 0 && -idx >= length+1) {
					return nil, fmt.Errorf("Cannot assign to list. Index out of range.")
				}
				return func(n *Node) error {
					container.Val.([]*Node)[idx] = n
					return nil
				}, nil
			default:
				return nil, fmt.Errorf("Cannot assign to list index. Invalid index.")
			}
		}
	case ObjectNT:
		{
			if assignee.R.Type != StringNT && assignee.R.Type != IdentifierNT {
				return nil, fmt.Errorf("Cannot assign. Invalid object field key.")
			}

			var key string
			if assignee.Type == BracketAccessNT && assignee.R.Type == IdentifierNT {
				val, err := Interpret(assignee.R, env)
				if err != nil {
					return nil, err
				}
				if val.Type != StringNT {
					return nil, fmt.Errorf("Cannot assign. Invalid object field key.")
				}
				key = val.Val.(string)
			} else {
				key = assignee.R.Val.(string)
			}

			return func(n *Node) error {
				container.Scope.Vars[key] = n
				return nil
			}, nil
		}
	default:
		return nil, fmt.Errorf("Invalid assignment target.")
	}
}

func interpretStmt(root *Node, env *Environment) (res *Node, err error) {
	for n := root; n != nil; n = n.R {
		if n.L != nil && n.L.Type == StmtNT {
			res, err = Interpret(n.L, &Environment{
				Parent: env,
				Consts: map[string]*Node{},
				Vars:   map[string]*Node{},
			})
		} else {
			res, err = Interpret(n.L, env)
		}

		if err != nil {
			break
		}
	}

	return res, err
}

func interpretMathOp(n *Node, env *Environment) (res *Node, err error) {
	if n.L == nil {
		return nil, fmt.Errorf("Missing first argument for operation \"%s\"", nodeTypeMap[n.Type])
	}
	if n.R == nil {
		return nil, fmt.Errorf("Missing second argument for operation \"%s\"", nodeTypeMap[n.Type])
	}

	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}
	rhs, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	l, r, t := maybeCastNumbers(lhs, rhs)
	switch n.Type {
	case AddNT:
		{
			switch t {
			case IntNT:
				return &Node{
					Type: IntNT,
					Val:  l.Val.(int64) + r.Val.(int64),
				}, nil
			case FloatNT:
				return &Node{
					Type: FloatNT,
					Val:  l.Val.(float64) + r.Val.(float64),
				}, nil
			case StringNT:
				return &Node{
					Type: StringNT,
					Val:  l.Val.(string) + r.Val.(string),
				}, nil
			case ListNT:
				combined := l.Val.([]*Node)
				for _, x := range r.Val.([]*Node) {
					combined = append(combined, x)
				}
				return &Node{
					Type: ListNT,
					Val:  combined,
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	case SubtNT:
		{
			switch t {
			case IntNT:
				return &Node{
					Type: IntNT,
					Val:  l.Val.(int64) - r.Val.(int64),
				}, nil
			case FloatNT:
				return &Node{
					Type: FloatNT,
					Val:  l.Val.(float64) - r.Val.(float64),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	case DivNT:
		{
			switch t {
			case IntNT:
				if r.Val.(int64) == 0 {
					return &Node{Type: FailNT}, nil
				}
				return &Node{
					Type: FloatNT,
					Val:  float64(l.Val.(int64)) / float64(r.Val.(int64)),
				}, nil
			case FloatNT:
				if r.Val.(float64) == 0 {
					return &Node{Type: FailNT}, nil
				}
				return &Node{
					Type: FloatNT,
					Val:  l.Val.(float64) / r.Val.(float64),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	case MultNT:
		{
			switch t {
			case IntNT:
				return &Node{
					Type: IntNT,
					Val:  l.Val.(int64) * r.Val.(int64),
				}, nil
			case FloatNT:
				return &Node{
					Type: FloatNT,
					Val:  l.Val.(float64) * r.Val.(float64),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	case ModuloNT:
		{
			switch t {
			case IntNT:
				if r.Val.(int64) == 0 {
					return &Node{Type: FailNT}, nil
				}
				return &Node{
					Type: IntNT,
					Val:  l.Val.(int64) % r.Val.(int64),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	}

	return nil, fmt.Errorf("Unknown binary operator")
}

func interpretPower(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}
	rhs, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	if rhs.Type != IntNT {
		return &Node{Type: FailNT}, nil
	}

	if lhs.Type == FloatNT {
		var total float64 = 1
		var i int64 = 0
		x := rhs.Val.(int64)
		if rhs.Val.(int64) < 0 {
			x = -x
		}
		for ; i < x; i++ {
			total *= lhs.Val.(float64)
		}
		return &Node{
			Type: FloatNT,
			Val:  total,
		}, nil
	}

	if lhs.Type == IntNT {
		var total int64 = 1
		var i int64 = 0
		x := rhs.Val.(int64)
		if rhs.Val.(int64) < 0 {
			x = -x
		}
		for ; i < x; i++ {
			total *= lhs.Val.(int64)
		}
		if rhs.Val.(int64) < 0 {
			return &Node{
				Type: FloatNT,
				Val:  1 / float64(total),
			}, nil
		}
		return &Node{
			Type: IntNT,
			Val:  total,
		}, nil
	}

	return &Node{Type: FailNT}, nil
}

// maybeCastNumbers casts numbers to floats if one is a float
func maybeCastNumbers(a, b *Node) (*Node, *Node, NodeType) {
	if a == nil || b == nil {
		return nil, nil, ErrorNT
	}

	switch a.Type {
	case IntNT:
		if b.Type == IntNT {
			return a, b, IntNT
		} else if b.Type == FloatNT {
			return &Node{
				Type: FloatNT,
				Val:  float64(a.Val.(int64)),
			}, b, FloatNT
		} else {
			return a, b, ErrorNT
		}
	case FloatNT:
		if b.Type == IntNT {
			return a, &Node{
				Type: FloatNT,
				Val:  float64(b.Val.(int64)),
			}, FloatNT
		} else if b.Type == FloatNT {
			return a, b, FloatNT
		} else {
			return a, b, ErrorNT
		}
	case StringNT:
		switch b.Type {
		case IntNT, FloatNT:
			return a, &Node{
				Type: StringNT,
				Val:  b.ToString(),
			}, StringNT
		case StringNT:
			return a, b, StringNT
		default:
			return a, b, ErrorNT
		}
	case ListNT:
		if b.Type == ListNT {
			return a, b, ListNT
		}
		return a, b, ErrorNT
	case BoolNT:
		if b.Type == BoolNT {
			return a, b, BoolNT
		}
		return a, b, ErrorNT
	case SuccessNT:
		if b.Type == SuccessNT {
			return a, b, SuccessNT
		}
		return a, b, ErrorNT
	case FailNT:
		if b.Type == FailNT {
			return a, b, FailNT
		}
		return a, b, ErrorNT
	case NullNT:
		if b.Type == NullNT {
			return a, b, NullNT
		}
		return a, b, ErrorNT
	default:
		return a, b, ErrorNT
	}
}

func interpretLogicOp(n *Node, env *Environment) (res *Node, err error) {
	if n.L == nil {
		return nil, fmt.Errorf("Missing first argument for operation \"%s\"", nodeTypeMap[n.Type])
	}
	if n.R == nil {
		return nil, fmt.Errorf("Missing second argument for operation \"%s\"", nodeTypeMap[n.Type])
	}

	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}
	rhs, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	switch n.Type {
	case LogicAndNT:
		// should and/or always return bool?
		if isTruthy(lhs) {
			return rhs, nil
		}
		return &Node{Type: BoolNT, Val: false}, nil
	case LogicOrNT:
		if isTruthy(lhs) {
			return lhs, nil
		}
		return rhs, nil
	case FallbackNT:
		if lhs.Type == FailNT {
			return rhs, nil
		}
		return lhs, nil
	}

	return nil, fmt.Errorf("Unknown logical operator")
}

func evalEquality(a, b *Node) (bool, error) {
	l, r, t := maybeCastNumbers(a, b)
	switch t {
	case IntNT:
		return l.Val.(int64) == r.Val.(int64), nil
	case FloatNT:
		return l.Val.(float64) == r.Val.(float64), nil
	case StringNT:
		return l.Val.(string) == r.Val.(string), nil
	case ListNT:
		if len(l.Val.([]*Node)) != len(r.Val.([]*Node)) {
			return false, nil
		}
		for i, n := range a.Val.([]*Node) {
			equal, err := evalEquality(n, r.Val.([]*Node)[i])
			if !equal || err != nil {
				return false, err
			}
		}
		return true, nil
	case BoolNT:
		return l.Val.(bool) == r.Val.(bool), nil
	case SuccessNT:
		return true, nil
	case FailNT:
		return true, nil
	case NullNT:
		return true, nil
	default:
		return false, fmt.Errorf("Cannot compare types")
	}
}

func interpretComparison(n *Node, env *Environment) (res *Node, err error) {
	if n.L == nil {
		return nil, fmt.Errorf("Missing first argument for operation \"%s\"", nodeTypeMap[n.Type])
	}
	if n.R == nil {
		return nil, fmt.Errorf("Missing second argument for operation \"%s\"", nodeTypeMap[n.Type])
	}

	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}
	rhs, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	switch n.Type {
	case EqualNT:
		equal, err := evalEquality(lhs, rhs)
		if err != nil {
			return &Node{Type: FailNT}, nil
		}
		return &Node{Type: BoolNT, Val: equal}, nil
	case NotEqualNT:
		equal, err := evalEquality(lhs, rhs)
		if err != nil {
			return &Node{Type: FailNT}, nil
		}
		return &Node{Type: BoolNT, Val: !equal}, nil
	case LessEqualNT:
		l, r, t := maybeCastNumbers(lhs, rhs)
		switch t {
		case IntNT:
			return &Node{Type: BoolNT, Val: l.Val.(int64) <= r.Val.(int64)}, nil
		case FloatNT:
			return &Node{Type: BoolNT, Val: l.Val.(float64) <= r.Val.(float64)}, nil
		default:
			return &Node{Type: FailNT}, nil
		}
	case GreaterEqualNT:
		l, r, t := maybeCastNumbers(lhs, rhs)
		switch t {
		case IntNT:
			return &Node{Type: BoolNT, Val: l.Val.(int64) >= r.Val.(int64)}, nil
		case FloatNT:
			return &Node{Type: BoolNT, Val: l.Val.(float64) >= r.Val.(float64)}, nil
		default:
			return &Node{Type: FailNT}, nil
		}
	case LessNT:
		l, r, t := maybeCastNumbers(lhs, rhs)
		switch t {
		case IntNT:
			return &Node{Type: BoolNT, Val: l.Val.(int64) < r.Val.(int64)}, nil
		case FloatNT:
			return &Node{Type: BoolNT, Val: l.Val.(float64) < r.Val.(float64)}, nil
		default:
			return &Node{Type: FailNT}, nil
		}
	case GreaterNT:
		l, r, t := maybeCastNumbers(lhs, rhs)
		switch t {
		case IntNT:
			return &Node{Type: BoolNT, Val: l.Val.(int64) > r.Val.(int64)}, nil
		case FloatNT:
			return &Node{Type: BoolNT, Val: l.Val.(float64) > r.Val.(float64)}, nil
		default:
			return &Node{Type: FailNT}, nil
		}
	}

	return nil, fmt.Errorf("Unknown comparison operator")
}

func interpretIn(n *Node, env *Environment) (res *Node, err error) {
	item, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}
	container, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	switch container.Type {
	case ListNT:
		for i := 0; i < len(container.Val.([]*Node)); i++ {
			equal, _ := evalEquality(item, container.Val.([]*Node)[i])
			if equal {
				return &Node{
					Type: BoolNT,
					Val:  true,
				}, nil
			}
		}

		return &Node{
			Type: BoolNT,
			Val:  false,
		}, nil
	case SetNT:
		set := container.Val.(map[Value]bool)
		return &Node{
			Type: BoolNT,
			Val:  set[item.toValue()],
		}, nil

	default:
		return &Node{Type: FailNT}, nil
	}
}

func isTruthy(n *Node) bool {
	if n == nil {
		return false
	}
	switch n.Type {
	case SuccessNT:
		return true
	case FailNT:
		return false
	case FloatNT:
		return n.Val.(float64) != 0
	case IntNT:
		return n.Val.(int64) != 0
	case BoolNT:
		return n.Val.(bool)
	case StringNT:
		return len(n.Val.(string)) != 0
	case NullNT:
		return false
	default:
		return true
	}
}

func interpretUnOp(n *Node, env *Environment) (res *Node, err error) {
	if n.R == nil {
		return nil, fmt.Errorf("Missing argument for unary operation \"%s\"", nodeTypeMap[n.Type])
	}

	arg, err := Interpret(n.R, env)
	if err != nil {
		return arg, err
	}
	switch n.Type {
	case LogicNotNT:
		return &Node{
			Type: BoolNT,
			Val:  !isTruthy(arg),
		}, nil
	case MaybeNT:
		if arg.Type == FailNT {
			return arg, nil
		}
		return &Node{Type: SuccessNT}, nil
	case CardinalityNT:
		{
			switch arg.Type {
			case ListNT:
				return &Node{
					Type: IntNT,
					Val:  int64(len(arg.Val.([]*Node))),
				}, nil
			case StringNT:
				return &Node{
					Type: IntNT,
					Val:  int64(len(arg.Val.(string))),
				}, nil
			case SetNT:
				return &Node{
					Type: IntNT,
					Val:  int64(len(arg.Val.(map[Value]bool))),
				}, nil
			case ObjectNT:
				return &Node{
					Type: IntNT,
					Val:  int64(len(arg.Val.(map[string]*Node))),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	case UnaryNegNT:
		{
			switch arg.Type {
			case IntNT:
				return &Node{
					Type: IntNT,
					Val:  -arg.Val.(int64),
				}, nil
			case FloatNT:
				return &Node{
					Type: FloatNT,
					Val:  -arg.Val.(float64),
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		}
	}

	return nil, fmt.Errorf("Unknown unary operator")
}

func interpretIf(n *Node, env *Environment) (res *Node, err error) {
	cond, result := n.L, n.R
	condRes, err := Interpret(cond, env)
	if err != nil {
		return nil, err
	}

	if isTruthy(condRes) {
		if result.Type == ThenNT {
			return Interpret(result.L, env)
		} else {
			return Interpret(result, env)
		}
	} else {
		if result.Type == ThenNT {
			return Interpret(result.R, env)
		} else {
			return &Node{Type: FailNT}, nil
		}
	}
}

func countArgs(params, args *Node) (p, a int) {
	for param := params; ; p++ {
		if param == nil || (param.Val == nil && param.L == nil) {
			break
		}
		param = param.R
	}

	for arg := args; ; a++ {
		if arg == nil || arg.L == nil {
			break
		}
		arg = arg.R
	}

	return p, a
}

func assignArg(arg, param *Node, scope *Environment) {
	if param == nil || (param.L == nil && param.Val == nil) {
		return
	}

	// identifier: 	param{Val: string, R: nextParam}
	// list:				param{L: List of idents, R: nextParam}
	// object:			param{L: KVPair{Val: key, L: nil | ident, R: nil | nextField}, R: nextParam}

	// plain parameter
	if param.Val != nil {
		scope.Vars[param.Val.(string)] = arg
		return
	}

	// destructured param
	switch param.L.Type {
	case ListNT:
		{
			if arg.Type != ListNT {
				for _, p := range param.L.Val.([]*Node) {
					scope.Vars[p.Val.(string)] = &Node{Type: FailNT}
				}
				return
			}
			as := arg.Val.([]*Node)
			for i, p := range param.L.Val.([]*Node) {
				if i >= len(as) {
					scope.Vars[p.Val.(string)] = &Node{Type: FailNT}
				}
				scope.Vars[p.Val.(string)] = as[i]
			}
			return
		}
	case KVPairNT, IdentifierNT:
		{
			if arg.Type != ObjectNT {
				// The arg is not an object
				for p := param.L; p != nil; p = p.R {
					if p.L != nil {
						assignArg(&Node{Type: FailNT}, p.L, scope)
					} else {
						scope.Vars[p.Val.(string)] = &Node{Type: FailNT}
					}
				}
				return
			}

			obj := arg.Scope.Vars
			for p := param.L; p != nil; p = p.R {
				old := p.Val.(string)
				if p.L != nil {
					// rename the object field
					new := p.L.Val.(string)
					val, ok := obj[old]
					if ok {
						// The field exists. Add to scope
						scope.Vars[new] = val
					} else {
						// The field does not exist on the arg
						scope.Vars[new] = &Node{Type: FailNT}
					}
				} else {
					// Add to scope without renaming
					scope.Vars[old] = obj[old]
				}
			}
		}

	}

}

func interpretCall(n *Node, env *Environment) (res *Node, err error) {
	callee := n.L
	var lambda *Node
	if callee.Type == IdentifierNT {
		lambda, err = resolveIdentifier(callee, env)
	} else {
		lambda, err = Interpret(callee, env)
	}

	if err != nil {
		return nil, err
	}

	// built-in functions
	if lambda.Func != nil {
		args := []*Node{}
		for arg := n.R; arg != nil && arg.L != nil; arg = arg.R {
			val, err := Interpret(arg.L, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		return lambda.Func(env, args...)
	}

	scope := &Environment{
		Consts: map[string]*Node{},
		Vars:   map[string]*Node{},
	}
	if lambda.Scope != nil {
		scope.Parent = lambda.Scope
	} else {
		scope.Parent = env
	}

	ps, as := countArgs(lambda.L, n.R)
	if ps > as {
		if callee.Type == IdentifierNT {
			return nil, fmt.Errorf("Too few arguments provided to function \"%s\". Expected %d, received %d.", callee.Val.(string), ps, as)
		}
		return nil, fmt.Errorf("Too few arguments provided to anonymous function. Expected %d, received %d.", ps, as)
	}

	if ps < as {
		if callee.Type == IdentifierNT {
			return nil, fmt.Errorf("Too many arguments provided to function \"%s\". Expected %d, received %d.", callee.Val.(string), ps, as)
		}
		return nil, fmt.Errorf("Too many arguments provided to anonymous function. Expected %d, received %d.", ps, as)
	}

	param, arg := lambda.L, n.R
	// assign arguments to function scope
	for param != nil && (param.Val != nil || param.L != nil) && arg != nil && arg.L != nil {
		val, err := Interpret(arg.L, env)
		if err != nil {
			return nil, err
		}

		assignArg(val, param, scope)
		param, arg = param.R, arg.R
	}

	// fmt.Println("\nScope:")
	// for k := range scope.Vars {
	// 	fmt.Printf("%s: \t%s\n", k, scope.Vars[k].ToString())
	// }

	if lambda.R.Type == StmtNT {
		res, err = interpretFunctionBody(lambda.R, scope)
	} else {
		res, err = Interpret(lambda.R, scope)
	}

	if err != nil {
		return res, err
	}

	if res.Type == LambdaNT {
		res.Scope = scope
	}
	return res, err
}

func interpretFunctionBody(start *Node, env *Environment) (res *Node, err error) {
	for n := start; n != nil; n = n.R {
		if n.L != nil && n.L.Type == StmtNT {
			res, err = Interpret(n.L, &Environment{
				Parent: env,
				Consts: map[string]*Node{},
				Vars:   map[string]*Node{},
			})
		} else {
			res, err = Interpret(n.L, env)
		}

		if err != nil {
			return res, err
		}

		if res.Type == ReturnStmtNT {
			return res.R, nil
		}
	}

	return res, err
}

func iterateCollection(n *Node) func() *Node {
	switch n.Type {
	case ListNT:
		list := n.Val.([]*Node)
		i := -1
		return func() *Node {
			if i < len(list)-1 {
				i++
				return list[i]
			}
			return nil
		}
	case ObjectNT:
		obj := n.Scope.Vars
		keys := []string{}
		for k := range obj {
			keys = append(keys, k)
		}
		i := -1
		return func() *Node {
			if i < len(keys)-1 {
				i++
				return &Node{
					Type: StringNT,
					Val:  keys[i],
				}
			}
			return nil
		}
	case SetNT:
		set := n.Val.(map[Value]bool)
		items := []*Node{}
		for k := range set {
			if set[k] {
				items = append(items, k.toNode())
			}
		}
		i := -1
		return func() *Node {
			if i < len(items)-1 {
				i++
				return items[i]
			}
			return nil
		}
	default:
		return func() *Node {
			return nil
		}
	}
}

func interpretMap(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if lhs.Type == FailNT || (lhs.Type != ListNT && lhs.Type != SetNT) {
		return &Node{Type: FailNT}, nil
	}

	callee := n.R
	var lambda *Node
	if callee.Type == IdentifierNT {
		lambda, err = resolveIdentifier(callee, env)
	} else {
		lambda, err = Interpret(callee, env)
	}

	if err != nil {
		return nil, err
	}

	if lambda.Type != LambdaNT {
		return &Node{Type: FailNT}, nil
	}

	newList := []*Node{}
	newSet := map[Value]bool{}

	next := iterateCollection(lhs)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		old, err := Interpret(item, env)
		if err != nil {
			return nil, err
		}

		env.Consts["index"] = &Node{Type: IntNT, Val: int64(i)}

		var new *Node
		if lambda.Func != nil {
			new, err = lambda.Func(env, item)
		} else {
			call := &Node{
				Type: CallNT,
				L:    lambda,
				R: &Node{
					Type: ArgNT,
					L:    old,
				},
			}
			new, err = Interpret(call, env)
		}

		if lhs.Type == ListNT {
			newList = append(newList, new)
		}
		if lhs.Type == SetNT {
			newSet[new.toValue()] = true
		}
	}
	env.Consts["_"] = nil
	env.Consts["index"] = nil

	if lhs.Type == SetNT {
		return &Node{
			Type: SetNT,
			Val:  newSet,
		}, nil
	}

	return &Node{
		Type: ListNT,
		Val:  newList,
	}, nil
}

func interpretWhere(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if lhs.Type == FailNT || (lhs.Type != ListNT && lhs.Type != SetNT) {
		return &Node{Type: FailNT}, nil
		// return nil, fmt.Errorf("Invalid argument provided to \"where\"")
	}

	callee := n.R
	var lambda *Node
	if callee.Type == IdentifierNT {
		lambda, err = resolveIdentifier(callee, env)
	} else {
		lambda, err = Interpret(callee, env)
	}

	if err != nil {
		return nil, err
	}

	if lambda.Type != LambdaNT {
		return &Node{Type: FailNT}, nil
		// return nil, fmt.Errorf("Invalid second argument provided to \"where\"")
	}

	newList := []*Node{}
	newSet := map[Value]bool{}

	next := iterateCollection(lhs)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		val, err := Interpret(item, env)
		if err != nil {
			return nil, err
		}

		env.Consts["index"] = &Node{Type: IntNT, Val: int64(i)}
		var result *Node
		if lambda.Func != nil {
			result, err = lambda.Func(env, item)
		} else {
			call := &Node{
				Type: CallNT,
				L:    lambda,
				R: &Node{
					Type: ArgNT,
					L:    val,
				},
			}
			result, err = Interpret(call, env)
		}

		if err != nil {
			return nil, err
		}

		if isTruthy(result) {
			if lhs.Type == ListNT {
				newList = append(newList, val)
			}
			if lhs.Type == SetNT {
				newSet[val.toValue()] = true
			}
		}
	}
	env.Consts["_"] = nil
	env.Consts["index"] = nil

	if lhs.Type == SetNT {
		return &Node{
			Type: SetNT,
			Val:  newSet,
		}, nil
	}

	return &Node{
		Type: ListNT,
		Val:  newList,
	}, nil
}

func interpretPipe(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if lhs.Type == FailNT {
		return lhs, nil
	}

	callee := n.R
	var lambda *Node
	if callee.Type == IdentifierNT {
		lambda, err = resolveIdentifier(callee, env)
	} else {
		lambda, err = Interpret(callee, env)
	}

	if err != nil {
		return nil, err
	}

	// built-in functions
	if lambda.Func != nil {
		return lambda.Func(env, lhs)
	}

	call := &Node{
		Type: CallNT,
		L:    lambda,
		R: &Node{
			Type: ArgNT,
			L:    lhs,
		},
	}

	return Interpret(call, env)
}

func interpretBracketAccess(n *Node, env *Environment) (res *Node, err error) {
	src, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	accessor, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	if src.Type == ListNT || src.Type == StringNT {
		return getByIndex(src, accessor)
	}

	if src.Type == ObjectNT {
		return getByName(src, accessor)
	}

	return &Node{Type: FailNT}, nil
}

func getByIndex(src, idxNode *Node) (res *Node, err error) {
	var idx int64
	switch idxNode.Type {
	case IntNT:
		idx = idxNode.Val.(int64)
	case FloatNT:
		idx = int64(idxNode.Val.(float64))
	default:
		return &Node{Type: FailNT}, nil
	}

	var length int64
	switch src.Type {
	case ListNT:
		length = int64(len(src.Val.([]*Node)))
	case StringNT:
		length = int64(len(src.Val.(string)))

	}

	if idx >= length {
		return &Node{Type: FailNT}, nil
	}

	if idx < 0 && -idx > length {
		return &Node{Type: FailNT}, nil
	}

	if idx < 0 {
		return src.Val.([]*Node)[length+idx], nil
	}

	if src.Type == StringNT {
		return &Node{
			Type: StringNT,
			Val:  string(src.Val.(string)[idx]),
		}, nil
	}
	return src.Val.([]*Node)[idx], nil
}

func getByName(src, nameNode *Node) (res *Node, err error) {
	if nameNode.Type != StringNT {
		return &Node{Type: FailNT}, nil
	}

	obj := src.Scope.Vars
	val, ok := obj[nameNode.Val.(string)]
	if !ok {
		return &Node{Type: FailNT}, nil
	}

	return val, nil
}

func interpretListSlice(n *Node, env *Environment) (res *Node, err error) {
	src, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if src.Type != ListNT && src.Type != StringNT {
		return nil, fmt.Errorf("Value is not a list and cannot be sliced")
	}

	startNode := n.R.L
	endNode := n.R.R
	if startNode == nil && endNode == nil {
		return Interpret(src, env)
	}

	var start int64
	var end int64
	switch src.Type {
	case ListNT:
		end = int64(len(src.Val.([]*Node)))
	case StringNT:
		end = int64(len(src.Val.(string)))

	}
	if startNode != nil {
		startVal, err := Interpret(startNode, env)
		if err != nil {
			return nil, err
		}

		switch startVal.Type {
		case IntNT:
			start = startVal.Val.(int64)
		case FloatNT:
			start = int64(startVal.Val.(float64))
		default:
			return &Node{Type: FailNT}, nil
		}
	}

	if endNode != nil {
		endVal, err := Interpret(endNode, env)
		if err != nil {
			return nil, err
		}

		switch endVal.Type {
		case IntNT:
			end = endVal.Val.(int64)
		case FloatNT:
			end = int64(endVal.Val.(float64))
		default:
			return &Node{Type: FailNT}, nil
		}
	}

	if src.Type == StringNT {
		if end > int64(len(src.Val.(string))) {
			end = int64(len(src.Val.(string)))
		}
		if start > end {
			start = end
		}
		return &Node{
			Type: StringNT,
			Val:  src.Val.(string)[int(start):int(end)],
		}, nil
	}

	list := []*Node{}
	for i := int(start); i < int(end) && i < len(src.Val.([]*Node)); i++ {
		list = append(list, src.Val.([]*Node)[i])
	}

	return &Node{
		Type: ListNT,
		Val:  list,
	}, nil
}

func interpretWhile(stmt *Node, env *Environment) (res *Node, err error) {
	for {
		cond, err := Interpret(stmt.L, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			break
		}
		stop := false

		scope := &Environment{
			Parent: env,
			Consts: map[string]*Node{},
			Vars:   map[string]*Node{},
		}

		for n := stmt.R; n != nil; n = n.R {
			if n.Type == StmtNT {
				res, err = Interpret(n.L, scope)
			} else {
				res, err = Interpret(n, scope)
			}

			if err != nil {
				return nil, err
			}

			if res.Type == BreakNT {
				stop = true
				break
			}

			if res.Type == ContinueNT {
				break
			}

			if res.Type == ReturnStmtNT {
				stop = true
				return res.R, nil
			}
		}

		if stop {
			break
		}
	}

	return res, err
}

func interpretFor(stmt *Node, env *Environment) (res *Node, err error) {
	src, err := Interpret(stmt.L.R, env)
	if err != nil {
		return nil, err
	}
	if src.Type != ListNT && src.Type != ObjectNT && src.Type != SetNT {
		return &Node{Type: FailNT}, nil
	}

	iter := stmt.L.L.Val.(string)

	// for each iteration
	next := iterateCollection(src)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		scope := &Environment{
			Parent: env,
			Consts: map[string]*Node{},
			Vars:   map[string]*Node{},
		}

		scope.Consts[iter] = item
		scope.Consts["index"] = &Node{Type: IntNT, Val: int64(i)}
		stop := false

		// for each statement in body
		for n := stmt.R; n != nil; n = n.R {
			if n.Type == StmtNT {
				res, err = Interpret(n.L, scope)
			} else {
				res, err = Interpret(n, scope)
			}

			if err != nil {
				return res, err
			}

			break_, continue_ := false, false
			switch res.Type {
			case BreakNT:
				stop = true
				break_ = true
			case ContinueNT:
				continue_ = true
			case ReturnStmtNT:
				stop = true
				return res.R, nil
			}

			if break_ {
				break
			}
			if continue_ {
				continue
			}
		}

		if stop {
			break
		}
	}
	env.Consts["index"] = nil

	return res, err
}

func interpretRange(n *Node, env *Environment) (res *Node, err error) {
	var start *Node
	if n.L != nil {
		start, err = Interpret(n.L, env)
	}

	if err != nil {
		return nil, err
	}

	if start != nil && start.Type != IntNT {
		return nil, fmt.Errorf("Invalid start value for range")
	}

	end, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

	rng := []*Node{}
	var i int64
	if start != nil {
		switch start.Type {
		case IntNT:
			i = start.Val.(int64)
		case FloatNT:
			i = int64(start.Val.(float64))
		default:
			return &Node{Type: FailNT}, nil
		}
	}

	var endVal int64
	switch end.Type {
	case IntNT:
		endVal = end.Val.(int64)
	case FloatNT:
		endVal = int64(end.Val.(float64))
	default:
		return &Node{Type: FailNT}, nil
	}

	if i >= endVal {

		return &Node{
			Type: ListNT,
			Val:  []*Node{},
		}, nil
	}
	for ; i < endVal; i++ {
		rng = append(rng, &Node{
			Type: IntNT,
			Val:  i,
		})
	}

	return &Node{
		Type: ListNT,
		Val:  rng,
	}, nil
}

func interpretKVPair(n *Node, env *Environment) (res *Node, err error) {
	obj := map[string]*Node{}

	curr := n
	for curr != nil {
		val, err := Interpret(curr.L, env)
		if err != nil {
			return nil, err
		}

		key := curr.Val.(string)
		obj[key] = val
		curr = curr.R
	}

	return &Node{
		Type: ObjectNT,
		Scope: &Environment{
			Vars: obj,
		},
	}, nil
}

func interpretFieldAccess(n *Node, env *Environment) (res *Node, err error) {
	lhs := n.L
	rhs := n.R

	obj, err := Interpret(lhs, env)
	if err != nil {
		return nil, err
	}

	if obj.Type == ObjectNT {
		val, ok := obj.Scope.Vars[rhs.Val.(string)]
		if !ok {
			return &Node{Type: FailNT}, nil
		}

		return Interpret(val, env)
	}

	if obj.Type == ModuleNT {
		val, ok := obj.Scope.Consts[rhs.Val.(string)]
		if !ok {
			return &Node{Type: FailNT}, nil
		}

		return Interpret(val, env)
	}

	return &Node{Type: FailNT}, nil
}

func importModule(n *Node, env *Environment) (res *Node, err error) {
	top := env
	for top.Parent != nil {
		top = top.Parent
	}

	path := n.Val.(string)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to import from path \"%s\": %s", path, err.Error())
	}

	ts := Scan(string(file))
	modRoot, err := Parse(ts)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse module at path \"%s\": %s", path, err.Error())
	}

	modEnv := &Environment{
		Parent: &Environment{
			Consts: map[string]*Node{},
		},
		Consts: map[string]*Node{},
		Vars:   map[string]*Node{},
	}
	_, err = Interpret(modRoot, modEnv)
	if err != nil {
		return nil, fmt.Errorf("Encountered error while importing \"%s\": %s", path, err.Error())
	}

	var modName string
	if n.R != nil {
		modName = n.R.Val.(string)
	} else {
		modName = getModuleName(path)
	}

	module := &Node{
		Type: ModuleNT,
		Val:  modName,
		Scope: &Environment{
			Consts: modEnv.Consts,
		},
	}

	top.Consts[modName] = module

	return &Node{Type: SuccessNT}, nil
}

func getModuleName(path string) string {
	pathPieces := strings.Split(path, "/")
	if len(pathPieces) == 0 {
		return ""
	}
	filename := pathPieces[len(pathPieces)-1]
	filenamePieces := strings.Split(filename, ".")
	if len(filenamePieces) == 0 {
		return ""
	}
	return filenamePieces[0]
}

func (n *Node) toValue() Value {
	switch n.Type {
	case IntNT:
		return Value{
			DataType: IntDT,
			Val:      n.Val.(int64),
		}
	case FloatNT:
		return Value{
			DataType: FloatDT,
			Val:      n.Val.(float64),
		}
	case StringNT:
		return Value{
			DataType: StringDT,
			Val:      n.Val.(string),
		}
	case BoolNT:
		return Value{
			DataType: BoolDT,
			Val:      n.Val.(bool),
		}
	case SuccessNT:
		return Value{
			DataType: ResultDT,
			Val:      true,
		}
	case FailNT:
		return Value{
			DataType: ResultDT,
			Val:      false,
		}
	default:
		return Value{
			DataType: ResultDT,
			Val:      false,
		}
	}
}

func (v Value) toNode() *Node {
	switch v.DataType {
	case IntDT:
		return &Node{
			Type: IntNT,
			Val:  v.Val.(int64),
		}
	case FloatDT:
		return &Node{
			Type: FloatNT,
			Val:  v.Val.(float64),
		}
	case StringDT:
		return &Node{
			Type: StringNT,
			Val:  v.Val.(string),
		}
	case BoolDT:
		return &Node{
			Type: BoolNT,
			Val:  v.Val.(bool),
		}
	case ResultDT:
		if v.Val.(bool) {
			return &Node{Type: SuccessNT}
		}
		return &Node{Type: FailNT}
	default:
		return &Node{Type: FailNT}
	}
}

func interpretSetItem(n *Node, env *Environment) (res *Node, err error) {
	set := map[Value]bool{}

	curr := n
	for curr != nil {
		val, err := Interpret(curr.L, env)
		if err != nil {
			return nil, err
		}

		set[val.toValue()] = true
		curr = curr.R
	}

	return &Node{
		Type: SetNT,
		Val:  set,
	}, nil
}
