// Package token defines constants representing the lexical tokens of the ARC
// assembly language and basic operations on tokens (printing, predicates).
package token

import "strings"

// Token is a lexical token of the ARC assembly language.
type Token int

// All available tokens.
const (
	// Special tokens
	specialBeg Token = iota
	ILLEGAL          // Illegal character
	EOF              // End of file
	WS               // Whitespace
	NL               // Newline
	COMMENT          // !
	specialEnd

	// Identifiers and type literals
	literalBeg
	IDENT // x, y, abc, foo_bar
	REG   // %r1, %r2, %pc
	INT   // 12345
	literalEnd

	// Operators
	operatorBeg
	PLUS  // +
	MINUS // -
	operatorEnd

	// Misc characters
	LBRACKET // [
	RBRACKET // ]
	COMMA    // ,
	COLON    // :

	// Keywords
	keywordBeg
	LOAD  // ld
	STORE // st
	ADD   // add
	ADDCC // addcc
	SUB   // sub
	SUBCC // subcc
	keywordEnd

	// Directives
	directiveBeg
	BEGIN // .begin
	END   // .end
	ORG   // .org
	directiveEnd
)

var tokens = [...]string{
	// Special tokens
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WHITESPACE",
	NL:      "NEWLINE",
	COMMENT: "COMMENT",

	// Identifiers and type literals
	IDENT: "IDENTIFIER",
	REG:   "REGISTER",
	INT:   "INTEGER",

	// Operators
	PLUS:  "+",
	MINUS: "-",

	// Misc characters
	LBRACKET: "[",
	RBRACKET: "]",
	COMMA:    ",",
	COLON:    ":",

	// Keywords
	LOAD:  "ld",
	STORE: "st",
	ADD:   "add",
	ADDCC: "addcc",
	SUB:   "sub",
	SUBCC: "subcc",

	// Directives
	BEGIN: ".begin",
	END:   ".end",
	ORG:   ".org",
}

var reservedWords map[string]Token

func init() {
	reservedWords = make(map[string]Token)
	for tok := keywordBeg + 1; tok < keywordEnd; tok++ {
		reservedWords[strings.ToLower(tokens[tok])] = tok
	}
	for tok := directiveBeg + 1; tok < directiveEnd; tok++ {
		reservedWords[strings.ToLower(tokens[tok])] = tok
	}
}

// String returns the string representation of the token.
func (t Token) String() string {
	return tokens[t]
}

// IsSpecial returns true for tokens corresponding to special tokens. It returns
// false otherwise.
func (t Token) IsSpecial() bool { return specialBeg < t && t < specialEnd }

// IsLiteral returns true for tokens corresponding to identifiers and basic type
// literals. It returns false otherwise.
func (t Token) IsLiteral() bool { return literalBeg < t && t < literalEnd }

// IsOperator returns true for tokens corresponding to operators. It returns
// false otherwise.
func (t Token) IsOperator() bool { return operatorBeg < t && t < operatorEnd }

// IsKeyword returns true for tokens corresponding to keywords. It returns false
// otherwise.
func (t Token) IsKeyword() bool { return keywordBeg < t && t < keywordEnd }

// IsDirective returns true for tokens corresponding to directives. It returns
// false otherwise.
func (t Token) IsDirective() bool { return directiveBeg < t && t < directiveEnd }

// Directives returns all tokens corresponding to directives.
func Directives() []Token {
	var buf []Token
	for i := directiveBeg + 1; i < directiveEnd; i++ {
		buf = append(buf, Token(i))
	}
	return buf
}

// Keywords returns all tokens corresponding to keywords.
func Keywords() []Token {
	var buf []Token
	for i := keywordBeg + 1; i < keywordEnd; i++ {
		buf = append(buf, Token(i))
	}
	return buf
}

// Lookup returns the token associated with a given string.
func Lookup(ident string) Token {
	if tok, ok := reservedWords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
