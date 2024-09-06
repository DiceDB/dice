grammar DSQL;

parse
 : sql_stmt_list EOF
 ;

sql_stmt_list
 : sql_stmt (';' sql_stmt)* ';'?
 ;

sql_stmt
 : select_stmt
 ;

select_stmt
 : K_SELECT select_core
   (K_FROM from_clause)?
   (K_WHERE where_clause)?
   (K_ORDER K_BY ordering_terms)?
   (K_LIMIT limit_clause)?
 ;

select_core
 : KEY_TOKEN
 | VALUE_TOKEN
 | KEY_TOKEN ',' VALUE_TOKEN
 | VALUE_TOKEN ',' KEY_TOKEN
 ;

from_clause
 : table_name
 ;

where_clause
 : scalar_expr
 ;


ordering_terms
 : ordering_term (',' ordering_term)*
 ;

ordering_term
 : IDENTIFIER (K_ASC | K_DESC)?
 ;

limit_clause
 : NUMERIC_LITERAL
 ;

scalar_expr
 : literal_value
 | KEY_TOKEN
 | VALUE_TOKEN
 | IDENTIFIER
 | unary_operator scalar_expr
 | scalar_expr binary_operator scalar_expr
 | '(' scalar_expr ')'
 | scalar_expr K_IS K_NULL
 | scalar_expr K_IS K_NOT K_NULL
 | scalar_expr K_LIKE scalar_expr
 | scalar_expr K_AND scalar_expr
 | scalar_expr K_OR scalar_expr
 ;

table_name
 : IDENTIFIER
 ;

literal_value
 : NUMERIC_LITERAL
 | STRING_LITERAL
 | K_NULL
 ;

unary_operator
 : '-' | '+' | '~' | K_NOT
 ;

binary_operator
 : '||' | '*' | '/' | '%' | '+' | '-' | '<<' | '>>' | '&' | '|'
 | '<' | '<=' | '>' | '>=' | '=' | '==' | '!=' | '<>'
 | K_AND | K_OR
 ;

keyword
 : K_SELECT
 | K_FROM
 | K_WHERE
 | K_ORDER
 | K_BY
 | K_LIMIT
 | K_ASC
 | K_DESC
 | K_IS
 | K_NOT
 | K_NULL
 | K_LIKE
 | K_AND
 | K_OR
 ;

/* Lexer rules */

KEY_TOKEN : '$key';
VALUE_TOKEN : '$value';

K_SELECT : S E L E C T;
K_FROM   : F R O M;
K_WHERE  : W H E R E;
K_ORDER  : O R D E R;
K_BY     : B Y;
K_LIMIT  : L I M I T;
K_ASC    : A S C;
K_DESC   : D E S C;
K_IS     : I S;
K_NOT    : N O T;
K_NULL   : N U L L;
K_LIKE   : L I K E;
K_AND    : A N D;
K_OR     : O R;

VALUE_TOKEN_TRAILING_DOT : VALUE_TOKEN '.';

IDENTIFIER
 : [a-zA-Z_$] (IDENTIFIER_CHAR | '[' | ']' | '(' | ')' | '{' | '}')*
 ;

fragment IDENTIFIER_CHAR
 : [a-zA-Z0-9_$*?:~!@#%^&+=|<>/\-.]
 ;

NUMERIC_LITERAL
 : DIGIT+ ( '.' DIGIT* )?
 | '.' DIGIT+
 ;

STRING_LITERAL
 : '\'' ( ~'\'' | '\'\'' )* '\''
 ;

WHITESPACE : [ \t\r\n]+ -> skip;

fragment DIGIT : [0-9];
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