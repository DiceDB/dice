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
		"", "';'", "','", "'('", "')'", "'.'", "'-'", "'+'", "'||'", "'/'",
		"'%'", "'<<'", "'>>'", "'&'", "'|'", "'<'", "'<='", "'>'", "'>='", "'='",
		"'=='", "'!='", "'<>'", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "'*'", "'$key'", "'$value'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "K_ALL", "K_AND", "K_ASC", "K_DESC", "K_DISTINCT",
		"K_FALSE", "K_FROM", "K_LIMIT", "K_NULL", "K_OR", "K_ORDER", "K_SELECT",
		"K_TRUE", "K_WHERE", "K_BY", "STAR", "KEY", "VALUE", "NUMERIC_LITERAL",
		"STRING_LITERAL", "BLOB_LITERAL", "SINGLE_LINE_COMMENT", "MULTILINE_COMMENT",
		"SPACES", "IDENTIFIER", "SIMPLE_IDENTIFIER", "UNEXPECTED_CHAR",
	}
	staticData.RuleNames = []string{
		"parse", "sql_stmt_list", "sql_stmt", "select_stmt", "select_clause",
		"from_clause", "where_clause", "order_by_clause", "limit_clause", "result_column",
		"table_name", "expr", "json_path", "json_path_part", "ordering_term",
		"unary_operator", "binary_operator", "function_name", "literal_value",
		"error",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 49, 174, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 1, 0, 1, 0, 5,
		0, 43, 8, 0, 10, 0, 12, 0, 46, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1, 51, 8, 1,
		10, 1, 12, 1, 54, 9, 1, 1, 1, 1, 1, 4, 1, 58, 8, 1, 11, 1, 12, 1, 59, 1,
		1, 5, 1, 63, 8, 1, 10, 1, 12, 1, 66, 9, 1, 1, 1, 5, 1, 69, 8, 1, 10, 1,
		12, 1, 72, 9, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 3, 3, 79, 8, 3, 1, 3, 3,
		3, 82, 8, 3, 1, 3, 3, 3, 85, 8, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5,
		1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 5, 7, 101, 8, 7, 10, 7,
		12, 7, 104, 9, 7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1,
		11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11,
		1, 11, 5, 11, 126, 8, 11, 10, 11, 12, 11, 129, 9, 11, 3, 11, 131, 8, 11,
		1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 3, 11, 139, 8, 11, 1, 11, 1,
		11, 1, 11, 1, 11, 5, 11, 145, 8, 11, 10, 11, 12, 11, 148, 9, 11, 1, 12,
		1, 12, 1, 12, 5, 12, 153, 8, 12, 10, 12, 12, 12, 156, 9, 12, 1, 13, 1,
		13, 1, 14, 1, 14, 3, 14, 162, 8, 14, 1, 15, 1, 15, 1, 16, 1, 16, 1, 17,
		1, 17, 1, 18, 1, 18, 1, 19, 1, 19, 1, 19, 0, 1, 22, 20, 0, 2, 4, 6, 8,
		10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 0, 6, 1, 0,
		38, 40, 2, 0, 42, 42, 47, 47, 1, 0, 25, 26, 1, 0, 6, 7, 4, 0, 6, 22, 24,
		24, 32, 32, 38, 38, 4, 0, 28, 28, 31, 31, 35, 35, 41, 43, 174, 0, 44, 1,
		0, 0, 0, 2, 52, 1, 0, 0, 0, 4, 73, 1, 0, 0, 0, 6, 75, 1, 0, 0, 0, 8, 86,
		1, 0, 0, 0, 10, 89, 1, 0, 0, 0, 12, 92, 1, 0, 0, 0, 14, 95, 1, 0, 0, 0,
		16, 105, 1, 0, 0, 0, 18, 108, 1, 0, 0, 0, 20, 110, 1, 0, 0, 0, 22, 138,
		1, 0, 0, 0, 24, 149, 1, 0, 0, 0, 26, 157, 1, 0, 0, 0, 28, 159, 1, 0, 0,
		0, 30, 163, 1, 0, 0, 0, 32, 165, 1, 0, 0, 0, 34, 167, 1, 0, 0, 0, 36, 169,
		1, 0, 0, 0, 38, 171, 1, 0, 0, 0, 40, 43, 3, 2, 1, 0, 41, 43, 3, 38, 19,
		0, 42, 40, 1, 0, 0, 0, 42, 41, 1, 0, 0, 0, 43, 46, 1, 0, 0, 0, 44, 42,
		1, 0, 0, 0, 44, 45, 1, 0, 0, 0, 45, 47, 1, 0, 0, 0, 46, 44, 1, 0, 0, 0,
		47, 48, 5, 0, 0, 1, 48, 1, 1, 0, 0, 0, 49, 51, 5, 1, 0, 0, 50, 49, 1, 0,
		0, 0, 51, 54, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0, 53, 55,
		1, 0, 0, 0, 54, 52, 1, 0, 0, 0, 55, 64, 3, 4, 2, 0, 56, 58, 5, 1, 0, 0,
		57, 56, 1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 59, 60, 1,
		0, 0, 0, 60, 61, 1, 0, 0, 0, 61, 63, 3, 4, 2, 0, 62, 57, 1, 0, 0, 0, 63,
		66, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 70, 1, 0, 0,
		0, 66, 64, 1, 0, 0, 0, 67, 69, 5, 1, 0, 0, 68, 67, 1, 0, 0, 0, 69, 72,
		1, 0, 0, 0, 70, 68, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 3, 1, 0, 0, 0,
		72, 70, 1, 0, 0, 0, 73, 74, 3, 6, 3, 0, 74, 5, 1, 0, 0, 0, 75, 76, 3, 8,
		4, 0, 76, 78, 3, 10, 5, 0, 77, 79, 3, 12, 6, 0, 78, 77, 1, 0, 0, 0, 78,
		79, 1, 0, 0, 0, 79, 81, 1, 0, 0, 0, 80, 82, 3, 14, 7, 0, 81, 80, 1, 0,
		0, 0, 81, 82, 1, 0, 0, 0, 82, 84, 1, 0, 0, 0, 83, 85, 3, 16, 8, 0, 84,
		83, 1, 0, 0, 0, 84, 85, 1, 0, 0, 0, 85, 7, 1, 0, 0, 0, 86, 87, 5, 34, 0,
		0, 87, 88, 3, 18, 9, 0, 88, 9, 1, 0, 0, 0, 89, 90, 5, 29, 0, 0, 90, 91,
		3, 20, 10, 0, 91, 11, 1, 0, 0, 0, 92, 93, 5, 36, 0, 0, 93, 94, 3, 22, 11,
		0, 94, 13, 1, 0, 0, 0, 95, 96, 5, 33, 0, 0, 96, 97, 5, 37, 0, 0, 97, 102,
		3, 28, 14, 0, 98, 99, 5, 2, 0, 0, 99, 101, 3, 28, 14, 0, 100, 98, 1, 0,
		0, 0, 101, 104, 1, 0, 0, 0, 102, 100, 1, 0, 0, 0, 102, 103, 1, 0, 0, 0,
		103, 15, 1, 0, 0, 0, 104, 102, 1, 0, 0, 0, 105, 106, 5, 30, 0, 0, 106,
		107, 5, 41, 0, 0, 107, 17, 1, 0, 0, 0, 108, 109, 7, 0, 0, 0, 109, 19, 1,
		0, 0, 0, 110, 111, 7, 1, 0, 0, 111, 21, 1, 0, 0, 0, 112, 113, 6, 11, -1,
		0, 113, 139, 3, 36, 18, 0, 114, 139, 5, 39, 0, 0, 115, 139, 5, 40, 0, 0,
		116, 139, 3, 24, 12, 0, 117, 118, 3, 30, 15, 0, 118, 119, 3, 22, 11, 4,
		119, 139, 1, 0, 0, 0, 120, 121, 3, 34, 17, 0, 121, 130, 5, 3, 0, 0, 122,
		127, 3, 22, 11, 0, 123, 124, 5, 2, 0, 0, 124, 126, 3, 22, 11, 0, 125, 123,
		1, 0, 0, 0, 126, 129, 1, 0, 0, 0, 127, 125, 1, 0, 0, 0, 127, 128, 1, 0,
		0, 0, 128, 131, 1, 0, 0, 0, 129, 127, 1, 0, 0, 0, 130, 122, 1, 0, 0, 0,
		130, 131, 1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132, 133, 5, 4, 0, 0, 133,
		139, 1, 0, 0, 0, 134, 135, 5, 3, 0, 0, 135, 136, 3, 22, 11, 0, 136, 137,
		5, 4, 0, 0, 137, 139, 1, 0, 0, 0, 138, 112, 1, 0, 0, 0, 138, 114, 1, 0,
		0, 0, 138, 115, 1, 0, 0, 0, 138, 116, 1, 0, 0, 0, 138, 117, 1, 0, 0, 0,
		138, 120, 1, 0, 0, 0, 138, 134, 1, 0, 0, 0, 139, 146, 1, 0, 0, 0, 140,
		141, 10, 3, 0, 0, 141, 142, 3, 32, 16, 0, 142, 143, 3, 22, 11, 4, 143,
		145, 1, 0, 0, 0, 144, 140, 1, 0, 0, 0, 145, 148, 1, 0, 0, 0, 146, 144,
		1, 0, 0, 0, 146, 147, 1, 0, 0, 0, 147, 23, 1, 0, 0, 0, 148, 146, 1, 0,
		0, 0, 149, 154, 3, 26, 13, 0, 150, 151, 5, 5, 0, 0, 151, 153, 3, 26, 13,
		0, 152, 150, 1, 0, 0, 0, 153, 156, 1, 0, 0, 0, 154, 152, 1, 0, 0, 0, 154,
		155, 1, 0, 0, 0, 155, 25, 1, 0, 0, 0, 156, 154, 1, 0, 0, 0, 157, 158, 5,
		47, 0, 0, 158, 27, 1, 0, 0, 0, 159, 161, 3, 22, 11, 0, 160, 162, 7, 2,
		0, 0, 161, 160, 1, 0, 0, 0, 161, 162, 1, 0, 0, 0, 162, 29, 1, 0, 0, 0,
		163, 164, 7, 3, 0, 0, 164, 31, 1, 0, 0, 0, 165, 166, 7, 4, 0, 0, 166, 33,
		1, 0, 0, 0, 167, 168, 5, 48, 0, 0, 168, 35, 1, 0, 0, 0, 169, 170, 7, 5,
		0, 0, 170, 37, 1, 0, 0, 0, 171, 172, 5, 49, 0, 0, 172, 39, 1, 0, 0, 0,
		16, 42, 44, 52, 59, 64, 70, 78, 81, 84, 102, 127, 130, 138, 146, 154, 161,
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
	DSQLParserEOF                 = antlr.TokenEOF
	DSQLParserT__0                = 1
	DSQLParserT__1                = 2
	DSQLParserT__2                = 3
	DSQLParserT__3                = 4
	DSQLParserT__4                = 5
	DSQLParserT__5                = 6
	DSQLParserT__6                = 7
	DSQLParserT__7                = 8
	DSQLParserT__8                = 9
	DSQLParserT__9                = 10
	DSQLParserT__10               = 11
	DSQLParserT__11               = 12
	DSQLParserT__12               = 13
	DSQLParserT__13               = 14
	DSQLParserT__14               = 15
	DSQLParserT__15               = 16
	DSQLParserT__16               = 17
	DSQLParserT__17               = 18
	DSQLParserT__18               = 19
	DSQLParserT__19               = 20
	DSQLParserT__20               = 21
	DSQLParserT__21               = 22
	DSQLParserK_ALL               = 23
	DSQLParserK_AND               = 24
	DSQLParserK_ASC               = 25
	DSQLParserK_DESC              = 26
	DSQLParserK_DISTINCT          = 27
	DSQLParserK_FALSE             = 28
	DSQLParserK_FROM              = 29
	DSQLParserK_LIMIT             = 30
	DSQLParserK_NULL              = 31
	DSQLParserK_OR                = 32
	DSQLParserK_ORDER             = 33
	DSQLParserK_SELECT            = 34
	DSQLParserK_TRUE              = 35
	DSQLParserK_WHERE             = 36
	DSQLParserK_BY                = 37
	DSQLParserSTAR                = 38
	DSQLParserKEY                 = 39
	DSQLParserVALUE               = 40
	DSQLParserNUMERIC_LITERAL     = 41
	DSQLParserSTRING_LITERAL      = 42
	DSQLParserBLOB_LITERAL        = 43
	DSQLParserSINGLE_LINE_COMMENT = 44
	DSQLParserMULTILINE_COMMENT   = 45
	DSQLParserSPACES              = 46
	DSQLParserIDENTIFIER          = 47
	DSQLParserSIMPLE_IDENTIFIER   = 48
	DSQLParserUNEXPECTED_CHAR     = 49
)

