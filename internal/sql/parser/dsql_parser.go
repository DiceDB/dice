// Code generated from internal/sql/parser/dsql.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // dsql

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type dsqlParser struct {
	*antlr.BaseParser
}

var DsqlParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func dsqlParserInit() {
	staticData := &DsqlParserStaticData
	staticData.LiteralNames = []string{
		"", "'SELECT'", "'WHERE'", "'ORDER'", "'BY'", "'LIMIT'", "'ASC'", "'DESC'",
		"'AND'", "'OR'", "'LIKE'", "'IS'", "'NOT'", "'NULL'", "'_key'", "'_value'",
		"'='", "'=='", "'!='", "'>'", "'>='", "'<'", "'<='", "'('", "')'", "','",
		"'.'", "'''", "'*'",
	}
	staticData.SymbolicNames = []string{
		"", "SELECT", "WHERE", "ORDER", "BY", "LIMIT", "ASC", "DESC", "AND",
		"OR", "LIKE", "IS", "NOT", "NULL", "KEY", "VALUE", "ASSIGN", "EQ", "NEQ",
		"GT", "GTE", "LT", "LTE", "LPAREN", "RPAREN", "COMMA", "DOT", "QUOTE",
		"STAR", "NUMBER", "NUMERIC_LITERAL", "STRING_LITERAL", "BACKTICK_LITERAL",
		"IDENTIFIER", "SINGLE_LINE_COMMENT", "MULTILINE_COMMENT", "SPACES",
		"UNEXPECTED_CHAR",
	}
	staticData.RuleNames = []string{
		"query", "selectStmt", "selectFields", "whereClause", "orderByClause",
		"limitClause", "field", "condition", "orderByField", "expression", "fieldWithString",
		"value", "comparisonOp",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 37, 132, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 3,
		1, 33, 8, 1, 1, 1, 3, 1, 36, 8, 1, 1, 1, 3, 1, 39, 8, 1, 1, 2, 1, 2, 1,
		2, 3, 2, 44, 8, 2, 1, 2, 3, 2, 47, 8, 2, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7,
		1, 7, 3, 7, 67, 8, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 5, 7, 75, 8,
		7, 10, 7, 12, 7, 78, 9, 7, 1, 8, 1, 8, 3, 8, 82, 8, 8, 1, 9, 1, 9, 1, 9,
		1, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 4, 10, 95, 8, 10,
		11, 10, 12, 10, 96, 3, 10, 99, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11,
		1, 11, 1, 11, 4, 11, 108, 8, 11, 11, 11, 12, 11, 109, 1, 11, 1, 11, 1,
		11, 3, 11, 115, 8, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12,
		1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 3, 12, 130, 8, 12, 1, 12, 0,
		1, 14, 13, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 0, 2, 1, 0, 14,
		15, 1, 0, 6, 7, 150, 0, 26, 1, 0, 0, 0, 2, 29, 1, 0, 0, 0, 4, 46, 1, 0,
		0, 0, 6, 48, 1, 0, 0, 0, 8, 51, 1, 0, 0, 0, 10, 55, 1, 0, 0, 0, 12, 58,
		1, 0, 0, 0, 14, 66, 1, 0, 0, 0, 16, 79, 1, 0, 0, 0, 18, 83, 1, 0, 0, 0,
		20, 98, 1, 0, 0, 0, 22, 114, 1, 0, 0, 0, 24, 129, 1, 0, 0, 0, 26, 27, 3,
		2, 1, 0, 27, 28, 5, 0, 0, 1, 28, 1, 1, 0, 0, 0, 29, 30, 5, 1, 0, 0, 30,
		32, 3, 4, 2, 0, 31, 33, 3, 6, 3, 0, 32, 31, 1, 0, 0, 0, 32, 33, 1, 0, 0,
		0, 33, 35, 1, 0, 0, 0, 34, 36, 3, 8, 4, 0, 35, 34, 1, 0, 0, 0, 35, 36,
		1, 0, 0, 0, 36, 38, 1, 0, 0, 0, 37, 39, 3, 10, 5, 0, 38, 37, 1, 0, 0, 0,
		38, 39, 1, 0, 0, 0, 39, 3, 1, 0, 0, 0, 40, 43, 3, 12, 6, 0, 41, 42, 5,
		25, 0, 0, 42, 44, 3, 12, 6, 0, 43, 41, 1, 0, 0, 0, 43, 44, 1, 0, 0, 0,
		44, 47, 1, 0, 0, 0, 45, 47, 5, 28, 0, 0, 46, 40, 1, 0, 0, 0, 46, 45, 1,
		0, 0, 0, 47, 5, 1, 0, 0, 0, 48, 49, 5, 2, 0, 0, 49, 50, 3, 14, 7, 0, 50,
		7, 1, 0, 0, 0, 51, 52, 5, 3, 0, 0, 52, 53, 5, 4, 0, 0, 53, 54, 3, 16, 8,
		0, 54, 9, 1, 0, 0, 0, 55, 56, 5, 5, 0, 0, 56, 57, 5, 29, 0, 0, 57, 11,
		1, 0, 0, 0, 58, 59, 7, 0, 0, 0, 59, 13, 1, 0, 0, 0, 60, 61, 6, 7, -1, 0,
		61, 67, 3, 18, 9, 0, 62, 63, 5, 23, 0, 0, 63, 64, 3, 14, 7, 0, 64, 65,
		5, 24, 0, 0, 65, 67, 1, 0, 0, 0, 66, 60, 1, 0, 0, 0, 66, 62, 1, 0, 0, 0,
		67, 76, 1, 0, 0, 0, 68, 69, 10, 4, 0, 0, 69, 70, 5, 8, 0, 0, 70, 75, 3,
		14, 7, 5, 71, 72, 10, 3, 0, 0, 72, 73, 5, 9, 0, 0, 73, 75, 3, 14, 7, 4,
		74, 68, 1, 0, 0, 0, 74, 71, 1, 0, 0, 0, 75, 78, 1, 0, 0, 0, 76, 74, 1,
		0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 15, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 79,
		81, 3, 20, 10, 0, 80, 82, 7, 1, 0, 0, 81, 80, 1, 0, 0, 0, 81, 82, 1, 0,
		0, 0, 82, 17, 1, 0, 0, 0, 83, 84, 3, 20, 10, 0, 84, 85, 3, 24, 12, 0, 85,
		86, 3, 22, 11, 0, 86, 19, 1, 0, 0, 0, 87, 99, 5, 14, 0, 0, 88, 99, 5, 15,
		0, 0, 89, 99, 5, 31, 0, 0, 90, 99, 5, 32, 0, 0, 91, 94, 5, 15, 0, 0, 92,
		93, 5, 26, 0, 0, 93, 95, 5, 33, 0, 0, 94, 92, 1, 0, 0, 0, 95, 96, 1, 0,
		0, 0, 96, 94, 1, 0, 0, 0, 96, 97, 1, 0, 0, 0, 97, 99, 1, 0, 0, 0, 98, 87,
		1, 0, 0, 0, 98, 88, 1, 0, 0, 0, 98, 89, 1, 0, 0, 0, 98, 90, 1, 0, 0, 0,
		98, 91, 1, 0, 0, 0, 99, 21, 1, 0, 0, 0, 100, 115, 5, 14, 0, 0, 101, 115,
		5, 15, 0, 0, 102, 115, 5, 31, 0, 0, 103, 115, 5, 32, 0, 0, 104, 107, 5,
		15, 0, 0, 105, 106, 5, 26, 0, 0, 106, 108, 5, 33, 0, 0, 107, 105, 1, 0,
		0, 0, 108, 109, 1, 0, 0, 0, 109, 107, 1, 0, 0, 0, 109, 110, 1, 0, 0, 0,
		110, 115, 1, 0, 0, 0, 111, 115, 5, 29, 0, 0, 112, 115, 5, 30, 0, 0, 113,
		115, 5, 13, 0, 0, 114, 100, 1, 0, 0, 0, 114, 101, 1, 0, 0, 0, 114, 102,
		1, 0, 0, 0, 114, 103, 1, 0, 0, 0, 114, 104, 1, 0, 0, 0, 114, 111, 1, 0,
		0, 0, 114, 112, 1, 0, 0, 0, 114, 113, 1, 0, 0, 0, 115, 23, 1, 0, 0, 0,
		116, 130, 5, 16, 0, 0, 117, 130, 5, 17, 0, 0, 118, 130, 5, 18, 0, 0, 119,
		130, 5, 19, 0, 0, 120, 130, 5, 20, 0, 0, 121, 130, 5, 21, 0, 0, 122, 130,
		5, 22, 0, 0, 123, 124, 5, 12, 0, 0, 124, 130, 5, 10, 0, 0, 125, 130, 5,
		10, 0, 0, 126, 127, 5, 11, 0, 0, 127, 130, 5, 12, 0, 0, 128, 130, 5, 11,
		0, 0, 129, 116, 1, 0, 0, 0, 129, 117, 1, 0, 0, 0, 129, 118, 1, 0, 0, 0,
		129, 119, 1, 0, 0, 0, 129, 120, 1, 0, 0, 0, 129, 121, 1, 0, 0, 0, 129,
		122, 1, 0, 0, 0, 129, 123, 1, 0, 0, 0, 129, 125, 1, 0, 0, 0, 129, 126,
		1, 0, 0, 0, 129, 128, 1, 0, 0, 0, 130, 25, 1, 0, 0, 0, 14, 32, 35, 38,
		43, 46, 66, 74, 76, 81, 96, 98, 109, 114, 129,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// dsqlParserInit initializes any static state used to implement dsqlParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewdsqlParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func DsqlParserInit() {
	staticData := &DsqlParserStaticData
	staticData.once.Do(dsqlParserInit)
}

// NewdsqlParser produces a new parser instance for the optional input antlr.TokenStream.
func NewdsqlParser(input antlr.TokenStream) *dsqlParser {
	DsqlParserInit()
	this := new(dsqlParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &DsqlParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "dsql.g4"

	return this
}

// dsqlParser tokens.
const (
	dsqlParserEOF                 = antlr.TokenEOF
	dsqlParserSELECT              = 1
	dsqlParserWHERE               = 2
	dsqlParserORDER               = 3
	dsqlParserBY                  = 4
	dsqlParserLIMIT               = 5
	dsqlParserASC                 = 6
	dsqlParserDESC                = 7
	dsqlParserAND                 = 8
	dsqlParserOR                  = 9
	dsqlParserLIKE                = 10
	dsqlParserIS                  = 11
	dsqlParserNOT                 = 12
	dsqlParserNULL                = 13
	dsqlParserKEY                 = 14
	dsqlParserVALUE               = 15
	dsqlParserASSIGN              = 16
	dsqlParserEQ                  = 17
	dsqlParserNEQ                 = 18
	dsqlParserGT                  = 19
	dsqlParserGTE                 = 20
	dsqlParserLT                  = 21
	dsqlParserLTE                 = 22
	dsqlParserLPAREN              = 23
	dsqlParserRPAREN              = 24
	dsqlParserCOMMA               = 25
	dsqlParserDOT                 = 26
	dsqlParserQUOTE               = 27
	dsqlParserSTAR                = 28
	dsqlParserNUMBER              = 29
	dsqlParserNUMERIC_LITERAL     = 30
	dsqlParserSTRING_LITERAL      = 31
	dsqlParserBACKTICK_LITERAL    = 32
	dsqlParserIDENTIFIER          = 33
	dsqlParserSINGLE_LINE_COMMENT = 34
	dsqlParserMULTILINE_COMMENT   = 35
	dsqlParserSPACES              = 36
	dsqlParserUNEXPECTED_CHAR     = 37
)

// dsqlParser rules.
const (
	dsqlParserRULE_query           = 0
	dsqlParserRULE_selectStmt      = 1
	dsqlParserRULE_selectFields    = 2
	dsqlParserRULE_whereClause     = 3
	dsqlParserRULE_orderByClause   = 4
	dsqlParserRULE_limitClause     = 5
	dsqlParserRULE_field           = 6
	dsqlParserRULE_condition       = 7
	dsqlParserRULE_orderByField    = 8
	dsqlParserRULE_expression      = 9
	dsqlParserRULE_fieldWithString = 10
	dsqlParserRULE_value           = 11
	dsqlParserRULE_comparisonOp    = 12
)

// IQueryContext is an interface to support dynamic dispatch.
type IQueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SelectStmt() ISelectStmtContext
	EOF() antlr.TerminalNode

	// IsQueryContext differentiates from other interfaces.
	IsQueryContext()
}

type QueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQueryContext() *QueryContext {
	var p = new(QueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_query
	return p
}

func InitEmptyQueryContext(p *QueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_query
}

func (*QueryContext) IsQueryContext() {}

func NewQueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QueryContext {
	var p = new(QueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_query

	return p
}

func (s *QueryContext) GetParser() antlr.Parser { return s.parser }

func (s *QueryContext) SelectStmt() ISelectStmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStmtContext)
}

func (s *QueryContext) EOF() antlr.TerminalNode {
	return s.GetToken(dsqlParserEOF, 0)
}

func (s *QueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QueryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterQuery(s)
	}
}

