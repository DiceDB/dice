// Code generated from DSQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // DSQL

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

type DSQLParser struct {
	*antlr.BaseParser
}

var DSQLParserStaticData struct {
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
	staticData := &DSQLParserStaticData
	staticData.LiteralNames = []string{
		"", "';'", "','", "'('", "')'", "'-'", "'+'", "'~'", "'||'", "'*'",
		"'/'", "'%'", "'<<'", "'>>'", "'&'", "'|'", "'<'", "'<='", "'>'", "'>='",
		"'='", "'=='", "'!='", "'<>'", "'$key'", "'$value'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "KEY_TOKEN", "VALUE_TOKEN", "K_SELECT",
		"K_FROM", "K_WHERE", "K_ORDER", "K_BY", "K_LIMIT", "K_ASC", "K_DESC",
		"K_IS", "K_NOT", "K_NULL", "K_LIKE", "K_AND", "K_OR", "VALUE_TOKEN_TRAILING_DOT",
		"IDENTIFIER", "NUMERIC_LITERAL", "STRING_LITERAL", "WHITESPACE",
	}
	staticData.RuleNames = []string{
		"parse", "sql_stmt_list", "sql_stmt", "select_stmt", "select_core",
		"from_clause", "where_clause", "ordering_terms", "ordering_term", "limit_clause",
		"scalar_expr", "table_name", "literal_value", "unary_operator", "binary_operator",
		"keyword",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 44, 145, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 5, 1, 39, 8, 1, 10, 1, 12, 1, 42, 9,
		1, 1, 1, 3, 1, 45, 8, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 53,
		8, 3, 1, 3, 1, 3, 3, 3, 57, 8, 3, 1, 3, 1, 3, 1, 3, 3, 3, 62, 8, 3, 1,
		3, 1, 3, 3, 3, 66, 8, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		3, 4, 76, 8, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 5, 7, 85, 8,
		7, 10, 7, 12, 7, 88, 9, 7, 1, 8, 1, 8, 3, 8, 92, 8, 8, 1, 9, 1, 9, 1, 10,
		1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1,
		10, 3, 10, 108, 8, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 10, 5, 10, 130, 8, 10, 10, 10, 12, 10, 133, 9, 10, 1, 11,
		1, 11, 1, 12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 0,
		1, 20, 16, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 0,
		5, 1, 0, 32, 33, 2, 0, 36, 36, 42, 43, 2, 0, 5, 7, 35, 35, 3, 0, 5, 6,
		8, 23, 38, 39, 1, 0, 26, 39, 150, 0, 32, 1, 0, 0, 0, 2, 35, 1, 0, 0, 0,
		4, 46, 1, 0, 0, 0, 6, 48, 1, 0, 0, 0, 8, 75, 1, 0, 0, 0, 10, 77, 1, 0,
		0, 0, 12, 79, 1, 0, 0, 0, 14, 81, 1, 0, 0, 0, 16, 89, 1, 0, 0, 0, 18, 93,
		1, 0, 0, 0, 20, 107, 1, 0, 0, 0, 22, 134, 1, 0, 0, 0, 24, 136, 1, 0, 0,
		0, 26, 138, 1, 0, 0, 0, 28, 140, 1, 0, 0, 0, 30, 142, 1, 0, 0, 0, 32, 33,
		3, 2, 1, 0, 33, 34, 5, 0, 0, 1, 34, 1, 1, 0, 0, 0, 35, 40, 3, 4, 2, 0,
		36, 37, 5, 1, 0, 0, 37, 39, 3, 4, 2, 0, 38, 36, 1, 0, 0, 0, 39, 42, 1,
		0, 0, 0, 40, 38, 1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41, 44, 1, 0, 0, 0, 42,
		40, 1, 0, 0, 0, 43, 45, 5, 1, 0, 0, 44, 43, 1, 0, 0, 0, 44, 45, 1, 0, 0,
		0, 45, 3, 1, 0, 0, 0, 46, 47, 3, 6, 3, 0, 47, 5, 1, 0, 0, 0, 48, 49, 5,
		26, 0, 0, 49, 52, 3, 8, 4, 0, 50, 51, 5, 27, 0, 0, 51, 53, 3, 10, 5, 0,
		52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0, 53, 56, 1, 0, 0, 0, 54, 55, 5,
		28, 0, 0, 55, 57, 3, 12, 6, 0, 56, 54, 1, 0, 0, 0, 56, 57, 1, 0, 0, 0,
		57, 61, 1, 0, 0, 0, 58, 59, 5, 29, 0, 0, 59, 60, 5, 30, 0, 0, 60, 62, 3,
		14, 7, 0, 61, 58, 1, 0, 0, 0, 61, 62, 1, 0, 0, 0, 62, 65, 1, 0, 0, 0, 63,
		64, 5, 31, 0, 0, 64, 66, 3, 18, 9, 0, 65, 63, 1, 0, 0, 0, 65, 66, 1, 0,
		0, 0, 66, 7, 1, 0, 0, 0, 67, 76, 5, 24, 0, 0, 68, 76, 5, 25, 0, 0, 69,
		70, 5, 24, 0, 0, 70, 71, 5, 2, 0, 0, 71, 76, 5, 25, 0, 0, 72, 73, 5, 25,
		0, 0, 73, 74, 5, 2, 0, 0, 74, 76, 5, 24, 0, 0, 75, 67, 1, 0, 0, 0, 75,
		68, 1, 0, 0, 0, 75, 69, 1, 0, 0, 0, 75, 72, 1, 0, 0, 0, 76, 9, 1, 0, 0,
		0, 77, 78, 3, 22, 11, 0, 78, 11, 1, 0, 0, 0, 79, 80, 3, 20, 10, 0, 80,
		13, 1, 0, 0, 0, 81, 86, 3, 16, 8, 0, 82, 83, 5, 2, 0, 0, 83, 85, 3, 16,
		8, 0, 84, 82, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0, 86, 84, 1, 0, 0, 0, 86, 87,
		1, 0, 0, 0, 87, 15, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 89, 91, 5, 41, 0, 0,
		90, 92, 7, 0, 0, 0, 91, 90, 1, 0, 0, 0, 91, 92, 1, 0, 0, 0, 92, 17, 1,
		0, 0, 0, 93, 94, 5, 42, 0, 0, 94, 19, 1, 0, 0, 0, 95, 96, 6, 10, -1, 0,
		96, 108, 3, 24, 12, 0, 97, 108, 5, 24, 0, 0, 98, 108, 5, 25, 0, 0, 99,
		108, 5, 41, 0, 0, 100, 101, 3, 26, 13, 0, 101, 102, 3, 20, 10, 8, 102,
		108, 1, 0, 0, 0, 103, 104, 5, 3, 0, 0, 104, 105, 3, 20, 10, 0, 105, 106,
		5, 4, 0, 0, 106, 108, 1, 0, 0, 0, 107, 95, 1, 0, 0, 0, 107, 97, 1, 0, 0,
		0, 107, 98, 1, 0, 0, 0, 107, 99, 1, 0, 0, 0, 107, 100, 1, 0, 0, 0, 107,
		103, 1, 0, 0, 0, 108, 131, 1, 0, 0, 0, 109, 110, 10, 7, 0, 0, 110, 111,
		3, 28, 14, 0, 111, 112, 3, 20, 10, 8, 112, 130, 1, 0, 0, 0, 113, 114, 10,
		3, 0, 0, 114, 115, 5, 37, 0, 0, 115, 130, 3, 20, 10, 4, 116, 117, 10, 2,
		0, 0, 117, 118, 5, 38, 0, 0, 118, 130, 3, 20, 10, 3, 119, 120, 10, 1, 0,
		0, 120, 121, 5, 39, 0, 0, 121, 130, 3, 20, 10, 2, 122, 123, 10, 5, 0, 0,
		123, 124, 5, 34, 0, 0, 124, 130, 5, 36, 0, 0, 125, 126, 10, 4, 0, 0, 126,
		127, 5, 34, 0, 0, 127, 128, 5, 35, 0, 0, 128, 130, 5, 36, 0, 0, 129, 109,
		1, 0, 0, 0, 129, 113, 1, 0, 0, 0, 129, 116, 1, 0, 0, 0, 129, 119, 1, 0,
		0, 0, 129, 122, 1, 0, 0, 0, 129, 125, 1, 0, 0, 0, 130, 133, 1, 0, 0, 0,
		131, 129, 1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132, 21, 1, 0, 0, 0, 133, 131,
		1, 0, 0, 0, 134, 135, 5, 41, 0, 0, 135, 23, 1, 0, 0, 0, 136, 137, 7, 1,
		0, 0, 137, 25, 1, 0, 0, 0, 138, 139, 7, 2, 0, 0, 139, 27, 1, 0, 0, 0, 140,
		141, 7, 3, 0, 0, 141, 29, 1, 0, 0, 0, 142, 143, 7, 4, 0, 0, 143, 31, 1,
		0, 0, 0, 12, 40, 44, 52, 56, 61, 65, 75, 86, 91, 107, 129, 131,
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

// DSQLParserInit initializes any static state used to implement DSQLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewDSQLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func DSQLParserInit() {
	staticData := &DSQLParserStaticData
	staticData.once.Do(dsqlParserInit)
}

// NewDSQLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewDSQLParser(input antlr.TokenStream) *DSQLParser {
	DSQLParserInit()
	this := new(DSQLParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &DSQLParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "DSQL.g4"

	return this
}

// DSQLParser tokens.
const (
	DSQLParserEOF                      = antlr.TokenEOF
	DSQLParserT__0                     = 1
	DSQLParserT__1                     = 2
	DSQLParserT__2                     = 3
	DSQLParserT__3                     = 4
	DSQLParserT__4                     = 5
	DSQLParserT__5                     = 6
	DSQLParserT__6                     = 7
	DSQLParserT__7                     = 8
	DSQLParserT__8                     = 9
	DSQLParserT__9                     = 10
	DSQLParserT__10                    = 11
	DSQLParserT__11                    = 12
	DSQLParserT__12                    = 13
	DSQLParserT__13                    = 14
	DSQLParserT__14                    = 15
	DSQLParserT__15                    = 16
	DSQLParserT__16                    = 17
	DSQLParserT__17                    = 18
	DSQLParserT__18                    = 19
	DSQLParserT__19                    = 20
	DSQLParserT__20                    = 21
	DSQLParserT__21                    = 22
	DSQLParserT__22                    = 23
	DSQLParserKEY_TOKEN                = 24
	DSQLParserVALUE_TOKEN              = 25
	DSQLParserK_SELECT                 = 26
	DSQLParserK_FROM                   = 27
	DSQLParserK_WHERE                  = 28
	DSQLParserK_ORDER                  = 29
	DSQLParserK_BY                     = 30
	DSQLParserK_LIMIT                  = 31
	DSQLParserK_ASC                    = 32
	DSQLParserK_DESC                   = 33
	DSQLParserK_IS                     = 34
	DSQLParserK_NOT                    = 35
	DSQLParserK_NULL                   = 36
	DSQLParserK_LIKE                   = 37
	DSQLParserK_AND                    = 38
	DSQLParserK_OR                     = 39
	DSQLParserVALUE_TOKEN_TRAILING_DOT = 40
	DSQLParserIDENTIFIER               = 41
	DSQLParserNUMERIC_LITERAL          = 42
	DSQLParserSTRING_LITERAL           = 43
	DSQLParserWHITESPACE               = 44
)

// DSQLParser rules.
const (
	DSQLParserRULE_parse           = 0
	DSQLParserRULE_sql_stmt_list   = 1
	DSQLParserRULE_sql_stmt        = 2
	DSQLParserRULE_select_stmt     = 3
	DSQLParserRULE_select_core     = 4
	DSQLParserRULE_from_clause     = 5
	DSQLParserRULE_where_clause    = 6
	DSQLParserRULE_ordering_terms  = 7
	DSQLParserRULE_ordering_term   = 8
	DSQLParserRULE_limit_clause    = 9
	DSQLParserRULE_scalar_expr     = 10
	DSQLParserRULE_table_name      = 11
	DSQLParserRULE_literal_value   = 12
	DSQLParserRULE_unary_operator  = 13
	DSQLParserRULE_binary_operator = 14
	DSQLParserRULE_keyword         = 15
)

// IParseContext is an interface to support dynamic dispatch.
type IParseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Sql_stmt_list() ISql_stmt_listContext
	EOF() antlr.TerminalNode

	// IsParseContext differentiates from other interfaces.
	IsParseContext()
}

type ParseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParseContext() *ParseContext {
	var p = new(ParseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_parse
	return p
}

func InitEmptyParseContext(p *ParseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_parse
}

func (*ParseContext) IsParseContext() {}

func NewParseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ParseContext {
	var p = new(ParseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_parse

	return p
}

func (s *ParseContext) GetParser() antlr.Parser { return s.parser }

func (s *ParseContext) Sql_stmt_list() ISql_stmt_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmt_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISql_stmt_listContext)
}

func (s *ParseContext) EOF() antlr.TerminalNode {
	return s.GetToken(DSQLParserEOF, 0)
}

func (s *ParseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ParseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterParse(s)
	}
}

func (s *ParseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitParse(s)
	}
}

func (p *DSQLParser) Parse() (localctx IParseContext) {
	localctx = NewParseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, DSQLParserRULE_parse)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(32)
		p.Sql_stmt_list()
	}
	{
		p.SetState(33)
		p.Match(DSQLParserEOF)
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

// ISql_stmt_listContext is an interface to support dynamic dispatch.
type ISql_stmt_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSql_stmt() []ISql_stmtContext
	Sql_stmt(i int) ISql_stmtContext

	// IsSql_stmt_listContext differentiates from other interfaces.
	IsSql_stmt_listContext()
}

