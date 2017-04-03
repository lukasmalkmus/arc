/*
Package parser implements an ARC assembly parser. The package exports simple
functions which can be used to parse ARC source code. It relies on the scanner
package which provides lexical analysis (tokenizing) of ARC source code.
*/
package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/LukasMa/arc/ast"
	"github.com/LukasMa/arc/internal"
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

	unresolvedIdents map[string]*ast.Identifier
	declaredLabels   map[string]*ast.LabelStatement
}

// New returns a new instance of Parser.
func New(r io.Reader) *Parser {
	// Init Parser with EOF token. This ensures functions must read the first
	// token themselves.
	p := &Parser{
		scanner: scanner.New(r),

		tok: token.EOF,
		lit: "",
		pos: token.Pos{Filename: ""},

		unresolvedIdents: make(map[string]*ast.Identifier),
		declaredLabels:   make(map[string]*ast.LabelStatement),
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

		tok: token.EOF,
		lit: "",
		pos: token.Pos{Filename: f.Name()},

		unresolvedIdents: make(map[string]*ast.Identifier),
		declaredLabels:   make(map[string]*ast.LabelStatement),
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
		return nil, err
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
	prog := &ast.Program{Filename: p.pos}
	errs := internal.MultiError{}

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

		// Add statement to the programs list of statements.
		prog.AddStatement(stmt)

		// Next token.
		p.scanIgnoreNewLine()
	}

	// Generate error for unresolved identifiers.
	for lit, ident := range p.unresolvedIdents {
		err := &ParseError{Pos: ident.Pos(), Message: fmt.Sprintf("unresolved IDENTIFIER %q", lit)}
		errs.Add(err)
	}

	// Sort errors.
	errs.Sort()

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
			return &ast.LabelStatement{}, nil
		}
		return p.parseLabelStatement()
	case token.LOAD:
		return p.parseLoadStatement()
	case token.STORE:
		return p.parseStoreStatement()
	case token.ADD:
		return p.parseAddStatement()
	case token.ADDCC:
		return p.parseAddCCStatement()
	case token.SUB:
		return p.parseSubStatement()
	case token.SUBCC:
		return p.parseSubCCStatement()
	case token.AND:
		return p.parseAndStatement()
	case token.ANDCC:
		return p.parseAndCCStatement()
	case token.OR:
		return p.parseOrStatement()
	case token.ORCC:
		return p.parseOrCCStatement()
	case token.ORN:
		return p.parseOrnStatement()
	case token.ORNCC:
		return p.parseOrnCCStatement()
	case token.XOR:
		return p.parseXorStatement()
	case token.XORCC:
		return p.parseXorCCStatement()
	}

	// We expect a comment, an identifier, a directive or a keyword.
	exp := []token.Token{token.COMMENT, token.IDENT}
	exp = append(exp, token.Directives()...)
	exp = append(exp, token.Keywords()...)

	return nil, p.newParseError(exp...)
}

