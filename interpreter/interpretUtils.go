package interpreter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var FAIL = &Node{Type: FailNT}
var SUCCESS = &Node{Type: SuccessNT}
var TRUE = &Node{Type: BoolNT, Val: true}
var FALSE = &Node{Type: BoolNT, Val: false}

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
		if len(l.Val.(List)) != len(r.Val.(List)) {
			return false, nil
		}
		for i, n := range a.Val.(List) {
			equal, err := evalEquality(n, r.Val.(List)[i])
			if !equal || err != nil {
				return false, err
			}
		}
		return true, nil
	case BoolNT:
		return l.Val.(bool) == r.Val.(bool), nil
	case SuccessNT, FailNT, NullNT:
		return true, nil
	default:
		return false, fmt.Errorf("Cannot compare types")
	}
}

// maybeCastNumbers casts numbers to floats if one is a float
func maybeCastNumbers(a, b *Node) (*Node, *Node, NodeType) {
	if a == nil || b == nil {
		return nil, nil, ErrorNT
	}

	if a.Type == b.Type {
		return a, b, a.Type
	}

	switch a.Type {
	case IntNT:
		if b.Type == FloatNT {
			return newFloat(float64(a.Val.(int64))), b, FloatNT
		} else {
			return a, b, ErrorNT
		}
	case FloatNT:
		if b.Type == IntNT {
			return a, newFloat(float64(b.Val.(int64))), FloatNT
		} else {
			return a, b, ErrorNT
		}
	case StringNT:
		switch b.Type {
		case IntNT, FloatNT:
			return a, newString(b.ToString()), StringNT
		default:
			return a, b, ErrorNT
		}
	case ListNT, BoolNT, SuccessNT, FailNT, NullNT:
		return a, b, ErrorNT
	default:
		return a, b, ErrorNT
	}
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

	if n.Line != 0 {
		return nil, fmt.Errorf("Line %d: \"%s\" is undefined", n.Line, ident)
	}
	return nil, fmt.Errorf("\"%s\" is undefined", ident)
}

func declareVar(n *Node, env *Environment) (res *Node, err error) {
	val, err := Interpret(n.R, env)
	if err != nil {
		return nil, err
	}

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

	return SUCCESS, nil
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

	return SUCCESS, nil
}

// getAssignmentTarget - returns a function that will assign its argument to the desired place.
// lhs refers to a potentially nested assignment target, like a list index, object field, or some combination
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

	return getNestedAssign(lhs, env)
}

func getDestructuredAssign(assignee *Node, env *Environment) (assignFunc func(*Node) error, err error) {
	switch assignee.Type {
	case ListNT:
		return func(n *Node) error {
			if n.Type != ListNT {
				for _, m := range assignee.Val.(List) {
					env.Consts[m.Val.(string)] = FAIL
				}
				return nil
			}

			for i, m := range assignee.Val.(List) {
				if i < len(n.Val.(List)) {
					env.Consts[m.Val.(string)] = n.Val.(List)[i]
				} else {
					env.Consts[m.Val.(string)] = FAIL
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
			length := len(container.Val.(List))
			if err != nil {
				return nil, err
			}
			var idx int
			if idxNode.Type == IntNT {
				idx = int(idxNode.Val.(int64))
			} else if idxNode.Type == FloatNT {
				idx = int(idxNode.Val.(float64))
			} else {
				return nil, fmt.Errorf("Cannot assign to list index. Invalid index.")
			}

			if idx < 0 {
				idx += length
			}
			if idx >= length || idx < 0 {
				return nil, fmt.Errorf("Cannot assign to list. Index out of range.")
			}
			return func(n *Node) error {
				container.Val.(List)[idx] = n
				return nil
			}, nil
		}
	case ObjectNT:
		{
			// field access
			if assignee.Type == FieldAccessNT {
				return func(n *Node) error {
					container.Val.(Object)[assignee.R.toValue()] = n
					return nil
				}, nil
			}

			// bracket access
			key, err := Interpret(assignee.R, env)
			if err != nil {
				return nil, err
			}

			return func(n *Node) error {
				container.Val.(Object)[key.toValue()] = n
				return nil
			}, nil
		}
	default:
		return nil, fmt.Errorf("Invalid assignment target.")
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

	// parameter: the identifier defined in a function definition
	// argument: the value assigned to a parameter at call time

	// identifier: 	param{Val: string, R: nextParam}
	// list:				param{L: List of idents, R: nextParam}
	// object:			param{L: ObjectItem{L: KVPair{L: key, R: rename} | ident, R: nil | nextField}, R: nextParam}

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
				for _, p := range param.L.Val.(List) {
					scope.Vars[p.Val.(string)] = FAIL
				}
				return
			}
			as := arg.Val.(List)
			for i, p := range param.L.Val.(List) {
				if i >= len(as) {
					scope.Vars[p.Val.(string)] = FAIL
				}
				scope.Vars[p.Val.(string)] = as[i]
			}
			return
		}
	case ObjectItemNT:
		{
			if arg.Type != ObjectNT {
				// The arg is not an object
				for p := param.L; p != nil; p = p.R {
					if p.L != nil {
						assignArg(FAIL, p.L, scope)
					} else {
						scope.Vars[p.Val.(string)] = FAIL
					}
				}
				return
			}

			obj := arg.Val.(Object)
			for p := param.L; p != nil; p = p.R {
				destrucItem := p.L

				var originalName *Node
				var newName string
				if destrucItem.Type == KVPairNT {
					// key-value pair, rename
					originalName = destrucItem.L
					newName = destrucItem.R.Val.(string)
				} else {
					// plain identifier, do not rename
					originalName = destrucItem
					newName = originalName.Val.(string)
				}

				val, ok := obj[originalName.toValue()]
				if ok {
					// The field exists. Add to scope
					scope.Vars[newName] = val
				} else {
					// The field does not exist on the arg
					scope.Vars[newName] = FAIL
				}
			}
		}

	}
}

