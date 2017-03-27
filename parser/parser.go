package parser

import (
	"fmt"
	"io"
	"os"
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
	pos token.Pos

	// Buffered token.
	buf struct {
		tok token.Token
		lit string
		pos token.Pos
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
		pos:     token.Pos{Filename: "", Line: 0},
	}
	return p
}

// NewFileParser returns a new instance of Parser, but will exclusively take an
// *os.File as argument instead of the more general io.Reader interface.
// Therefore it will enhance token positions with the filename.
func NewFileParser(f *os.File) *Parser {
	// Init Parser with EOF token. This ensures functions must read the first
	// token themselves.
	p := &Parser{
		scanner: scanner.NewFileScanner(f),
		tok:     token.EOF,
		lit:     "",
		pos:     token.Pos{Filename: f.Name(), Line: 0},
	}
	return p
}

// Parse parses a string into a Program AST object.
func Parse(s string) (*ast.Program, error) { return New(strings.NewReader(s)).Parse() }

// ParseFile parses the content of a file into a Program AST object. An error is
// returned if opening of the file or parsing fails.
func ParseFile(filename string) (*ast.Program, error) {
	// Read source file.
	src, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading source file %q: %s", filename, err.Error())
	}
	defer src.Close()

	return NewFileParser(src).Parse()
}

// ParseStatement parses a string into a Statement AST object.
func ParseStatement(s string) (ast.Statement, error) {
	return New(strings.NewReader(s)).ParseStatement()
}

// Parse parses the content of the underlying reader into a Program AST object.
func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	errs := MultiError{}

	// Read the first token. Linebreaks might prepend a statement. Those are
	// skipped.
	p.scanIgnoreNewLine()

	// Parse input line by line.
	for p.tok != token.EOF {
		// Parse statement. An error will be added to the list of errors.
		stmt, err := p.parseStatement(true)
		if err != nil {
			errs.Add(err)
			p.skipStatement()
			continue
		}
		prog.Statements = append(prog.Statements, stmt)

		// Next token.
		p.scanIgnoreNewLine()
	}

	return prog, errs.Return()
}

// ParseStatement parses lexical tokens into a Statement AST object.
func (p *Parser) ParseStatement() (ast.Statement, error) {
	// Read the first token.
	p.next()
	return p.parseStatement(true)
}

// parseStatement parses lexical tokens into a Statement AST object. Parsing
// identifiers into LabelStatement AST objects can be turned off by passing
// false. This is useful for avoiding recursive parsing of labels. Labels can't
// reference another label.
func (p *Parser) parseStatement(withLabel bool) (ast.Statement, error) {
	switch p.tok {
	case token.COMMENT:
		return p.parseCommentStatement()
	case token.BEGIN:
		return p.parseBeginStatement()
	case token.END:
		return p.parseEndStatement()
	case token.ORG:
		return p.parseOrgStatement()
	case token.IDENT:
		if !withLabel {
			return nil, nil
		}
		return p.parseLabelStatement()
	case token.LOAD:
		return p.parseLoadStatement()
	case token.STORE:
		return p.parseStoreStatement()
	}

	// We expect a comment, an identifier, a directive or a keyword.
	exp := []token.Token{token.COMMENT, token.IDENT}
	exp = append(exp, token.Directives()...)
	exp = append(exp, token.Keywords()...)

	return nil, p.newParseError(exp...)
}

// parseCommentStatement parses an CommentStatement AST object.
func (p *Parser) parseCommentStatement() (*ast.CommentStatement, error) {
	stmt := &ast.CommentStatement{Text: p.lit}

	// The comment should end after its literal value.
	err := p.expectStatementEnd()

	// Return the successfully parsed statement.
	return stmt, err
}

// parseBeginStatement parses an BeginStatement AST object.
func (p *Parser) parseBeginStatement() (*ast.BeginStatement, error) {
	stmt := &ast.BeginStatement{}

	// The directive should end after its literal value.
	err := p.expectStatementEndOrComment()

	// Return the successfully parsed statement.
	return stmt, err
}

// parseEndStatement parses an EndStatement AST object.
func (p *Parser) parseEndStatement() (*ast.EndStatement, error) {
	stmt := &ast.EndStatement{}

	// The directive should end after its literal value.
	err := p.expectStatementEndOrComment()

	// Return the successfully parsed statement.
	return stmt, err
}

// parseOrgStatement parses an OrgStatement AST object.
func (p *Parser) parseOrgStatement() (*ast.OrgStatement, error) {
	stmt := &ast.OrgStatement{}

	// The directive should be followed by an integer.
	val, err := p.parseInteger()
	if err != nil {
		return nil, err
	}
	stmt.Value = val

	// Finally we should see the end of the directive.
	err = p.expectStatementEndOrComment()

	// Return the successfully parsed statement.
	return stmt, err
}

func (p *Parser) parseLabelStatement() (*ast.LabelStatement, error) {
	stmt := &ast.LabelStatement{}

	stmt.Ident = &ast.Identifier{Name: p.lit}

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
			exp := []token.Token{token.INT}
			exp = append(exp, token.Keywords()...)
			return nil, p.newParseError(exp...)
		}
		stmt.Reference = refStmt
	}

	// Finally we should see the end of the statement.
	err := p.expectStatementEndOrComment()

	return stmt, err
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

	// Next we should see the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	err = p.expectStatementEndOrComment()

	// Return the successfully parsed statement.
	return stmt, err
}