// DSQLParser rules.
const (
	DSQLParserRULE_parse           = 0
	DSQLParserRULE_sql_stmt_list   = 1
	DSQLParserRULE_sql_stmt        = 2
	DSQLParserRULE_select_stmt     = 3
	DSQLParserRULE_select_clause   = 4
	DSQLParserRULE_from_clause     = 5
	DSQLParserRULE_where_clause    = 6
	DSQLParserRULE_order_by_clause = 7
	DSQLParserRULE_limit_clause    = 8
	DSQLParserRULE_result_column   = 9
	DSQLParserRULE_table_name      = 10
	DSQLParserRULE_expr            = 11
	DSQLParserRULE_json_path       = 12
	DSQLParserRULE_json_path_part  = 13
	DSQLParserRULE_ordering_term   = 14
	DSQLParserRULE_unary_operator  = 15
	DSQLParserRULE_binary_operator = 16
	DSQLParserRULE_function_name   = 17
	DSQLParserRULE_literal_value   = 18
	DSQLParserRULE_error           = 19
)

// IParseContext is an interface to support dynamic dispatch.
type IParseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllSql_stmt_list() []ISql_stmt_listContext
	Sql_stmt_list(i int) ISql_stmt_listContext
	AllError_() []IErrorContext
	Error_(i int) IErrorContext

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