func (s *QueryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitQuery(s)
	}
}

func (p *dsqlParser) Query() (localctx IQueryContext) {
	localctx = NewQueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, dsqlParserRULE_query)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(26)
		p.SelectStmt()
	}
	{
		p.SetState(27)
		p.Match(dsqlParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectStmtContext is an interface to support dynamic dispatch.
type ISelectStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SELECT() antlr.TerminalNode
	SelectFields() ISelectFieldsContext
	WhereClause() IWhereClauseContext
	OrderByClause() IOrderByClauseContext
	LimitClause() ILimitClauseContext

	// IsSelectStmtContext differentiates from other interfaces.
	IsSelectStmtContext()
}

type SelectStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectStmtContext() *SelectStmtContext {
	var p = new(SelectStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_selectStmt
	return p
}

func InitEmptySelectStmtContext(p *SelectStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_selectStmt
}

func (*SelectStmtContext) IsSelectStmtContext() {}

func NewSelectStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectStmtContext {
	var p = new(SelectStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_selectStmt

	return p
}

func (s *SelectStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectStmtContext) SELECT() antlr.TerminalNode {
	return s.GetToken(dsqlParserSELECT, 0)
}

func (s *SelectStmtContext) SelectFields() ISelectFieldsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectFieldsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectFieldsContext)
}

func (s *SelectStmtContext) WhereClause() IWhereClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhereClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhereClauseContext)
}

