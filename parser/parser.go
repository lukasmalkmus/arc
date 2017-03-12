package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LukasMa/arc/ast"
	"github.com/LukasMa/arc/scanner"
	"github.com/LukasMa/arc/token"
)

// Parser represents a parser.
type Parser struct {
	scanner *scanner.Scanner

	// Current token.
	tok token.Token
	lit string

	// Buffered token.
	buf struct {
		tok token.Token
		lit string
		n   int
	}
}

// New returns a new instance of Parser.
func New(r io.Reader) *Parser {
	// Init Parser with EOF token. This ensures functions must read the first
	// token themselves.
	p := &Parser{
		scanner: scanner.New(r),
		tok:     token.EOF,
		lit:     "",
	}
	return p
}

// Parse parses a string into a Program AST object.
func Parse(s string) (*ast.Program, error) { return New(strings.NewReader(s)).Parse() }

// ParseStatement parses a string into a Statement AST object.
func ParseStatement(s string) (ast.Statement, error) {
	return New(strings.NewReader(s)).ParseStatement()
}

// Parse parses lexical tokens into a Program AST object.
func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}

	// Read first token.
	p.next()

	// Fast forward to first non-whitespace, non-newline, non-comment token.
	for p.tok != token.EOF {
		// Skip linebreaks and comments. First token must be .begin directive.
		if p.tok == token.NL || p.tok == token.COMMENT {
			p.next()
			continue
		} else if p.tok == token.BEGIN {
			p.next()
			break
		}
		return nil, p.newParseError(token.BEGIN)
	}

	for p.tok != token.EOF {
		// Linebreaks might prepend a statement. Those are skipped.
		if p.tok == token.NL {
			p.next()
			continue
		} else if p.tok == token.END {
			p.next()
			break
		}

		stmt, err := p.parseStatement(true)
		if err != nil {
			return nil, err
		}
		prog.Statements = append(prog.Statements, stmt)

		// Next token.
		p.next()
	}

	// Last token must be .end directive.
	for p.tok != token.EOF {
		// Skip linebreaks and comments. First token must be .begin directive.
		if p.tok == token.NL || p.tok == token.COMMENT {
			p.next()
			continue
		}
		return nil, p.newParseError(token.EOF)
	}

	return prog, nil
}

// ParseStatement parses lexical tokens into a Statement AST object.
func (p *Parser) ParseStatement() (ast.Statement, error) {
	// Inspect the first token.
	p.next()
	return p.parseStatement(true)
}

// parseStatement parses lexical tokens into a Statement AST object. Parsing
// identifiers into LabelStatement AST objects can be turned off by passing
// true. This will return an empty LabelStatement.
func (p *Parser) parseStatement(withLabel bool) (ast.Statement, error) {
	switch p.tok {
	case token.COMMENT:
		// nop
	case token.LOAD:
		return p.parseLoadStatement()
	case token.STORE:
		return p.parseStoreStatement()
	case token.IDENT:
		if !withLabel {
			return &ast.LabelStatement{}, nil
		}
		return p.parseLabelStatement()
	}

	return nil, p.newParseError(token.Keywords()...)
}

