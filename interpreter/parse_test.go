package interpreter

import (
	"strings"
	"testing"
)

func removeWhitespace(s string) string {
	res := strings.ReplaceAll(s, " ", "")
	res = strings.ReplaceAll(res, "\t", "")
	res = strings.ReplaceAll(res, "\n", "")
	return res
}

type SingleNodeTest struct {
	input string
	NodeType
	sExpr string // S-expression representing AST
}

func runSingleNodeTest(test SingleNodeTest, t *testing.T) {
	tkns := Scan(test.input)
	ast, err := Parse(tkns)

	if err != nil {
		t.Fatalf(`Failed to parse "%s": %s`, test.input, err.Error())
	}

	if ast == nil || ast.L == nil {
		t.Fatalf(`Failed to parse "%s": Parse returned nil`, test.input)
	}

	node := ast.L
	if node.Type != test.NodeType {
		t.Fatalf(`Parsed "%s" incorrectly. 
		Expected: %s
		Received: %s`,
			test.input, nodeTypeMap[test.NodeType], nodeTypeMap[node.Type])
	}

	if removeWhitespace(node.ToString()) != removeWhitespace(test.sExpr) {
		t.Fatalf(`Parsed "%s" incorrectly. 
		Expected: %s
		Received: %s`,
			test.input, test.sExpr, node.ToString())
	}
}

