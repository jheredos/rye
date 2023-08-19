package interpreter

import (
	"fmt"
	"strconv"
)

// Nodify is a function that takes any number of parse results and transforms them into an AST node
type Nodify func(...ParseRes) *Node

// utils
var takeFirst Nodify = func(res ...ParseRes) *Node {
	if len(res) == 0 {
		return nil
	}
	// fmt.Println("takeFirst: ", res[0].node.ToString())
	return res[0].node
}

var takeSecond Nodify = func(res ...ParseRes) *Node {
	if len(res) < 2 {
		return nil
	}
	// fmt.Println("takeSecond: ", res[1].node.ToString())
	return res[1].node
}

func allOk(res ...ParseRes) bool {
	for _, r := range res {
		if !r.ok || r.node == nil {
			return false
		}
	}
	return true
}

func getParsed(res []ParseRes) (res1 ParseRes, ok bool) {
	if len(res) < 1 {
		return ParseRes{}, false
	}
	return res[0], true
}

func get2Results(res []ParseRes) (res1, res2 ParseRes, ok bool) {
	if len(res) < 2 || !allOk(res...) {
		return ParseRes{}, ParseRes{}, false
	}

	return res[0], res[1], true
}

func get3Results(res []ParseRes) (res1, res2, res3 ParseRes, ok bool) {
	if len(res) < 3 || !allOk(res...) {
		return ParseRes{}, ParseRes{}, ParseRes{}, false
	}

	return res[0], res[1], res[2], true
}

// invertFirst wraps the first result in a Not
func invertFirst(n Nodify) Nodify {
	return func(res ...ParseRes) *Node {
		if len(res) >= 1 {
			res[0].node = &Node{
				Type: LogicNotNT,
				R:    res[0].node,
			}
		}
		return n(res...)
	}
}

// invertFirst wraps the second result in a Not
func invertSecond(n Nodify) Nodify {
	return func(res ...ParseRes) *Node {
		if len(res) >= 2 {
			res[1].node = &Node{
				Type: LogicNotNT,
				R:    res[1].node,
			}
		}
		return n(res...)
	}
}

func alterNodeType(p Parser, nt NodeType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			res.node.Type = nt
		}
		return res
	}
}

func nestNode(p Parser, nt NodeType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			// fmt.Println("nestNode: ", nodeTypeMap[nt])
			// fmt.Printf("\t%s\n", res.node.ToString())
			res.node = &Node{
				Type: nt,
				L:    res.node,
			}
			// fmt.Printf("\t%s\n", res.node.ToString())
		}
		return res
	}
}

func nestRight(p Parser, nt NodeType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			res.node = &Node{
				Type: nt,
				R:    res.node,
			}
		}
		return res
	}
}

func listify(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			res.node = &Node{
				Type: ListNT,
				Val:  []*Node{res.node},
			}
		}
		return res
	}
}

func maybeFunc(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if !res.ok {
			return res
		}

		forbidden := map[NodeType]bool{
			MapNT:   true,
			WhereNT: true,
			PipeNT:  true,
			StmtNT:  true,
		}

		underscore := false
		q, q2 := []*Node{res.node}, []*Node{}
		for len(q) > 0 {
			for _, n := range q {
				if forbidden[n.Type] {
					return res
				}

				if n.Type == UnderscoreNT {
					underscore = true
				}

				if n.L != nil {
					q2 = append(q2, n.L)
				}
				if n.R != nil {
					q2 = append(q2, n.R)
				}
			}
			q, q2 = q2, []*Node{}
		}

		if underscore {
			n := &Node{
				Type: LambdaNT,
				L: &Node{
					Type: ParamNT,
					Val:  "_",
				},
				R: res.node,
			}

			return ParseRes{
				ok:     true,
				node:   n,
				tokens: res.tokens,
			}
		}
		return res
	}
}

