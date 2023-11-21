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

// trim skips all new lines before attempting a parser
func trim(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return p(curr, nil)
		}

		// for len(curr.tokens) > 0 && curr.tokens[0].Type == NewLineTT {
		// 	curr.tokens = curr.tokens[1:]
		// }

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
	FindTT:         FindNT,
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
	DotDotDotTT:    SplatNT,
}

// pOperator creates a parser for a binary operator, finding the appropriate node based on a token type
func pOperator(tt TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		tokens := curr.tokens
		if tt != NewLineTT {
			for len(tokens) > 0 && tokens[0].Type == NewLineTT {
				tokens = tokens[1:]
			}
		}
		if len(tokens) == 0 {
			return fail("Tokens exhausted")
		}

		if tokens[0].Type == tt {
			op, ok := operatorMap[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok:     true,
				node:   &Node{Type: op},
				tokens: tokens[1:],
			}
		}
		return fail("No match")
	}
}

func pOperatorUnary(tt TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		tokens := curr.tokens
		if tt != NewLineTT {
			for len(tokens) > 0 && tokens[0].Type == NewLineTT {
				tokens = tokens[1:]
			}
		}
		if len(tokens) == 0 {
			return fail("Tokens exhausted")
		}

		if tokens[0].Type == tt {
			op, ok := map[TokenType]NodeType{
				MinusTT:        UnaryNegNT,
				BangTT:         LogicNotNT,
				HashTT:         CardinalityNT,
				QuestionMarkTT: MaybeNT,
				DotDotDotTT:    SplatNT,
			}[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok:     true,
				node:   &Node{Type: op},
				tokens: tokens[1:],
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

		tokens := curr.tokens
		if tt != NewLineTT {
			for len(tokens) > 0 && tokens[0].Type == NewLineTT {
				tokens = tokens[1:]
			}
		}

		if len(tokens) == 0 {
			return fail("Tokens exhausted")
		}
		if tokens[0].Type == tt {
			res := ParseRes{
				ok:     true,
				parsed: &tokens[0],
				tokens: tokens[1:],
			}
			if n != nil {
				res.node = n(res)
			}

			return res
		}
		return ParseRes{
			ok: false,
			err: fmt.Sprintf(
				"Line %d: Parsing error. Expected %s, received %s \"%s\"",
				tokens[0].Line,
				tt.ToString(),
				tokens[0].Type.ToString(),
				tokens[0].Lexeme),
			tokens: curr.tokens,
		}
	}
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
		if !curr.ok {
			return curr
		}

		tokens := curr.tokens
		for len(tokens) > 0 && tokens[0].Type == NewLineTT {
			tokens = tokens[1:]
		}

		if len(tokens) == 0 {
			return fail("Tokens exhausted")
		}

		if tokens[0].Type == tt {
			if tt == EqualTT {
				return ParseRes{
					ok:     true,
					node:   &Node{Type: AssignmentNT},
					tokens: tokens[1:],
				}
			}

			nt, ok := assignOpMap[tt]
			if !ok {
				return fail("Unknown operator")
			}
			return ParseRes{
				ok: true,
				node: &Node{
					Type: AugAssignNT,
					R: &Node{
						Type: nt,
					},
				},
				tokens: tokens[1:],
			}
		}
		return fail("No match")
	}
}