type Sql_stmt_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySql_stmt_listContext() *Sql_stmt_listContext {
	var p = new(Sql_stmt_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_sql_stmt_list
	return p
}

func InitEmptySql_stmt_listContext(p *Sql_stmt_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_sql_stmt_list
}

func (*Sql_stmt_listContext) IsSql_stmt_listContext() {}

func NewSql_stmt_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sql_stmt_listContext {
	var p = new(Sql_stmt_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_sql_stmt_list

	return p
}

func (s *Sql_stmt_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Sql_stmt_listContext) AllSql_stmt() []ISql_stmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISql_stmtContext); ok {
			len++
		}
	}

	tst := make([]ISql_stmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISql_stmtContext); ok {
			tst[i] = t.(ISql_stmtContext)
			i++
		}
	}

	return tst
}

func (s *Sql_stmt_listContext) Sql_stmt(i int) ISql_stmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmtContext); ok {
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

	return t.(ISql_stmtContext)
}

func (s *Sql_stmt_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sql_stmt_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sql_stmt_listContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterSql_stmt_list(s)
	}
}

func (s *Sql_stmt_listContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitSql_stmt_list(s)
	}
}

func (p *DSQLParser) Sql_stmt_list() (localctx ISql_stmt_listContext) {
	localctx = NewSql_stmt_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, DSQLParserRULE_sql_stmt_list)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(35)
		p.Sql_stmt()
	}
	p.SetState(40)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(36)
				p.Match(DSQLParserT__0)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(37)
				p.Sql_stmt()
			}

		}
		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserT__0 {
		{
			p.SetState(43)
			p.Match(DSQLParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
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

// ISql_stmtContext is an interface to support dynamic dispatch.
type ISql_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Select_stmt() ISelect_stmtContext

	// IsSql_stmtContext differentiates from other interfaces.
	IsSql_stmtContext()
}

type Sql_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySql_stmtContext() *Sql_stmtContext {
	var p = new(Sql_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_sql_stmt
	return p
}

func InitEmptySql_stmtContext(p *Sql_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_sql_stmt
}

func (*Sql_stmtContext) IsSql_stmtContext() {}

func NewSql_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sql_stmtContext {
	var p = new(Sql_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_sql_stmt

	return p
}

func (s *Sql_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Sql_stmtContext) Select_stmt() ISelect_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmtContext)
}

func (s *Sql_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sql_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sql_stmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterSql_stmt(s)
	}
}

func (s *Sql_stmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitSql_stmt(s)
	}
}

