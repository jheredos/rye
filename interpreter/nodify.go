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

// negateSecond wraps the second result in a Not
func negateSecond(n Nodify) Nodify {
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

// nestLeft nests the parsed node on the left side of a node of the type provided
func nestLeft(p Parser, nt NodeType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			res.node = &Node{
				Type: nt,
				L:    res.node,
			}
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
				Val:  List{res.node},
			}
		}
		return res
	}
}

func nAlways(nt NodeType) Nodify {
	return func(_ ...ParseRes) *Node {
		return &Node{Type: nt}
	}
}

// maybeFunc checks if a subtree should be converted to a function (i.e. it contains an underscore)
// and can be (it does not contain any compound expressions or statements)
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
var nAtom func(NodeType) Nodify = func(nt NodeType) Nodify {
	return func(res ...ParseRes) *Node {
		res1, ok := getParsed(res)
		if !ok {
			return nil
		}

		var val interface{}
		switch nt {
		case IntNT:
			val, _ = strconv.ParseInt(res1.parsed.Lexeme, 10, 64)
		case FloatNT:
			val, _ = strconv.ParseFloat(res1.parsed.Lexeme, 64)
		case IdentifierNT:
			val = res1.parsed.Lexeme
		case BoolNT:
			if res1.parsed.Lexeme == "true" {
				val = true
			} else {
				val = false
			}
		case StringNT:
			val = res1.parsed.Lexeme
		case UnderscoreNT:
			val = "_"
		case IndexNT:
			val = "index"
		case NullNT, FailNT, SuccessNT:
			val = nil
		}

		return &Node{
			Type: nt,
			Val:  val,
			Line: res1.parsed.Line,
		}
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
		L:    k.node,
		R:    v.node,
	}
}

var nObject Nodify = func(res ...ParseRes) *Node {
	return &Node{Type: ObjectNT, Val: Object{}}
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

	return &Node{
		Type: op.node.Type,
		R:    rhs.node,
	}
}

// nUnaryPost creates a node with a unary postfix operator and its argument
var nUnaryPost Nodify = func(res ...ParseRes) *Node {
	lhs, op, ok := get2Results(res)
	if !ok {
		fmt.Println("nUnaryPost failed")
		return nil
	}

	return &Node{
		Type: op.node.Type,
		R:    lhs.node,
	}
}

// Binary
var nRhs Nodify = func(res ...ParseRes) *Node {

	//									A
	//	A 	+		B		=		 \
	//										B

	op, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nRhs failed :(")
		return nil
	}

	return &Node{
		Type: op.node.Type,
		Val:  op.node.Val,
		L:    op.node.L,
		R:    rhs.node,
	}
}

var nLhs Nodify = func(res ...ParseRes) *Node {

	//									A
	//	A 	+		B		=	 /
	//								B

	a, b, ok := get2Results(res)
	if !ok {
		fmt.Println("nLhs failed :(")
		return nil
	}

	return &Node{
		Type: a.node.Type,
		L:    b.node,
	}
}

var nBinary Nodify = func(res ...ParseRes) *Node {

	//  		B					B
	// A	+	 \		=  / \
	//				C     A   C

	lhs, rest, ok := get2Results(res)
	if !ok {
		fmt.Println("nBinary failed :(")
		return nil
	}

	return &Node{
		Type: rest.node.Type,
		L:    lhs.node,
		R:    rest.node.R,
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

	return &Node{
		Type: op.node.Type,
		L:    op.node.L,
		R:    rhs.node,
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
			Type: ThenBranchNT,
			L:    ifNode.R, // the then branch
			R:    fallback, // the else branch
		},
	}
}

var nAssignmentRhs Nodify = func(res ...ParseRes) *Node {
	op, rhs, ok := get2Results(res)
	if !ok {
		fmt.Println("nAssignmentRhs failed :(")
		return nil
	}

	// compound assignment (+=, -=, etc.)
	if op.node.R != nil {
		op.node.R.R = rhs.node
		return op.node
	}

	// simple assignment
	op.node.R = rhs.node
	return op.node
}

var nAssignment Nodify = func(res ...ParseRes) *Node {
	target, op, ok := get2Results(res)
	if !ok {
		fmt.Println("nAssignment failed :(")
		return nil
	}

	op.node.L = target.node
	if op.node.Type == AugAssignNT {
		op.node.Type = AssignmentNT
		op.node.R.L = target.node
	}

	op.node.Line = op.tokens[0].Line
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
		fmt.Println("nRightAssoc failed :(")
		return nil
	}

	o1 := prev.node
	r1 := prev.node.R
	o2 := rhs.node
	r2 := rhs.node.R

	o1.R = &Node{
		Type: o2.Type,
		L:    r1,
		R:    r2,
	}

	return o1.R
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
		fmt.Println("nLeftAssoc failed :(")
		return nil
	}

	o1 := prev.node
	o2 := rhs.node
	o2.L = o1

	return o2
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

	return curr.node
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
		Val:  List{},
	}
}

var nListHead Nodify = func(res ...ParseRes) *Node {
	head, tail, ok := get2Results(res)
	if !ok {
		fmt.Println("nListHead failed :(")
		return nil
	}

	if head.node.Type != ListNT {
		fmt.Println("nListHead failed :(")
		return nil
	}

	if tail.node.Type != ListNT {
		h := head.node.Val.(List)
		return &Node{
			Type: ListNT,
			Val:  append(h, tail.node),
		}
	}

	h, t := head.node.Val.(List), tail.node.Val.(List)

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
		list := prev.node.Val.(List)

		return &Node{
			Type: ListNT,
			Val:  append(list, curr.node),
		}
	}

	return &Node{
		Type: ListNT,
		Val:  List{prev.node, curr.node},
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
