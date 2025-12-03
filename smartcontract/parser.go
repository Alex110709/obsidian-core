// Parser for OCL (Obsidian Contract Language)
package smartcontract

import (
	"fmt"
	"strconv"
)

// AST Node types
type Node interface {
	String() string
}

type Program struct {
	Statements []Node
}

func (p *Program) String() string {
	result := ""
	for _, stmt := range p.Statements {
		result += stmt.String() + "\n"
	}
	return result
}

type ContractDecl struct {
	Name       string
	Statements []Node
}

func (c *ContractDecl) String() string {
	result := fmt.Sprintf("contract %s:\n", c.Name)
	for _, stmt := range c.Statements {
		result += "    " + stmt.String() + "\n"
	}
	return result
}

type FunctionDecl struct {
	Name       string
	Parameters []string
	Body       []Node
}

func (f *FunctionDecl) String() string {
	params := ""
	for i, param := range f.Parameters {
		if i > 0 {
			params += ", "
		}
		params += param
	}
	result := fmt.Sprintf("def %s(%s):\n", f.Name, params)
	for _, stmt := range f.Body {
		result += "        " + stmt.String() + "\n"
	}
	return result
}

type IfStmt struct {
	Condition Node
	ThenBody  []Node
	ElseBody  []Node
}

func (i *IfStmt) String() string {
	result := fmt.Sprintf("if %s:\n", i.Condition.String())
	for _, stmt := range i.ThenBody {
		result += "    " + stmt.String() + "\n"
	}
	if len(i.ElseBody) > 0 {
		result += "else:\n"
		for _, stmt := range i.ElseBody {
			result += "    " + stmt.String() + "\n"
		}
	}
	return result
}

type AssignStmt struct {
	Target Node
	Value  Node
}

func (a *AssignStmt) String() string {
	return fmt.Sprintf("%s = %s", a.Target.String(), a.Value.String())
}

type ReturnStmt struct {
	Value Node
}

func (r *ReturnStmt) String() string {
	if r.Value == nil {
		return "return"
	}
	return fmt.Sprintf("return %s", r.Value.String())
}

type PassStmt struct{}

func (p *PassStmt) String() string {
	return "pass"
}

type ExprStmt struct {
	Expression Node
}

func (e *ExprStmt) String() string {
	return e.Expression.String()
}

type BinaryExpr struct {
	Left  Node
	Op    string
	Right Node
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op, b.Right.String())
}

type UnaryExpr struct {
	Op    string
	Right Node
}

func (u *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", u.Op, u.Right.String())
}

type CallExpr struct {
	Function  Node
	Arguments []Node
}

func (c *CallExpr) String() string {
	args := ""
	for i, arg := range c.Arguments {
		if i > 0 {
			args += ", "
		}
		args += arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Function.String(), args)
}

type AttributeExpr struct {
	Object    Node
	Attribute string
}

func (a *AttributeExpr) String() string {
	return fmt.Sprintf("%s.%s", a.Object.String(), a.Attribute)
}

type Identifier struct {
	Name string
}

func (i *Identifier) String() string {
	return i.Name
}

type NumberLiteral struct {
	Value int64
}

func (n *NumberLiteral) String() string {
	return strconv.FormatInt(n.Value, 10)
}

type StringLiteral struct {
	Value string
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type BoolLiteral struct {
	Value bool
}

func (b *BoolLiteral) String() string {
	if b.Value {
		return "True"
	}
	return "False"
}

// Parser
type Parser struct {
	tokens  []Token
	pos     int
	current Token
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) Parse() (*Program, error) {
	p.advance()
	program := &Program{}

	for p.current.Type != TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, stmt)
	}

	return program, nil
}

func (p *Parser) parseStatement() (Node, error) {
	switch p.current.Type {
	case TokenContract:
		return p.parseContractDecl()
	case TokenDef:
		return p.parseFunctionDecl()
	case TokenIf:
		return p.parseIfStmt()
	case TokenReturn:
		return p.parseReturnStmt()
	case TokenPass:
		p.expect(TokenPass)
		p.expect(TokenNewline)
		return &PassStmt{}, nil
	case TokenIdentifier:
		// Could be assignment or expression statement
		if p.peek().Type == TokenAssign {
			return p.parseAssignStmt()
		}
		fallthrough
	default:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.expect(TokenNewline)
		return &ExprStmt{Expression: expr}, nil
	}
}

func (p *Parser) parseContractDecl() (Node, error) {
	p.expect(TokenContract)
	name := p.current.Value
	p.expect(TokenIdentifier)
	p.expect(TokenColon)
	p.expect(TokenNewline)

	statements := []Node{}
	p.expect(TokenIndent)
	for p.current.Type != TokenDedent && p.current.Type != TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	p.expect(TokenDedent)

	return &ContractDecl{Name: name, Statements: statements}, nil
}

