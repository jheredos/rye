package interpreter

import "fmt"

// Primaries and atoms
var pPrimary, pPrimaryRhs, pAtom, pCollection, pIdentifier, pCall, pGroup Parser
var pList, pListItem, pListItems, pSplatExpr, pEmptyList, pObject, pObjectItems, pObjectItem, pKVPair, pSet, pSetItem, pSetItems Parser
var pArgs, pCallRhs, pBracketAccess, pListSlice, pSlice, pFieldAccess Parser

// Unary expressions (and power)
var pUnPostOp, pPowerRhs, pUnPreOp Parser
var pUnaryPre, pPower, pUnaryPost Parser

// Binary expressions
var pRange, pRangeRhs, pRangeEnd Parser

// Arithmetic
var pTermOp, pSumOp, pComparisonOp, pEqualityOp Parser
var pTerm, pTermRhs, pSum, pSumRhs, pComparison, pComparisonRhs, pEquality, pEqualityRhs Parser

// Logical
var pConjunction, pConjunctionRhs, pDisjunction, pDisjunctionRhs, pInExpr, pInExprRhs, pFallback, pFallbackRhs Parser

// Conditional
var pCondExpr, pCondElseExpr, pCondRhs, pIfRhs, pUnlessRhs, pElseRhs Parser

// Match
// var pMatchExpr Parser

// Lambdas
var pLambda, pLambdaRhs, pEmptyParams, pParams, pParam Parser
var pListDestruc, pObjDestruc, pObjPairDestruc Parser

// Simple expressions
var pExpr, pSimpleExpr Parser

// Compound expressions
var pCompoundExpr, pCompoundExprRhs, pMapExprRhs, pWhereExprRhs, pPipeExprRhs, pFindExprRhs, pCompoundExprArg Parser

// Statements
var pCompoundStmt, pSimpleStmt, pStmtBody, pStmt, pStmts Parser
var pIfStmt, pUnlessStmt, pElseStmt, pCondStmt Parser
var pWhileStmt, pUntilStmt, pForStmt, pForAssign, pLoopStmt Parser

// Simple statements
var pVarDecl, pConstDecl, pDeclTarget, pDeclRhs, pAssignment, pAssignTarget, pAssignRhs, pAssignOp, pDecl Parser
var pImportStmt, pReturnStmt Parser
var pProgram Parser