func (s *ParseContext) EOF() antlr.TerminalNode {
	return s.GetToken(DSQLParserEOF, 0)
}

func (s *ParseContext) AllSql_stmt_list() []ISql_stmt_listContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISql_stmt_listContext); ok {
			len++
		}
	}

	tst := make([]ISql_stmt_listContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISql_stmt_listContext); ok {
			tst[i] = t.(ISql_stmt_listContext)
			i++
		}
	}

	return tst
}

func (s *ParseContext) Sql_stmt_list(i int) ISql_stmt_listContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmt_listContext); ok {
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

	return t.(ISql_stmt_listContext)
}

func (s *ParseContext) AllError_() []IErrorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IErrorContext); ok {
			len++
		}
	}

	tst := make([]IErrorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IErrorContext); ok {
			tst[i] = t.(IErrorContext)
			i++
		}
	}

	return tst
}

func (s *ParseContext) Error_(i int) IErrorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IErrorContext); ok {
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

	return t.(IErrorContext)
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
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&562967133290498) != 0 {
		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case DSQLParserT__0, DSQLParserK_SELECT:
			{
				p.SetState(40)
				p.Sql_stmt_list()
			}

		case DSQLParserUNEXPECTED_CHAR:
			{
				p.SetState(41)
				p.Error_()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(46)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(47)
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
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DSQLParserT__0 {
		{
			p.SetState(49)
			p.Match(DSQLParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(55)
		p.Sql_stmt()
	}
	p.SetState(64)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			p.SetState(57)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for ok := true; ok; ok = _la == DSQLParserT__0 {
				{
					p.SetState(56)
					p.Match(DSQLParserT__0)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(59)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(61)
				p.Sql_stmt()
			}

		}
		p.SetState(66)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(67)
				p.Match(DSQLParserT__0)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(72)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
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
		p.SetState(73)
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
	Select_clause() ISelect_clauseContext
	From_clause() IFrom_clauseContext
	Where_clause() IWhere_clauseContext
	Order_by_clause() IOrder_by_clauseContext
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

func (s *Select_stmtContext) Select_clause() ISelect_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_clauseContext)
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

func (s *Select_stmtContext) Order_by_clause() IOrder_by_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrder_by_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrder_by_clauseContext)
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
		p.SetState(75)
		p.Select_clause()
	}
	{
		p.SetState(76)
		p.From_clause()
	}
	p.SetState(78)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_WHERE {
		{
			p.SetState(77)
			p.Where_clause()
		}

	}
	p.SetState(81)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_ORDER {
		{
			p.SetState(80)
			p.Order_by_clause()
		}

	}
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_LIMIT {
		{
			p.SetState(83)
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

// ISelect_clauseContext is an interface to support dynamic dispatch.
type ISelect_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_SELECT() antlr.TerminalNode
	Result_column() IResult_columnContext

	// IsSelect_clauseContext differentiates from other interfaces.
	IsSelect_clauseContext()
}

type Select_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelect_clauseContext() *Select_clauseContext {
	var p = new(Select_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_clause
	return p
}

func InitEmptySelect_clauseContext(p *Select_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_select_clause
}

func (*Select_clauseContext) IsSelect_clauseContext() {}

func NewSelect_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_clauseContext {
	var p = new(Select_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_select_clause

	return p
}

func (s *Select_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_clauseContext) K_SELECT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_SELECT, 0)
}

func (s *Select_clauseContext) Result_column() IResult_columnContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IResult_columnContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IResult_columnContext)
}