// Atoms
var nInt Nodify = func(res ...ParseRes) *Node {
	res1, ok := getParsed(res)
	if !ok {
		return nil
	}

	val, _ := strconv.ParseInt(res1.parsed.Lexeme, 10, 64)
	return &Node{
		Type: IntNT,
		Val:  val,
	}
}

var nFloat Nodify = func(res ...ParseRes) *Node {
	res1, ok := getParsed(res)
	if !ok {
		return nil
	}

	val, _ := strconv.ParseFloat(res1.parsed.Lexeme, 64)
	return &Node{
		Type: FloatNT,
		Val:  val,
	}
}

var nIdentifier Nodify = func(res ...ParseRes) *Node {
	res1, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: IdentifierNT,
		Val:  res1.parsed.Lexeme,
	}
}

var nParam Nodify = func(res ...ParseRes) *Node {
	res1, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: ParamNT,
		Val:  res1.parsed.Lexeme,
	}
}

var nString Nodify = func(res ...ParseRes) *Node {
	res1, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: StringNT,
		Val:  res1.parsed.Lexeme,
	}
}

var nTrue Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: BoolNT,
		Val:  true,
	}
}

var nFalse Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: BoolNT,
		Val:  false,
	}
}

var nNull Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: NullNT,
	}
}

var nFail Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: FailNT,
	}
}

var nSuccess Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{
		Type: SuccessNT,
	}
}

var nUnderscore Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{Type: UnderscoreNT, Val: "_"}
}

var nIndex Nodify = func(res ...ParseRes) *Node {
	_, ok := getParsed(res)
	if !ok {
		return nil
	}

	return &Node{Type: IndexNT, Val: "index"}
}

var nSlice Nodify = func(res ...ParseRes) *Node {
	if len(res) == 1 {
		// full slice: x[..]
		if res[0].node == nil {
			return &Node{Type: SliceNT}
		}
		// start included: x[3..]
		return &Node{
			Type: SliceNT,
			R:    res[0].node,
		}
	}
	if len(res) == 2 {
		// start and end included
		return &Node{
			Type: SliceNT,
			L:    res[0].node,
			R:    res[1].node,
		}
	}
	return nil
}

var nKVPair Nodify = func(res ...ParseRes) *Node {
	k, v, ok := get2Results(res)

	if !ok {
		fmt.Println("nKVPair failed :(")
		return nil
	}

	return &Node{
		Type: KVPairNT,
		Val:  k.node.Val.(string),
		L:    v.node,
	}
}

var nObject Nodify = func(res ...ParseRes) *Node {
	return &Node{Type: ObjectNT}
}

var nImport Nodify = func(res ...ParseRes) *Node {
	if !res[0].ok {
		return nil
	}

	return &Node{
		Type: ImportNT,
		Val:  res[1].node.Val.(string),
	}
}

// Unary
// nUnaryPre creates a node with a unary prefix operator and its argument
var nUnaryPre Nodify = func(res ...ParseRes) *Node {
	op, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nUnaryPre failed")
		return nil
	}

	switch op.node.Type {
	case UnaryNegNT:
		return &Node{
			Type: op.node.Type,
			R:    rhs.node,
		}
	case LogicNotNT:
		return &Node{
			Type: op.node.Type,
			R:    rhs.node,
		}
	case CardinalityNT:
		return &Node{
			Type: op.node.Type,
			R:    rhs.node,
		}
	default:
		fmt.Println("nUnaryPre failed")
		return nil
	}
}

// nUnaryPost creates a node with a unary postfix operator and its argument
var nUnaryPost Nodify = func(res ...ParseRes) *Node {
	lhs, op, ok := get2Results(res)
	if !ok {
		fmt.Println("nUnaryPost failed")
		return nil
	}

	switch op.node.Type {
	case MaybeNT:
		return &Node{
			Type: op.node.Type,
			R:    lhs.node,
		}
	default:
		fmt.Println("nUnaryPost failed")
		return nil
	}
}