func (s *SelectStmtContext) OrderByClause() IOrderByClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderByClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderByClauseContext)
}

func (s *SelectStmtContext) LimitClause() ILimitClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILimitClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILimitClauseContext)
}

func (s *SelectStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterSelectStmt(s)
	}
}

func (s *SelectStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitSelectStmt(s)
	}
}

func (p *dsqlParser) SelectStmt() (localctx ISelectStmtContext) {
	localctx = NewSelectStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, dsqlParserRULE_selectStmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(29)
		p.Match(dsqlParserSELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(30)
		p.SelectFields()
	}
	p.SetState(32)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == dsqlParserWHERE {
		{
			p.SetState(31)
			p.WhereClause()
		}

	}
	p.SetState(35)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == dsqlParserORDER {
		{
			p.SetState(34)
			p.OrderByClause()
		}

	}
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == dsqlParserLIMIT {
		{
			p.SetState(37)
			p.LimitClause()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectFieldsContext is an interface to support dynamic dispatch.
type ISelectFieldsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllField() []IFieldContext
	Field(i int) IFieldContext
	COMMA() antlr.TerminalNode
	STAR() antlr.TerminalNode

	// IsSelectFieldsContext differentiates from other interfaces.
	IsSelectFieldsContext()
}

type SelectFieldsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectFieldsContext() *SelectFieldsContext {
	var p = new(SelectFieldsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_selectFields
	return p
}

func InitEmptySelectFieldsContext(p *SelectFieldsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_selectFields
}

func (*SelectFieldsContext) IsSelectFieldsContext() {}

func NewSelectFieldsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectFieldsContext {
	var p = new(SelectFieldsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_selectFields

	return p
}

func (s *SelectFieldsContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectFieldsContext) AllField() []IFieldContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFieldContext); ok {
			len++
		}
	}

	tst := make([]IFieldContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFieldContext); ok {
			tst[i] = t.(IFieldContext)
			i++
		}
	}

	return tst
}