func (s *Select_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterSelect_clause(s)
	}
}

func (s *Select_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitSelect_clause(s)
	}
}

func (p *DSQLParser) Select_clause() (localctx ISelect_clauseContext) {
	localctx = NewSelect_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, DSQLParserRULE_select_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(86)
		p.Match(DSQLParserK_SELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(87)
		p.Result_column()
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
	K_FROM() antlr.TerminalNode
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

func (s *From_clauseContext) K_FROM() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_FROM, 0)
}

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
		p.SetState(89)
		p.Match(DSQLParserK_FROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(90)
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
	K_WHERE() antlr.TerminalNode
	Expr() IExprContext

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

func (s *Where_clauseContext) K_WHERE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_WHERE, 0)
}

func (s *Where_clauseContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
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
		p.SetState(92)
		p.Match(DSQLParserK_WHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(93)
		p.expr(0)
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

// IOrder_by_clauseContext is an interface to support dynamic dispatch.
type IOrder_by_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_ordering_term returns the _ordering_term rule contexts.
	Get_ordering_term() IOrdering_termContext

	// Set_ordering_term sets the _ordering_term rule contexts.
	Set_ordering_term(IOrdering_termContext)

	// GetOrdering_terms returns the ordering_terms rule context list.
	GetOrdering_terms() []IOrdering_termContext

	// SetOrdering_terms sets the ordering_terms rule context list.
	SetOrdering_terms([]IOrdering_termContext)

	// Getter signatures
	K_ORDER() antlr.TerminalNode
	K_BY() antlr.TerminalNode
	AllOrdering_term() []IOrdering_termContext
	Ordering_term(i int) IOrdering_termContext

	// IsOrder_by_clauseContext differentiates from other interfaces.
	IsOrder_by_clauseContext()
}

type Order_by_clauseContext struct {
	antlr.BaseParserRuleContext
	parser         antlr.Parser
	_ordering_term IOrdering_termContext
	ordering_terms []IOrdering_termContext
}

func NewEmptyOrder_by_clauseContext() *Order_by_clauseContext {
	var p = new(Order_by_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_order_by_clause
	return p
}

func InitEmptyOrder_by_clauseContext(p *Order_by_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_order_by_clause
}

func (*Order_by_clauseContext) IsOrder_by_clauseContext() {}

func NewOrder_by_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Order_by_clauseContext {
	var p = new(Order_by_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_order_by_clause

	return p
}

func (s *Order_by_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Order_by_clauseContext) Get_ordering_term() IOrdering_termContext { return s._ordering_term }

func (s *Order_by_clauseContext) Set_ordering_term(v IOrdering_termContext) { s._ordering_term = v }

func (s *Order_by_clauseContext) GetOrdering_terms() []IOrdering_termContext { return s.ordering_terms }

func (s *Order_by_clauseContext) SetOrdering_terms(v []IOrdering_termContext) { s.ordering_terms = v }

func (s *Order_by_clauseContext) K_ORDER() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_ORDER, 0)
}

func (s *Order_by_clauseContext) K_BY() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_BY, 0)
}

