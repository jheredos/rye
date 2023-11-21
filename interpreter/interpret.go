package interpreter

import (
	"fmt"
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
		return copyNode(n), nil
	case LambdaNT:
		return copyNode(n), nil
	case ObjectNT:
		if n.Val == nil {
			return newObject(Object{}), nil
		}
		return n, nil
	case ModuleNT:
		return n, nil
	case ListNT:
		return interpretList(n, env)
	case ObjectItemNT:
		return interpretObjectItem(n, env)
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
	case FindNT:
		return interpretFind(n, env)
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

func interpretStmt(root *Node, env *Environment) (res *Node, err error) {
	for n := root; n != nil; n = n.R {
		if n.L != nil && n.L.Type == StmtNT {
			res, err = Interpret(n.L, newScope(env))
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
				return newInt(l.Val.(int64) + r.Val.(int64)), nil
			case FloatNT:
				return newFloat(l.Val.(float64) + r.Val.(float64)), nil
			case StringNT:
				return newString(l.Val.(string) + r.Val.(string)), nil
			case ListNT:
				return newList(append(l.Val.(List), r.Val.(List)...)), nil
			default:
				return FAIL, nil
			}
		}
	case SubtNT:
		{
			switch t {
			case IntNT:
				return newInt(l.Val.(int64) - r.Val.(int64)), nil
			case FloatNT:
				return newFloat(l.Val.(float64) - r.Val.(float64)), nil
			default:
				return FAIL, nil
			}
		}
	case DivNT:
		{
			switch t {
			case IntNT:
				if r.Val.(int64) == 0 {
					return FAIL, nil
				}
				return newFloat(float64(l.Val.(int64)) / float64(r.Val.(int64))), nil
			case FloatNT:
				if r.Val.(float64) == 0 {
					return FAIL, nil
				}
				return newFloat(l.Val.(float64) / r.Val.(float64)), nil
			default:
				return FAIL, nil
			}
		}
	case MultNT:
		{
			switch t {
			case IntNT:
				return newInt(l.Val.(int64) * r.Val.(int64)), nil
			case FloatNT:
				return newFloat(l.Val.(float64) * r.Val.(float64)), nil
			default:
				return FAIL, nil
			}
		}
	case ModuloNT:
		{
			switch t {
			case IntNT:
				if r.Val.(int64) == 0 {
					return FAIL, nil
				}
				return newInt(l.Val.(int64) % r.Val.(int64)), nil
			default:
				return FAIL, nil
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
		return FAIL, nil
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

		return newFloat(total), nil
	} else if lhs.Type == IntNT {
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
			return newFloat(1 / float64(total)), nil
		}
		return newInt(total), nil
	}

	return FAIL, nil
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
		return FALSE, nil
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

	// ==, !=
	switch n.Type {
	case EqualNT:
		equal, err := evalEquality(lhs, rhs)
		if err != nil {
			return FAIL, nil
		}
		return newBool(equal), nil
	case NotEqualNT:
		equal, err := evalEquality(lhs, rhs)
		if err != nil {
			return FAIL, nil
		}
		return newBool(!equal), nil
	}

	// <, >, <=, >=
	l, r, t := maybeCastNumbers(lhs, rhs)
	switch n.Type {
	case LessEqualNT:
		switch t {
		case IntNT:
			return newBool(l.Val.(int64) <= r.Val.(int64)), nil
		case FloatNT:
			return newBool(l.Val.(float64) <= r.Val.(float64)), nil
		}
	case GreaterEqualNT:
		switch t {
		case IntNT:
			return newBool(l.Val.(int64) >= r.Val.(int64)), nil
		case FloatNT:
			return newBool(l.Val.(float64) >= r.Val.(float64)), nil
		}
	case LessNT:
		switch t {
		case IntNT:
			return newBool(l.Val.(int64) < r.Val.(int64)), nil
		case FloatNT:
			return newBool(l.Val.(float64) < r.Val.(float64)), nil
		}
	case GreaterNT:
		switch t {
		case IntNT:
			return newBool(l.Val.(int64) > r.Val.(int64)), nil
		case FloatNT:
			return newBool(l.Val.(float64) > r.Val.(float64)), nil
		}
	}

	return FAIL, nil
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
		for i := 0; i < len(container.Val.(List)); i++ {
			equal, _ := evalEquality(item, container.Val.(List)[i])
			if equal {
				return TRUE, nil
			}
		}

		return FALSE, nil
	case SetNT:
		set := container.Val.(Set)
		return newBool(set[item.toValue()]), nil

	default:
		return FAIL, nil
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
		return newBool(!isTruthy(arg)), nil
	case MaybeNT:
		if arg.Type == FailNT {
			return arg, nil
		}
		return SUCCESS, nil
	case CardinalityNT:
		{
			var cardinality int
			switch arg.Type {
			case ListNT:
				cardinality = len(arg.Val.(List))
			case StringNT:
				cardinality = len(arg.Val.(string))
			case SetNT:
				cardinality = len(arg.Val.(Set))
			case ObjectNT:
				cardinality = len(arg.Val.(Object))
			default:
				return FAIL, nil
			}
			return newInt(int64(cardinality)), nil
		}
	case UnaryNegNT:
		{
			switch arg.Type {
			case IntNT:
				return newInt(-arg.Val.(int64)), nil
			case FloatNT:
				return newFloat(-arg.Val.(float64)), nil
			default:
				return FAIL, nil
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
		if result.Type == ThenBranchNT {
			// expr has an else branch
			return Interpret(result.L, env)
		} else {
			// expr does not have an else branch
			return Interpret(result, env)
		}
	} else {
		if result.Type == ThenBranchNT {
			return Interpret(result.R, env)
		} else {
			return FAIL, nil
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

	parent := env
	if lambda.Scope != nil {
		parent = lambda.Scope
	}

	scope := newScope(parent)

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
			res, err = Interpret(n.L, newScope(env))
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

func interpretMap(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if lhs.Type == FailNT || (lhs.Type != ListNT && lhs.Type != SetNT) {
		return FAIL, nil
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
		return FAIL, nil
	}

	resList := List{}
	resSet := Set{}

	next := iterateCollection(lhs)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		old, err := Interpret(item, env)
		if err != nil {
			return nil, err
		}

		env.Consts["index"] = newInt(int64(i))

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
			resList = append(resList, new)
		}
		if lhs.Type == SetNT {
			resSet[new.toValue()] = true
		}
	}
	env.Consts["_"] = nil
	env.Consts["index"] = nil

	if lhs.Type == SetNT {
		return newSet(resSet), nil
	}

	return newList(resList), nil
}

func interpretWhere(n *Node, env *Environment) (res *Node, err error) {
	lhs, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if lhs.Type == FailNT || (lhs.Type != ListNT && lhs.Type != SetNT) {
		return FAIL, nil
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
		return FAIL, nil
	}

	resList := List{}
	resSet := Set{}

	next := iterateCollection(lhs)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		val, err := Interpret(item, env)
		if err != nil {
			return nil, err
		}

		env.Consts["index"] = newInt(int64(i))
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
				resList = append(resList, val)
			}
			if lhs.Type == SetNT {
				resSet[val.toValue()] = true
			}
		}
	}
	env.Consts["_"] = nil
	env.Consts["index"] = nil

	if lhs.Type == SetNT {
		return newSet(resSet), nil
	}

	return newList(resList), nil
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

func interpretFind(n *Node, env *Environment) (res *Node, err error) {
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

	next := iterateCollection(lhs)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		val, err := Interpret(item, env)
		if err != nil {
			return nil, err
		}

		env.Consts["index"] = newInt(int64(i))
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
			env.Consts["_"] = nil
			env.Consts["index"] = nil
			return item, nil
		}
	}

	env.Consts["_"] = nil
	env.Consts["index"] = nil

	return FAIL, nil
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

	return FAIL, nil
}

