package interpreter

import (
	"fmt"
	"strconv"
)

// Scan ...
func Scan(src string) []Token {
	tokens := make([]Token, 0)
	return scan(tokens, src, 1)
}

func scan(scanned []Token, remaining string, line int) []Token {
	if len(remaining) == 0 {
		scanned = append(scanned, Token{NewLineTT, line, ""})
		return append(scanned, Token{EOFTT, line, "\x00"})
	}

	r := remaining[0]
	switch r {
	// whitespace
	case '\n':
		scanned = append(scanned, Token{NewLineTT, line, ""})
		return scan(scanned, remaining[1:], line+1)
	case '\t', '\r', ' ':
		return scan(scanned, remaining[1:], line)

	// 1 character
	case '(', ')', '{', '}', '[', ']', ';', ',', '?', '^', '#', '_':
		if tt, ok := scanOneRune(r); ok {
			if tt == RightBraceTT {
				scanned = append(scanned, Token{NewLineTT, line, ""}) // insert newline at end of block
			}
			scanned = append(scanned, Token{tt, line, string(r)})
			return scan(scanned, remaining[1:], line)
		}
		fmt.Printf("Scanning error on line %d: Unexpected character \"%s\"\n", line, string(r))
		return nil

	// 1-2 characters
	case '!', '=', '>', '<', ':', '-', '+', '/', '*', '%', '|':
		if tt, ok := scanTwoRune(r, remaining[1]); ok {
			if tt == CommentTT {
				remaining = scanComment(remaining)
				return scan(scanned, remaining, line)
			}
			scanned = append(scanned, Token{tt, line, string(r) + string(remaining[1])})
			return scan(scanned, remaining[2:], line)
		} else if tt, ok = scanOneRune(r); ok {
			scanned = append(scanned, Token{tt, line, string(r)})
			return scan(scanned, remaining[1:], line)
		} else {
			fmt.Printf("Scanning error on line %d: Unexpected character \"%s\"\n", line, string(r))
			return nil
		}
	case '.':
		if len(remaining) > 1 {
			n := remaining[1]
			if n == '.' {
				// ...
				if len(remaining) > 2 && remaining[2] == '.' {
					scanned = append(scanned, Token{DotDotDotTT, line, "..."})
					return scan(scanned, remaining[3:], line)
				}
				// ..
				scanned = append(scanned, Token{DotDotTT, line, string(r) + string(n)})
				return scan(scanned, remaining[2:], line)
			} else if isDigit(n) {
				// float
				ds, remaining := scanDigits(remaining[1:])
				scanned = append(scanned, Token{FloatTT, line, "." + ds})
				return scan(scanned, remaining, line)
			} else {
				// .
				scanned = append(scanned, Token{DotTT, line, string(r)})
				return scan(scanned, remaining[1:], line)
			}
		}

	// string
	case '"':
		t, remaining, ln := scanString(remaining, line)
		if ln == -1 {
			fmt.Printf("Scanning error: Unterminated string starting on line %d\n", line)
			return nil
		}
		scanned = append(scanned, t)
		return scan(scanned, remaining[1:], ln)
	default:
		// numbers
		if isDigit(r) {
			n, remaining := scanDigits(remaining)
			// check if float
			if len(remaining) > 0 && remaining[0] == '.' {
				// check range operator
				if len(remaining) > 1 && remaining[1] == '.' {
					scanned = append(scanned, Token{IntTT, line, n})
					return scan(scanned, remaining, line)
				}
				m, remaining := scanDigits(remaining[1:])
				n += "." + m
				scanned = append(scanned, Token{FloatTT, line, n})
				return scan(scanned, remaining, line)
			}
			scanned = append(scanned, Token{IntTT, line, n})
			return scan(scanned, remaining, line)
		}
		// identifiers
		if isAlpha(r) {
			s, remaining := scanIdentifier(remaining)
			if tt, ok := scanKeyword(s); ok {
				scanned = append(scanned, Token{tt, line, s})
				return scan(scanned, remaining, line)
			}
			scanned = append(scanned, Token{IdentifierTT, line, s})
			return scan(scanned, remaining, line)
		}
		// error
		fmt.Printf("Scanning error: Unexpected character \"%s\" on line %d\n", string(r), line)
		return nil
	}

	return nil
}