func TestParseAtom(t *testing.T) {
	tests := []SingleNodeTest{
		{"x", IdentifierNT, "x"},
		{"42", IntNT, "42"},
		{"true", BoolNT, "true"},
		{"fail", FailNT, "fail"},
		{`"foo"`, StringNT, `"foo"`},
		{"3.14", FloatNT, "3.14"},
		{"_", UnderscoreNT, "_"},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

func TestParseCollection(t *testing.T) {
	tests := []SingleNodeTest{
		// lists
		{"[]", ListNT, "[]"},
		{"[1]", ListNT, "[1]"},
		{"[1,2,3,4]", ListNT, "[1,2,3,4]"},
		// sets
		{`{"apple"}`, SetItemNT, `(set-item "apple")`},
		{`{"apple", "banana"}`, SetItemNT, `(set-item "apple" (set-item "banana"))`},
		// objects
		{"{}", ObjectNT, "{}"},
		{"{a: 1}", ObjectItemNT, `(object-item (: a 1))`},
		{"{a: 1, b: true}", ObjectItemNT, `(object-item (: a 1) (object-item (: b true)))`},
		{`{a: 1, b: true, c: {d: "foo"}}`, ObjectItemNT, `
			(object-item (: a 1) 
			(object-item (: b true) 
			(object-item (: c 
				(object-item (: d "foo"))
			))))
		`},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

func TestParsePrimary(t *testing.T) {
	tests := []SingleNodeTest{
		// call
		{`f()`, CallNT, `(call f (arg))`},
		{`f(1)`, CallNT, `(call f (arg 1))`},
		{`f(1, "two")`, CallNT, `(call f (arg 1 (arg "two")))`},
		// list slice
		{`myList[1..]`, ListSliceNT, `(slice-access myList (slice 1 NIL_PTR))`},
		{`myList[..5]`, ListSliceNT, `(slice-access myList (slice NIL_PTR 5))`},
		{`myList[2..x]`, ListSliceNT, `(slice-access myList (slice 2 x))`},
		// bracket access
		{`myList[-1]`, BracketAccessNT, `(bracket-access myList (- 1))`},
		{`myObj["foo"]`, BracketAccessNT, `(bracket-access myObj "foo")`},
		// dot access
		{`myObj.foo`, FieldAccessNT, `(field-access myObj foo)`},
		// chained
		{`hof(1)(2)`, CallNT, `
			(call 
				(call hof (arg 1)) 
				(arg 2)
			)
		`},
		{`myMatrix[2][4]`, BracketAccessNT, `
			(bracket-access 
				(bracket-access myMatrix 2) 
				4
			)
		`},
		{`foo(x).y`, FieldAccessNT, `
			(field-access 
				(call foo (arg x)) 
				y
			)
		`},
		{`foo[-1].bar("baz")`, CallNT, `
			(call 
				(field-access 
					(bracket-access foo (- 1)) 
					bar) 
				(arg "baz")
			)
		`},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

// Unary expressions
func TestParseUnary(t *testing.T) {
	tests := []SingleNodeTest{
		// individual operators
		{"-3", UnaryNegNT, "(- 3)"},
		{"!false", LogicNotNT, "(! false)"},
		{"#[]", CardinalityNT, "(# [])"},
		{"result?", MaybeNT, "(? result)"},
		{"[...xs]", ListNT, "[(... xs)]"},
		// combined
		{"!result?", LogicNotNT, "(! (? result))"},
		{"-#foo", UnaryNegNT, "(- (# foo))"},
		{"!-#foo?", LogicNotNT, "(! (- (# (? foo))))"},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

// Binary expressions
type BinaryTest struct {
	input          string
	root, lhs, rhs NodeType
	sExpr          string
}

func runBinaryTest(test BinaryTest, t *testing.T) {
	tkns := Scan(test.input)
	ast, err := Parse(tkns)

	if err != nil {
		t.Fatalf(`Failed to parse "%s": %s`, test.input, err.Error())
	}

	if ast == nil || ast.L == nil {
		t.Fatalf(`Failed to parse "%s": Parse returned nil`, test.input)
	}

	root := ast.L
	if root.Type != test.root {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. Expected %s, received %s`, test.input, test.root.ToString(), root.Type.ToString())
	}

	lhs, rhs := root.L, root.R
	if lhs == nil {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. Missing lhs`, test.input)
	}

	if rhs == nil {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. Missing rhs`, test.input)
	}

	if lhs.Type != test.lhs {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. Expected lhs of type %s, received %s`, test.input, test.lhs.ToString(), lhs.Type.ToString())
	}

	if rhs.Type != test.rhs {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. Expected rhs of type %s, received %s`, test.input, test.rhs.ToString(), rhs.Type.ToString())
	}

	if removeWhitespace(root.ToString()) != removeWhitespace(test.sExpr) {
		t.Fatalf(`Parsed binary expression "%s" incorrectly. 
		Expected: %s
		Received: %s`, test.input, test.sExpr, root.ToString())
	}
}

// Arithmetic
func TestParseArithmetic(t *testing.T) {
	tests := []BinaryTest{
		{`2 + 2`, AddNT, IntNT, IntNT, `(+ 2 2)`},
		{`1 + 2 * 3 - 4.5 ^ 6`, SubtNT, AddNT, PowerNT, `(- (+ 1 (* 2 3)) (^ 4.5 6))`},
		{`(1 + 2) * (3 - 4.5) / 6`, DivNT, MultNT, IntNT, `(/ (* (+ 1 2) (- 3 4.5)) 6)`},
		{`((1 + 2 * 3) - 4) / 5 % 6 + 7`, AddNT, ModuloNT, IntNT, `(+ (% (/ (- (+ 1 (* 2 3)) 4) 5) 6) 7)`},
		{`x == 2`, EqualNT, IdentifierNT, IntNT, `(== x 2)`},
		{`x >= 0 != y < 0`, NotEqualNT, GreaterEqualNT, LessNT, `(!= (>= x 0) (< y 0))`},
	}

	for _, test := range tests {
		runBinaryTest(test, t)
	}
}

// Logical
func TestParseLogical(t *testing.T) {
	tests := []BinaryTest{
		{`true and !false`, LogicAndNT, BoolNT, LogicNotNT, `(and true (! false))`},
		{`foo or bar`, LogicOrNT, IdentifierNT, IdentifierNT, `(or foo bar)`},
		{`fail | success`, FallbackNT, FailNT, SuccessNT, `(| fail success)`},
		{`x in [1,2]`, InNT, IdentifierNT, ListNT, `(in x [1, 2])`},
		{`a and b or c and d`, LogicOrNT, LogicAndNT, LogicAndNT, `(or (and a b) (and c d))`},
		{`a or b or c and d | e and f`, FallbackNT, LogicOrNT, LogicAndNT, `(| (or (or a b) (and c d)) (and e f))`},
	}

	for _, test := range tests {
		runBinaryTest(test, t)
	}
}

// Conditional
func TestParseConditionalExpr(t *testing.T) {
	tests := []SingleNodeTest{
		{`x if y == z`, IfNT, `(if (== y z) x)`},
		{`x unless y <= 0`, IfNT, `(if (! (<= y 0)) x)`},
		{`1 if foo(x) else -1`, IfNT, `
			(if 
				(call foo (arg x)) 
				(then-branch 
					1 
					(- 1)
				)
			)`},
		{`"foo" unless bar? else "baz"`, IfNT, `
			(if 
				(! (? bar)) 
				(then-branch "foo" "baz")
			)`},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

// Lambdas
func TestParseLambda(t *testing.T) {
	// param values are not shown since they do not have an AST node of their own, but are stored as the Val of the param node
	tests := []SingleNodeTest{
		// single expression
		{`() => 1`, LambdaNT, `(lambda (param) 1)`},
		{`x => x + 1`, LambdaNT, `(lambda (param) (+ x 1))`},
		{`(x) => x ^ 2`, LambdaNT, `(lambda (param) (^ x 2))`},
		{`(x, y) => x if x > y else y`, LambdaNT, `(lambda (param (param)) (if (> x y) (then-branch x y)))`},
		{`x => y => x + y`, LambdaNT, `(lambda (param) (lambda (param) (+ x y)))`},
		// block body
		{`
			(x, y) => {
				z := x + y
				return z
			}
		`, LambdaNT, `
			(lambda (param (param)) 
				(const z (+ x y))
				(return z)
			)
		`},
		// destructured params...
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

// Compound expressions
func TestParseCompoundExpr(t *testing.T) {
	tests := []BinaryTest{
		{`1..100 where _ % 2 == 0`, WhereNT, RangeNT, LambdaNT, `(where (range 1 100) (lambda (param) (== (% _ 2) 0)))`},
		{`[2,4,6] map _ ^ 2`, MapNT, ListNT, LambdaNT, `(map [2, 4, 6] (lambda (param) (^ _ 2)))`},
		{`foo() then print`, PipeNT, CallNT, IdentifierNT, `(|> (call foo (arg)) print)`},
		{`ns map _ + index where _ > 0 then #_`, PipeNT, WhereNT, LambdaNT, `
			(|> 
				(where 
					(map 
						ns 
						(lambda (param) (+ _ index))) 
					(lambda (param) (> _ 0))) 
				(lambda (param) (# _))
			)`},
	}

	for _, test := range tests {
		runBinaryTest(test, t)
	}
}

// Assignment
func TestParseAssignment(t *testing.T) {
	tests := []BinaryTest{
		{`x := 1`, ConstDeclNT, IdentifierNT, IntNT, `(const x 1)`},
		{`var y := 2`, VarDeclNT, IdentifierNT, IntNT, `(var y 2)`},
		{`y += 1`, AssignmentNT, IdentifierNT, AddNT, `(= y (+ y 1))`},
		{`f := x => x + 1`, ConstDeclNT, IdentifierNT, LambdaNT, `(const f (lambda (param) (+ x 1)))`},
		{`z.a = "foo"`, AssignmentNT, FieldAccessNT, StringNT, `(= (field-access z a) "foo")`},
		{`z.a[3] = "bar"`, AssignmentNT, BracketAccessNT, StringNT, `(= (bracket-access (field-access z a) 3) "bar")`},
	}

	for _, test := range tests {
		runBinaryTest(test, t)
	}
}

// Conditional statements
func TestParseConditionalStmt(t *testing.T) {
	tests := []SingleNodeTest{
		{`if foo(): return true`, IfNT, `(if (call foo (arg)) (return true))`},
		{`unless x != y: print("equal")`, IfNT, `(if (! (!= x y)) (call print (arg "equal")))`},
		{`if foo() {
				x += 1
				return x
			}`, IfNT, `
			(if (call foo (arg)) 
        (= x (+ x 1))
        (return x)
			)`},
		{`if foo() {
				x += 1
				return x
			} else {
				return fail
			}`, IfNT, `
			(if (call foo (arg)) 
				(then-branch 
					(= x (+ x 1))
					(return x) 
        	(return fail)
				)
			)`},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}

// Loops
func TestParseLoop(t *testing.T) {
	tests := []SingleNodeTest{
		{`for x in 1..10: print(x)`, ForStmtNT, `(for (const x (range 1 10)) (call print (arg x)))`},
		{`while true { print("foo") }`, WhileStmtNT, `(while true (call print (arg "foo")))`},
	}

	for _, test := range tests {
		runSingleNodeTest(test, t)
	}
}