func interpretListSlice(n *Node, env *Environment) (res *Node, err error) {
	src, err := Interpret(n.L, env)
	if err != nil {
		return nil, err
	}

	if src.Type != ListNT && src.Type != StringNT {
		return FAIL, nil
		// return nil, fmt.Errorf("Value is not a list and cannot be sliced")
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
		end = int64(len(src.Val.(List)))
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
			return FAIL, nil
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
			return FAIL, nil
		}
	}

	if src.Type == StringNT {
		if end > int64(len(src.Val.(string))) {
			end = int64(len(src.Val.(string)))
		}
		if start > end {
			start = end
		}
		return newString(src.Val.(string)[int(start):int(end)]), nil
	}

	list := List{}
	for i := int(start); i < int(end) && i < len(src.Val.(List)); i++ {
		list = append(list, src.Val.(List)[i])
	}

	return newList(list), nil
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

		scope := newScope(env)

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
	iterator, iteratee := stmt.L.L, stmt.L.R
	src, err := Interpret(iteratee, env)
	if err != nil {
		return nil, err
	}
	if src.Type != ListNT && src.Type != ObjectNT && src.Type != SetNT {
		return FAIL, nil
	}

	iter := iterator.Val.(string)

	// for each iteration
	next := iterateCollection(src)
	for item, i := next(), 0; item != nil; item, i = next(), i+1 {
		scope := newScope(env)

		scope.Consts[iter] = item
		scope.Consts["index"] = newInt(int64(i))
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

	rng := List{}
	var i int64
	if start != nil {
		switch start.Type {
		case IntNT:
			i = start.Val.(int64)
		case FloatNT:
			i = int64(start.Val.(float64))
		default:
			return FAIL, nil
		}
	}

	var endVal int64
	switch end.Type {
	case IntNT:
		endVal = end.Val.(int64)
	case FloatNT:
		endVal = int64(end.Val.(float64))
	default:
		return FAIL, nil
	}

	if i >= endVal {

		return newList(List{}), nil
	}
	for ; i < endVal; i++ {
		rng = append(rng, newInt(i))
	}

	return newList(rng), nil
}