func (p *DSQLParser) Sql_stmt() (localctx ISql_stmtContext) {
	localctx = NewSql_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, DSQLParserRULE_sql_stmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(46)
		p.Select_stmt()
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

// ISelect_stmtContext is an interface to support dynamic dispatch.
type ISelect_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_SELECT() antlr.TerminalNode
	Select_core() ISelect_coreContext
	K_FROM() antlr.TerminalNode
	From_clause() IFrom_clauseContext
	K_WHERE() antlr.TerminalNode
	Where_clause() IWhere_clauseContext
	K_ORDER() antlr.TerminalNode
	K_BY() antlr.TerminalNode
	Ordering_terms() IOrdering_termsContext
	K_LIMIT() antlr.TerminalNode
	Limit_clause() ILimit_clauseContext

	// IsSelect_stmtContext differentiates from other interfaces.
	IsSelect_stmtContext()
}

type Select_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelect_stmtContext() *Select_stmtContext {
	var p = new(Select_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_stmt
	return p
}

func InitEmptySelect_stmtContext(p *Select_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_stmt
}

func (*Select_stmtContext) IsSelect_stmtContext() {}

func NewSelect_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_stmtContext {
	var p = new(Select_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_select_stmt

	return p
}

func (s *Select_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_stmtContext) K_SELECT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_SELECT, 0)
}

func (s *Select_stmtContext) Select_core() ISelect_coreContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_coreContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_coreContext)
}

func (s *Select_stmtContext) K_FROM() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_FROM, 0)
}

func (s *Select_stmtContext) From_clause() IFrom_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFrom_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFrom_clauseContext)
}

func (s *Select_stmtContext) K_WHERE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_WHERE, 0)
}

func (s *Select_stmtContext) Where_clause() IWhere_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhere_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhere_clauseContext)
}

func (s *Select_stmtContext) K_ORDER() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_ORDER, 0)
}

func (s *Select_stmtContext) K_BY() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_BY, 0)
}

func (s *Select_stmtContext) Ordering_terms() IOrdering_termsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrdering_termsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrdering_termsContext)
}

func (s *Select_stmtContext) K_LIMIT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_LIMIT, 0)
}

func (s *Select_stmtContext) Limit_clause() ILimit_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILimit_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILimit_clauseContext)
}

func (s *Select_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_stmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterSelect_stmt(s)
	}
}

func (s *Select_stmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitSelect_stmt(s)
	}
}

