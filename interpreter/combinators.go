package interpreter

// Then is a combinator corresponding to (A B) in a grammar. It takes a Nodify function with two
// arguments: result A and result B
var Then func(Parser, Parser, Nodify) Parser = func(a Parser, b Parser, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		resA := a(curr, nil)
		if resA.ok {
			resB := b(resA, nil)
			if resB.ok {
				if n != nil {
					resB.node = n(resA, resB)
				}
				return resB
			}

			return ParseRes{
				ok:     false,
				err:    resB.err,
				tokens: curr.tokens,
			}
		}

		return ParseRes{
			ok:     false,
			err:    resA.err,
			tokens: curr.tokens,
		}
	}
}

var ThenNot func(Parser, Parser) Parser = func(a Parser, b Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		resA := a(curr, nil)
		if resA.ok {
			resB := b(resA, nil)
			if resB.ok {
				return fail("ThenNot failed")
			}

			return resA
		}

		return ParseRes{
			ok:     false,
			err:    resA.err,
			tokens: curr.tokens,
		}
	}
}

// Then is a combinator corresponding to (A B) in a grammar. It takes a Nodify function with two
// arguments: result A and result B
var ThenMaybe func(Parser, Parser, Nodify) Parser = func(a Parser, b Parser, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		resA := a(curr, nil)
		if resA.ok {
			resB := b(resA, nil)
			if resB.ok {
				if n != nil {
					resB.node = n(resA, resB)
				}
				return resB
			}
			return resA
		}

		return ParseRes{
			ok:     false,
			err:    resA.err,
			tokens: curr.tokens,
		}
	}
}

var ThenPeek func(Parser, Parser, Nodify) Parser = func(a Parser, b Parser, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		resA := a(curr, nil)
		if resA.ok {
			resB := b(resA, nil)
			if resB.ok {
				return resA
			}
		}

		return ParseRes{
			ok:     false,
			err:    resA.err,
			tokens: curr.tokens,
		}
	}
}

// Either is a combinator aligning with (A | B) in a grammar
var Either func(Parser, Parser) Parser = func(a Parser, b Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		resA := a(curr, nil)
		if resA.ok {
			return resA
		}

		resB := b(curr, nil)
		if resB.ok {
			return resB
		}

		return ParseRes{
			ok:     false,
			tokens: curr.tokens,
		}
	}
}

// Choice = (A | B | C ...)
var Choice func(...Parser) Parser = func(ps ...Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		for _, p := range ps {
			res := p(curr, nil)
			if res.ok {
				return res
			}
		}

		return ParseRes{
			ok:     false,
			tokens: curr.tokens,
		}
	}
}

// Star = A*
var Star func(Parser, Nodify) Parser = func(p Parser, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		prev := curr
		res := p(curr, n)
		for res.ok {
			res = p(res, nil)
			if res.ok {
				res.node = n(prev, res)
				prev = res
			}
		}

		return ParseRes{
			ok:     true,
			node:   prev.node,
			tokens: prev.tokens,
		}
	}
}

// Plus = A+
var Plus func(Parser, Nodify) Parser = func(p Parser, n Nodify) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if !res.ok {
			return ParseRes{
				ok:     false,
				tokens: curr.tokens,
			}
		}

		prev := res
		for res.ok {
			res = p(res, nil)
			if res.ok {
				res.node = n(prev, res)
				prev = res
			}
		}

		return prev
	}
}

var Wrapped func(TokenType, Parser, TokenType) Parser = func(left TokenType, target Parser, right TokenType) Parser {
	return Then(
		pToken(left, nil),
		Then(
			target,
			pToken(right, nil),
			takeFirst,
		),
		takeSecond,
	)
}

var InBraces func(Parser) Parser = func(p Parser) Parser {
	return Wrapped(LeftBraceTT, p, RightBraceTT)
}

var InBrackets func(Parser) Parser = func(p Parser) Parser {
	return Wrapped(LeftBracketTT, p, RightBracketTT)
}

var InParens func(Parser) Parser = func(p Parser) Parser {
	return Wrapped(LeftParenTT, p, RightParenTT)
}

var CommaSeparated func(Parser) Parser = func(p Parser) Parser {
	return ThenMaybe(
		p,
		Plus(
			Then(
				pToken(CommaTT, nil),
				p,
				takeSecond,
			),
			nLinked,
		),
		nRhs,
	)
}

var Peek func(Parser) Parser = func(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if !curr.ok {
			return curr
		}

		resA := p(curr, nil)
		if resA.ok {
			return ParseRes{
				ok:     true,
				tokens: curr.tokens,
			}
		}

		return ParseRes{
			ok:     false,
			err:    resA.err,
			tokens: curr.tokens,
		}
	}
}