func (s *SelectFieldsContext) Field(i int) IFieldContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *SelectFieldsContext) COMMA() antlr.TerminalNode {
	return s.GetToken(dsqlParserCOMMA, 0)
}

func (s *SelectFieldsContext) STAR() antlr.TerminalNode {
	return s.GetToken(dsqlParserSTAR, 0)
}

func (s *SelectFieldsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectFieldsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectFieldsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterSelectFields(s)
	}
}

func (s *SelectFieldsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitSelectFields(s)
	}
}

func (p *dsqlParser) SelectFields() (localctx ISelectFieldsContext) {
	localctx = NewSelectFieldsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, dsqlParserRULE_selectFields)
	var _la int

	p.SetState(46)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case dsqlParserKEY, dsqlParserVALUE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(40)
			p.Field()
		}
		p.SetState(43)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == dsqlParserCOMMA {
			{
				p.SetState(41)
				p.Match(dsqlParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(42)
				p.Field()
			}

		}

	case dsqlParserSTAR:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(45)
			p.Match(dsqlParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWhereClauseContext is an interface to support dynamic dispatch.
type IWhereClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WHERE() antlr.TerminalNode
	Condition() IConditionContext

	// IsWhereClauseContext differentiates from other interfaces.
	IsWhereClauseContext()
}

type WhereClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhereClauseContext() *WhereClauseContext {
	var p = new(WhereClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_whereClause
	return p
}

func InitEmptyWhereClauseContext(p *WhereClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_whereClause
}

func (*WhereClauseContext) IsWhereClauseContext() {}

func NewWhereClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WhereClauseContext {
	var p = new(WhereClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_whereClause

	return p
}

func (s *WhereClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *WhereClauseContext) WHERE() antlr.TerminalNode {
	return s.GetToken(dsqlParserWHERE, 0)
}

func (s *WhereClauseContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *WhereClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WhereClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WhereClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterWhereClause(s)
	}
}

func (s *WhereClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitWhereClause(s)
	}
}

func (p *dsqlParser) WhereClause() (localctx IWhereClauseContext) {
	localctx = NewWhereClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, dsqlParserRULE_whereClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Match(dsqlParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(49)
		p.condition(0)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrderByClauseContext is an interface to support dynamic dispatch.
type IOrderByClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ORDER() antlr.TerminalNode
	BY() antlr.TerminalNode
	OrderByField() IOrderByFieldContext

	// IsOrderByClauseContext differentiates from other interfaces.
	IsOrderByClauseContext()
}

type OrderByClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrderByClauseContext() *OrderByClauseContext {
	var p = new(OrderByClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_orderByClause
	return p
}

func InitEmptyOrderByClauseContext(p *OrderByClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_orderByClause
}

func (*OrderByClauseContext) IsOrderByClauseContext() {}

func NewOrderByClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrderByClauseContext {
	var p = new(OrderByClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_orderByClause

	return p
}

func (s *OrderByClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *OrderByClauseContext) ORDER() antlr.TerminalNode {
	return s.GetToken(dsqlParserORDER, 0)
}

func (s *OrderByClauseContext) BY() antlr.TerminalNode {
	return s.GetToken(dsqlParserBY, 0)
}

func (s *OrderByClauseContext) OrderByField() IOrderByFieldContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderByFieldContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderByFieldContext)
}

func (s *OrderByClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderByClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrderByClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterOrderByClause(s)
	}
}

func (s *OrderByClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitOrderByClause(s)
	}
}

