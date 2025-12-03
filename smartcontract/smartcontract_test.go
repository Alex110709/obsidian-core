package smartcontract

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `contract MyContract`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	if len(tokens) == 0 {
		t.Fatal("No tokens generated")
	}

	// Check first token is contract
	if tokens[0].Type != TokenContract {
		t.Errorf("Expected TokenContract, got %v", tokens[0].Type)
	}
}

func TestParser(t *testing.T) {
	input := `x = 5
`

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Lexer error: %v", err)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parser error: %v", err)
	}

	if len(ast.Statements) == 0 {
		t.Fatal("No statements parsed")
	}
}

func TestVM(t *testing.T) {
	// Simple program: push 5, push 3, add, return
	program := []Instruction{
		{Op: OpPushInt, Arg: int64(5)},
		{Op: OpPushInt, Arg: int64(3)},
		{Op: OpAdd},
		{Op: OpReturn},
	}

	vm := NewVM(program)
	result, err := vm.Execute()
	if err != nil {
		t.Fatalf("VM error: %v", err)
	}

	if result.Type != ValueInt || result.Int != 8 {
		t.Errorf("Expected 8, got %v", result)
	}
}
