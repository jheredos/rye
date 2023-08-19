package interpreter

import "fmt"

// ParseRes holds the state of a parse: success or failure, remaining tokens, current node in the AST
type ParseRes struct {
	ok     bool
	err    string
	node   *Node
	parsed *Token
	tokens []Token
}

// Parser is a function that takes a parse state (ParseRes) and Nodify function that transforms
// parse results into an AST node
type Parser func(ParseRes, Nodify) ParseRes

func fail(message string) ParseRes {
	return ParseRes{
		ok:  false,
		err: message,
	}
}

// skipNewLines skips all new lines before attempting a parser
func skipNewLines(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return p(curr, nil)
		}

		for len(curr.tokens) > 0 && curr.tokens[0].Type == NewLineTT {
			curr.tokens = curr.tokens[1:]
		}

		return p(curr, nil)
	}
}

// Atoms
var operatorMap map[TokenType]NodeType = map[TokenType]NodeType{
	BangEqualTT:    NotEqualNT,
	DotDotTT:       RangeNT,
	EqualTT:        AssignmentNT,
	EqualEqualTT:   EqualNT,
	GreaterTT:      GreaterNT,
	GreaterEqualTT: GreaterEqualNT,
	LessTT:         LessNT,
	LessEqualTT:    LessEqualNT,
	BarTT:          FallbackNT,
	PlusTT:         AddNT,
	MinusTT:        SubtNT,
	StarTT:         MultNT,
	SlashTT:        DivNT,
	ModuloTT:       ModuloNT,
	CaratTT:        PowerNT,
	InTT:           InNT,
	AndTT:          LogicAndNT,
	OrTT:           LogicOrNT,
	PipeTT:         PipeNT,
	MapTT:          MapNT,
	WhereTT:        WhereNT,
	IfTT:           IfNT,
	UnlessTT:       IfNT,
	ArrowTT:        LambdaNT,
	ColonEqualTT:   ConstDeclNT, // TODO: split colon and equal operators to allow types in between
	LeftArrowTT:    ConstDeclNT,
	WhileTT:        WhileStmtNT,
	UntilTT:        WhileStmtNT,
	ForTT:          ForStmtNT,
	BreakTT:        BreakNT,
	ContinueTT:     ContinueNT,
	IndexTT:        IndexNT,
}

// pOperator creates a parser for a binary operator, finding the appropriate node based on a token type
func pOperator(tt TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}
		if len(curr.tokens) == 0 {
			return fail("Tokens exhausted")
		}
		if curr.tokens[0].Type == tt {
			op, ok := operatorMap[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok:     true,
				node:   &Node{Type: op},
				tokens: curr.tokens[1:],
			}
		}
		return fail("No match")
	}
}

func pOperatorUnary(tt TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if len(curr.tokens) == 0 {
			return fail("Tokens exhausted")
		}
		if curr.tokens[0].Type == tt {
			op, ok := map[TokenType]NodeType{
				MinusTT:        UnaryNegNT,
				BangTT:         LogicNotNT,
				HashTT:         CardinalityNT,
				QuestionMarkTT: MaybeNT,
			}[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok:     true,
				node:   &Node{Type: op},
				tokens: curr.tokens[1:],
			}
		}
		return fail("No match")
	}
}

// pToken creates a parser for any token type, using the nodify function provided
func pToken(tt TokenType, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}
		if len(curr.tokens) == 0 {
			return fail("Tokens exhausted")
		}
		if curr.tokens[0].Type == tt {
			res := ParseRes{
				ok:     true,
				parsed: &curr.tokens[0],
				tokens: curr.tokens[1:],
			}
			if n != nil {
				res.node = n(res)
			}

			return res
		}
		return ParseRes{
			ok:     false,
			err:    fmt.Sprintf("Expected %s", tt.ToString()),
			tokens: curr.tokens,
		}
	}
}

var pEOF Parser = func(curr ParseRes, _ Nodify) ParseRes {
	if len(curr.tokens) == 0 {
		return fail("Missing EOF")
	}

	for i := 0; i < len(curr.tokens); i++ {
		if curr.tokens[i].Type == EOFTT {
			return ParseRes{
				ok:     true,
				tokens: curr.tokens[i:],
			}
		}
	}

	return fail("Missing EOF")
}

var assignOpMap map[TokenType]NodeType = map[TokenType]NodeType{
	MinusEqualTT:  SubtNT,
	PlusEqualTT:   AddNT,
	SlashEqualTT:  DivNT,
	StarEqualTT:   MultNT,
	ModuloEqualTT: ModuloNT,
	BarEqualTT:    FallbackNT,
}

func pAssignOperator(tt TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if len(curr.tokens) == 0 {
			return fail("Tokens exhausted")
		}
		if curr.tokens[0].Type == tt {
			if tt == EqualTT {
				return ParseRes{
					ok:     true,
					node:   &Node{Type: AssignmentNT},
					tokens: curr.tokens[1:],
				}
			}

			nt, ok := assignOpMap[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok: true,
				node: &Node{
					Type: AssignmentNT,
					R: &Node{
						Type: nt,
					},
				},
				tokens: curr.tokens[1:],
			}
		}
		return fail("No match")
	}
}

var emptyParser Parser = func(res ParseRes, _ Nodify) ParseRes {
	return res
}
