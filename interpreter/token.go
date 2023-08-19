package interpreter

import "fmt"

// Token ...
type Token struct {
	Type   TokenType
	Line   int
	Lexeme string
}

// TokenType ...
type TokenType uint8

// TokenType values
const (
	// single character
	LeftParenTT TokenType = iota
	RightParenTT
	LeftBraceTT
	RightBraceTT
	LeftBracketTT
	RightBracketTT
	ColonTT
	CommaTT
	DotTT
	MinusTT
	PlusTT
	SemicolonTT
	NewLineTT
	SlashTT
	StarTT
	ModuloTT
	QuestionMarkTT
	BarTT
	HashTT
	CaratTT

	// 1-2 characters
	ArrowTT
	LeftArrowTT
	BangTT
	BangEqualTT
	DotDotTT
	EqualTT
	EqualEqualTT
	GreaterTT
	GreaterEqualTT
	LessTT
	LessEqualTT
	ColonEqualTT
	MinusEqualTT
	PlusEqualTT
	SlashEqualTT
	StarEqualTT
	ModuloEqualTT
	BarEqualTT
	PipeTT

	// Literals
	IdentifierTT
	StringTT
	IntTT
	FloatTT
	CharTT

	// Keywords
	AndTT
	BreakTT
	ContinueTT
	ElseTT
	FalseTT
	ForTT
	IfTT
	UnlessTT
	NullTT
	OrTT
	ReturnTT
	TrueTT
	WhileTT
	UntilTT
	FailTT
	SuccessTT
	MapTT
	WhereTT
	InTT
	VarTT
	UnderscoreTT
	IndexTT

	ImportTT
	AsTT

	CommentTT

	EOFTT
)

var tokenDescriptors map[TokenType]string = map[TokenType]string{
	LeftParenTT:    "(",
	RightParenTT:   ")",
	LeftBraceTT:    "{",
	RightBraceTT:   "}",
	LeftBracketTT:  "[",
	RightBracketTT: "]",
	ColonTT:        ":",
	CommaTT:        ",",
	DotTT:          ".",
	MinusTT:        "-",
	PlusTT:         "+",
	SemicolonTT:    ";",
	NewLineTT:      "new line",
	SlashTT:        "/",
	StarTT:         "*",
	ModuloTT:       "%",
	ArrowTT:        "=>",
	BangTT:         "!",
	BangEqualTT:    "!=",
	DotDotTT:       "..",
	EqualTT:        "=",
	EqualEqualTT:   "==",
	GreaterTT:      ">",
	GreaterEqualTT: ">=",
	LessTT:         "<",
	LessEqualTT:    "<=",
	ColonEqualTT:   ":=",
	MinusEqualTT:   "-=",
	PlusEqualTT:    "+=",
	SlashEqualTT:   "/=",
	StarEqualTT:    "*=",
	ModuloEqualTT:  "%=",
	BarEqualTT:     "|=",
	IdentifierTT:   "identifier",
	StringTT:       "string literal",
	IntTT:          "integer literal",
	FloatTT:        "float literal",
	AndTT:          "and",
	ElseTT:         "else",
	FalseTT:        "false",
	ForTT:          "for",
	IfTT:           "if",
	NullTT:         "null",
	OrTT:           "or",
	ReturnTT:       "return",
	TrueTT:         "true",
	WhileTT:        "while",
	CommentTT:      "comment",
	EOFTT:          "EOF",
	QuestionMarkTT: "?",
	BarTT:          "|",
	PipeTT:         "|>",
	UnlessTT:       "unless",
	UntilTT:        "until",
	FailTT:         "fail",
	SuccessTT:      "success",
	MapTT:          "map",
	WhereTT:        "where",
	CharTT:         "char",
	InTT:           "in",
	VarTT:          "var",
	IndexTT:        "index",
	ImportTT:       "import",
	AsTT:           "as",
}

// ToString returns a string representation of a token in the form <Line#: Type "Lexeme">
func (t Token) ToString() string {
	return fmt.Sprintf("%d: %s \"%s\"", t.Line, tokenDescriptors[t.Type], t.Lexeme)
}

func (tt TokenType) ToString() string {
	return tokenDescriptors[tt]
}
