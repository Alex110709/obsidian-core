// Package smartcontract provides smart contract functionality for Obsidian
package smartcontract

import (
	"fmt"
)

// Language: Obsidian Contract Language (OCL) - Python-like syntax for smart contracts

// Token types
type TokenType int

const (
	// Keywords
	TokenContract TokenType = iota
	TokenDef
	TokenIf
	TokenElif
	TokenElse
	TokenFor
	TokenWhile
	TokenReturn
	TokenSelf
	TokenTrue
	TokenFalse
	TokenNone
	TokenPass

	// Operators
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenGreater
	TokenLessEqual
	TokenGreaterEqual
	TokenAssign
	TokenDot
	TokenComma
	TokenColon
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenLBrace
	TokenRBrace

	// Literals
	TokenIdentifier
	TokenNumber
	TokenString

	// Special
	TokenIndent
	TokenDedent
	TokenNewline
	TokenEOF
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// Lexer tokenizes OCL source code
type Lexer struct {
	input  string
	pos    int
	line   int
	indent []int
	tokens []Token
}

// NewLexer creates a new lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		indent: []int{0},
		tokens: []Token{},
	}
}

// Tokenize converts source code to tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		char := l.input[l.pos]

		switch {
		case char == ' ':
			l.handleIndent()
		case char == '\n':
			l.handleNewline()
		case char == '#':
			l.skipComment()
		case isLetter(char) || char == '_':
			l.readIdentifier()
		case isDigit(char):
			l.readNumber()
		case char == '"':
			l.readString()
		default:
			if token := l.readOperator(); token.Type != TokenEOF {
				l.tokens = append(l.tokens, token)
			} else {
				return nil, fmt.Errorf("unexpected character: %c", char)
			}
		}
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line})
	return l.tokens, nil
}

// Helper functions for lexer
func (l *Lexer) handleIndent() {
	// Simplified indent handling
	l.pos++
}

func (l *Lexer) handleNewline() {
	l.line++
	l.pos++
	l.tokens = append(l.tokens, Token{Type: TokenNewline, Line: l.line})
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
}

func (l *Lexer) readIdentifier() {
	start := l.pos
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	value := l.input[start:l.pos]

	var tokenType TokenType
	switch value {
	case "contract":
		tokenType = TokenContract
	case "def":
		tokenType = TokenDef
	case "if":
		tokenType = TokenIf
	case "elif":
		tokenType = TokenElif
	case "else":
		tokenType = TokenElse
	case "for":
		tokenType = TokenFor
	case "while":
		tokenType = TokenWhile
	case "return":
		tokenType = TokenReturn
	case "self":
		tokenType = TokenSelf
	case "True":
		tokenType = TokenTrue
	case "False":
		tokenType = TokenFalse
	case "None":
		tokenType = TokenNone
	case "pass":
		tokenType = TokenPass
	default:
		tokenType = TokenIdentifier
	}

	l.tokens = append(l.tokens, Token{Type: tokenType, Value: value, Line: l.line})
}

func (l *Lexer) readNumber() {
	start := l.pos
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}
	value := l.input[start:l.pos]
	l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: value, Line: l.line})
}

func (l *Lexer) readString() {
	l.pos++ // skip opening quote
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		l.pos++
	}
	if l.pos >= len(l.input) {
		// Error: unterminated string
		return
	}
	value := l.input[start:l.pos]
	l.pos++ // skip closing quote
	l.tokens = append(l.tokens, Token{Type: TokenString, Value: value, Line: l.line})
}

func (l *Lexer) readOperator() Token {
	char := l.input[l.pos]
	l.pos++

	switch char {
	case '_':
		return Token{Type: TokenIdentifier, Value: "_", Line: l.line}
	case '+':
		return Token{Type: TokenPlus, Value: "+", Line: l.line}
	case '-':
		return Token{Type: TokenMinus, Value: "-", Line: l.line}
	case '*':
		return Token{Type: TokenMultiply, Value: "*", Line: l.line}
	case '/':
		return Token{Type: TokenDivide, Value: "/", Line: l.line}
	case '=':
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenEqual, Value: "==", Line: l.line}
		}
		return Token{Type: TokenAssign, Value: "=", Line: l.line}
	case '!':
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenNotEqual, Value: "!=", Line: l.line}
		}
	case '<':
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenLessEqual, Value: "<=", Line: l.line}
		}
		return Token{Type: TokenLess, Value: "<", Line: l.line}
	case '>':
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenGreaterEqual, Value: ">=", Line: l.line}
		}
		return Token{Type: TokenGreater, Value: ">", Line: l.line}
	case '.':
		return Token{Type: TokenDot, Value: ".", Line: l.line}
	case ',':
		return Token{Type: TokenComma, Value: ",", Line: l.line}
	case ':':
		return Token{Type: TokenColon, Value: ":", Line: l.line}
	case '(':
		return Token{Type: TokenLParen, Value: "(", Line: l.line}
	case ')':
		return Token{Type: TokenRParen, Value: ")", Line: l.line}
	case '[':
		return Token{Type: TokenLBracket, Value: "[", Line: l.line}
	case ']':
		return Token{Type: TokenRBracket, Value: "]", Line: l.line}
	case '{':
		return Token{Type: TokenLBrace, Value: "{", Line: l.line}
	case '}':
		return Token{Type: TokenRBrace, Value: "}", Line: l.line}
	}

	return Token{Type: TokenEOF}
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}
