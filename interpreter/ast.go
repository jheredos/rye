package interpreter

import "fmt"

type Node struct {
	Type NodeType
	Val  interface{}
	Func
	L, R  *Node
	Scope *Environment
	Line  int
}

type Func func(*Environment, ...*Node) (*Node, error)

type List []*Node
type Set map[Value]bool
type Object map[Value]*Node

type Value struct {
	DataType
	Val interface{}
}

type DataType uint8

const (
	IntDT DataType = iota
	FloatDT
	BoolDT
	StringDT
	ResultDT

	LambdaDT
	ListDT
	SetDT
	ObjectDT
)

type NodeType uint8

const (
	UnknownNT NodeType = iota
	ErrorNT

	ProgramNT
	LineNT
	StmtNT
	BlockNT
	ConstDeclNT
	VarDeclNT
	ReturnStmtNT
	WhileStmtNT
	ForStmtNT
	BreakNT
	ContinueNT
	ExprNT
	IfNT
	ThenBranchNT
	AssignmentNT
	AugAssignNT
	LambdaNT
	ParamNT
	ArgNT
	LogicOrNT
	LogicAndNT
	EqualNT
	NotEqualNT
	LessNT
	LessEqualNT
	GreaterNT
	GreaterEqualNT
	AddNT
	SubtNT
	MultNT
	DivNT
	FallbackNT
	ModuloNT
	MapNT
	WhereNT
	PipeNT
	FindNT
	FoldNT
	BindNT
	InNT
	PowerNT
	UnderscoreNT
	IndexNT
	SliceNT
	KVPairNT
	SetItemNT
	ObjectItemNT

	LogicNotNT
	UnaryNegNT
	CardinalityNT
	MaybeNT
	SplatNT

	ListNT
	SetNT
	ObjectNT

	SuccessNT
	FailNT

	ImportNT
	ModuleNT

	CallNT
	RangeNT
	BracketAccessNT
	ListSliceNT
	FieldAccessNT
	IdentifierNT
	FloatNT
	IntNT
	BoolNT
	StringNT
	CharNT
	NullNT
	EOFNT
)

