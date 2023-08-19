package interpreter

import "fmt"

// Primaries and atoms
var pPrimary, pPrimaryRhs, pAtom, pCall, pGroup Parser
var pList, pListItems, pEmptyList, pObject, pObjectItems, pKVPair, pSet, pSetItems Parser
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

// Lambdas
var pLambda, pLambdaRhs, pEmptyParams, pParenParams, pParams, pParam, pSingleParam, pParamsRhs Parser
var pParamDestruc, pListDestruc, pObjDestruc, pObjPairDestruc Parser

// Simple expressions
var pExpr, pSimpleExpr Parser

// Compound expressions
var pCompoundExpr, pCompoundExprRhs, pMapExprRhs, pWhereExprRhs, pPipeExprRhs, pCompoundExprArg Parser

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
	pGroup = Then(
		Then(
			pToken(LeftParenTT, nil),
			// This nonsense deals with circular dependencies. Passing the Parser itself, before defining, will pass nil
			func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) },
			takeSecond),
		pToken(RightParenTT, nil),
		takeFirst)

	// Collections
	// Lists
	pListItems = ThenMaybe(
		listify(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) }),
		Plus(
			Then(
				pToken(CommaTT, nil),
				func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) },
				takeSecond,
			), nListTail),
		nListHead,
	)
	pEmptyList = Then(
		pToken(LeftBracketTT, nil),
		pToken(RightBracketTT, nil),
		nEmptyList,
	)
	pList = Choice(
		pEmptyList,
		Wrapped(
			LeftBracketTT,
			pListItems,
			RightBracketTT,
		),
	)

	// Objects
	pKVPair = Then(
		Choice(
			pToken(IdentifierTT, nIdentifier),
			pToken(StringTT, nString),
		),
		Then(
			pToken(ColonTT, nil),
			func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) },
			takeSecond,
		),
		nKVPair,
	)
	pObjectItems = ThenMaybe(
		skipNewLines(pKVPair),
		Plus(
			Then(
				pToken(CommaTT, nil),
				skipNewLines(pKVPair),
				takeSecond,
			),
			nLinked,
		),
		nRhs,
	)
	pObject = Choice(
		Then(pToken(LeftBraceTT, nil), skipNewLines(pToken(RightBraceTT, nil)), nObject),
		Then(
			pToken(LeftBraceTT, nil),
			Then(
				pObjectItems,
				skipNewLines(pToken(RightBraceTT, nil)),
				takeFirst,
			),
			takeSecond,
		),
	)

	// Set
	pSetItems = ThenMaybe(
		nestNode(func(r ParseRes, n Nodify) ParseRes { return pSimpleExpr(r, n) }, SetItemNT),
		Plus(
			Then(
				pToken(CommaTT, nil),
				skipNewLines(nestNode(func(r ParseRes, n Nodify) ParseRes { return pSimpleExpr(r, n) }, SetItemNT)),
				takeSecond,
			),
			nLinked,
		),
		nRhs,
	)
	pSet = Then(
		pToken(LeftBraceTT, nil),
		Then(
			pSetItems,
			skipNewLines(pToken(RightBraceTT, nil)),
			takeFirst,
		),
		takeSecond,
	)

	pAtom = Choice(
		pToken(IdentifierTT, nIdentifier),
		pToken(TrueTT, nTrue),
		pToken(FalseTT, nFalse),
		pToken(NullTT, nNull),
		pToken(FailTT, nFail),
		pToken(SuccessTT, nSuccess),
		pToken(StringTT, nString),
		pToken(IntTT, nInt),
		pToken(FloatTT, nFloat),
		pToken(UnderscoreTT, nUnderscore),
		pToken(IndexTT, nIndex),
		pList,
		pObject,
		pSet,
		// pTuple,
		pGroup,
	)

	pArgs = Then(ThenMaybe(
		nestNode(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) }, ArgNT),
		Plus(
			Then(
				pToken(CommaTT, nil),
				nestNode(func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) }, ArgNT),
				takeSecond,
			),
			nLinked,
		),
		nRhs,
	), pToken(RightParenTT, nil), takeFirst)
	pCallRhs = nestRight(Then(
		pToken(LeftParenTT, nil),
		Choice(nestNode(pToken(RightParenTT, nil), ArgNT), pArgs),
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
	pListSlice = nestRight(Then(
		pToken(LeftBracketTT, nil),
		Then(pSlice, pToken(RightBracketTT, nil), takeFirst),
		takeSecond,
	), ListSliceNT)
	pBracketAccess = nestRight(Then(
		pToken(LeftBracketTT, nil),
		Then(
			func(r ParseRes, n Nodify) ParseRes { return pSimpleExpr(r, n) },
			pToken(RightBracketTT, nil),
			takeFirst),
		takeSecond,
	), BracketAccessNT)

	pFieldAccess = nestRight(Then(
		pToken(DotTT, nil),
		Choice(pToken(IdentifierTT, nIdentifier), pToken(UnderscoreTT, nUnderscore)),
		takeSecond,
	), FieldAccessNT)

	pPrimaryRhs = Plus(Choice(pCallRhs, pListSlice, pBracketAccess, pFieldAccess), nLeftAssoc)

	pPrimary = ThenMaybe(pAtom, pPrimaryRhs, nEndLeftAssoc) // function call, list access, object access, etc.

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
	pUnaryPre = Either(Then(Plus(pUnPreOp, nUnaryNested), pPower, nUnaryNested), pPower)

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

	pSumOp = Either(pOperator(PlusTT), pOperator(MinusTT))
	pSumRhs = Plus(Then(pSumOp, pTerm, nRhs), nLeftAssoc)
	pSum = ThenMaybe(pTerm, pSumRhs, nEndLeftAssoc)

	pComparisonOp = Choice(pOperator(LessEqualTT), pOperator(GreaterEqualTT), pOperator(LessTT), pOperator(GreaterTT))
	pComparisonRhs = Plus(Then(pComparisonOp, pSum, nRhs), nLeftAssoc)
	pComparison = ThenMaybe(pSum, pComparisonRhs, nEndLeftAssoc)

	pEqualityOp = Either(pOperator(EqualEqualTT), pOperator(BangEqualTT))
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
	pUnlessRhs = Then(pOperator(UnlessTT), pFallback, invertSecond(nLhs))
	pIfRhs = Then(pOperator(IfTT), pFallback, nLhs)
	pCondRhs = Choice(pIfRhs, pUnlessRhs)
	pCondExpr = ThenMaybe(pFallback, pCondRhs, nBinaryFlip)
	pCondElseExpr = ThenMaybe(pCondExpr, pElseRhs, nElse)

	// Lambdas
	pListDestruc = Wrapped(
		LeftBracketTT,
		ThenMaybe(
			listify(pToken(IdentifierTT, nIdentifier)),
			Plus(
				Then(
					pToken(CommaTT, nil),
					pToken(IdentifierTT, nIdentifier),
					takeSecond,
				), nListTail),
			nListHead,
		),
		RightBracketTT,
	)
	pObjPairDestruc = ThenMaybe(
		pToken(IdentifierTT, nIdentifier),
		Then(
			pToken(ColonTT, nil),
			pToken(IdentifierTT, nIdentifier), // pParam
			takeSecond,
		),
		nKVPair,
	)
	pObjDestruc = Wrapped(
		LeftBraceTT,
		ThenMaybe(
			skipNewLines(pObjPairDestruc),
			Plus(
				Then(
					pToken(CommaTT, nil),
					skipNewLines(pObjPairDestruc), // pParam
					takeSecond,
				),
				nLinked),
			nRhs,
		),
		RightBraceTT,
	)

	pParam = Choice(pToken(IdentifierTT, nParam), nestNode(pListDestruc, ParamNT), nestNode(pObjDestruc, ParamNT))
	pParamsRhs = Plus(Then(pToken(CommaTT, nil), pParam, takeSecond), nLinked)
	pParams = ThenMaybe(pParam, pParamsRhs, nRhs)
	pParenParams = Wrapped(LeftParenTT, pParams, RightParenTT)
	// Then(Then(pToken(LeftParenTT, nil), pParams, takeSecond), pToken(RightParenTT, nil), takeFirst)
	pEmptyParams = Then(pToken(LeftParenTT, nil), pToken(RightParenTT, nil), func(res ...ParseRes) *Node { return &Node{Type: ParamNT} })
	pSingleParam = pToken(IdentifierTT, nParam)

	pLambdaRhs = Then(pOperator(ArrowTT),
		skipNewLines(Choice(
			func(r ParseRes, n Nodify) ParseRes { return pExpr(r, n) },
			skipNewLines(Then(Then(pToken(LeftBraceTT, nil),
				skipNewLines(func(r ParseRes, n Nodify) ParseRes { return pStmts(r, n) }), takeSecond),
				skipNewLines(pToken(RightBraceTT, nil)), takeFirst)))),
		nRhs)
	pLambda = Then(Choice(pSingleParam, pEmptyParams, pParenParams), pLambdaRhs, nBinary)

	pSimpleExpr = Choice(pLambda, pCondElseExpr)

	// Compound expressions
	pCompoundExprArg = Choice(pLambda, maybeFunc(pCondElseExpr))
	pPipeExprRhs = Then(pOperator(PipeTT), pCompoundExprArg, nRhs)
	pWhereExprRhs = Then(pOperator(WhereTT), pCompoundExprArg, nRhs)
	pMapExprRhs = Then(pOperator(MapTT), pCompoundExprArg, nRhs)
	pCompoundExprRhs = Plus(skipNewLines(Choice(pPipeExprRhs, pWhereExprRhs, pMapExprRhs)), nLeftAssoc)
	pCompoundExpr = ThenMaybe(pSimpleExpr, pCompoundExprRhs, nEndLeftAssoc)

	pExpr = Choice(pCompoundExpr)

	// Statements

	// Assignment and declaration
	pAssignOp = Choice(pAssignOperator(EqualTT), pAssignOperator(PlusEqualTT), pAssignOperator(MinusEqualTT), pAssignOperator(StarEqualTT), pAssignOperator(SlashEqualTT), pAssignOperator(ModuloEqualTT), pAssignOperator(BarEqualTT))
	pAssignRhs = Then(pAssignOp, skipNewLines(pExpr), nAssignmentRhs)
	pAssignTarget = ThenMaybe(
		pToken(IdentifierTT, nIdentifier),
		Plus(Choice(pBracketAccess, pFieldAccess), nLeftAssoc),
		nEndLeftAssoc,
	)
	pAssignment = Then(pAssignTarget, pAssignRhs, nAssignment)

	pDeclRhs = Then(pOperator(ColonEqualTT), maybeFunc(skipNewLines(pExpr)), nRhs)
	pDeclTarget = Choice(
		pListDestruc,
		pToken(IdentifierTT, nIdentifier),
		pObjDestruc)
	pConstDecl = Then(pDeclTarget, pDeclRhs, nBinary)
	pVarDecl = Then(Then(pToken(VarTT, nil), pDeclTarget, takeSecond), alterNodeType(pDeclRhs, VarDeclNT), nBinary)
	pDecl = Choice(pVarDecl, pConstDecl)

	pReturnStmt = nestRight(Then(pToken(ReturnTT, nil), pExpr, takeSecond), ReturnStmtNT)
	pImportStmt = ThenMaybe(
		Then(pToken(ImportTT, nil), pToken(StringTT, nString), nImport),
		Then(pToken(AsTT, nil), pToken(IdentifierTT, nIdentifier), takeSecond),
		nRhs,
	)

	pSimpleStmt = Choice(pReturnStmt, pOperator(BreakTT), pOperator(ContinueTT), pDecl, pAssignment)

	pStmtBody = Choice(
		Then(pToken(ColonTT, nil), skipNewLines(func(r ParseRes, n Nodify) ParseRes { return pStmt(r, n) }), takeSecond),
		skipNewLines(Then(Then(pToken(LeftBraceTT, nil),
			skipNewLines(func(r ParseRes, n Nodify) ParseRes { return pStmts(r, n) }), takeSecond),
			skipNewLines(pToken(RightBraceTT, nil)), takeFirst)),
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
		Then(pOperator(UnlessTT), pExpr, invertSecond(nLhs)),
		pStmtBody, nRhs)
	pCondStmt = Then(Choice(pIfStmt, pUnlessStmt), Maybe(pElseStmt), nElse)

	// Loop statements
	pWhileStmt = Then(Then(pOperator(WhileTT), pExpr, nLhs), pStmtBody, nRhs)
	pUntilStmt = Then(Then(pOperator(UntilTT), pExpr, invertSecond(nLhs)), pStmtBody, nRhs)
	pForAssign = Then(pDeclTarget, Then(pOperator(LeftArrowTT), pExpr, nRhs), nBinary)
	pForStmt = Then(Then(pOperator(ForTT), pForAssign, nLhs), pStmtBody, nRhs)
	pLoopStmt = Choice(pWhileStmt, pUntilStmt, pForStmt)

	pCompoundStmt = Choice(pCondStmt, pLoopStmt)

	pStmt = nestNode(ThenMaybe(Choice(pImportStmt, pCompoundStmt, pSimpleStmt, pExpr), Either(pToken(NewLineTT, nil), pToken(SemicolonTT, nil)), takeFirst), StmtNT)
	pStmts = Plus(skipNewLines(pStmt), nLinked)

	pProgram = Then(
		pStmts,
		skipNewLines(pToken(EOFTT, nil)),
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