func (p *dsqlParser) OrderByClause() (localctx IOrderByClauseContext) {
	localctx = NewOrderByClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, dsqlParserRULE_orderByClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(51)
		p.Match(dsqlParserORDER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(52)
		p.Match(dsqlParserBY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(53)
		p.OrderByField()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILimitClauseContext is an interface to support dynamic dispatch.
type ILimitClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LIMIT() antlr.TerminalNode
	NUMBER() antlr.TerminalNode

	// IsLimitClauseContext differentiates from other interfaces.
	IsLimitClauseContext()
}

type LimitClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLimitClauseContext() *LimitClauseContext {
	var p = new(LimitClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_limitClause
	return p
}

func InitEmptyLimitClauseContext(p *LimitClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_limitClause
}

func (*LimitClauseContext) IsLimitClauseContext() {}

func NewLimitClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LimitClauseContext {
	var p = new(LimitClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_limitClause

	return p
}

func (s *LimitClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *LimitClauseContext) LIMIT() antlr.TerminalNode {
	return s.GetToken(dsqlParserLIMIT, 0)
}

func (s *LimitClauseContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(dsqlParserNUMBER, 0)
}

func (s *LimitClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LimitClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LimitClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterLimitClause(s)
	}
}

func (s *LimitClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitLimitClause(s)
	}
}

func (p *dsqlParser) LimitClause() (localctx ILimitClauseContext) {
	localctx = NewLimitClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, dsqlParserRULE_limitClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(55)
		p.Match(dsqlParserLIMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(56)
		p.Match(dsqlParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldContext is an interface to support dynamic dispatch.
type IFieldContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KEY() antlr.TerminalNode
	VALUE() antlr.TerminalNode

	// IsFieldContext differentiates from other interfaces.
	IsFieldContext()
}

type FieldContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldContext() *FieldContext {
	var p = new(FieldContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_field
	return p
}

func InitEmptyFieldContext(p *FieldContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_field
}

func (*FieldContext) IsFieldContext() {}

func NewFieldContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldContext {
	var p = new(FieldContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_field

	return p
}

func (s *FieldContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldContext) KEY() antlr.TerminalNode {
	return s.GetToken(dsqlParserKEY, 0)
}

func (s *FieldContext) VALUE() antlr.TerminalNode {
	return s.GetToken(dsqlParserVALUE, 0)
}

func (s *FieldContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterField(s)
	}
}

func (s *FieldContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitField(s)
	}
}

func (p *dsqlParser) Field() (localctx IFieldContext) {
	localctx = NewFieldContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, dsqlParserRULE_field)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(58)
		_la = p.GetTokenStream().LA(1)

		if !(_la == dsqlParserKEY || _la == dsqlParserVALUE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IConditionContext is an interface to support dynamic dispatch.
type IConditionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	LPAREN() antlr.TerminalNode
	AllCondition() []IConditionContext
	Condition(i int) IConditionContext
	RPAREN() antlr.TerminalNode
	AND() antlr.TerminalNode
	OR() antlr.TerminalNode

	// IsConditionContext differentiates from other interfaces.
	IsConditionContext()
}

type ConditionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConditionContext() *ConditionContext {
	var p = new(ConditionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_condition
	return p
}

func InitEmptyConditionContext(p *ConditionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_condition
}

func (*ConditionContext) IsConditionContext() {}

func NewConditionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConditionContext {
	var p = new(ConditionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_condition

	return p
}

func (s *ConditionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConditionContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(dsqlParserLPAREN, 0)
}

func (s *ConditionContext) AllCondition() []IConditionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConditionContext); ok {
			len++
		}
	}

	tst := make([]IConditionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConditionContext); ok {
			tst[i] = t.(IConditionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionContext) Condition(i int) IConditionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(dsqlParserRPAREN, 0)
}

func (s *ConditionContext) AND() antlr.TerminalNode {
	return s.GetToken(dsqlParserAND, 0)
}

func (s *ConditionContext) OR() antlr.TerminalNode {
	return s.GetToken(dsqlParserOR, 0)
}

func (s *ConditionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConditionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterCondition(s)
	}
}

func (s *ConditionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitCondition(s)
	}
}

func (p *dsqlParser) Condition() (localctx IConditionContext) {
	return p.condition(0)
}

func (p *dsqlParser) condition(_p int) (localctx IConditionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewConditionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IConditionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 14
	p.EnterRecursionRule(localctx, 14, dsqlParserRULE_condition, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(66)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case dsqlParserKEY, dsqlParserVALUE, dsqlParserSTRING_LITERAL, dsqlParserBACKTICK_LITERAL:
		{
			p.SetState(61)
			p.Expression()
		}

	case dsqlParserLPAREN:
		{
			p.SetState(62)
			p.Match(dsqlParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(63)
			p.condition(0)
		}
		{
			p.SetState(64)
			p.Match(dsqlParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(76)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(74)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext()) {
			case 1:
				localctx = NewConditionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, dsqlParserRULE_condition)
				p.SetState(68)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(69)
					p.Match(dsqlParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(70)
					p.condition(5)
				}

			case 2:
				localctx = NewConditionContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, dsqlParserRULE_condition)
				p.SetState(71)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(72)
					p.Match(dsqlParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(73)
					p.condition(4)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(78)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrderByFieldContext is an interface to support dynamic dispatch.
type IOrderByFieldContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FieldWithString() IFieldWithStringContext
	ASC() antlr.TerminalNode
	DESC() antlr.TerminalNode

	// IsOrderByFieldContext differentiates from other interfaces.
	IsOrderByFieldContext()
}

type OrderByFieldContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrderByFieldContext() *OrderByFieldContext {
	var p = new(OrderByFieldContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_orderByField
	return p
}

func InitEmptyOrderByFieldContext(p *OrderByFieldContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_orderByField
}

func (*OrderByFieldContext) IsOrderByFieldContext() {}

func NewOrderByFieldContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrderByFieldContext {
	var p = new(OrderByFieldContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_orderByField

	return p
}

func (s *OrderByFieldContext) GetParser() antlr.Parser { return s.parser }

func (s *OrderByFieldContext) FieldWithString() IFieldWithStringContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldWithStringContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldWithStringContext)
}

func (s *OrderByFieldContext) ASC() antlr.TerminalNode {
	return s.GetToken(dsqlParserASC, 0)
}

func (s *OrderByFieldContext) DESC() antlr.TerminalNode {
	return s.GetToken(dsqlParserDESC, 0)
}

func (s *OrderByFieldContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderByFieldContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrderByFieldContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterOrderByField(s)
	}
}

func (s *OrderByFieldContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitOrderByField(s)
	}
}

func (p *dsqlParser) OrderByField() (localctx IOrderByFieldContext) {
	localctx = NewOrderByFieldContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, dsqlParserRULE_orderByField)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(79)
		p.FieldWithString()
	}
	p.SetState(81)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == dsqlParserASC || _la == dsqlParserDESC {
		{
			p.SetState(80)
			_la = p.GetTokenStream().LA(1)

			if !(_la == dsqlParserASC || _la == dsqlParserDESC) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FieldWithString() IFieldWithStringContext
	ComparisonOp() IComparisonOpContext
	Value() IValueContext

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) FieldWithString() IFieldWithStringContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldWithStringContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldWithStringContext)
}

func (s *ExpressionContext) ComparisonOp() IComparisonOpContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonOpContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonOpContext)
}