func interpretList(n *Node, env *Environment) (res *Node, err error) {
	list := List{}

	for _, m := range n.Val.(List) {
		switch m.Type {
		case SplatNT, RangeNT, MapNT, WhereNT:
			var arg *Node
			var err error
			if m.Type == SplatNT {
				arg, err = Interpret(m.R, env)
			} else {
				arg, err = Interpret(m, env)
			}
			if err != nil {
				return nil, err
			}

			switch arg.Type {
			case ListNT:
				list = append(list, arg.Val.(List)...)
			case SetNT:
				set := arg.Val.(Set)
				for k := range set {
					if set[k] {
						list = append(list, k.toNode())
					}
				}
			default:
				list = append(list, FAIL)
			}
			continue
		default:
			val, err := Interpret(m, env)
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		}
	}

	return newList(list), nil
}

func interpretObjectItem(n *Node, env *Environment) (res *Node, err error) {
	obj := Object{}

	curr := n
	for curr != nil {
		node := curr.L

		switch node.Type {
		case KVPairNT:
			key := node.L
			if node.L.Type != IdentifierNT {
				var err error
				key, err = Interpret(node.L, env)
				if err != nil {
					return nil, err
				}
			}

			val, err := Interpret(node.R, env)
			if err != nil {
				return nil, err
			}

			obj[key.toValue()] = val
		case SplatNT:
			arg, err := Interpret(node.R, env)
			if err != nil {
				return nil, err
			}

			if arg.Type == ObjectNT {
				for k, v := range arg.Val.(Object) {
					obj[k] = v
				}
			}
		}

		curr = curr.R
	}

	return newObject(obj), nil
}

func interpretFieldAccess(n *Node, env *Environment) (res *Node, err error) {
	lhs, rhs := n.L, n.R

	obj, err := Interpret(lhs, env)
	if err != nil {
		return nil, err
	}

	if obj.Type == ObjectNT {
		val, ok := obj.Val.(Object)[rhs.toValue()]
		if !ok {
			return FAIL, nil
		}

		return Interpret(val, env)
	}

	if obj.Type == ModuleNT {
		val, ok := obj.Scope.Consts[rhs.Val.(string)]
		if !ok {
			return FAIL, nil
		}

		return Interpret(val, env)
	}

	return FAIL, nil
}

func interpretSetItem(n *Node, env *Environment) (res *Node, err error) {
	set := Set{}

	curr := n
	for curr != nil {
		// handle spread
		switch curr.L.Type {
		case SplatNT, RangeNT, MapNT, WhereNT:
			var arg *Node
			var err error
			if curr.L.Type == SplatNT {
				arg, err = Interpret(curr.L.R, env)
			} else {
				arg, err = Interpret(curr.L, env)
			}
			if err != nil {
				return nil, err
			}

			switch arg.Type {
			case ListNT:
				for _, m := range arg.Val.(List) {
					set[m.toValue()] = true
				}
			case SetNT:
				s := arg.Val.(Set)
				for m := range s {
					if s[m] {
						set[m] = true
					}
				}
			default:
				set[(FAIL).toValue()] = true
			}

			curr = curr.R
			continue
		}

		// all other set items
		val, err := Interpret(curr.L, env)
		if err != nil {
			return nil, err
		}

		set[val.toValue()] = true
		curr = curr.R
	}

	return newSet(set), nil
}