func init() {

	// Simple expressions
	// Primaries and atoms
	pIdentifier = pToken(IdentifierTT, nAtom(IdentifierNT))
	// This nonsense deals with circular dependencies. Passing the Parser itself, before defining, will pass nil
	pGroup = InParens(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) })

	// Collections
	// Lists
	pSplatExpr = Then(pOperatorUnary(DotDotDotTT), func(r ParseRes, n Nodify) ParseRes { return pPrimary(r, n) }, nUnaryPre)
	pListItem = Choice(pSplatExpr, func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) })
	pListItems = ThenMaybe(
		listify(pListItem),
		Plus(
			Then(
				pToken(CommaTT, nil),
				(pListItem),
				takeSecond,
			), nListTail),
		nListHead,
	)
	pEmptyList = Then(
		pToken(LeftBracketTT, nil),
		(pToken(RightBracketTT, nil)),
		nEmptyList,
	)
	pList = Choice(
		pEmptyList,
		InBrackets(pListItems),
	)

	// Objects
	pKVPair = Then(
		Choice(
			pIdentifier,
			pToken(StringTT, nAtom(StringNT)),
			pGroup,
		),
		Then(
			pToken(ColonTT, nil),
			func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) },
			takeSecond,
		),
		nKVPair,
	)
	pObjectItem = nestLeft(Choice(pSplatExpr, pKVPair), ObjectItemNT)
	pObjectItems = CommaSeparated(pObjectItem)
	pObject = Choice(
		Then(pToken(LeftBraceTT, nil), (pToken(RightBraceTT, nil)), nObject),
		InBraces(pObjectItems),
	)

	// Set
	pSetItem = Choice(pSplatExpr, func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) })
	pSetItems = CommaSeparated(nestLeft(pSetItem, SetItemNT))
	pSet = InBraces(pSetItems)

	pAtom = Choice(
		pIdentifier,
		pToken(TrueTT, nAtom(BoolNT)),
		pToken(FalseTT, nAtom(BoolNT)),
		pToken(NullTT, nAtom(NullNT)),
		pToken(FailTT, nAtom(FailNT)),
		pToken(SuccessTT, nAtom(SuccessNT)),
		pToken(StringTT, nAtom(StringNT)),
		pToken(IntTT, nAtom(IntNT)),
		pToken(FloatTT, nAtom(FloatNT)),
		pToken(UnderscoreTT, nAtom(UnderscoreNT)),
		pToken(IndexTT, nAtom(IndexNT)),
		// pTuple,
		pGroup,
	)

	pCollection = Choice(
		pList,
		pObject,
		pSet,
	)

	pArgs = Then(
		// KVPairs for named params?
		CommaSeparated(nestLeft(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) }, ArgNT)),
		pToken(RightParenTT, nil),
		takeFirst,
	)
	pCallRhs = nestRight(Then(
		pToken(LeftParenTT, nil),
		Choice(nestLeft(pToken(RightParenTT, nil), ArgNT), pArgs),
		takeSecond,
	), CallNT)

	pSlice = Choice(
		Then(
			func(r ParseRes, n Nodify) ParseRes { return pUnaryPre(r, n) },
			ThenMaybe(
				pToken(DotDotTT, nil),
				func(r ParseRes, n Nodify) ParseRes { return pUnaryPre(r, n) },
				takeSecond,
			),
			nSlice,
		),
		ThenMaybe(
			pToken(DotDotTT, nSlice),
			func(r ParseRes, n Nodify) ParseRes { return pUnaryPre(r, n) },
			nRhs,
		),
	)
	pListSlice = nestRight(InBrackets(pSlice), ListSliceNT)
	pBracketAccess = nestRight(
		InBrackets(func(r ParseRes, n Nodify) ParseRes { return pSimpleExpr(r, n) }),
		BracketAccessNT)

	pFieldAccess = nestRight(Then(
		pToken(DotTT, nil),
		Choice(pIdentifier, pToken(UnderscoreTT, nAtom(UnderscoreNT))),
		takeSecond,
	), FieldAccessNT)

	pPrimaryRhs = Plus(Choice(pCallRhs, pListSlice, pBracketAccess, pFieldAccess), nLeftAssoc)

	pPrimary = ThenMaybe(
		Choice(pAtom, pCollection),
		pPrimaryRhs,
		nEndLeftAssoc,
	) // function call, list access, object access, etc.

	// Unary expressions
	pUnPostOp = Choice(pOperatorUnary(QuestionMarkTT))
	pUnaryPost = ThenMaybe(pPrimary, pUnPostOp, nUnaryPost)
	pPowerRhs = Plus(
		Then(pOperator(CaratTT),
			// circular dependency
			func(r ParseRes, n Nodify) ParseRes { return pUnaryPre(r, n) },
			nRhs), nRightAssoc)
	pPower = ThenMaybe(pUnaryPost, pPowerRhs, nBinary)
	pUnPreOp = Choice(pOperatorUnary(BangTT), pOperatorUnary(MinusTT), pOperatorUnary(HashTT))
	pUnaryPre = Choice(Then(Plus(pUnPreOp, nUnaryNested), pPower, nUnaryNested), pPower)

	// Binary expressions
	// Range
	pRangeRhs = Then(pOperator(DotDotTT), pUnaryPre, takeSecond)
	pRangeEnd = Then(pOperator(DotDotTT), pUnaryPre, nRangeEnd)
	pRange = Choice(
		pRangeEnd,
		ThenMaybe(
			pUnaryPre,
			Then(pOperator(DotDotTT), pUnaryPre, takeSecond),
			nRange,
		),
	)

	// Arithmetic expressions
	pTermOp = Choice(pOperator(StarTT), pOperator(SlashTT), pOperator(ModuloTT))
	pTermRhs = Plus(Then(pTermOp, pUnaryPre, nRhs), nLeftAssoc)
	pTerm = Choice(
		pRangeEnd,
		Then(pUnaryPre, pTermRhs, nEndLeftAssoc),
		ThenMaybe(pUnaryPre, pRangeRhs, nRange),
	)

	pSumOp = Choice(pOperator(PlusTT), pOperator(MinusTT))
	pSumRhs = Plus(Then(pSumOp, pTerm, nRhs), nLeftAssoc)
	pSum = ThenMaybe(pTerm, pSumRhs, nEndLeftAssoc)

	pComparisonOp = Choice(pOperator(LessEqualTT), pOperator(GreaterEqualTT), pOperator(LessTT), pOperator(GreaterTT))
	pComparisonRhs = Plus(Then(pComparisonOp, pSum, nRhs), nLeftAssoc)
	pComparison = ThenMaybe(pSum, pComparisonRhs, nEndLeftAssoc)

	pEqualityOp = Choice(pOperator(EqualEqualTT), pOperator(BangEqualTT))
	pEqualityRhs = Plus(Then(pEqualityOp, pComparison, nRhs), nLeftAssoc)
	pEquality = ThenMaybe(pComparison, pEqualityRhs, nEndLeftAssoc)

	// Logical expressions
	pInExprRhs = Plus(Then(pOperator(InTT), pEquality, nRhs), nLeftAssoc)
	pInExpr = ThenMaybe(pEquality, pInExprRhs, nEndLeftAssoc)

	pConjunctionRhs = Plus(Then(pOperator(AndTT), pInExpr, nRhs), nLeftAssoc)
	pConjunction = ThenMaybe(pInExpr, pConjunctionRhs, nEndLeftAssoc)

	pDisjunctionRhs = Plus(Then(pOperator(OrTT), pConjunction, nRhs), nLeftAssoc)
	pDisjunction = ThenMaybe(pConjunction, pDisjunctionRhs, nEndLeftAssoc)

	pFallbackRhs = Plus(Then(pOperator(BarTT), pDisjunction, nRhs), nLeftAssoc)
	pFallback = ThenMaybe(pDisjunction, pFallbackRhs, nEndLeftAssoc)

	// Conditional expressions
	pElseRhs = Then(pToken(ElseTT, nil), func(r ParseRes, n Nodify) ParseRes { return pCondElseExpr(r, n) }, takeSecond)
	pUnlessRhs = Then(pOperator(UnlessTT), pFallback, negateSecond(nLhs))
	pIfRhs = Then(pOperator(IfTT), pFallback, nLhs)
	pCondRhs = ThenNot(
		Choice(pIfRhs, pUnlessRhs),
		Choice(pToken(ColonTT, nil), pToken(LeftBraceTT, nil)))
	pCondExpr = ThenMaybe(pFallback, pCondRhs, nBinaryFlip)
	pCondElseExpr = ThenMaybe(pCondExpr, pElseRhs, nElse)

	// Lambdas
	pListDestruc = InBrackets(
		ThenMaybe(
			listify(pIdentifier),
			Plus(
				Then(
					pToken(CommaTT, nil),
					pIdentifier,
					takeSecond,
				), nListTail),
			nListHead,
		))
	pObjPairDestruc = nestLeft(ThenMaybe(
		pIdentifier,
		Then(
			pToken(ColonTT, nil),
			pIdentifier, // pParam
			takeSecond,
		),
		nKVPair,
	), ObjectItemNT)
	pObjDestruc = InBraces(CommaSeparated(pObjPairDestruc))

	pParam = Choice(pToken(IdentifierTT, nParam), nestLeft(pListDestruc, ParamNT), nestLeft(pObjDestruc, ParamNT))
	pParams =
		Choice(
			// single identifier: x => ...
			pToken(IdentifierTT, nParam),
			// empty params: () => ...
			Then(
				pToken(LeftParenTT, nil),
				pToken(RightParenTT, nil),
				nAlways(ParamNT)),
			// comma-separated params: (x,y) => ...
			InParens(CommaSeparated(pParam)),
		)
	pLambdaRhs = Then((pOperator(ArrowTT)),
		Choice(
			(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) }),
			InBraces(func(r ParseRes, n Nodify) ParseRes { return pStmts(r, n) }),
		),
		nRhs)
	pLambda = Then(pParams, pLambdaRhs, nBinary)

	pSimpleExpr = Choice(pLambda, pCondElseExpr)

	// Compound expressions
	pCompoundExprArg = Choice(pLambda, maybeFunc(pCondElseExpr))
	pPipeExprRhs = Then(pOperator(PipeTT), pCompoundExprArg, nRhs)
	pWhereExprRhs = Then(pOperator(WhereTT), pCompoundExprArg, nRhs)
	pMapExprRhs = Then(pOperator(MapTT), pCompoundExprArg, nRhs)
	pFindExprRhs = Then(pOperator(FindTT), pCompoundExprArg, nRhs)
	pCompoundExprRhs = Plus((Choice(pPipeExprRhs, pWhereExprRhs, pMapExprRhs, pFindExprRhs)), nLeftAssoc)
	pCompoundExpr = ThenMaybe(pSimpleExpr, pCompoundExprRhs, nEndLeftAssoc)

	pExpr = Choice(pCompoundExpr)

	// Statements

	// Assignment and declaration
	pAssignOp = Choice(
		pAssignOperator(EqualTT),
		pAssignOperator(PlusEqualTT),
		pAssignOperator(MinusEqualTT),
		pAssignOperator(StarEqualTT),
		pAssignOperator(SlashEqualTT),
		pAssignOperator(ModuloEqualTT),
		pAssignOperator(BarEqualTT),
	)
	pAssignRhs = Then(pAssignOp, pExpr, nAssignmentRhs)
	pAssignTarget = ThenMaybe(
		pIdentifier,
		Plus(Choice(pBracketAccess, pFieldAccess), nLeftAssoc),
		nEndLeftAssoc,
	)
	pAssignment = Then(pAssignTarget, pAssignRhs, nAssignment)

	pDeclRhs = Then(pOperator(ColonEqualTT), maybeFunc(pExpr), nRhs)
	pDeclTarget = Choice(
		pListDestruc,
		pIdentifier,
		pObjDestruc)
	pConstDecl = Then(pDeclTarget, pDeclRhs, nBinary)
	pVarDecl = Then(Then(pToken(VarTT, nil), pDeclTarget, takeSecond), alterNodeType(pDeclRhs, VarDeclNT), nBinary)
	pDecl = Choice(pVarDecl, pConstDecl)

	pReturnStmt = nestRight(Then(pToken(ReturnTT, nil), pExpr, takeSecond), ReturnStmtNT)
	pImportStmt = ThenMaybe(
		Then(pToken(ImportTT, nil), pToken(StringTT, nAtom(StringNT)), nImport),
		Then(pToken(AsTT, nil), pIdentifier, takeSecond),
		nRhs,
	)

	pSimpleStmt = Choice(pReturnStmt, pOperator(BreakTT), pOperator(ContinueTT), pDecl, pAssignment)

	pStmtBody = Choice(
		Then(
			pToken(ColonTT, nil),
			func(r ParseRes, n Nodify) ParseRes { return pStmt(r, n) },
			takeSecond),
		InBraces(func(r ParseRes, n Nodify) ParseRes { return pStmts(r, n) }),
	)

	// Conditional statements
	pElseStmt = Then(pToken(ElseTT, nil), Choice(
		func(r ParseRes, n Nodify) ParseRes { return pCondStmt(r, n) },
		pStmtBody,
	), takeSecond)
	pIfStmt = Then(
		Then(pOperator(IfTT), pExpr, nLhs),
		pStmtBody, nRhs)
	pUnlessStmt = Then(
		Then(pOperator(UnlessTT), pExpr, negateSecond(nLhs)),
		pStmtBody, nRhs)
	pCondStmt = ThenMaybe(Choice(pIfStmt, pUnlessStmt), pElseStmt, nElse)

	// Loop statements
	pWhileStmt = Then(Then(pOperator(WhileTT), pExpr, nLhs), pStmtBody, nRhs)
	pUntilStmt = Then(Then(pOperator(UntilTT), pExpr, negateSecond(nLhs)), pStmtBody, nRhs)
	pForAssign = alterNodeType(Then(pDeclTarget, Then(pOperator(InTT), pExpr, nRhs), nBinary), ConstDeclNT)
	pForStmt = Then(Then(pOperator(ForTT), pForAssign, nLhs), pStmtBody, nRhs)
	pLoopStmt = Choice(pWhileStmt, pUntilStmt, pForStmt)

	pCompoundStmt = Choice(pCondStmt, pLoopStmt)

	pStmt = nestLeft(
		Then(
			Choice(pImportStmt, pCompoundStmt, pSimpleStmt, pExpr),
			Choice(
				Peek(pToken(NewLineTT, nil)),
				Peek(pToken(RightBraceTT, nil)),
				pToken(SemicolonTT, nil)),
			takeFirst),
		StmtNT)
	pStmts = Plus(pStmt, nLinked)

	pProgram = Then(
		pStmts,
		pToken(EOFTT, nil),
		takeFirst,
	)
}

// Parse parses a slice of tokens
func Parse(ts []Token) (*Node, error) {
	start := ParseRes{
		ok:     true,
		tokens: ts,
	}
	res := pProgram(start, nil)

	if !res.ok {
		return nil, fmt.Errorf(res.err)
	}

	// fmt.Printf("AST:	%s\n", res.node.ToString())

	return res.node, nil
}