func (s *Order_by_clauseContext) AllOrdering_term() []IOrdering_termContext {
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

func (s *Order_by_clauseContext) Ordering_term(i int) IOrdering_termContext {
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

func (s *Order_by_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Order_by_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Order_by_clauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterOrder_by_clause(s)
	}
}

func (s *Order_by_clauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitOrder_by_clause(s)
	}
}

func (p *DSQLParser) Order_by_clause() (localctx IOrder_by_clauseContext) {
	localctx = NewOrder_by_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, DSQLParserRULE_order_by_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(95)
		p.Match(DSQLParserK_ORDER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(96)
		p.Match(DSQLParserK_BY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(97)

		var _x = p.Ordering_term()

		localctx.(*Order_by_clauseContext)._ordering_term = _x
	}
	localctx.(*Order_by_clauseContext).ordering_terms = append(localctx.(*Order_by_clauseContext).ordering_terms, localctx.(*Order_by_clauseContext)._ordering_term)
	p.SetState(102)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DSQLParserT__1 {
		{
			p.SetState(98)
			p.Match(DSQLParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(99)

			var _x = p.Ordering_term()

			localctx.(*Order_by_clauseContext)._ordering_term = _x
		}
		localctx.(*Order_by_clauseContext).ordering_terms = append(localctx.(*Order_by_clauseContext).ordering_terms, localctx.(*Order_by_clauseContext)._ordering_term)

		p.SetState(104)
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

// ILimit_clauseContext is an interface to support dynamic dispatch.
type ILimit_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	K_LIMIT() antlr.TerminalNode
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

func (s *Limit_clauseContext) K_LIMIT() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_LIMIT, 0)
}

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
	p.EnterRule(localctx, 16, DSQLParserRULE_limit_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(105)
		p.Match(DSQLParserK_LIMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(106)
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

// IResult_columnContext is an interface to support dynamic dispatch.
type IResult_columnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STAR() antlr.TerminalNode
	KEY() antlr.TerminalNode
	VALUE() antlr.TerminalNode

	// IsResult_columnContext differentiates from other interfaces.
	IsResult_columnContext()
}

type Result_columnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyResult_columnContext() *Result_columnContext {
	var p = new(Result_columnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_result_column
	return p
}

func InitEmptyResult_columnContext(p *Result_columnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_result_column
}

func (*Result_columnContext) IsResult_columnContext() {}

func NewResult_columnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Result_columnContext {
	var p = new(Result_columnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_result_column

	return p
}

func (s *Result_columnContext) GetParser() antlr.Parser { return s.parser }

func (s *Result_columnContext) STAR() antlr.TerminalNode {
	return s.GetToken(DSQLParserSTAR, 0)
}

func (s *Result_columnContext) KEY() antlr.TerminalNode {
	return s.GetToken(DSQLParserKEY, 0)
}

func (s *Result_columnContext) VALUE() antlr.TerminalNode {
	return s.GetToken(DSQLParserVALUE, 0)
}

func (s *Result_columnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Result_columnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Result_columnContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterResult_column(s)
	}
}

func (s *Result_columnContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitResult_column(s)
	}
}

func (p *DSQLParser) Result_column() (localctx IResult_columnContext) {
	localctx = NewResult_columnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, DSQLParserRULE_result_column)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(108)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&1924145348608) != 0) {
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

// ITable_nameContext is an interface to support dynamic dispatch.
type ITable_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	STRING_LITERAL() antlr.TerminalNode

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

func (s *Table_nameContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(DSQLParserSTRING_LITERAL, 0)
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
	p.EnterRule(localctx, 20, DSQLParserRULE_table_name)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(110)
		_la = p.GetTokenStream().LA(1)

		if !(_la == DSQLParserSTRING_LITERAL || _la == DSQLParserIDENTIFIER) {
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

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Literal_value() ILiteral_valueContext
	KEY() antlr.TerminalNode
	VALUE() antlr.TerminalNode
	Json_path() IJson_pathContext
	Unary_operator() IUnary_operatorContext
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	Function_name() IFunction_nameContext
	Binary_operator() IBinary_operatorContext

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) Literal_value() ILiteral_valueContext {
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

func (s *ExprContext) KEY() antlr.TerminalNode {
	return s.GetToken(DSQLParserKEY, 0)
}

func (s *ExprContext) VALUE() antlr.TerminalNode {
	return s.GetToken(DSQLParserVALUE, 0)
}

func (s *ExprContext) Json_path() IJson_pathContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJson_pathContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IJson_pathContext)
}

func (s *ExprContext) Unary_operator() IUnary_operatorContext {
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

func (s *ExprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *ExprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *ExprContext) Function_name() IFunction_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunction_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunction_nameContext)
}

func (s *ExprContext) Binary_operator() IBinary_operatorContext {
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

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (p *DSQLParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *DSQLParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 22
	p.EnterRecursionRule(localctx, 22, DSQLParserRULE_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case DSQLParserK_FALSE, DSQLParserK_NULL, DSQLParserK_TRUE, DSQLParserNUMERIC_LITERAL, DSQLParserSTRING_LITERAL, DSQLParserBLOB_LITERAL:
		{
			p.SetState(113)
			p.Literal_value()
		}

	case DSQLParserKEY:
		{
			p.SetState(114)
			p.Match(DSQLParserKEY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserVALUE:
		{
			p.SetState(115)
			p.Match(DSQLParserVALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserIDENTIFIER:
		{
			p.SetState(116)
			p.Json_path()
		}

	case DSQLParserT__5, DSQLParserT__6:
		{
			p.SetState(117)
			p.Unary_operator()
		}
		{
			p.SetState(118)
			p.expr(4)
		}

	case DSQLParserSIMPLE_IDENTIFIER:
		{
			p.SetState(120)
			p.Function_name()
		}
		{
			p.SetState(121)
			p.Match(DSQLParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(130)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&439291670954184) != 0 {
			{
				p.SetState(122)
				p.expr(0)
			}
			p.SetState(127)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == DSQLParserT__1 {
				{
					p.SetState(123)
					p.Match(DSQLParserT__1)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(124)
					p.expr(0)
				}

				p.SetState(129)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(132)
			p.Match(DSQLParserT__3)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DSQLParserT__2:
		{
			p.SetState(134)
			p.Match(DSQLParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(135)
			p.expr(0)
		}
		{
			p.SetState(136)
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
	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewExprContext(p, _parentctx, _parentState)
			p.PushNewRecursionContext(localctx, _startState, DSQLParserRULE_expr)
			p.SetState(140)

			if !(p.Precpred(p.GetParserRuleContext(), 3)) {
				p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
				goto errorExit
			}
			{
				p.SetState(141)
				p.Binary_operator()
			}
			{
				p.SetState(142)
				p.expr(4)
			}

		}
		p.SetState(148)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
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

// IJson_pathContext is an interface to support dynamic dispatch.
type IJson_pathContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_json_path_part returns the _json_path_part rule contexts.
	Get_json_path_part() IJson_path_partContext

	// Set_json_path_part sets the _json_path_part rule contexts.
	Set_json_path_part(IJson_path_partContext)

	// GetJson_path_parts returns the json_path_parts rule context list.
	GetJson_path_parts() []IJson_path_partContext

	// SetJson_path_parts sets the json_path_parts rule context list.
	SetJson_path_parts([]IJson_path_partContext)

	// Getter signatures
	AllJson_path_part() []IJson_path_partContext
	Json_path_part(i int) IJson_path_partContext

	// IsJson_pathContext differentiates from other interfaces.
	IsJson_pathContext()
}

type Json_pathContext struct {
	antlr.BaseParserRuleContext
	parser          antlr.Parser
	_json_path_part IJson_path_partContext
	json_path_parts []IJson_path_partContext
}

func NewEmptyJson_pathContext() *Json_pathContext {
	var p = new(Json_pathContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_json_path
	return p
}

func InitEmptyJson_pathContext(p *Json_pathContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_json_path
}

func (*Json_pathContext) IsJson_pathContext() {}

func NewJson_pathContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Json_pathContext {
	var p = new(Json_pathContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_json_path

	return p
}

func (s *Json_pathContext) GetParser() antlr.Parser { return s.parser }

func (s *Json_pathContext) Get_json_path_part() IJson_path_partContext { return s._json_path_part }

func (s *Json_pathContext) Set_json_path_part(v IJson_path_partContext) { s._json_path_part = v }

func (s *Json_pathContext) GetJson_path_parts() []IJson_path_partContext { return s.json_path_parts }

func (s *Json_pathContext) SetJson_path_parts(v []IJson_path_partContext) { s.json_path_parts = v }

func (s *Json_pathContext) AllJson_path_part() []IJson_path_partContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IJson_path_partContext); ok {
			len++
		}
	}

	tst := make([]IJson_path_partContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IJson_path_partContext); ok {
			tst[i] = t.(IJson_path_partContext)
			i++
		}
	}

	return tst
}

func (s *Json_pathContext) Json_path_part(i int) IJson_path_partContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJson_path_partContext); ok {
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

	return t.(IJson_path_partContext)
}

func (s *Json_pathContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Json_pathContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Json_pathContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterJson_path(s)
	}
}

func (s *Json_pathContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitJson_path(s)
	}
}

func (p *DSQLParser) Json_path() (localctx IJson_pathContext) {
	localctx = NewJson_pathContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, DSQLParserRULE_json_path)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(149)

		var _x = p.Json_path_part()

		localctx.(*Json_pathContext)._json_path_part = _x
	}
	localctx.(*Json_pathContext).json_path_parts = append(localctx.(*Json_pathContext).json_path_parts, localctx.(*Json_pathContext)._json_path_part)
	p.SetState(154)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 14, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(150)
				p.Match(DSQLParserT__4)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(151)

				var _x = p.Json_path_part()

				localctx.(*Json_pathContext)._json_path_part = _x
			}
			localctx.(*Json_pathContext).json_path_parts = append(localctx.(*Json_pathContext).json_path_parts, localctx.(*Json_pathContext)._json_path_part)

		}
		p.SetState(156)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 14, p.GetParserRuleContext())
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
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IJson_path_partContext is an interface to support dynamic dispatch.
type IJson_path_partContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsJson_path_partContext differentiates from other interfaces.
	IsJson_path_partContext()
}

type Json_path_partContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyJson_path_partContext() *Json_path_partContext {
	var p = new(Json_path_partContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_json_path_part
	return p
}

func InitEmptyJson_path_partContext(p *Json_path_partContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_json_path_part
}

func (*Json_path_partContext) IsJson_path_partContext() {}

func NewJson_path_partContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Json_path_partContext {
	var p = new(Json_path_partContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_json_path_part

	return p
}

func (s *Json_path_partContext) GetParser() antlr.Parser { return s.parser }

func (s *Json_path_partContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(DSQLParserIDENTIFIER, 0)
}

func (s *Json_path_partContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Json_path_partContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Json_path_partContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterJson_path_part(s)
	}
}

func (s *Json_path_partContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitJson_path_part(s)
	}
}

func (p *DSQLParser) Json_path_part() (localctx IJson_path_partContext) {
	localctx = NewJson_path_partContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, DSQLParserRULE_json_path_part)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(157)
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

// IOrdering_termContext is an interface to support dynamic dispatch.
type IOrdering_termContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expr() IExprContext
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

func (s *Ordering_termContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
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
	p.EnterRule(localctx, 28, DSQLParserRULE_ordering_term)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(159)
		p.expr(0)
	}
	p.SetState(161)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DSQLParserK_ASC || _la == DSQLParserK_DESC {
		{
			p.SetState(160)
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

// IUnary_operatorContext is an interface to support dynamic dispatch.
type IUnary_operatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
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
	p.EnterRule(localctx, 30, DSQLParserRULE_unary_operator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(163)
		_la = p.GetTokenStream().LA(1)

		if !(_la == DSQLParserT__5 || _la == DSQLParserT__6) {
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
	STAR() antlr.TerminalNode
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

func (s *Binary_operatorContext) STAR() antlr.TerminalNode {
	return s.GetToken(DSQLParserSTAR, 0)
}

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
	p.EnterRule(localctx, 32, DSQLParserRULE_binary_operator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(165)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&279198040000) != 0) {
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

// IFunction_nameContext is an interface to support dynamic dispatch.
type IFunction_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SIMPLE_IDENTIFIER() antlr.TerminalNode

	// IsFunction_nameContext differentiates from other interfaces.
	IsFunction_nameContext()
}

type Function_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunction_nameContext() *Function_nameContext {
	var p = new(Function_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_function_name
	return p
}

func InitEmptyFunction_nameContext(p *Function_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_function_name
}

func (*Function_nameContext) IsFunction_nameContext() {}

func NewFunction_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Function_nameContext {
	var p = new(Function_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_function_name

	return p
}

func (s *Function_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Function_nameContext) SIMPLE_IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(DSQLParserSIMPLE_IDENTIFIER, 0)
}

func (s *Function_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Function_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Function_nameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterFunction_name(s)
	}
}

func (s *Function_nameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitFunction_name(s)
	}
}

func (p *DSQLParser) Function_name() (localctx IFunction_nameContext) {
	localctx = NewFunction_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, DSQLParserRULE_function_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(167)
		p.Match(DSQLParserSIMPLE_IDENTIFIER)
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
	BLOB_LITERAL() antlr.TerminalNode
	K_NULL() antlr.TerminalNode
	K_TRUE() antlr.TerminalNode
	K_FALSE() antlr.TerminalNode

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

func (s *Literal_valueContext) BLOB_LITERAL() antlr.TerminalNode {
	return s.GetToken(DSQLParserBLOB_LITERAL, 0)
}

func (s *Literal_valueContext) K_NULL() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_NULL, 0)
}

func (s *Literal_valueContext) K_TRUE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_TRUE, 0)
}