func (s *ExpressionContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (p *dsqlParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, dsqlParserRULE_expression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(83)
		p.FieldWithString()
	}
	{
		p.SetState(84)
		p.ComparisonOp()
	}
	{
		p.SetState(85)
		p.Value()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldWithStringContext is an interface to support dynamic dispatch.
type IFieldWithStringContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KEY() antlr.TerminalNode
	VALUE() antlr.TerminalNode
	STRING_LITERAL() antlr.TerminalNode
	BACKTICK_LITERAL() antlr.TerminalNode
	AllDOT() []antlr.TerminalNode
	DOT(i int) antlr.TerminalNode
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode

	// IsFieldWithStringContext differentiates from other interfaces.
	IsFieldWithStringContext()
}

type FieldWithStringContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldWithStringContext() *FieldWithStringContext {
	var p = new(FieldWithStringContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_fieldWithString
	return p
}

func InitEmptyFieldWithStringContext(p *FieldWithStringContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_fieldWithString
}

func (*FieldWithStringContext) IsFieldWithStringContext() {}

func NewFieldWithStringContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldWithStringContext {
	var p = new(FieldWithStringContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_fieldWithString

	return p
}

func (s *FieldWithStringContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldWithStringContext) KEY() antlr.TerminalNode {
	return s.GetToken(dsqlParserKEY, 0)
}

func (s *FieldWithStringContext) VALUE() antlr.TerminalNode {
	return s.GetToken(dsqlParserVALUE, 0)
}

func (s *FieldWithStringContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(dsqlParserSTRING_LITERAL, 0)
}

func (s *FieldWithStringContext) BACKTICK_LITERAL() antlr.TerminalNode {
	return s.GetToken(dsqlParserBACKTICK_LITERAL, 0)
}

func (s *FieldWithStringContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(dsqlParserDOT)
}

func (s *FieldWithStringContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(dsqlParserDOT, i)
}

func (s *FieldWithStringContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(dsqlParserIDENTIFIER)
}

func (s *FieldWithStringContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(dsqlParserIDENTIFIER, i)
}

func (s *FieldWithStringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldWithStringContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldWithStringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterFieldWithString(s)
	}
}

func (s *FieldWithStringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitFieldWithString(s)
	}
}

func (p *dsqlParser) FieldWithString() (localctx IFieldWithStringContext) {
	localctx = NewFieldWithStringContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, dsqlParserRULE_fieldWithString)
	var _la int

	p.SetState(98)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(87)
			p.Match(dsqlParserKEY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(88)
			p.Match(dsqlParserVALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(89)
			p.Match(dsqlParserSTRING_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(90)
			p.Match(dsqlParserBACKTICK_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(91)
			p.Match(dsqlParserVALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(94)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == dsqlParserDOT {
			{
				p.SetState(92)
				p.Match(dsqlParserDOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(93)
				p.Match(dsqlParserIDENTIFIER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(96)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KEY() antlr.TerminalNode
	VALUE() antlr.TerminalNode
	STRING_LITERAL() antlr.TerminalNode
	BACKTICK_LITERAL() antlr.TerminalNode
	AllDOT() []antlr.TerminalNode
	DOT(i int) antlr.TerminalNode
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	NUMBER() antlr.TerminalNode
	NUMERIC_LITERAL() antlr.TerminalNode
	NULL() antlr.TerminalNode

	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) KEY() antlr.TerminalNode {
	return s.GetToken(dsqlParserKEY, 0)
}

func (s *ValueContext) VALUE() antlr.TerminalNode {
	return s.GetToken(dsqlParserVALUE, 0)
}

func (s *ValueContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(dsqlParserSTRING_LITERAL, 0)
}

func (s *ValueContext) BACKTICK_LITERAL() antlr.TerminalNode {
	return s.GetToken(dsqlParserBACKTICK_LITERAL, 0)
}

func (s *ValueContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(dsqlParserDOT)
}

func (s *ValueContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(dsqlParserDOT, i)
}

func (s *ValueContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(dsqlParserIDENTIFIER)
}

func (s *ValueContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(dsqlParserIDENTIFIER, i)
}

func (s *ValueContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(dsqlParserNUMBER, 0)
}

func (s *ValueContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(dsqlParserNUMERIC_LITERAL, 0)
}

func (s *ValueContext) NULL() antlr.TerminalNode {
	return s.GetToken(dsqlParserNULL, 0)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitValue(s)
	}
}

func (p *dsqlParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, dsqlParserRULE_value)
	var _alt int

	p.SetState(114)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(100)
			p.Match(dsqlParserKEY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(101)
			p.Match(dsqlParserVALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(102)
			p.Match(dsqlParserSTRING_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(103)
			p.Match(dsqlParserBACKTICK_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(104)
			p.Match(dsqlParserVALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(107)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(105)
					p.Match(dsqlParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(106)
					p.Match(dsqlParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			default:
				p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
				goto errorExit
			}

			p.SetState(109)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(111)
			p.Match(dsqlParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(112)
			p.Match(dsqlParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(113)
			p.Match(dsqlParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IComparisonOpContext is an interface to support dynamic dispatch.
type IComparisonOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ASSIGN() antlr.TerminalNode
	EQ() antlr.TerminalNode
	NEQ() antlr.TerminalNode
	GT() antlr.TerminalNode
	GTE() antlr.TerminalNode
	LT() antlr.TerminalNode
	LTE() antlr.TerminalNode
	NOT() antlr.TerminalNode
	LIKE() antlr.TerminalNode
	IS() antlr.TerminalNode

	// IsComparisonOpContext differentiates from other interfaces.
	IsComparisonOpContext()
}

type ComparisonOpContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyComparisonOpContext() *ComparisonOpContext {
	var p = new(ComparisonOpContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_comparisonOp
	return p
}

func InitEmptyComparisonOpContext(p *ComparisonOpContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = dsqlParserRULE_comparisonOp
}

func (*ComparisonOpContext) IsComparisonOpContext() {}

func NewComparisonOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ComparisonOpContext {
	var p = new(ComparisonOpContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = dsqlParserRULE_comparisonOp

	return p
}

func (s *ComparisonOpContext) GetParser() antlr.Parser { return s.parser }

func (s *ComparisonOpContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(dsqlParserASSIGN, 0)
}

func (s *ComparisonOpContext) EQ() antlr.TerminalNode {
	return s.GetToken(dsqlParserEQ, 0)
}

func (s *ComparisonOpContext) NEQ() antlr.TerminalNode {
	return s.GetToken(dsqlParserNEQ, 0)
}

func (s *ComparisonOpContext) GT() antlr.TerminalNode {
	return s.GetToken(dsqlParserGT, 0)
}

func (s *ComparisonOpContext) GTE() antlr.TerminalNode {
	return s.GetToken(dsqlParserGTE, 0)
}

func (s *ComparisonOpContext) LT() antlr.TerminalNode {
	return s.GetToken(dsqlParserLT, 0)
}

func (s *ComparisonOpContext) LTE() antlr.TerminalNode {
	return s.GetToken(dsqlParserLTE, 0)
}

func (s *ComparisonOpContext) NOT() antlr.TerminalNode {
	return s.GetToken(dsqlParserNOT, 0)
}

func (s *ComparisonOpContext) LIKE() antlr.TerminalNode {
	return s.GetToken(dsqlParserLIKE, 0)
}

func (s *ComparisonOpContext) IS() antlr.TerminalNode {
	return s.GetToken(dsqlParserIS, 0)
}

func (s *ComparisonOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonOpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ComparisonOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.EnterComparisonOp(s)
	}
}

func (s *ComparisonOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(dsqlListener); ok {
		listenerT.ExitComparisonOp(s)
	}
}

func (p *dsqlParser) ComparisonOp() (localctx IComparisonOpContext) {
	localctx = NewComparisonOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, dsqlParserRULE_comparisonOp)
	p.SetState(129)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(116)
			p.Match(dsqlParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(117)
			p.Match(dsqlParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(118)
			p.Match(dsqlParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(119)
			p.Match(dsqlParserGT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(120)
			p.Match(dsqlParserGTE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(121)
			p.Match(dsqlParserLT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(122)
			p.Match(dsqlParserLTE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(123)
			p.Match(dsqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(124)
			p.Match(dsqlParserLIKE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(125)
			p.Match(dsqlParserLIKE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 10:
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(126)
			p.Match(dsqlParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.Match(dsqlParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 11:
		p.EnterOuterAlt(localctx, 11)
		{
			p.SetState(128)
			p.Match(dsqlParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *dsqlParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 7:
		var t *ConditionContext = nil
		if localctx != nil {
			t = localctx.(*ConditionContext)
		}
		return p.Condition_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *dsqlParser) Condition_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
