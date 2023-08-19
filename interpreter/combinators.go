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

// Maybe = A?
var Maybe func(Parser) Parser = func(p Parser) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		res := p(curr, nil)
		if res.ok {
			return res
		}

		return ParseRes{
			ok:     true,
			tokens: curr.tokens,
		}
	}
}

var Wrapped func(TokenType, Parser, TokenType) Parser = func(left TokenType, target Parser, right TokenType) Parser {
	return func(curr ParseRes, _ Nodify) ParseRes {
		if len(curr.tokens) == 0 || curr.tokens[0].Type != left {
			return ParseRes{
				ok:     false,
				tokens: curr.tokens,
			}
		}

		tokens := curr.tokens[1:]
		for len(tokens) > 0 && tokens[0].Type == NewLineTT {
			tokens = tokens[1:]
		}

		res := target(ParseRes{
			ok:     true,
			tokens: tokens,
		}, nil)
		if !res.ok {
			return curr
		}

		tokens = res.tokens
		for len(tokens) > 0 && tokens[0].Type == NewLineTT {
			tokens = tokens[1:]
		}

		if len(tokens) == 0 || tokens[0].Type != right {
			return ParseRes{
				ok:     false,
				tokens: curr.tokens,
			}
		}

		return ParseRes{
			ok:     true,
			tokens: tokens[1:],
			node:   res.node,
		}
	}
}