func (s *Literal_valueContext) K_FALSE() antlr.TerminalNode {
	return s.GetToken(DSQLParserK_FALSE, 0)
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
	p.EnterRule(localctx, 36, DSQLParserRULE_literal_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(169)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&15429938446336) != 0) {
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

// IErrorContext is an interface to support dynamic dispatch.
type IErrorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UNEXPECTED_CHAR() antlr.TerminalNode

	// IsErrorContext differentiates from other interfaces.
	IsErrorContext()
}

type ErrorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyErrorContext() *ErrorContext {
	var p = new(ErrorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_error
	return p
}

func InitEmptyErrorContext(p *ErrorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DSQLParserRULE_error
}

func (*ErrorContext) IsErrorContext() {}

func NewErrorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ErrorContext {
	var p = new(ErrorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DSQLParserRULE_error

	return p
}

func (s *ErrorContext) GetParser() antlr.Parser { return s.parser }

func (s *ErrorContext) UNEXPECTED_CHAR() antlr.TerminalNode {
	return s.GetToken(DSQLParserUNEXPECTED_CHAR, 0)
}

func (s *ErrorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ErrorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ErrorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.EnterError(s)
	}
}

func (s *ErrorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DSQLListener); ok {
		listenerT.ExitError(s)
	}
}

func (p *DSQLParser) Error_() (localctx IErrorContext) {
	localctx = NewErrorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, DSQLParserRULE_error)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(171)
		p.Match(DSQLParserUNEXPECTED_CHAR)
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

func (p *DSQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 11:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *DSQLParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