func scanTwoRune(a byte, b byte) (TokenType, bool) {
	twoRunes := map[string]TokenType{
		"=>":  ArrowTT,
		"<-":  LeftArrowTT,
		"!=":  BangEqualTT,
		"==":  EqualEqualTT,
		">=":  GreaterEqualTT,
		"<=":  LessEqualTT,
		":=":  ColonEqualTT,
		"-=":  MinusEqualTT,
		"+=":  PlusEqualTT,
		"/=":  SlashEqualTT,
		"*=":  StarEqualTT,
		"%=":  ModuloEqualTT,
		"..":  DotDotTT,
		"...": DotDotDotTT,
		"//":  CommentTT,
		"|=":  BarEqualTT,
		"|>":  PipeTT,
	}
	tt, ok := twoRunes[string(a)+string(b)]
	return tt, ok
}

func scanOneRune(r byte) (TokenType, bool) {
	oneRune := map[byte]TokenType{
		'(': LeftParenTT,
		')': RightParenTT,
		'{': LeftBraceTT,
		'}': RightBraceTT,
		'[': LeftBracketTT,
		']': RightBracketTT,
		':': ColonTT,
		',': CommaTT,
		'.': DotTT,
		'-': MinusTT,
		'+': PlusTT,
		';': SemicolonTT,
		'/': SlashTT,
		'*': StarTT,
		'%': ModuloTT,
		'!': BangTT,
		'=': EqualTT,
		'>': GreaterTT,
		'<': LessTT,
		'?': QuestionMarkTT,
		'|': BarTT,
		'#': HashTT,
		'^': CaratTT,
		'_': UnderscoreTT,
	}
	tt, ok := oneRune[r]
	return tt, ok
}

func scanDigits(rem string) (string, string) {
	for i := 0; i < len(rem); i++ {
		if !isDigit(rem[i]) {
			return rem[:i], rem[i:]
		}
	}
	return rem, ""
}

func scanIdentifier(rem string) (string, string) {
	for i := 0; true; i++ {
		if !isAlphaNumeric(rem[i]) {
			return rem[:i], rem[i:]
		}
		if i == len(rem)-1 {
			return rem, ""
		}
	}
	return "", ""
}

func scanKeyword(s string) (TokenType, bool) {
	keywords := map[string]TokenType{
		"and":      AndTT,
		"break":    BreakTT,
		"continue": ContinueTT,
		"else":     ElseTT,
		"false":    FalseTT,
		"for":      ForTT,
		"if":       IfTT,
		"null":     NullTT,
		"or":       OrTT,
		"return":   ReturnTT,
		"true":     TrueTT,
		"while":    WhileTT,
		"until":    UntilTT,
		"unless":   UnlessTT,
		"fail":     FailTT,
		"success":  SuccessTT,
		"map":      MapTT,
		"where":    WhereTT,
		"in":       InTT,
		"var":      VarTT,
		"_":        UnderscoreTT,
		"index":    IndexTT,
		"import":   ImportTT,
		"as":       AsTT,
		"then":     PipeTT,
		"find":     FindTT,
		"fold":     FoldTT,
		"bind":     PipeTT, //BindTT,
		"each":     MapTT,
	}
	tt, ok := keywords[s]
	return tt, ok
}

func scanComment(rem string) string {
	for i := 0; i < len(rem); i++ {
		if rem[i] == '\n' {
			return rem[i:]
		}
	}
	return ""
}

func scanString(rem string, line int) (Token, string, int) {
	for i := 1; i < len(rem); i++ {
		if rem[i] == '\n' {
			line++
		}
		if rem[i] == '\\' {
			i++
		} else if rem[i] == '"' {
			val, _ := strconv.Unquote(fmt.Sprintf(`"%s"`, rem[1:i]))
			return Token{StringTT, line, val}, rem[i:], line
		}
	}
	return Token{}, "", -1
}

func isAlpha(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isDigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func isAlphaNumeric(r byte) bool {
	return isAlpha(r) || isDigit(r)
}
