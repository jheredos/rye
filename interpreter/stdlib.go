package interpreter

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var randSrc = rand.New(rand.NewSource(time.Now().UnixNano()))

var StdLib map[string]*Node = map[string]*Node{
	// I/O utils
	"print": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			strs := []any{}
			for _, arg := range args {
				if arg.Type == StringNT {
					strs = append(strs, fmt.Sprintf("%s", arg.Val))
				} else {
					strs = append(strs, arg.ToString())
				}
			}
			fmt.Println(strs...)
			return &Node{Type: SuccessNT}, nil
		},
	},
	"readInput": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"readInput\". Expected 1, received %d.", len(args))
			}

			prompt := args[0]
			if prompt.Type != StringNT {
				return &Node{Type: FailNT}, nil
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print(prompt.Val.(string))
			inp, err := reader.ReadString('\n')
			if err != nil {
				return &Node{Type: FailNT}, nil
			}

			return &Node{
				Type: StringNT,
				Val:  inp,
			}, nil
		},
	},
	"readFile": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"input\". Expected 1, received %d.", len(args))
			}

			path := args[0]
			if path.Type != StringNT {
				fmt.Println("path is not a string: fail")
				return &Node{Type: FailNT}, nil
			}

			file, err := ioutil.ReadFile(path.Val.(string))
			if err != nil {
				return &Node{Type: FailNT}, nil
			}

			return &Node{
				Type: StringNT,
				Val:  string(file),
			}, nil
		},
	},
	// "readJson": {
	// 	Type: LambdaNT,
	// 	Func: func(_ *Environment, args ...*Node) (*Node, error) {
	// 		if len(args) < 1 {
	// 			return nil, fmt.Errorf("Wrong number of arguments for \"sum\". Expected 1+, received %d.", len(args))
	// 		}

	// 	},
	// },
	// math utils
	"sum": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"sum\". Expected 1+, received %d.", len(args))
			}

			if args[0].Type == ListNT {
				args = args[0].Val.(List)
			}

			allInts := true
			for _, n := range args {
				if n.Type != IntNT {
					allInts = false
					break
				}
			}

			if allInts {
				var total int64
				for _, n := range args {
					val, err := castInt(n)
					if err != nil {
						return &Node{Type: FailNT}, nil
					}
					total += val
				}
				return &Node{Type: IntNT, Val: total}, nil
			}

			var total float64
			for _, n := range args {
				val, err := castFloat(n)
				if err != nil {
					return &Node{Type: FailNT}, nil
				}
				total += val
			}

			return &Node{Type: FloatNT, Val: total}, nil
		},
	},
	"max": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"max\". Expected 1+, received %d.", len(args))
			}

			if len(args) == 1 {
				if args[0].Type == ListNT {
					args = args[0].Val.(List)
				} else {
					return &Node{Type: FailNT}, nil
				}
			}

			allInts := true
			var floatMax float64
			var intMax int64
			switch args[0].Type {
			case FloatNT:
				allInts = false
				floatMax = args[0].Val.(float64)
				intMax = int64(floatMax)
			case IntNT:
				intMax = args[0].Val.(int64)
				floatMax = float64(intMax)
			default:
				return &Node{Type: FailNT}, nil
			}

			for _, n := range args[1:] {
				switch n.Type {
				case FloatNT:
					if n.Val.(float64) > floatMax {
						floatMax = n.Val.(float64)
						intMax = int64(floatMax)
					}
					allInts = false
				case IntNT:
					if n.Val.(int64) > intMax {
						intMax = n.Val.(int64)
						floatMax = float64(intMax)
					}
				default:
					return &Node{Type: FailNT}, nil
				}
			}

			if allInts {
				return &Node{Type: FloatNT, Val: floatMax}, nil
			}

			return &Node{Type: IntNT, Val: intMax}, nil
		},
	},
	"min": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"min\". Expected 1+, received %d.", len(args))
			}

			if len(args) == 1 {
				if args[0].Type == ListNT {
					args = args[0].Val.(List)
				} else {
					return &Node{Type: FailNT}, nil
				}
			}

			allInts := true
			var floatMax float64
			var intMax int64
			switch args[0].Type {
			case FloatNT:
				allInts = false
				floatMax = args[0].Val.(float64)
				intMax = int64(floatMax)
			case IntNT:
				intMax = args[0].Val.(int64)
				floatMax = float64(intMax)
			default:
				return &Node{Type: FailNT}, nil
			}

			for _, n := range args[1:] {
				switch n.Type {
				case FloatNT:
					if n.Val.(float64) < floatMax {
						floatMax = n.Val.(float64)
						intMax = int64(floatMax)
					}
					allInts = false
				case IntNT:
					if n.Val.(int64) < intMax {
						intMax = n.Val.(int64)
						floatMax = float64(intMax)
					}
				default:
					return &Node{Type: FailNT}, nil
				}
			}

			if allInts {
				return &Node{Type: FloatNT, Val: floatMax}, nil
			}

			return &Node{Type: IntNT, Val: intMax}, nil
		},
	},
	"random": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("Wrong number of arguments for \"random\". Expected 0, received %d.", len(args))
			}

			return &Node{
				Type: FloatNT,
				Val:  randSrc.Float64(),
			}, nil
		},
	},
	// string utils
	"split": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"split\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != StringNT || args[1].Type != StringNT {
				return &Node{Type: FailNT}, nil
			}

			strs := strings.Split(args[0].Val.(string), args[1].Val.(string))
			ns := List{}
			for _, s := range strs {
				ns = append(ns, &Node{
					Type: StringNT,
					Val:  s,
				})
			}

			return &Node{
				Type: ListNT,
				Val:  ns,
			}, nil
		},
	},
	"join": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"join\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != ListNT || args[1].Type != StringNT {
				return &Node{Type: FailNT}, nil
			}

			strs := []string{}
			for _, n := range args[0].Val.(List) {
				if n.Type != StringNT {
					return &Node{Type: FailNT}, nil
				}
				strs = append(strs, n.Val.(string))
			}

			return &Node{
				Type: StringNT,
				Val:  strings.Join(strs, args[1].Val.(string)),
			}, nil
		},
	},
	"uppercase": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"uppercase\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != StringNT {
				return &Node{Type: FailNT}, nil
			}

			return &Node{
				Type: StringNT,
				Val:  strings.ToUpper(args[0].Val.(string)),
			}, nil
		},
	},
	"lowercase": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"lowercase\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != StringNT {
				return &Node{Type: FailNT}, nil
			}

			return &Node{
				Type: StringNT,
				Val:  strings.ToLower(args[0].Val.(string)),
			}, nil
		},
	},
	// type casts and utils
	"typeof": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of values for \"typeof\". Expected 1, received %d.", len(args))
			}

			switch args[0].Type {
			case LambdaNT:
				return &Node{
					Type: StringNT,
					Val:  "Lambda",
				}, nil
			case ListNT:
				return &Node{
					Type: StringNT,
					Val:  "List",
				}, nil
			case SetNT:
				return &Node{
					Type: StringNT,
					Val:  "Set",
				}, nil
			case ObjectNT:
				return &Node{
					Type: StringNT,
					Val:  "Object",
				}, nil
			case SuccessNT, FailNT:
				return &Node{
					Type: StringNT,
					Val:  "Result",
				}, nil
			case FloatNT:
				return &Node{
					Type: StringNT,
					Val:  "Float",
				}, nil
			case IntNT:
				return &Node{
					Type: StringNT,
					Val:  "Int",
				}, nil
			case BoolNT:
				return &Node{
					Type: StringNT,
					Val:  "Bool",
				}, nil
			case StringNT:
				return &Node{
					Type: StringNT,
					Val:  "String",
				}, nil
			case NullNT:
				return &Node{
					Type: StringNT,
					Val:  "Null",
				}, nil
			case ModuleNT:
				return &Node{
					Type: StringNT,
					Val:  "Module",
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		},
	},
	"Int": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"Int\". Expected 1, received %d.", len(args))
			}

			switch args[0].Type {
			case IntNT:
				return args[0], nil
			case FloatNT:
				return &Node{
					Type: IntNT,
					Val:  int64(args[0].Val.(float64)),
				}, nil
			case StringNT:
				val, err := strconv.ParseInt(args[0].Val.(string), 10, 64)
				if err != nil {
					return &Node{Type: FailNT}, nil
				}
				return &Node{
					Type: IntNT,
					Val:  val,
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		},
	},
	"Float": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"Float\". Expected 1, received %d.", len(args))
			}

			switch args[0].Type {
			case IntNT:
				return &Node{
					Type: FloatNT,
					Val:  float64(args[0].Val.(int64)),
				}, nil
			case FloatNT:
				return args[0], nil
			case StringNT:
				val, err := strconv.ParseFloat(args[0].Val.(string), 64)
				if err != nil {
					return &Node{Type: FailNT}, nil
				}
				return &Node{
					Type: FloatNT,
					Val:  val,
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		},
	},
	"String": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"String\". Expected 1, received %d.", len(args))
			}

			switch args[0].Type {
			case StringNT:
				return args[0], nil
			case LambdaNT:
				return &Node{
					Type: StringNT,
					Val:  "<lambda>",
				}, nil
			default:
				return &Node{
					Type: StringNT,
					Val:  args[0].ToString(),
				}, nil
			}
		},
	},
	"Set": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"Set\". Expected 1, received %d.", len(args))
			}

			set := Set{}

			switch args[0].Type {
			case SetNT:
				return args[0], nil
			case ListNT:
				for _, n := range args[0].Val.(List) {
					set[n.toValue()] = true
				}
				return &Node{
					Type: SetNT,
					Val:  set,
				}, nil
			case IntNT, FloatNT, StringNT, BoolNT, SuccessNT, FailNT:
				set[args[0].toValue()] = true
				return &Node{
					Type: SetNT,
					Val:  set,
				}, nil
			default:
				return &Node{Type: FailNT}, nil
			}
		},
	},
	"List": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"List\". Expected 1+, received %d.", len(args))
			}

			list := List{}
			if len(args) > 1 {
				list = append(list, args...)
				return &Node{
					Type: ListNT,
					Val:  list,
				}, nil
			}

			switch args[0].Type {
			case ListNT:
				return args[0], nil
			case SetNT:
				{
					set := args[0].Val.(Set)
					for v := range set {
						if !set[v] {
							continue
						}
						list = append(list, v.toNode())
					}
					return &Node{
						Type: ListNT,
						Val:  list,
					}, nil
				}
			default:
				list = append(list, args[0])
				return &Node{
					Type: ListNT,
					Val:  list,
				}, nil
			}
		},
	},
	// set utils
	"union": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"union\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != SetNT || args[1].Type != SetNT {
				return &Node{Type: FailNT}, nil
			}

			union := Set{}
			a, b := args[0].Val.(Set), args[1].Val.(Set)
			for n := range a {
				union[n] = true
			}

			for n := range b {
				union[n] = true
			}

			return &Node{Type: SetNT, Val: union}, nil
		},
	},
	"intersection": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"intersection\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != SetNT || args[1].Type != SetNT {
				return &Node{Type: FailNT}, nil
			}

			intersection := Set{}
			a, b := args[0].Val.(Set), args[1].Val.(Set)
			for n := range a {
				if b[n] {
					intersection[n] = true
				}
			}

			return &Node{Type: SetNT, Val: intersection}, nil
		},
	},
	"difference": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"difference\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != SetNT || args[1].Type != SetNT {
				return &Node{Type: FailNT}, nil
			}

			difference := Set{}
			a, b := args[0].Val.(Set), args[1].Val.(Set)
			for n := range a {
				if !b[n] {
					difference[n] = true
				}
			}

			return &Node{Type: SetNT, Val: difference}, nil
		},
	},
	"add": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"add\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != SetNT {
				return &Node{Type: FailNT}, nil
			}

			set := args[0].Val.(Set)
			set[args[1].toValue()] = true

			return &Node{
				Type: SetNT,
				Val:  set,
			}, nil
		},
	},
	"remove": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"remove\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != SetNT {
				return &Node{Type: FailNT}, nil
			}

			set := args[0].Val.(Set)
			set[args[1].toValue()] = false

			return &Node{
				Type: SetNT,
				Val:  set,
			}, nil
		},
	},
	// object utils
	"keys": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"keys\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != ObjectNT {
				return &Node{Type: FailNT}, nil
			}

			keys := List{}
			for k := range args[0].Val.(Object) {
				keys = append(keys, &Node{
					Type: StringNT,
					Val:  k,
				})
			}

			return &Node{Type: ListNT, Val: keys}, nil
		},
	},
	"values": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of values for \"values\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != ObjectNT {
				return &Node{Type: FailNT}, nil
			}

			vals := List{}
			for _, v := range args[0].Val.(Object) {
				vals = append(vals, v)
			}

			return &Node{Type: ListNT, Val: vals}, nil
		},
	},
	// list utils
	"flat": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"flat\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != ListNT {
				return &Node{Type: FailNT}, nil
			}

			flattened := List{}
			for _, n := range args[0].Val.(List) {
				if n.Type == ListNT {
					flattened = append(flattened, n.Val.(List)...)
				} else {
					flattened = append(flattened, n)
				}
			}

			return &Node{
				Type: ListNT,
				Val:  flattened,
			}, nil
		},
	},
	"find": {
		Type: LambdaNT,
		Func: func(env *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"find\". Expected 2, received %d.", len(args))
			}

			list := args[0]
			if list.Type != ListNT {
				return &Node{Type: FailNT}, nil
			}

			predicate := args[1]
			if predicate.Type != LambdaNT {
				return &Node{Type: FailNT}, nil
			}

			for _, n := range list.Val.(List) {
				call := &Node{
					Type: CallNT,
					L:    predicate,
					R: &Node{
						Type: ArgNT,
						L:    n,
					},
				}
				val, err := Interpret(call, env)
				if err != nil {
					return nil, err
				}

				if isTruthy(val) {
					return n, nil
				}
			}

			return &Node{Type: FailNT}, nil
		},
	},
	"findIndex": {
		Type: LambdaNT,
		Func: func(env *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"findIndex\". Expected 2, received %d.", len(args))
			}

			list := args[0]
			if list.Type != ListNT {
				return &Node{Type: FailNT}, nil
			}

			predicate := args[1]
			if predicate.Type != LambdaNT {
				return &Node{Type: FailNT}, nil
			}

			for i, n := range list.Val.(List) {
				call := &Node{
					Type: CallNT,
					L:    predicate,
					R: &Node{
						Type: ArgNT,
						L:    n,
					},
				}
				val, err := Interpret(call, env)
				if err != nil {
					return nil, err
				}

				if isTruthy(val) {
					return &Node{
						Type: IntNT,
						Val:  int64(i),
					}, nil
				}
			}

			return &Node{Type: FailNT}, nil
		},
	},
	// "fold": {
	// 	Type: LambdaNT,
	// 	Func: func(env *Environment, args ...*Node) (*Node, error) {
	// 		// list, startingVal, func
	// 		if len(args) != 3 {
	// 			return nil, fmt.Errorf("Wrong number of arguments for \"fold\". Expected 3, received %d.\n\"fold\" takes a list, a starting value, and a binary function that takes the accumulator and the current value and returns a value.", len(args))
	// 		}

	// 		list := args[0]
	// 		if list.Type != ListNT {
	// 			return &Node{Type: FailNT}, nil
	// 		}

	// 		accumulator := args[1]

	// 		fn := args[2]
	// 		if fn.Type != LambdaNT {
	// 			return &Node{Type: FailNT}, nil
	// 		}

	// 		for _, n := range list.Val.(List) {
	// 			call := &Node{
	// 				Type: CallNT,
	// 				L:    fn,
	// 				R: &Node{
	// 					Type: ArgNT,
	// 					L:    accumulator,
	// 					R: &Node{
	// 						Type: ArgNT,
	// 						L:    n,
	// 					},
	// 				},
	// 			}

	// 			val, err := Interpret(call, env)
	// 			if err != nil {
	// 				return nil, err
	// 			}

	// 			accumulator = val
	// 		}

	// 		return accumulator, nil
	// 	},
	// },
	"append": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Wrong number of arguments for \"append\". Expected 2, received %d.", len(args))
			}

			if args[0].Type != ListNT {
				return &Node{Type: FailNT}, nil
			}

			return &Node{
				Type: ListNT,
				Val:  append(args[0].Val.(List), args[1]),
			}, nil
		},
	},
	"reverse": {
		Type: LambdaNT,
		Func: func(_ *Environment, args ...*Node) (*Node, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Wrong number of arguments for \"reverse\". Expected 1, received %d.", len(args))
			}

			if args[0].Type != ListNT {
				return &Node{Type: FailNT}, nil
			}

			list := args[0].Val.(List)
			rev := make(List, len(list))
			for i, n := range list {
				rev[len(list)-i-1] = n
			}

			return &Node{
				Type: ListNT,
				Val:  rev,
			}, nil
		},
	},
}

func castFloat(n *Node) (float64, error) {
	if n.Type == FloatNT {
		return n.Val.(float64), nil
	}
	if n.Type == IntNT {
		return float64(n.Val.(int64)), nil
	}
	return 0, fmt.Errorf("Cannot cast to Float")
}

func castInt(n *Node) (int64, error) {
	if n.Type == IntNT {
		return n.Val.(int64), nil
	}
	if n.Type == FloatNT {
		return int64(n.Val.(float64)), nil
	}
	return 0, fmt.Errorf("Cannot cast to Int")
}