func (p *Parser) parseFunctionDecl() (Node, error) {
	p.expect(TokenDef)
	name := p.current.Value
	p.expect(TokenIdentifier)
	p.expect(TokenLParen)

	parameters := []string{}
	if p.current.Type != TokenRParen {
		param := p.current.Value
		p.expect(TokenIdentifier)
		parameters = append(parameters, param)

		for p.current.Type == TokenComma {
			p.expect(TokenComma)
			param = p.current.Value
			p.expect(TokenIdentifier)
			parameters = append(parameters, param)
		}
	}
	p.expect(TokenRParen)
	p.expect(TokenColon)
	p.expect(TokenNewline)

	body := []Node{}
	p.expect(TokenIndent)
	for p.current.Type != TokenDedent && p.current.Type != TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
	}
	p.expect(TokenDedent)

	return &FunctionDecl{Name: name, Parameters: parameters, Body: body}, nil
}

func (p *Parser) parseIfStmt() (Node, error) {
	p.expect(TokenIf)
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.expect(TokenColon)
	p.expect(TokenNewline)

	thenBody := []Node{}
	p.expect(TokenIndent)
	for p.current.Type != TokenDedent && p.current.Type != TokenEOF && p.current.Type != TokenElse {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		thenBody = append(thenBody, stmt)
	}

	var elseBody []Node
	if p.current.Type == TokenElse {
		p.expect(TokenElse)
		p.expect(TokenColon)
		p.expect(TokenNewline)
		p.expect(TokenIndent)
		for p.current.Type != TokenDedent && p.current.Type != TokenEOF {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			elseBody = append(elseBody, stmt)
		}
		p.expect(TokenDedent)
	} else {
		p.expect(TokenDedent)
	}

	return &IfStmt{Condition: condition, ThenBody: thenBody, ElseBody: elseBody}, nil
}

func (p *Parser) parseAssignStmt() (Node, error) {
	target, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.expect(TokenAssign)
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.expect(TokenNewline)
	return &AssignStmt{Target: target, Value: value}, nil
}

func (p *Parser) parseReturnStmt() (Node, error) {
	p.expect(TokenReturn)
	var value Node
	if p.current.Type != TokenNewline {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.expect(TokenNewline)
	return &ReturnStmt{Value: value}, nil
}

func (p *Parser) parseExpression() (Node, error) {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(precedence int) (Node, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		op := p.getBinaryOp()
		if op == "" || p.getPrecedence(op) <= precedence {
			break
		}

		p.advance()
		right, err := p.parseBinaryExpr(p.getPrecedence(op))
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}

	return left, nil
}

func (p *Parser) parseUnaryExpr() (Node, error) {
	if p.current.Type == TokenMinus {
		p.expect(TokenMinus)
		right, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "-", Right: right}, nil
	}

	return p.parsePrimaryExpr()
}

func (p *Parser) parsePrimaryExpr() (Node, error) {
	switch p.current.Type {
	case TokenIdentifier:
		name := p.current.Value
		p.expect(TokenIdentifier)
		if p.current.Type == TokenDot {
			p.expect(TokenDot)
			attr := p.current.Value
			p.expect(TokenIdentifier)
			return &AttributeExpr{Object: &Identifier{Name: name}, Attribute: attr}, nil
		}
		if p.current.Type == TokenLParen {
			return p.parseCallExpr(&Identifier{Name: name})
		}
		return &Identifier{Name: name}, nil
	case TokenNumber:
		value, _ := strconv.ParseInt(p.current.Value, 10, 64)
		p.expect(TokenNumber)
		return &NumberLiteral{Value: value}, nil
	case TokenString:
		value := p.current.Value
		p.expect(TokenString)
		return &StringLiteral{Value: value}, nil
	case TokenTrue:
		p.expect(TokenTrue)
		return &BoolLiteral{Value: true}, nil
	case TokenFalse:
		p.expect(TokenFalse)
		return &BoolLiteral{Value: false}, nil
	case TokenLParen:
		p.expect(TokenLParen)
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.expect(TokenRParen)
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v", p.current)
	}
}

func (p *Parser) parseCallExpr(function Node) (Node, error) {
	p.expect(TokenLParen)
	arguments := []Node{}
	if p.current.Type != TokenRParen {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)

		for p.current.Type == TokenComma {
			p.expect(TokenComma)
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
		}
	}
	p.expect(TokenRParen)
	return &CallExpr{Function: function, Arguments: arguments}, nil
}

func (p *Parser) getBinaryOp() string {
	switch p.current.Type {
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenMultiply:
		return "*"
	case TokenDivide:
		return "/"
	case TokenEqual:
		return "=="
	case TokenNotEqual:
		return "!="
	case TokenLess:
		return "<"
	case TokenGreater:
		return ">"
	case TokenLessEqual:
		return "<="
	case TokenGreaterEqual:
		return ">="
	default:
		return ""
	}
}

func (p *Parser) getPrecedence(op string) int {
	switch op {
	case "*", "/":
		return 7
	case "+", "-":
		return 6
	case "<", ">", "<=", ">=":
		return 5
	case "==", "!=":
		return 4
	default:
		return 0
	}
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
		p.pos++
	}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) expect(tokenType TokenType) {
	if p.current.Type != tokenType {
		panic(fmt.Sprintf("expected %v, got %v", tokenType, p.current.Type))
	}
	p.advance()
}