type Environment struct {
	Parent *Environment
	Vars   map[string]*Node
	Consts map[string]*Node
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
	case StringNT, IdentifierNT:
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

var nodeTypeMap map[NodeType]string = map[NodeType]string{
	ProgramNT:       "program",
	LineNT:          "line",
	StmtNT:          "stmt",
	BlockNT:         "block",
	VarDeclNT:       "var",
	ConstDeclNT:     "const",
	ReturnStmtNT:    "return",
	WhileStmtNT:     "while",
	ForStmtNT:       "for",
	ExprNT:          "expr",
	IfNT:            "if",
	ThenBranchNT:    "then-branch",
	AssignmentNT:    "=",
	LambdaNT:        "lambda",
	ParamNT:         "param",
	ArgNT:           "arg",
	LogicOrNT:       "or",
	LogicAndNT:      "and",
	EqualNT:         "==",
	NotEqualNT:      "!=",
	LessNT:          "<",
	LessEqualNT:     "<=",
	GreaterNT:       ">",
	GreaterEqualNT:  ">=",
	AddNT:           "+",
	SubtNT:          "-",
	MultNT:          "*",
	DivNT:           "/",
	FallbackNT:      "|",
	ModuloNT:        "%",
	LogicNotNT:      "!",
	UnaryNegNT:      "-",
	ListNT:          "list",
	SetNT:           "set",
	ObjectNT:        "obj",
	SuccessNT:       "success",
	FailNT:          "fail",
	CallNT:          "call",
	RangeNT:         "range",
	BracketAccessNT: "bracket-access",
	FieldAccessNT:   "field-access",
	IdentifierNT:    "IDENT",
	FloatNT:         "FLOAT",
	IntNT:           "INT",
	BoolNT:          "BOOL",
	StringNT:        "STRING",
	CharNT:          "CHAR",
	NullNT:          "null",
	EOFNT:           "",
	CardinalityNT:   "#",
	MaybeNT:         "?",
	MapNT:           "map",
	WhereNT:         "where",
	InNT:            "in",
	PowerNT:         "^",
	PipeNT:          "|>",
	UnderscoreNT:    "_",
	BreakNT:         "break",
	ContinueNT:      "continue",
	SliceNT:         "slice",
	ListSliceNT:     "slice-access",
	KVPairNT:        ":",
	SetItemNT:       "set-item",
	ImportNT:        "import",
	ModuleNT:        "module",
	SplatNT:         "...",
	ObjectItemNT:    "object-item",
	FindNT:          "find",
	FoldNT:          "fold",
}

func (nt NodeType) ToString() string {
	return nodeTypeMap[nt]
}

func unOp2String(n *Node) string {
	return fmt.Sprintf("(%s %s)", n.Type.ToString(), n.R.ToString())
}

func binOp2String(n *Node) string {
	return fmt.Sprintf("(%s %s %s)", n.Type.ToString(), n.L.ToString(), n.R.ToString())
}

func linked2String(n *Node) string {
	if n.L == nil && n.R == nil {
		return fmt.Sprintf("(%s)", n.Type.ToString())
	}

	if n.L == nil {
		return fmt.Sprintf("(%s %s)", n.Type.ToString(), n.R.ToString())
	}

	if n.R == nil {
		return fmt.Sprintf("(%s %s)", n.Type.ToString(), n.L.ToString())
	}

	return fmt.Sprintf("(%s %s %s)", n.Type.ToString(), n.L.ToString(), n.R.ToString())
}

func Display(n *Node) string {
	if n == nil {
		return "NIL_PTR"
	}
	switch n.Type {
	case FloatNT, IntNT, CharNT, BoolNT, IdentifierNT, StringNT, ListNT, ObjectNT, SetNT, NullNT, UnderscoreNT, FailNT, SuccessNT:
		return n.ToString()
	case LambdaNT:
		return "<lambda>"
	default:
		return "success"
	}
}

func (n *Node) ToString() string {
	if n == nil {
		return "NIL_PTR"
	}
	switch n.Type {
	// atoms
	case FloatNT, IntNT, CharNT, BoolNT, IdentifierNT:
		return fmt.Sprintf("%v", n.Val)
	case StringNT:
		return fmt.Sprintf("\"%v\"", n.Val)
	case ListNT:
		list := n.Val.(List)
		res := "["
		for i, m := range list {
			if i > 0 {
				res += ", "
			}
			res += fmt.Sprintf(m.ToString())
		}
		res += "]"
		return res
	case ObjectNT:
		obj := n.Val.(Object)
		res := "{"
		for k, v := range obj {
			if len(res) > 1 {
				res += ", "
			}
			res += k.toNode().ToString()
			res += ": "
			res += v.ToString()
		}
		res += "}"
		return res
	case SetNT:
		set := n.Val.(Set)
		res := "{"
		for k := range set {
			if !set[k] {
				continue
			}
			if len(res) > 1 {
				res += ", "
			}
			res += k.toNode().ToString()
		}
		res += "}"
		return res
	case NullNT:
		return "null"
	case UnderscoreNT:
		return "_"
	case FailNT:
		return "fail"
	case SuccessNT:
		return "success"
	case IndexNT:
		return "index"
	case BreakNT, ContinueNT:
		return n.Type.ToString()
	case ModuleNT:
		return fmt.Sprintf("(module %s)", n.Val.(string))
	case ImportNT:
		if n.R != nil {
			return fmt.Sprintf("(import %s %s)\n", n.Val.(string), n.L.Val.(string))
		}
		return fmt.Sprintf("(import %s)\n", n.Val.(string))
	case StmtNT:
		if n.R != nil {
			return fmt.Sprintf("\n%s%s", n.L.ToString(), n.R.ToString())
		}
		return fmt.Sprintf("\n%s", n.L.ToString())
	// unary
	case UnaryNegNT, LogicNotNT, CardinalityNT, MaybeNT, ReturnStmtNT, SplatNT:
		return unOp2String(n)
	// binary
	case MultNT, DivNT, AddNT, SubtNT, ModuloNT, NotEqualNT, EqualNT, GreaterNT, GreaterEqualNT, LessNT, LessEqualNT, FallbackNT, LogicOrNT, LogicAndNT, MapNT, WhereNT, InNT, PowerNT, IfNT, ThenBranchNT, LambdaNT, PipeNT, AssignmentNT, VarDeclNT, ConstDeclNT, WhileStmtNT, ForStmtNT, CallNT, BracketAccessNT, ListSliceNT, FieldAccessNT, RangeNT, SliceNT, KVPairNT, FindNT, FoldNT:
		return binOp2String(n)
	case ParamNT, ArgNT, SetItemNT, ObjectItemNT:
		return linked2String(n)

	default:
		return "UNKNOWN"
	}
}
