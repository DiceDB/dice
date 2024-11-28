grammar dsql;

options {
    caseInsensitive = true;
}

query
    : selectStmt EOF
    ;

selectStmt
    : SELECT selectFields whereClause? orderByClause? limitClause?
    ;

selectFields
    : field (COMMA field)?
    | STAR
    ;

whereClause
    : WHERE condition
    ;

orderByClause
    : ORDER BY orderByField
    ;

limitClause
    : LIMIT NUMBER
    ;

field
    : KEY
    | VALUE
    ;

// AND has higher precedence than OR
condition
    : condition AND condition
    | condition OR condition
    | expression
    | LPAREN condition RPAREN
    ;

orderByField
    : fieldWithString (ASC | DESC)?
    ;

expression
    : fieldWithString comparisonOp value
    ;

fieldWithString
    : KEY
    | VALUE
    | STRING_LITERAL
    | BACKTICK_LITERAL
    | VALUE (DOT IDENTIFIER)+
    ;

value
    : KEY
    | VALUE
    | STRING_LITERAL
    | BACKTICK_LITERAL
    | VALUE (DOT IDENTIFIER)+
    | NUMBER
    | NUMERIC_LITERAL
    | NULL
    ;

comparisonOp
    : ASSIGN
    | EQ
    | NEQ
    | GT
    | GTE
    | LT
    | LTE
    | NOT LIKE
    | LIKE
    | IS NOT
    | IS
    ;


SELECT : 'SELECT';
WHERE  : 'WHERE';
ORDER  : 'ORDER';
BY     : 'BY';
LIMIT  : 'LIMIT';
ASC    : 'ASC';
DESC   : 'DESC';
AND    : 'AND';
OR     : 'OR';
LIKE   : 'LIKE';
IS     : 'IS';
NOT    : 'NOT';
NULL   : 'NULL';
KEY    : '_key';
VALUE  : '_value';

ASSIGN : '=';
EQ     : '==';
NEQ    : '!=';
GT     : '>';
GTE    : '>=';
LT     : '<';
LTE    : '<=';

LPAREN : '(';
RPAREN : ')';
COMMA  : ',';
DOT    : '.';
QUOTE  : '\'';
STAR   : '*';

NUMBER : DIGIT+;

NUMERIC_LITERAL: ((DIGIT+ ('.' DIGIT*)?) | ('.' DIGIT+)) ('E' [-+]? DIGIT+)? | '0x' HEX_DIGIT+;

STRING_LITERAL: '\'' ( ~'\'' | '\'\'')* '\'';

BACKTICK_LITERAL: '`' (~'`' | '``')* '`';

IDENTIFIER : [A-Z0-9_]+;

SINGLE_LINE_COMMENT: '--' ~[\r\n]* (('\r'? '\n') | EOF) -> channel(HIDDEN);

MULTILINE_COMMENT: '/*' .*? '*/' -> channel(HIDDEN);

SPACES: [ \u000B\t\r\n] -> channel(HIDDEN);

UNEXPECTED_CHAR: .;

fragment HEX_DIGIT : [0-9A-F];
fragment DIGIT     : [0-9];