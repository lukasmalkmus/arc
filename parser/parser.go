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
	tok     token.Token
	lit     string
	buf     struct {
		tok token.Token
		lit string
		n   int
	}
}

// New returns a new instance of Parser.
func New(r io.Reader) *Parser {
	return &Parser{scanner: scanner.New(r)}
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

	p.next()
	for p.tok != token.EOF {
		// Linebreaks and Comments might prepend a statement. Those are skipped.
		if p.tok == token.NL || p.tok == token.COMMENT {
			p.next()
			continue
		}

		p.unscan()
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		prog.Statements = append(prog.Statements, stmt)

		// Next token.
		p.next()
	}

	return prog, nil
}

// ParseStatement parses lexical tokens into a Statement AST object.
func (p *Parser) ParseStatement() (ast.Statement, error) {
	// Inspect the first token.
	p.next()
	switch p.tok {
	case token.LOAD:
		return p.parseLoadStatement()
	}

	return nil, newParseError(p.tok, p.lit, token.Keywords()...)
}

// parseLoadStatement parses an LoadStatement AST object.
func (p *Parser) parseLoadStatement() (*ast.LoadStatement, error) {
	stmt := &ast.LoadStatement{}

	// First token should be the source memory location.
	src, err := p.parseMemoryLocation()
	if err != nil {
		return nil, err
	}
	stmt.Source = src

	// Next we should see a comma as seperator between source and destination.
	if p.next(); p.tok != token.COMMA {
		return nil, newParseError(p.tok, p.lit, token.COMMA)
	}

	// Next we should read the destination identifier.
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

// parseIdent parses an identifier and creates a Identifier AST object.
func (p *Parser) parseIdent() (*ast.Identifier, error) {
	if p.next(); p.tok != token.IDENT {
		return nil, newParseError(p.tok, p.lit, token.IDENT)
	}
	return &ast.Identifier{Name: p.lit}, nil
}

// parseSIMM13 parses an SIMM13 integer.
func (p *Parser) parseSIMM13() (uint64, error) {
	p.next()
	if p.tok != token.INT {
		return 0, newParseError(p.tok, p.lit, token.INT)
	}

	val, err := strconv.ParseUint(p.lit, 0, 13)
	if err != nil {
		return 0, &ParseError{Message: fmt.Sprintf("found INT %q is not a valid SIMM13", p.lit)}
	}
	return val, nil
}

// parseMemoryLocation parses a memory location and creates a MemoryLocation AST
// object.
func (p *Parser) parseMemoryLocation() (*ast.MemoryLocation, error) {
	memLoc := &ast.MemoryLocation{}

	// We either expect a left bracket which opens a direct or an offset
	// expression or a bare identifier which indicates an indirect expression.
	if p.next(); p.tok == token.LBRACKET {
		// Opening bracket is followed by identifier.
		src, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		memLoc.Base = src

		// After the identifier we expect a closing bracket which indicates a
		// direct expression or an operator which indicates an offset
		// expression.
		if p.next(); p.tok == token.RBRACKET {
			memLoc.Mode = ast.Direct
			return memLoc, nil
		}

		// Must be Offset mode.
		memLoc.Mode = ast.Offset

		// We expect the operator.
		if !p.tok.IsOperator() {
			return nil, newParseError(p.tok, p.lit, token.PLUS, token.MINUS)
		}
		memLoc.Operator = p.lit

		// We expect the offset value.
		val, err := p.parseSIMM13()
		if err != nil {
			return nil, err
		}
		memLoc.Offset = val

		// Finally, the expression must close.
		if err := p.expectClosing(); err != nil {
			return nil, err
		}

		return memLoc, nil
	}

	// No opening bracket means we should see a register in indirect addressing
	// mode.
	p.unscan()
	src, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	memLoc.Base = src
	memLoc.Mode = ast.Indirect

	return memLoc, nil
}

// expectClosing expectes the end of a expression and will error if the next
// token is not a RBRACKET token. This method will unscan the read token.
func (p *Parser) expectClosing() error {
	if p.next(); p.tok != token.RBRACKET {
		return newParseError(p.tok, p.lit, token.RBRACKET)
	}
	return nil
}

// expectStatementEnd expectes the end of a statement and will error if the next
// token is not a NL (newline) or EOF token.
func (p *Parser) expectStatementEnd() error {
	if p.next(); p.tok != token.NL && p.tok != token.EOF {
		return newParseError(p.tok, p.lit, token.NL, token.EOF)
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
func newParseError(foundTok token.Token, foundLit string, expected ...token.Token) *ParseError {
	return &ParseError{FoundTok: foundTok, FoundLit: foundLit, Expected: expected}
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