func iterateCollection(n *Node) func() *Node {
	switch n.Type {
	case ListNT:
		list := n.Val.(List)
		i := -1
		return func() *Node {
			if i < len(list)-1 {
				i++
				return list[i]
			}
			return nil
		}
	case ObjectNT:
		obj := n.Val.(Object)
		keys := []*Node{}
		for k := range obj {
			keys = append(keys, k.toNode())
		}
		i := -1
		return func() *Node {
			if i < len(keys)-1 {
				i++
				return keys[i]
			}
			return nil
		}
	case SetNT:
		set := n.Val.(Set)
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

func getByIndex(src, idxNode *Node) (res *Node, err error) {
	var idx int64
	switch idxNode.Type {
	case IntNT:
		idx = idxNode.Val.(int64)
	case FloatNT:
		idx = int64(idxNode.Val.(float64))
	default:
		return FAIL, nil
	}

	var length int64
	switch src.Type {
	case ListNT:
		length = int64(len(src.Val.(List)))
	case StringNT:
		length = int64(len(src.Val.(string)))

	}

	if idx >= length {
		return FAIL, nil
	}

	if idx < 0 && -idx > length {
		return FAIL, nil
	}

	if idx < 0 {
		return src.Val.(List)[length+idx], nil
	}

	if src.Type == StringNT {
		return newString(string(src.Val.(string)[idx])), nil
	}
	return src.Val.(List)[idx], nil
}

func getByName(src, nameNode *Node) (res *Node, err error) {
	obj := src.Val.(Object)

	val, ok := obj[nameNode.toValue()]
	if !ok {
		return FAIL, nil
	}

	return val, nil
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

func importModule(n *Node, env *Environment) (res *Node, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	top := env
	for top.Parent != nil {
		top = top.Parent
	}

	pathVal := n.Val.(string)
	pathElems := strings.Split(pathVal, "/")

	path := pwd + "/"
	for _, elem := range pathElems[:len(pathElems)-1] {
		if elem == "." {
			continue
		}
		path += elem + "/"
	}
	path += pathElems[len(pathElems)-1]

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to import from path \"%s\": %s", path, err.Error())
	}

	ts := Scan(string(file))
	modRoot, err := Parse(ts)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse module at path \"%s\": %s", path, err.Error())
	}

	modEnv := newScope(&Environment{Consts: map[string]*Node{}})

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

	return SUCCESS, nil
}

func newScope(parent *Environment) *Environment {
	return &Environment{
		Parent: parent,
		Consts: map[string]*Node{},
		Vars:   map[string]*Node{},
	}
}

func copyNode(n *Node) *Node {
	return &Node{
		Type:  n.Type,
		Val:   n.Val,
		Func:  n.Func,
		Scope: n.Scope,
		L:     n.L,
		R:     n.R,
	}
}

func newInt(val int64) *Node {
	return &Node{
		Type: IntNT,
		Val:  val,
	}
}

func newFloat(val float64) *Node {
	return &Node{
		Type: FloatNT,
		Val:  val,
	}
}

func newBool(val bool) *Node {
	return &Node{
		Type: BoolNT,
		Val:  val,
	}
}

func newString(val string) *Node {
	return &Node{
		Type: StringNT,
		Val:  val,
	}
}

func newSet(val Set) *Node {
	return &Node{
		Type: SetNT,
		Val:  val,
	}
}

func newObject(val Object) *Node {
	return &Node{
		Type: ObjectNT,
		Val:  val,
	}
}

func newList(val List) *Node {
	return &Node{
		Type: ListNT,
		Val:  val,
	}
}