var nNegative Nodify = func(res ...ParseRes) *Node {
	_, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nNegative failed :(")
		return nil
	}

	switch rhs.node.Type {
	case FloatNT:
		return &Node{
			Type: FloatNT,
			Val:  -rhs.node.Val.(float64),
		}
	case IntNT:
		return &Node{
			Type: IntNT,
			Val:  -rhs.node.Val.(int64),
		}
	default:
		return &Node{
			Type: UnaryNegNT,
			R:    rhs.node,
		}
	}
}

// Binary

var nRhs Nodify = func(res ...ParseRes) *Node {

	//									A
	//	A 	+		B		=		 \
	//										B

	op, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nRhs failed :( (bad results)")
		return nil
	}

	switch op.node.Type {
	case LogicOrNT, LogicAndNT, EqualNT, NotEqualNT, LessNT, LessEqualNT, GreaterNT, GreaterEqualNT, AddNT, SubtNT, MultNT, DivNT, ModuloNT, FallbackNT, PowerNT, ParamNT, LambdaNT, MapNT, WhereNT, PipeNT, VarDeclNT, ConstDeclNT, IfNT, WhileStmtNT, ForStmtNT, CallNT, ArgNT, SliceNT, KVPairNT, ImportNT, SetItemNT, InNT, IdentifierNT:
		// fmt.Printf("op:      %s\n", op.node.ToString())
		// fmt.Printf("rhs:     %s\n", rhs.node.ToString())
		n := &Node{
			Type: op.node.Type,
			Val:  op.node.Val,
			L:    op.node.L,
			R:    rhs.node,
		}
		// fmt.Printf("result:  %s\n", n.ToString())
		return n
	default:
		fmt.Println("nRhs failed :( (unknown node type)")
		return nil
	}
}

var nLhs Nodify = func(res ...ParseRes) *Node {

	//									A
	//	A 	+		B		=	 /
	//								B

	a, b, ok := get2Results(res)
	if !ok {
		fmt.Println("nLhs failed :( (bad results)")
		return nil
	}

	switch a.node.Type {
	case IfNT, WhileStmtNT, ForStmtNT:
		n := &Node{
			Type: a.node.Type,
			L:    b.node,
		}

		return n
	default:
		fmt.Println("nLhs failed :( (unknown node type)")
		return nil
	}
}

var nBinary Nodify = func(res ...ParseRes) *Node {

	//  		B					B
	// A	+	 \		=  / \
	//				C     A   C

	lhs, rest, ok := get2Results(res)
	if !ok {
		fmt.Println("nBinary failed :( (bad results)")
		return nil
	}

	switch rest.node.Type {
	case LogicOrNT, LogicAndNT, EqualNT, NotEqualNT, LessNT, LessEqualNT, GreaterNT, GreaterEqualNT, AddNT, SubtNT, MultNT, DivNT, ModuloNT, FallbackNT, PowerNT, LambdaNT, ConstDeclNT, VarDeclNT:
		n := &Node{
			Type: rest.node.Type,
			L:    lhs.node,
			R:    rest.node.R,
		}
		// fmt.Printf("nBinary: %s\n", n.ToString())
		return n
	default:
		fmt.Println("nBinary failed :( (unknown node type)")
		return nil
	}
}

var nBinaryFlip Nodify = func(res ...ParseRes) *Node {

	//  		  B		  B
	// A	+  /	=  / \
	//		  C     C   A

	rhs, op, ok := get2Results(res)
	if !ok {
		fmt.Println("nBinaryFlip failed :(")
		return nil
	}

	switch op.node.Type {
	case IfNT:
		n := &Node{
			Type: op.node.Type,
			L:    op.node.L,
			R:    rhs.node,
		}
		return n
	default:
		fmt.Println("nBinaryFlip failed :(")
		return nil
	}
}