func (p *DSQLParser) Select_stmt() (localctx ISelect_stmtContext) {
	localctx = NewSelect_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, DSQLParserRULE_select_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Match(DSQLParserK_SELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(49)
		p.Select_core()
	}
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_FROM {
		{
			p.SetState(50)
			p.Match(DSQLParserK_FROM)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(51)
			p.From_clause()
		}

	}
	p.SetState(56)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_WHERE {
		{
			p.SetState(54)
			p.Match(DSQLParserK_WHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(55)
			p.Where_clause()
		}

	}
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_ORDER {
		{
			p.SetState(58)
			p.Match(DSQLParserK_ORDER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(59)
			p.Match(DSQLParserK_BY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(60)
			p.Ordering_terms()
		}

	}
	p.SetState(65)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_LIMIT {
		{
			p.SetState(63)
			p.Match(DSQLParserK_LIMIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(64)
			p.Limit_clause()
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

// ISelect_coreContext is an interface to support dynamic dispatch.
type ISelect_coreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	KEY_TOKEN() antlr.TerminalNode
	VALUE_TOKEN() antlr.TerminalNode

	// IsSelect_coreContext differentiates from other interfaces.
	IsSelect_coreContext()
}

type Select_coreContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelect_coreContext() *Select_coreContext {
	var p = new(Select_coreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_core
	return p
}

func InitEmptySelect_coreContext(p *Select_coreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_core
}

func (*Select_coreContext) IsSelect_coreContext() {}

func NewSelect_coreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_coreContext {
	var p = new(Select_coreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_select_core

	return p
}

func (s *Select_coreContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_coreContext) KEY_TOKEN() antlr.TerminalNode {
	return s.GetToken(DSQLParserKEY_TOKEN, 0)
}

func (s *Select_coreContext) VALUE_TOKEN() antlr.TerminalNode {
	return s.GetToken(DSQLParserVALUE_TOKEN, 0)
}

func (s *Select_coreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_coreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_coreContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterSelect_core(s)
	}
}

func (s *Select_coreContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitSelect_core(s)
	}
}

func (p *DSQLParser) Select_core() (localctx ISelect_coreContext) {
	localctx = NewSelect_coreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, DSQLParserRULE_select_core)
	p.SetState(75)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(67)
			p.Match(DSQLParserKEY_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(68)
			p.Match(DSQLParserVALUE_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(69)
			p.Match(DSQLParserKEY_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(70)
			p.Match(DSQLParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(71)
			p.Match(DSQLParserVALUE_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(72)
			p.Match(DSQLParserVALUE_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(73)
			p.Match(DSQLParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(74)
			p.Match(DSQLParserKEY_TOKEN)
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

// IFrom_clauseContext is an interface to support dynamic dispatch.
type IFrom_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Table_name() ITable_nameContext

	// IsFrom_clauseContext differentiates from other interfaces.
	IsFrom_clauseContext()
}

type From_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFrom_clauseContext() *From_clauseContext {
	var p = new(From_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_from_clause
	return p
}

func InitEmptyFrom_clauseContext(p *From_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_from_clause
}

func (*From_clauseContext) IsFrom_clauseContext() {}

func NewFrom_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *From_clauseContext {
	var p = new(From_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_from_clause

	return p
}

func (s *From_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *From_clauseContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *From_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *From_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *From_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterFrom_clause(s)
	}
}

func (s *From_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitFrom_clause(s)
	}
}

func (p *DSQLParser) From_clause() (localctx IFrom_clauseContext) {
	localctx = NewFrom_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, DSQLParserRULE_from_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(77)
		p.Table_name()
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

// IWhere_clauseContext is an interface to support dynamic dispatch.
type IWhere_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Scalar_expr() IScalar_exprContext

	// IsWhere_clauseContext differentiates from other interfaces.
	IsWhere_clauseContext()
}

type Where_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhere_clauseContext() *Where_clauseContext {
	var p = new(Where_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_where_clause
	return p
}

func InitEmptyWhere_clauseContext(p *Where_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_where_clause
}

func (*Where_clauseContext) IsWhere_clauseContext() {}

func NewWhere_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Where_clauseContext {
	var p = new(Where_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_where_clause

	return p
}

func (s *Where_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Where_clauseContext) Scalar_expr() IScalar_exprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScalar_exprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScalar_exprContext)
}

func (s *Where_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Where_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Where_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterWhere_clause(s)
	}
}

func (s *Where_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitWhere_clause(s)
	}
}

func (p *DSQLParser) Where_clause() (localctx IWhere_clauseContext) {
	localctx = NewWhere_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, DSQLParserRULE_where_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(79)
		p.scalar_expr(0)
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

// IOrdering_termsContext is an interface to support dynamic dispatch.
type IOrdering_termsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllOrdering_term() []IOrdering_termContext
	Ordering_term(i int) IOrdering_termContext

	// IsOrdering_termsContext differentiates from other interfaces.
	IsOrdering_termsContext()
}

type Ordering_termsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrdering_termsContext() *Ordering_termsContext {
	var p = new(Ordering_termsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_ordering_terms
	return p
}

func InitEmptyOrdering_termsContext(p *Ordering_termsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_ordering_terms
}

func (*Ordering_termsContext) IsOrdering_termsContext() {}

func NewOrdering_termsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Ordering_termsContext {
	var p = new(Ordering_termsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_ordering_terms

	return p
}

func (s *Ordering_termsContext) GetParser() antlr.Parser { return s.parser }

func (s *Ordering_termsContext) AllOrdering_term() []IOrdering_termContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrdering_termContext); ok {
			len++
		}
	}

	tst := make([]IOrdering_termContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrdering_termContext); ok {
			tst[i] = t.(IOrdering_termContext)
			i++
		}
	}

	return tst
}

func (s *Ordering_termsContext) Ordering_term(i int) IOrdering_termContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrdering_termContext); ok {
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

	return t.(IOrdering_termContext)
}

func (s *Ordering_termsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Ordering_termsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Ordering_termsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterOrdering_terms(s)
	}
}

func (s *Ordering_termsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitOrdering_terms(s)
	}
}

func (p *DSQLParser) Ordering_terms() (localctx IOrdering_termsContext) {
	localctx = NewOrdering_termsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, DSQLParserRULE_ordering_terms)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.Ordering_term()
	}
	p.SetState(86)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DSQLParserT__1 {
		{
			p.SetState(82)
			p.Match(DSQLParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(83)
			p.Ordering_term()
		}

		p.SetState(88)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
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

// IOrdering_termContext is an interface to support dynamic dispatch.
type IOrdering_termContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	K_ASC() antlr.TerminalNode
	K_DESC() antlr.TerminalNode

	// IsOrdering_termContext differentiates from other interfaces.
	IsOrdering_termContext()
}

type Ordering_termContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrdering_termContext() *Ordering_termContext {
	var p = new(Ordering_termContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_ordering_term
	return p
}

func InitEmptyOrdering_termContext(p *Ordering_termContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_ordering_term
}

func (*Ordering_termContext) IsOrdering_termContext() {}

func NewOrdering_termContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Ordering_termContext {
	var p = new(Ordering_termContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_ordering_term

	return p
}

func (s *Ordering_termContext) GetParser() antlr.Parser { return s.parser }

func (s *Ordering_termContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(DSQLParserIDENTIFIER, 0)
}

func (s *Ordering_termContext) K_ASC() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_ASC, 0)
}

func (s *Ordering_termContext) K_DESC() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_DESC, 0)
}

func (s *Ordering_termContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Ordering_termContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Ordering_termContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterOrdering_term(s)
	}
}

func (s *Ordering_termContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitOrdering_term(s)
	}
}

func (p *DSQLParser) Ordering_term() (localctx IOrdering_termContext) {
	localctx = NewOrdering_termContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, DSQLParserRULE_ordering_term)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(89)
		p.Match(DSQLParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(91)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_ASC || _la == DSQLParserK_DESC {
		{
			p.SetState(90)
			_la = p.GetTokenStream().LA(1)

			if !(_la == DSQLParserK_ASC || _la == DSQLParserK_DESC) {
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

// ILimit_clauseContext is an interface to support dynamic dispatch.
type ILimit_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMERIC_LITERAL() antlr.TerminalNode

	// IsLimit_clauseContext differentiates from other interfaces.
	IsLimit_clauseContext()
}

type Limit_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLimit_clauseContext() *Limit_clauseContext {
	var p = new(Limit_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_limit_clause
	return p
}

func InitEmptyLimit_clauseContext(p *Limit_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_limit_clause
}

func (*Limit_clauseContext) IsLimit_clauseContext() {}

func NewLimit_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Limit_clauseContext {
	var p = new(Limit_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_limit_clause

	return p
}

func (s *Limit_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Limit_clauseContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(DSQLParserNUMERIC_LITERAL, 0)
}

func (s *Limit_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Limit_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Limit_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterLimit_clause(s)
	}
}

func (s *Limit_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitLimit_clause(s)
	}
}

func (p *DSQLParser) Limit_clause() (localctx ILimit_clauseContext) {
	localctx = NewLimit_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, DSQLParserRULE_limit_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(93)
		p.Match(DSQLParserNUMERIC_LITERAL)
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

// IScalar_exprContext is an interface to support dynamic dispatch.
type IScalar_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Literal_value() ILiteral_valueContext
	KEY_TOKEN() antlr.TerminalNode
	VALUE_TOKEN() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	Unary_operator() IUnary_operatorContext
	AllScalar_expr() []IScalar_exprContext
	Scalar_expr(i int) IScalar_exprContext
	Binary_operator() IBinary_operatorContext
	K_LIKE() antlr.TerminalNode
	K_AND() antlr.TerminalNode
	K_OR() antlr.TerminalNode
	K_IS() antlr.TerminalNode
	K_NULL() antlr.TerminalNode
	K_NOT() antlr.TerminalNode

	// IsScalar_exprContext differentiates from other interfaces.
	IsScalar_exprContext()
}

type Scalar_exprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScalar_exprContext() *Scalar_exprContext {
	var p = new(Scalar_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_scalar_expr
	return p
}

func InitEmptyScalar_exprContext(p *Scalar_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_scalar_expr
}

func (*Scalar_exprContext) IsScalar_exprContext() {}

func NewScalar_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Scalar_exprContext {
	var p = new(Scalar_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_scalar_expr

	return p
}

func (s *Scalar_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Scalar_exprContext) Literal_value() ILiteral_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteral_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteral_valueContext)
}

func (s *Scalar_exprContext) KEY_TOKEN() antlr.TerminalNode {
	return s.GetToken(DSQLParserKEY_TOKEN, 0)
}

func (s *Scalar_exprContext) VALUE_TOKEN() antlr.TerminalNode {
	return s.GetToken(DSQLParserVALUE_TOKEN, 0)
}

func (s *Scalar_exprContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(DSQLParserIDENTIFIER, 0)
}

func (s *Scalar_exprContext) Unary_operator() IUnary_operatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnary_operatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnary_operatorContext)
}

func (s *Scalar_exprContext) AllScalar_expr() []IScalar_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IScalar_exprContext); ok {
			len++
		}
	}

	tst := make([]IScalar_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IScalar_exprContext); ok {
			tst[i] = t.(IScalar_exprContext)
			i++
		}
	}

	return tst
}

