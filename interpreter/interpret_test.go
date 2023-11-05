package interpreter

import "testing"

type ExprTest struct {
	input        string
	resultType   NodeType
	resultString string // S-expression representing AST
}

func runExprTest(test ExprTest, t *testing.T) {
	tkns := Scan(test.input)
	ast, err := Parse(tkns)

	if err != nil {
		t.Fatalf(`Failed to parse "%s": %s`, test.input, err.Error())
	}

	env := &Environment{
		Parent: &Environment{
			Consts: StdLib,
		},
		Consts: map[string]*Node{},
		Vars:   map[string]*Node{},
	}

	res, err := Interpret(ast, env)

	if err != nil {
		t.Fatalf(`Failed to evaluate "%s": %s`, test.input, err.Error())
	}

	if res.Type != test.resultType || removeWhitespace(res.ToString()) != removeWhitespace(test.resultString) {
		t.Fatalf(`Evaluated "%s" incorrectly:
			Expected: %s (type %s)
			Received: %s (type %s)
			`,
			test.input,
			test.resultString,
			test.resultType.ToString(),
			res.ToString(),
			res.Type.ToString(),
		)
	}
}

func TestInterpretSimpleExpr(t *testing.T) {
	tests := []ExprTest{
		// arithmetic, logic, conditional
		{`1`, IntNT, `1`},
		{`2 + 2`, IntNT, `4`},
		{`2 + 2 == 4`, BoolNT, `true`},
		{`2.0 ^ 3 != 8`, BoolNT, `false`},
		{`2.0 ^ -3 < .2`, BoolNT, `false`},
		{`1 + 2 * (3 - 4) <= 5 / 6.7`, BoolNT, `true`},
		{`false and true or true and !null`, BoolNT, `true`},
		{`"foo" if false`, FailNT, `fail`},
		{`"foo" if "bar"? else "baz"`, StringNT, `"foo"`},
		// collections, dot/bracket/slice access
		{`[1, 2, 3]`, ListNT, `[1, 2, 3]`},
		{`[1, 2, 3] + [4, 5, 6]`, ListNT, `[1, 2, 3, 4, 5, 6]`},
		{`#[1, 2, 3]`, IntNT, `3`},
		{`[1, 2, 3][5]`, FailNT, `fail`},
		{`[1, 2, 3][-1]`, IntNT, `3`},
		{`"cherry" in {"apple", "banana"}`, BoolNT, `false`},
		{`{ a: true }.a`, BoolNT, `true`},
		{`{ a: [{}, { "foo": {"bar"} }] }.a[1].foo`, SetNT, `{ "bar" }`},
		{`10 in 2..20`, BoolNT, `true`},
		{`(..10)[3..7]`, ListNT, `[3,4,5,6]`},
		{`[3.14][1..]`, ListNT, `[]`},
		{`"foobarbaz"[3..6]`, StringNT, `"bar"`},
		// lambdas, calls
		{`print("hello, world")`, SuccessNT, `success`},
		{`x => x + 1`, LambdaNT, `(lambda (param) (+ x 1))`},
		{`((a, b) => a if a > b else b)(-5, 7)`, IntNT, `7`},
	}

	for _, test := range tests {
		runExprTest(test, t)
	}
}

func TestInterpretStmt(t *testing.T) {
	tests := []ExprTest{
		// declarations, assignment
		{`
			x := 1
			x
		`, IntNT, `1`},
		{`
			var x := 1
			x = "one"
			x
		`, StringNT, `"one"`},
		{`
			var x := 1
			x += 2
			x *= 3
			x -= 4
			x /= 2
			x
		`, FloatNT, `2.5`},
		{`
			var foo := {}
			foo.bar = [1,2,3]
			foo.bar
			// foo.bar[1] = { baz: false }
		`, // known bug! updating a list inside an object (a Go slice inside a map) will require some workarounds: https://stackoverflow.com/questions/69475165/golang-does-not-update-array-in-a-map
			ListNT, `[1,2,3]`},
	}

	for _, test := range tests {
		runExprTest(test, t)
	}
}