// nElse is a special case for Nodify functions. If an else branch is not present, return on the
// then branch, otherwise return a Then node with the then branch on the L and the else branch on
// the R
var nElse Nodify = func(res ...ParseRes) *Node {
	if len(res) != 2 {
		fmt.Println("nElse failed :(")
		return nil
	}

	if res[1].node == nil {
		// no else, return just the then branch
		return res[0].node
	}

	ifNode := res[0].node
	fallback := res[1].node

	return &Node{
		Type: IfNT,
		L:    ifNode.L, // the condition
		R: &Node{
			Type: ThenNT,
			L:    ifNode.R, // the then branch
			R:    fallback, // the else branch
		},
	}
}

var nAssignmentRhs Nodify = func(res ...ParseRes) *Node {
	op, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nAssignmentRhs failed :( (bad results)")
		return nil
	}

	// compound assignment
	if op.node.R != nil {
		op.node.R.R = rhs.node
		return op.node
	}

	// simple assignment
	if op.node.Type == AssignmentNT {
		op.node.R = rhs.node
		return op.node
	}

	fmt.Println("nAssignmentRhs failed :( (unknown node type)")
	return nil
}

var nAssignment Nodify = func(res ...ParseRes) *Node {
	target, op, ok := get2Results(res)
	if !ok {
		fmt.Println("nAssignment failed :( (bad results)")
		return nil
	}

	op.node.L = target.node
	switch op.node.R.Type {
	case SubtNT, AddNT, DivNT, MultNT, ModuloNT, FallbackNT:
		// compound assignment
		if op.node.R.L == nil {
			op.node.R.L = target.node
		}
	}

	return op.node
}

// nRightAssoc expects 2 result structs: the previous result  and the rhs
var nRightAssoc Nodify = func(res ...ParseRes) *Node {

	//		 O1					 O2							O1
	//	  /  \		+		   \			=		 /  \
	// (L1)		R1					R2			(L1)   O2
	//																	/  \
	//																R1    R2

	prev, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nRightAssoc failed :( (bad results)")
		return nil
	}

	switch rhs.node.Type {
	case PowerNT, LambdaNT:
		// fmt.Printf("nRightAssoc:\n")
		// fmt.Printf("\tprev:   %s\n", prev.node.ToString())
		// fmt.Printf("\trhs:    %s\n", rhs.node.ToString())
		o1 := prev.node
		r1 := prev.node.R
		o2 := rhs.node
		r2 := rhs.node.R

		o1.R = &Node{
			Type: o2.Type,
			L:    r1,
			R:    r2,
		}
		// fmt.Printf("result: \t%s\n\n", o1.ToString())
		return o1.R
	default:
		fmt.Println("nRightAssoc failed :( (unknown node type)")
		return nil
	}
}

// nLeftAssoc expects 2 result structs: the previous result and the rhs
var nLeftAssoc Nodify = func(res ...ParseRes) *Node {

	//		 O1					 O2							O2
	//	  /  \		+		   \			=		 /  \
	//  L1		R1					R2			 O1    R2
	//														/  \
	//													L1    R1

	prev, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nLeftAssoc failed :( (bad results)")
		return nil
	}

	switch rhs.node.Type {
	case LogicOrNT, LogicAndNT, EqualNT, NotEqualNT, LessNT, LessEqualNT, GreaterNT, GreaterEqualNT, AddNT, SubtNT, MultNT, DivNT, ModuloNT, FallbackNT, MapNT, WhereNT, PipeNT, CallNT, BracketAccessNT, FieldAccessNT, SliceNT, ListSliceNT:
		// fmt.Printf("nLeftAssoc:\n")
		// fmt.Printf("\tprev:   %s\n", prev.node.ToString())
		// fmt.Printf("\trhs:    %s\n", rhs.node.ToString())
		o1 := prev.node
		o2 := rhs.node

		o2.L = o1
		// fmt.Printf("result: \t%s\n\n", o2.ToString())
		return o2
	default:
		fmt.Printf("nLeftAssoc failed :( (unknown node type \"%s\")\n", nodeTypeMap[rhs.node.Type])
		return nil
	}
}