func (s *Scalar_exprContext) Scalar_expr(i int) IScalar_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScalar_exprContext); ok {
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

	return t.(IScalar_exprContext)
}

func (s *Scalar_exprContext) Binary_operator() IBinary_operatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBinary_operatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBinary_operatorContext)
}

func (s *Scalar_exprContext) K_LIKE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_LIKE, 0)
}

func (s *Scalar_exprContext) K_AND() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_AND, 0)
}

func (s *Scalar_exprContext) K_OR() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_OR, 0)
}

func (s *Scalar_exprContext) K_IS() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_IS, 0)
}

func (s *Scalar_exprContext) K_NULL() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NULL, 0)
}

func (s *Scalar_exprContext) K_NOT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NOT, 0)
}

func (s *Scalar_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Scalar_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Scalar_exprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterScalar_expr(s)
	}
}

func (s *Scalar_exprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitScalar_expr(s)
	}
}

func (p *DSQLParser) Scalar_expr() (localctx IScalar_exprContext) {
	return p.scalar_expr(0)
}

func (p *DSQLParser) scalar_expr(_p int) (localctx IScalar_exprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewScalar_exprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IScalar_exprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 20
	p.EnterRecursionRule(localctx, 20, DSQLParserRULE_scalar_expr, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case DSQLParserK_NULL, DSQLParserNUMERIC_LITERAL, DSQLParserSTRING_LITERAL:
		{
			p.SetState(96)
			p.Literal_value()
		}

	case DSQLParserKEY_TOKEN:
		{
			p.SetState(97)
			p.Match(DSQLParserKEY_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserVALUE_TOKEN:
		{
			p.SetState(98)
			p.Match(DSQLParserVALUE_TOKEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserIDENTIFIER:
		{
			p.SetState(99)
			p.Match(DSQLParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserT__4, DSQLParserT__5, DSQLParserT__6, DSQLParserK_NOT:
		{
			p.SetState(100)
			p.Unary_operator()
		}
		{
			p.SetState(101)
			p.scalar_expr(8)
		}

	case DSQLParserT__2:
		{
			p.SetState(103)
			p.Match(DSQLParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(104)
			p.scalar_expr(0)
		}
		{
			p.SetState(105)
			p.Match(DSQLParserT__3)
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
	p.SetState(131)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(129)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext()) {
			case 1:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(109)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(110)
					p.Binary_operator()
				}
				{
					p.SetState(111)
					p.scalar_expr(8)
				}

			case 2:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(113)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(114)
					p.Match(DSQLParserK_LIKE)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(115)
					p.scalar_expr(4)
				}

			case 3:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(116)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(117)
					p.Match(DSQLParserK_AND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(118)
					p.scalar_expr(3)
				}

			case 4:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(119)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				{
					p.SetState(120)
					p.Match(DSQLParserK_OR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(121)
					p.scalar_expr(2)
				}

			case 5:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(122)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(123)
					p.Match(DSQLParserK_IS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(124)
					p.Match(DSQLParserK_NULL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case 6:
				localctx = NewScalar_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_scalar_expr)
				p.SetState(125)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(126)
					p.Match(DSQLParserK_IS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(127)
					p.Match(DSQLParserK_NOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(128)
					p.Match(DSQLParserK_NULL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(133)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
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

// ITable_nameContext is an interface to support dynamic dispatch.
type ITable_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsTable_nameContext differentiates from other interfaces.
	IsTable_nameContext()
}

type Table_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_nameContext() *Table_nameContext {
	var p = new(Table_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_table_name
	return p
}

func InitEmptyTable_nameContext(p *Table_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_table_name
}

func (*Table_nameContext) IsTable_nameContext() {}

func NewTable_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_nameContext {
	var p = new(Table_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_table_name

	return p
}

func (s *Table_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(DSQLParserIDENTIFIER, 0)
}

func (s *Table_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_nameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterTable_name(s)
	}
}

func (s *Table_nameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitTable_name(s)
	}
}

func (p *DSQLParser) Table_name() (localctx ITable_nameContext) {
	localctx = NewTable_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, DSQLParserRULE_table_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(134)
		p.Match(DSQLParserIDENTIFIER)
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

// ILiteral_valueContext is an interface to support dynamic dispatch.
type ILiteral_valueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMERIC_LITERAL() antlr.TerminalNode
	STRING_LITERAL() antlr.TerminalNode
	K_NULL() antlr.TerminalNode

	// IsLiteral_valueContext differentiates from other interfaces.
	IsLiteral_valueContext()
}

type Literal_valueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteral_valueContext() *Literal_valueContext {
	var p = new(Literal_valueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_literal_value
	return p
}

func InitEmptyLiteral_valueContext(p *Literal_valueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_literal_value
}

func (*Literal_valueContext) IsLiteral_valueContext() {}

func NewLiteral_valueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Literal_valueContext {
	var p = new(Literal_valueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_literal_value

	return p
}

func (s *Literal_valueContext) GetParser() antlr.Parser { return s.parser }

func (s *Literal_valueContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(DSQLParserNUMERIC_LITERAL, 0)
}

func (s *Literal_valueContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(DSQLParserSTRING_LITERAL, 0)
}

func (s *Literal_valueContext) K_NULL() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NULL, 0)
}

func (s *Literal_valueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Literal_valueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Literal_valueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterLiteral_value(s)
	}
}

func (s *Literal_valueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitLiteral_value(s)
	}
}

func (p *DSQLParser) Literal_value() (localctx ILiteral_valueContext) {
	localctx = NewLiteral_valueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, DSQLParserRULE_literal_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(136)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&13262859010048) != 0) {
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

// IUnary_operatorContext is an interface to support dynamic dispatch.
type IUnary_operatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_NOT() antlr.TerminalNode

	// IsUnary_operatorContext differentiates from other interfaces.
	IsUnary_operatorContext()
}

type Unary_operatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnary_operatorContext() *Unary_operatorContext {
	var p = new(Unary_operatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_unary_operator
	return p
}

func InitEmptyUnary_operatorContext(p *Unary_operatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_unary_operator
}

func (*Unary_operatorContext) IsUnary_operatorContext() {}

func NewUnary_operatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Unary_operatorContext {
	var p = new(Unary_operatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_unary_operator

	return p
}

func (s *Unary_operatorContext) GetParser() antlr.Parser { return s.parser }

func (s *Unary_operatorContext) K_NOT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NOT, 0)
}

func (s *Unary_operatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Unary_operatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Unary_operatorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterUnary_operator(s)
	}
}

func (s *Unary_operatorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitUnary_operator(s)
	}
}

func (p *DSQLParser) Unary_operator() (localctx IUnary_operatorContext) {
	localctx = NewUnary_operatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, DSQLParserRULE_unary_operator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(138)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&34359738592) != 0) {
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

// IBinary_operatorContext is an interface to support dynamic dispatch.
type IBinary_operatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_AND() antlr.TerminalNode
	K_OR() antlr.TerminalNode

	// IsBinary_operatorContext differentiates from other interfaces.
	IsBinary_operatorContext()
}

type Binary_operatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBinary_operatorContext() *Binary_operatorContext {
	var p = new(Binary_operatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_binary_operator
	return p
}

func InitEmptyBinary_operatorContext(p *Binary_operatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_binary_operator
}

func (*Binary_operatorContext) IsBinary_operatorContext() {}

func NewBinary_operatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Binary_operatorContext {
	var p = new(Binary_operatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_binary_operator

	return p
}

func (s *Binary_operatorContext) GetParser() antlr.Parser { return s.parser }

func (s *Binary_operatorContext) K_AND() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_AND, 0)
}

func (s *Binary_operatorContext) K_OR() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_OR, 0)
}

func (s *Binary_operatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Binary_operatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Binary_operatorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterBinary_operator(s)
	}
}

func (s *Binary_operatorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitBinary_operator(s)
	}
}

func (p *DSQLParser) Binary_operator() (localctx IBinary_operatorContext) {
	localctx = NewBinary_operatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, DSQLParserRULE_binary_operator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(140)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&824650497888) != 0) {
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

// IKeywordContext is an interface to support dynamic dispatch.
type IKeywordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_SELECT() antlr.TerminalNode
	K_FROM() antlr.TerminalNode
	K_WHERE() antlr.TerminalNode
	K_ORDER() antlr.TerminalNode
	K_BY() antlr.TerminalNode
	K_LIMIT() antlr.TerminalNode
	K_ASC() antlr.TerminalNode
	K_DESC() antlr.TerminalNode
	K_IS() antlr.TerminalNode
	K_NOT() antlr.TerminalNode
	K_NULL() antlr.TerminalNode
	K_LIKE() antlr.TerminalNode
	K_AND() antlr.TerminalNode
	K_OR() antlr.TerminalNode

	// IsKeywordContext differentiates from other interfaces.
	IsKeywordContext()
}

type KeywordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyKeywordContext() *KeywordContext {
	var p = new(KeywordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_keyword
	return p
}

func InitEmptyKeywordContext(p *KeywordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_keyword
}

func (*KeywordContext) IsKeywordContext() {}

func NewKeywordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *KeywordContext {
	var p = new(KeywordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_keyword

	return p
}

func (s *KeywordContext) GetParser() antlr.Parser { return s.parser }

func (s *KeywordContext) K_SELECT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_SELECT, 0)
}

func (s *KeywordContext) K_FROM() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_FROM, 0)
}