// parseCommentStatement parses an CommentStatement AST object.
func (p *Parser) parseCommentStatement() (*ast.CommentStatement, error) {
	stmt := &ast.CommentStatement{Position: p.pos, Text: p.lit}

	// The comment should end after its literal value.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseBeginStatement parses an BeginStatement AST object.
func (p *Parser) parseBeginStatement() (*ast.BeginStatement, error) {
	stmt := &ast.BeginStatement{Position: p.pos}

	// Finally we should see the end of the directive.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseEndStatement parses an EndStatement AST object.
func (p *Parser) parseEndStatement() (*ast.EndStatement, error) {
	stmt := &ast.EndStatement{Position: p.pos}

	// Finally we should see the end of the directive.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseOrgStatement parses an OrgStatement AST object.
func (p *Parser) parseOrgStatement() (*ast.OrgStatement, error) {
	stmt := &ast.OrgStatement{Position: p.pos}

	// The directive should be followed by an integer.
	val, err := p.parseInteger()
	if err != nil {
		return nil, err
	}
	stmt.Value = val

	// Finally we should see the end of the directive.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

func (p *Parser) parseLabelStatement() (*ast.LabelStatement, error) {
	stmt := &ast.LabelStatement{Position: p.pos}

	// Create label identifier.
	stmt.Ident = &ast.Identifier{Position: p.pos, Name: p.lit}

	// Is the label already declared? If so, an error is thrown.
	decl, prs := p.declaredLabels[stmt.Ident.Name]
	if prs {
		msg := fmt.Sprintf("label %q already declared: previous declaration at %s", stmt.Ident, decl.Pos().NoFile())
		err := &ParseError{Message: msg, Pos: stmt.Pos()}
		return nil, err
	}

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
		// Unscan because parsing the referenced statement already consumed the
		// statement end.
		p.unscan()
	}

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Declare label and remove its identifier from the list of
	// unresolved identifiers.
	p.declaredLabels[stmt.Ident.Name] = stmt
	delete(p.unresolvedIdents, stmt.Ident.Name)

	return stmt, nil
}

// parseLoadStatement parses an LoadStatement AST object.
func (p *Parser) parseLoadStatement() (*ast.LoadStatement, error) {
	stmt := &ast.LoadStatement{Position: p.pos}

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
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseStoreStatement parses an StoreStatement AST object.
func (p *Parser) parseStoreStatement() (*ast.StoreStatement, error) {
	stmt := &ast.StoreStatement{Position: p.pos}

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
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseAddStatement parses an AddStatement AST object.
func (p *Parser) parseAddStatement() (*ast.AddStatement, error) {
	stmt := &ast.AddStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseAddCCStatement parses an AddCCStatement AST object.
func (p *Parser) parseAddCCStatement() (*ast.AddCCStatement, error) {
	// Parse usual add statement.
	addStmt, err := p.parseAddStatement()
	if err != nil {
		return nil, err
	}

	// Transform to addcc.
	stmt := &ast.AddCCStatement{
		Position:    addStmt.Position,
		Source:      addStmt.Source,
		Operand:     addStmt.Operand,
		Destination: addStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseSubStatement parses an SubStatement AST object.
func (p *Parser) parseSubStatement() (*ast.SubStatement, error) {
	stmt := &ast.SubStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseSubCCStatement parses an SubCCStatement AST object.
func (p *Parser) parseSubCCStatement() (*ast.SubCCStatement, error) {
	// Parse usual sub statement.
	subStmt, err := p.parseSubStatement()
	if err != nil {
		return nil, err
	}

	// Transform to subcc.
	stmt := &ast.SubCCStatement{
		Position:    subStmt.Position,
		Source:      subStmt.Source,
		Operand:     subStmt.Operand,
		Destination: subStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseAndStatement parses an AndStatement AST object.
func (p *Parser) parseAndStatement() (*ast.AndStatement, error) {
	stmt := &ast.AndStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseAndCCStatement parses an AndCCStatement AST object.
func (p *Parser) parseAndCCStatement() (*ast.AndCCStatement, error) {
	// Parse usual and statement.
	andStmt, err := p.parseAndStatement()
	if err != nil {
		return nil, err
	}

	// Transform to andcc.
	stmt := &ast.AndCCStatement{
		Position:    andStmt.Position,
		Source:      andStmt.Source,
		Operand:     andStmt.Operand,
		Destination: andStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseOrStatement parses an OrStatement AST object.
func (p *Parser) parseOrStatement() (*ast.OrStatement, error) {
	stmt := &ast.OrStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseOrCCStatement parses an OrCCStatement AST object.
func (p *Parser) parseOrCCStatement() (*ast.OrCCStatement, error) {
	// Parse usual or statement.
	orStmt, err := p.parseOrStatement()
	if err != nil {
		return nil, err
	}

	// Transform to orcc.
	stmt := &ast.OrCCStatement{
		Position:    orStmt.Position,
		Source:      orStmt.Source,
		Operand:     orStmt.Operand,
		Destination: orStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseOrnStatement parses an OrnStatement AST object.
func (p *Parser) parseOrnStatement() (*ast.OrnStatement, error) {
	stmt := &ast.OrnStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseOrnCCStatement parses an OrnCCStatement AST object.
func (p *Parser) parseOrnCCStatement() (*ast.OrnCCStatement, error) {
	// Parse usual orn statement.
	ornStmt, err := p.parseOrnStatement()
	if err != nil {
		return nil, err
	}

	// Transform to orncc.
	stmt := &ast.OrnCCStatement{
		Position:    ornStmt.Position,
		Source:      ornStmt.Source,
		Operand:     ornStmt.Operand,
		Destination: ornStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseXorStatement parses an XorStatement AST object.
func (p *Parser) parseXorStatement() (*ast.XorStatement, error) {
	stmt := &ast.XorStatement{Position: p.pos}

	// First we should see the source register.
	reg, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Source = reg

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// Then we should see the second operand.
	second, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	stmt.Operand = second

	// Next we should see a comma as seperator between destination and source.
	if p.next(); p.tok != token.COMMA {
		return nil, p.newParseError(token.COMMA)
	}

	// The last valueable information is the destination register.
	dest, err := p.parseRegister()
	if err != nil {
		return nil, err
	}
	stmt.Destination = dest

	// Finally we should see the end of the statement.
	if err := p.expectStatementEndOrComment(); err != nil {
		return nil, err
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseXorCCStatement parses an XorCCStatement AST object.
func (p *Parser) parseXorCCStatement() (*ast.XorCCStatement, error) {
	// Parse usual xor statement.
	xorStmt, err := p.parseXorStatement()
	if err != nil {
		return nil, err
	}

	// Transform xto orcc.
	stmt := &ast.XorCCStatement{
		Position:    xorStmt.Position,
		Source:      xorStmt.Source,
		Operand:     xorStmt.Operand,
		Destination: xorStmt.Destination,
	}

	// Return the successfully parsed statement.
	return stmt, nil
}

// parseIdent parses an identifier and creates an Identifier AST object.
func (p *Parser) parseIdent() (*ast.Identifier, error) {
	if p.next(); p.tok != token.IDENT {
		return nil, p.newParseError(token.IDENT)
	}

	// If the identifier has not been declared yet, we add it to the list of
	// unresolved identifiers.
	ident := &ast.Identifier{Position: p.pos, Name: p.lit}
	if _, prs := p.declaredLabels[p.lit]; !prs {
		p.unresolvedIdents[p.lit] = ident
	}
	return ident, nil
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
			Message: fmt.Sprintf("INTEGER %q overflows 32 bit integer width", p.lit),
			Pos:     p.pos,
		}
	}
	return ast.Integer(i), nil
}

// parseSIMM13 parses an SIMM13 integer.
func (p *Parser) parseSIMM13() (ast.Integer, error) {
	if p.next(); p.tok != token.INT {
		return 0, p.newParseError(token.INT)
	}
	val, err := strconv.ParseUint(p.lit, 0, 13)
	if err != nil {
		return 0, &ParseError{
			Message: fmt.Sprintf("INTEGER %q is not a valid SIMM13", p.lit),
			Pos:     p.pos,
		}
	}
	return ast.Integer(val), nil
}

// parseExpression parses an expression and creates an Expression AST object.
func (p *Parser) parseExpression() (*ast.Expression, error) {
	exp := &ast.Expression{}

	// A left square bracket indicates the beginning of an expression.
	if p.next(); p.tok != token.LBRACKET {
		return nil, p.newParseError(token.LBRACKET)
	}

	// Opening bracket is followed by identifer or register. Checking errors of
	// the parse functions isn't required here, because we have already checked
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

// parseOperand parses an operand and creates an Operand AST object.
func (p *Parser) parseOperand() (ast.Operand, error) {
	var op ast.Operand

	// Checking errors of the parseRegister function isn't required here,
	// because we have already checked for the correct token. But the
	// parseInteger function needs checking because the literal can still be
	// overflowing the integer width.
	if p.next(); p.tok == token.REG {
		p.unscan()
		reg, _ := p.parseRegister()
		op = reg
	} else if p.tok == token.INT {
		p.unscan()
		i, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		op = i
	} else {
		return nil, p.newParseError(token.INT, token.REG)
	}

	return op, nil
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
		act = tok.String() + ` "` + e.FoundLit + `"`
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