// nEndLeftAssoc traverses the left-side of a left-associative tree to append its leftmost argument, returning the root
var nEndLeftAssoc Nodify = func(res ...ParseRes) *Node {
	lhs, root, ok := get2Results(res)
	if !ok {
		fmt.Println("nEndLeftAssoc failed :(")
		return nil
	}

	n := root.node
	for n.L != nil {
		n = n.L
	}

	n.L = lhs.node
	return root.node
}

// nLinked creates a linked list of nodes
var nLinked Nodify = func(res ...ParseRes) *Node {
	if len(res) < 1 {
		fmt.Println("nLinked failed :(")
		return nil
	}
	curr := res[0]

	if len(res) < 2 {
		return curr.node
	}

	next := res[1]
	n := curr.node
	for n.R != nil {
		n = n.R
	}
	n.R = next.node

	// fmt.Println("nLinked", curr.node.ToString())
	return curr.node
}

var nStmt Nodify = func(res ...ParseRes) *Node {
	if len(res) < 1 || !res[0].ok {
		fmt.Println("nStmt failed :(")
		return nil
	}

	// fmt.Printf("%s\n", res[0].node.ToString())
	return &Node{
		Type: StmtNT,
		L:    res[0].node,
	}
}

// nUnaryNested combines an inner unary operation with an outer
var nUnaryNested Nodify = func(res ...ParseRes) *Node {
	curr, next, ok := get2Results(res)
	if !ok {
		fmt.Println("nUnaryNested failed :(")
		return nil
	}

	n := curr.node
	for n.R != nil {
		n = n.R
	}
	n.R = next.node

	return curr.node
}

var nEmptyList Nodify = func(res ...ParseRes) *Node {
	return &Node{
		Type: ListNT,
		Val:  []*Node{},
	}
}

var nListHead Nodify = func(res ...ParseRes) *Node {
	// fmt.Println("nListHead")
	head, tail, ok := get2Results(res)
	if !ok {
		fmt.Println("nListHead failed :( (bad results)")
		return nil
	}

	if head.node.Type != ListNT {
		fmt.Println("nListHead failed :(")
		return nil
	}

	if tail.node.Type != ListNT {
		h := head.node.Val.([]*Node)
		return &Node{
			Type: ListNT,
			Val:  append(h, tail.node),
		}
	}

	// fmt.Println("there are 2 lists, all good\n")
	h, t := head.node.Val.([]*Node), tail.node.Val.([]*Node)

	return &Node{
		Type: ListNT,
		Val:  append(h, t...),
	}
}

var nListTail Nodify = func(res ...ParseRes) *Node {
	prev, curr, ok := get2Results(res)
	if !ok {
		fmt.Println("nListTail failed :(")
		return nil
	}

	if prev.node.Type == ListNT {
		list := prev.node.Val.([]*Node)
		list = append(list, curr.node)

		return &Node{
			Type: ListNT,
			Val:  list,
		}
	}

	return &Node{
		Type: ListNT,
		Val:  []*Node{prev.node, curr.node},
	}
}

// nRangeEnd, e.g. "..5", "..x", etc.
var nRangeEnd Nodify = func(res ...ParseRes) *Node {
	_, end, ok := get2Results(res)
	if !ok {
		fmt.Println("nRangeEnd failed :(")
		return nil
	}

	return &Node{
		Type: RangeNT,
		R:    end.node,
	}
}

// nRange, e.g. "..5", "..x", etc.
var nRange Nodify = func(res ...ParseRes) *Node {
	start, end, ok := get2Results(res)
	if !ok {
		fmt.Println("nRange failed :(")
		return nil
	}

	return &Node{
		Type: RangeNT,
		L:    start.node,
		R:    end.node,
	}
}