func (s *KeywordContext) K_WHERE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_WHERE, 0)
}

func (s *KeywordContext) K_ORDER() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_ORDER, 0)
}

func (s *KeywordContext) K_BY() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_BY, 0)
}

func (s *KeywordContext) K_LIMIT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_LIMIT, 0)
}

func (s *KeywordContext) K_ASC() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_ASC, 0)
}

func (s *KeywordContext) K_DESC() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_DESC, 0)
}

func (s *KeywordContext) K_IS() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_IS, 0)
}

func (s *KeywordContext) K_NOT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NOT, 0)
}

func (s *KeywordContext) K_NULL() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NULL, 0)
}

func (s *KeywordContext) K_LIKE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_LIKE, 0)
}

func (s *KeywordContext) K_AND() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_AND, 0)
}

func (s *KeywordContext) K_OR() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_OR, 0)
}

func (s *KeywordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *KeywordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *KeywordContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterKeyword(s)
	}
}

func (s *KeywordContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitKeyword(s)
	}
}

func (p *DSQLParser) Keyword() (localctx IKeywordContext) {
	localctx = NewKeywordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, DSQLParserRULE_keyword)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(142)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&1099444518912) != 0) {
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

func (p *DSQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 10:
		var t *Scalar_exprContext = nil
		if localctx != nil {
			t = localctx.(*Scalar_exprContext)
		}
		return p.Scalar_expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *DSQLParser) Scalar_expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 7)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 1)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 4)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