// parseLoadStatement parses an LoadStatement AST object.
func (p *Parser) parseLoadStatement() (*ast.LoadStatement, error) {
	stmt := &ast.LoadStatement{}

	// First we should see the source memory location.
	src, err := p.parseMemoryLocation()
	if err != nil {
		return nil, err
	}
	stmt.Source = src

	// Next we should see a comma as seperator between source and destination.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Next we should see the destination identifier.
	dest, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEnd(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseStoreStatement parses an StoreStatement AST object.
func (p *Parser) parseStoreStatement() (*ast.StoreStatement, error) {
	stmt := &ast.StoreStatement{}

	// First we should see the destination identifier.
	src, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	stmt.Source = src

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Next we should see the source memory location.
	dest, err := p.parseMemoryLocation()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEnd(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

func (p *Parser) parseLabelStatement() (*ast.LabelStatement, error) {
	stmt := &ast.LabelStatement{}

	stmt.Ident = &ast.Identifier{Value: p.lit}

	// Labels end with a colon (assignment).
	if p.next(); p.tok != token.COLON {
		return nil, p.newParseError(token.COLON)
	}

	// We either want an integer or a statement.
	// TODO: We need a string datatype!
	if p.next(); p.tok == token.INT {
		p.unscan()
		ref, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		stmt.Reference = ref
	} else {
		ref, err := p.parseStatement(false)
		if err != nil {
			return nil, err
		}
		refStmt, valid := ref.(ast.Reference)
		if !valid {
			toks := append([]token.Token{token.INT}, token.Keywords()...)
			return nil, p.newParseError(toks...)
		}
		stmt.Reference = refStmt
	}

	// Finally we should see the end of the statement.
	if err := p.expectStatementEnd(); err != nil {
		return nil, err
	}

	return stmt, nil
}

// parseExpression parses an expression and creates an Expression AST object.
func (p *Parser) parseExpression() (*ast.Expression, error) {
	exp := &ast.Expression{}

	// A left square bracket indicates the beginning of an expression.
	if p.next(); p.tok != token.LBRACKET {
		return nil, p.newParseError(token.LBRACKET)
	}

	// Opening bracket is followed by identifier.
	ident, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	exp.Ident = ident

	// After the identifier we expect a closing bracket which indicates a
	// direct expression or an operator which indicates an offset
	// expression.
	if p.next(); p.tok != token.RBRACKET {
		// If we don't see the closing square bracket, we expect to see an
		// operator.
		if p.tok != token.PLUS && p.tok != token.MINUS {
			return nil, p.newParseError(token.PLUS, token.MINUS, token.RBRACKET)
		}
		exp.Operator = p.lit

		// We expect the offset value.
		val, err := p.parseSIMM13()
		if err != nil {
			return nil, err
		}
		exp.Offset = val
	} else {
		p.unscan()
	}

	// The expression must close with a right square bracket.
	if p.next(); p.tok != token.RBRACKET {
		return nil, p.newParseError(token.RBRACKET)
	}

	return exp, nil
}

// parseIdent parses an identifier and creates an Identifier AST object.
func (p *Parser) parseIdent() (*ast.Identifier, error) {
	if p.next(); p.tok != token.IDENT {
		return nil, p.newParseError(token.IDENT)
	}
	return &ast.Identifier{Value: p.lit}, nil
}

// parseInteger parses an integer and returns an Integer AST object.
func (p *Parser) parseInteger() (ast.Integer, error) {
	if p.next(); p.tok != token.INT {
		return 0, p.newParseError(token.INT)
	}
	i, err := strconv.ParseInt(p.lit, 10, 32)
	if err != nil {
		return 0, &ParseError{Message: fmt.Sprintf("integer %s overflows 32 bit integer", p.lit)}
	}
	return ast.Integer(i), nil
}

// parseMemoryLocation parses a memory location and creates an Expression or
// Identifier AST object.
func (p *Parser) parseMemoryLocation() (ast.MemoryLocation, error) {
	var memLoc ast.MemoryLocation

	// We either expect a left bracket which opens a direct or an offset
	// expression or a bare identifier which indicates an indirect expression.
	if p.next(); p.tok == token.LBRACKET {
		p.unscan()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		memLoc = exp
	} else if p.tok == token.IDENT {
		memLoc = &ast.Identifier{Value: p.lit}
	} else {
		return nil, p.newParseError(token.LBRACKET, token.IDENT)
	}

	return memLoc, nil
}

// parseSIMM13 parses an SIMM13 integer.
func (p *Parser) parseSIMM13() (ast.Integer, error) {
	if p.next(); p.tok != token.INT {
		return 0, p.newParseError(token.INT)
	}
	val, err := strconv.ParseUint(p.lit, 0, 13)
	if err != nil {
		return 0, &ParseError{Message: fmt.Sprintf("integer %s is not a valid SIMM13", p.lit)}
	}
	return ast.Integer(val), nil
}

// expectStatementEnd expectes the end of a statement and will error if the next
// token is not a NL (newline) or EOF token.
func (p *Parser) expectStatementEnd() error {
	if p.next(); p.tok != token.NL && p.tok != token.EOF {
		return p.newParseError(token.NL, token.EOF)
	}
	return nil
}

// scan returns the next token from the underlying scanner. If a token has been
// unscanned then read that instead.
func (p *Parser) scan() {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		p.tok, p.lit = p.buf.tok, p.buf.lit
		return
	}

	// Otherwise read the next token from the scanner.
	p.tok, p.lit = p.scanner.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = p.tok, p.lit
}

// next scans the next non-whitespace token.
func (p *Parser) next() {
	if p.scan(); p.tok == token.WS {
		p.scan()
	}
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Message  string
	FoundTok token.Token
	FoundLit string
	Expected []token.Token
}

// newParseError returns a new instance of ParseError.
func (p *Parser) newParseError(expected ...token.Token) *ParseError {
	return &ParseError{FoundTok: p.tok, FoundLit: p.lit, Expected: expected}
}

// Error returns the string representation of the error.
func (e *ParseError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	act := ""
	if tok := e.FoundTok; tok.IsSpecial() && tok != token.ILLEGAL {
		act = tok.String()
	} else if tok := e.FoundTok; tok.IsLiteral() || tok == token.ILLEGAL {
		act = tok.String() + ` ("` + e.FoundLit + `")`
	} else {
		act = `"` + tok.String() + `"`
	}

	exp := make([]string, 0)
	for _, tok := range e.Expected {
		if tok.IsSpecial() || tok.IsLiteral() {
			exp = append(exp, tok.String())
		} else {
			exp = append(exp, `"`+tok.String()+`"`)
		}
	}

	return fmt.Sprintf("found %s, expected %s", act, strings.Join(exp, ", "))
}