// parseStoreStatement parses an StoreStatement AST object.
func (p *Parser) parseStoreStatement() (*ast.StoreStatement, error) {
	stmt := &ast.StoreStatement{}

	// First we should see the source register.
	src, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = src

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Next we should see the destination memory location.
	dest, err := p.parseMemoryLocation()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	err = p.expectStatementEndOrComment()

	// Return the successfully parsed statement.
	return stmt, err
}

// parseExpression parses an expression and creates an Expression AST object.
func (p *Parser) parseExpression() (*ast.Expression, error) {
	exp := &ast.Expression{}

	// A left square bracket indicates the beginning of an expression.
	if p.next(); p.tok != token.LBRACKET {
		return nil, p.newParseError(token.LBRACKET)
	}

	// Opening bracket is followed by identifer or register. Checking errors of
	// the parse functions isn't required here, becasue we have already checked
	// for the correct token.
	if p.next(); p.tok == token.IDENT {
		p.unscan()
		ident, _ := p.parseIdent()
		exp.Base = ident
	} else if p.tok == token.REG {
		p.unscan()
		reg, _ := p.parseRegister()
		exp.Base = reg
	} else {
		return nil, p.newParseError(token.IDENT, token.REG)
	}

	// After the base we expect a closing bracket which indicates a direct
	// expression or an operator which indicates an offset expression.
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
	return &ast.Identifier{Name: p.lit}, nil
}

// parseRegister parses a register and creates a Register AST object.
func (p *Parser) parseRegister() (*ast.Register, error) {
	if p.next(); p.tok != token.REG {
		return nil, p.newParseError(token.REG)
	}
	return &ast.Register{Name: p.lit}, nil
}

// parseInteger parses an integer and returns an Integer AST object.
func (p *Parser) parseInteger() (ast.Integer, error) {
	if p.next(); p.tok != token.INT {
		return 0, p.newParseError(token.INT)
	}
	i, err := strconv.ParseInt(p.lit, 10, 32)
	if err != nil {
		return 0, &ParseError{
			Message: fmt.Sprintf("integer %s overflows 32 bit integer", p.lit),
			Pos:     p.pos,
		}
	}
	return ast.Integer(i), nil
}

// parseMemoryLocation parses a memory location and creates an Expression or
// Identifier AST object.
func (p *Parser) parseMemoryLocation() (ast.MemoryLocation, error) {
	var memLoc ast.MemoryLocation

	// We either expect a left bracket which opens a direct or an offset
	// expression or a register which indicates an indirect expression.
	if p.next(); p.tok == token.LBRACKET {
		p.unscan()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		memLoc = exp
	} else if p.tok == token.REG {
		memLoc = &ast.Register{Name: p.lit}
	} else {
		return nil, p.newParseError(token.LBRACKET, token.REG)
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
		return 0, &ParseError{
			Message: fmt.Sprintf("integer %s is not a valid SIMM13", p.lit),
			Pos:     p.pos,
		}
	}
	return ast.Integer(val), nil
}

// expectStatementEnd expectes the end of a statement. It will error if the next
// token is not a NL (newline) or EOF token.
func (p *Parser) expectStatementEnd() error {
	if p.next(); p.tok != token.NL && p.tok != token.EOF {
		return p.newParseError(token.NL, token.EOF)
	}
	return nil
}

// expectStatementEndOrComment expectes the end of a statement or a suffixing
// comment. It will error if the next token is not a comment, NL (newline) or
// EOF token.
func (p *Parser) expectStatementEndOrComment() error {
	if p.next(); p.tok == token.COMMENT {
		p.unscan()
	} else if p.tok != token.NL && p.tok != token.EOF {
		return p.newParseError(token.COMMENT, token.NL, token.EOF)
	}
	return nil
}

// scan returns the next token from the underlying scanner. If a token has been
// unscanned then read that instead.
func (p *Parser) scan() {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		p.tok, p.lit, p.pos = p.buf.tok, p.buf.lit, p.buf.pos
		return
	}

	// Otherwise read the next token from the scanner.
	p.tok, p.lit, p.pos = p.scanner.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit, p.buf.pos = p.tok, p.lit, p.pos
}

// scanIgnoreNewLineComment scans the next non-whitespace, non-newline token.
func (p *Parser) scanIgnoreNewLine() {
	if p.next(); p.tok == token.NL {
		p.next()
	}
}

// skipStatement scans until it encounters a new statement (indicated by
// newline).
func (p *Parser) skipStatement() {
	for p.tok != token.NL && p.tok != token.EOF {
		p.next()
	}
	p.next()
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
	Pos      token.Pos
	Expected []token.Token
}

// newParseError returns a new instance of ParseError.
func (p *Parser) newParseError(expected ...token.Token) *ParseError {
	return &ParseError{FoundTok: p.tok, FoundLit: p.lit, Pos: p.pos, Expected: expected}
}

// Error returns the string representation of the error. It implements the error
// interface.
func (e ParseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Pos, e.Message)
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

	return fmt.Sprintf("%s: found %s, expected %s", e.Pos, act, strings.Join(exp, ", "))
}

// MultiError is a collection of multiple errors. It implements the error
// interface.
type MultiError struct {
	errs []error
}

func (m MultiError) Error() string {
	errs := []string{}
	for _, err := range m.errs {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "\n")
}

// Add adds one or more errors.
func (m *MultiError) Add(es ...error) {
	for _, e := range es {
		m.errs = append(m.errs, e)
	}
}

// Return returns the MultiError itself if errors are set, otherwise nil.
func (m *MultiError) Return() error {
	if len(m.errs) > 0 {
		return m
	}
	return nil
}
