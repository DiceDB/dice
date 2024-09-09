grammar DSQL;

// Root rule to parse multiple SQL statements or errors
parse
 : ( sql_stmt_list | error )* EOF
 ;

// List of SQL statements separated by semicolons
sql_stmt_list
 : ';'* sql_stmt ( ';'+ sql_stmt )* ';'*
 ;

// A single SQL statement (currently only supporting SELECT statements)
sql_stmt
 : select_stmt
 ;

// The SELECT statement with sub-rules for each clause
select_stmt
 : select_clause from_clause where_clause? order_by_clause? limit_clause?
 ;

// The SELECT clause
select_clause
 : K_SELECT result_column
 ;

// The FROM clause
from_clause
 : K_FROM table_name
 ;

// Optional WHERE clause
where_clause
 : K_WHERE expr
 ;

// Optional ORDER BY clause
order_by_clause
 : K_ORDER K_BY ordering_terms+=ordering_term ( ',' ordering_terms+=ordering_term)*
 ;

// Optional LIMIT clause
limit_clause
 : K_LIMIT NUMERIC_LITERAL
 ;

// Defines what can be selected: all columns (*), KEY, or VALUE
result_column
 : STAR
 | KEY
 | VALUE
 ;

// Table names, which can be an identifier or a string literal
table_name
 : IDENTIFIER
 | STRING_LITERAL
 ;

// Expressions used in WHERE, ORDER BY, and other clauses
expr
 : literal_value    #LiteralValue
 | KEY              #Key
 | VALUE            #Value
 | json_path        #JsonPathStmt
 | unary_expr       #UnaryExprStmt
 | binary_expr      #BinaryExprStmt
 | '(' expr ')'     #ParenthesizedExpr
 | function_call    #FunctionCallStmt
 ;

unary_expr
 : unary_operator expr
 ;

binary_expr
 : expr '||' expr
 ;

function_call
 : function_name '(' ( arg_list+=expr ( ',' arg_list+=expr )* )? ')'
 ;

// JSONPath-like access starting with $value
json_path
 : json_path_parts+=json_path_part ( '.' json_path_parts+=json_path_part )*
 ;

// Part of the JSON path, can be a field or a wildcard (*)
json_path_part
 : IDENTIFIER                   // Accessing a field
 ;

// Ordering term with optional ASC or DESC
ordering_term
 : expr ( K_ASC | K_DESC )?
 ;

// Unary operators like + and -
unary_operator
 : '-' | '+'
 ;

// Binary operators for string concatenation, arithmetic, bitwise operations, and comparisons
binary_operator
 : '||'                           // String concatenation
 | '*' | '/' | '%'                // Multiplication, division, modulus
 | '+' | '-'                      // Addition, subtraction
 | '<<' | '>>' | '&' | '|'        // Bitwise shift and bitwise operations
 | '<' | '<=' | '>' | '>='        // Comparison operators
 | '=' | '==' | '!=' | '<>'       // Equality and inequality operators
 | K_AND | K_OR                   // Logical AND and OR
 ;

// Function names, which can be an identifier
function_name
 : SIMPLE_IDENTIFIER
 ;

// Literal values for numbers, strings, blobs, true/false
literal_value
 : NUMERIC_LITERAL
 | STRING_LITERAL
 | BLOB_LITERAL
 | K_NULL
 | K_TRUE
 | K_FALSE
 ;

// Keywords (case-insensitive)
K_ALL : A L L;
K_AND : A N D;
K_ASC : A S C;
K_DESC : D E S C;
K_DISTINCT : D I S T I N C T;
K_FALSE : F A L S E;
K_FROM : F R O M;
K_LIMIT : L I M I T;
K_NULL : N U L L;
K_OR : O R;
K_ORDER : O R D E R;
K_SELECT : S E L E C T;
K_TRUE : T R U E;
K_WHERE : W H E R E;
K_BY : B Y;

// Special identifiers for $key and $value
STAR: '*';
KEY : '$key';
VALUE : '$value';

// Numeric literals with optional decimals and exponents
NUMERIC_LITERAL
 : DIGIT+ ( '.' DIGIT* )? ( E [-+]? DIGIT+ )?
 | '.' DIGIT+ ( E [-+]? DIGIT+ )?
 ;

// String literals enclosed in single quotes
STRING_LITERAL
 : '\'' ( ~'\'' | '\'\'' )* '\''
 ;

// Binary large object literals (blob) starting with X
BLOB_LITERAL
 : X STRING_LITERAL
 ;

// Single-line comments starting with --
SINGLE_LINE_COMMENT
 : '--' ~[\r\n]* -> channel(HIDDEN)
 ;

// Multiline comments enclosed by /* and */
MULTILINE_COMMENT
 : '/*' .*? ( '*/' | EOF ) -> channel(HIDDEN)
 ;

// Whitespace characters to be ignored
SPACES
 : [ \u000B\t\r\n] -> channel(HIDDEN)
 ;

// A fragment for matching digits (0-9)
fragment DIGIT : [0-9];

// Identifiers (quoted or alphanumeric)
IDENTIFIER
 : IDENTIFIER_CHAR (IDENTIFIER_CHAR | '[' | ']' | '(' | ')' | '{' | '}')*
 ;

// Identifiers (quoted or alphanumeric)
SIMPLE_IDENTIFIER
 : [a-zA-Z][a-zA-Z0-9_]*
 ;

fragment IDENTIFIER_CHAR
 : [a-zA-Z0-9_$*?:~!@#%^&+=|<>/\-.]
 ;

// Fragments for case-insensitive keyword matching (e.g., A can be a or A)
fragment A : [aA];
fragment B : [bB];
fragment C : [cC];
fragment D : [dD];
fragment E : [eE];
fragment F : [fF];
fragment G : [gG];
fragment H : [hH];
fragment I : [iI];
fragment J : [jJ];
fragment K : [kK];
fragment L : [lL];
fragment M : [mM];
fragment N : [nN];
fragment O : [oO];
fragment P : [pP];
fragment Q : [qQ];
fragment R : [rR];
fragment S : [sS];
fragment T : [tT];
fragment U : [uU];
fragment V : [vV];
fragment W : [wW];
fragment X : [xX];
fragment Y : [yY];
fragment Z : [zZ];

// Error rule for unexpected characters
error : UNEXPECTED_CHAR ;

// Matches any unexpected character
UNEXPECTED_CHAR
 : .
 ;