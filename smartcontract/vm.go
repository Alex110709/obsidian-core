// VM for executing OCL bytecode
package smartcontract

import (
	"fmt"
)

// Value types
type ValueType int

const (
	ValueInt ValueType = iota
	ValueStr
	ValueBool
	ValueNone
)

type Value struct {
	Type ValueType
	Int  int64
	Str  string
	Bool bool
}

// VM
type VM struct {
	stack   []Value
	program []Instruction
	pc      int
	vars    map[string]Value
}

// Instruction types
type OpCode int

const (
	OpPushInt OpCode = iota
	OpPushStr
	OpPushBool
	OpPushNone
	OpLoadVar
	OpStoreVar
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpEq
	OpNe
	OpLt
	OpGt
	OpLe
	OpGe
	OpCall
	OpReturn
	OpJump
	OpJumpIfFalse
	OpGetAttr
	OpSetAttr
)

type Instruction struct {
	Op  OpCode
	Arg interface{}
}

// NewVM creates a new VM
func NewVM(program []Instruction) *VM {
	return &VM{
		stack:   []Value{},
		program: program,
		pc:      0,
		vars:    make(map[string]Value),
	}
}

// Execute runs the program
func (vm *VM) Execute() (Value, error) {
	for vm.pc < len(vm.program) {
		inst := vm.program[vm.pc]
		vm.pc++

		switch inst.Op {
		case OpPushInt:
			val, _ := inst.Arg.(int64)
			vm.stack = append(vm.stack, Value{Type: ValueInt, Int: val})
		case OpPushStr:
			val, _ := inst.Arg.(string)
			vm.stack = append(vm.stack, Value{Type: ValueStr, Str: val})
		case OpPushBool:
			val, _ := inst.Arg.(bool)
			vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: val})
		case OpPushNone:
			vm.stack = append(vm.stack, Value{Type: ValueNone})
		case OpLoadVar:
			name, _ := inst.Arg.(string)
			val, ok := vm.vars[name]
			if !ok {
				return Value{}, fmt.Errorf("undefined variable: %s", name)
			}
			vm.stack = append(vm.stack, val)
		case OpStoreVar:
			name, _ := inst.Arg.(string)
			if len(vm.stack) == 0 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			val := vm.stack[len(vm.stack)-1]
			vm.stack = vm.stack[:len(vm.stack)-1]
			vm.vars[name] = val
		case OpAdd:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueInt, Int: a.Int + b.Int})
			} else if a.Type == ValueStr && b.Type == ValueStr {
				vm.stack = append(vm.stack, Value{Type: ValueStr, Str: a.Str + b.Str})
			} else {
				return Value{}, fmt.Errorf("invalid operands for +")
			}
		case OpSub:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueInt, Int: a.Int - b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for -")
			}
		case OpMul:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueInt, Int: a.Int * b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for *")
			}
		case OpDiv:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt && b.Int != 0 {
				vm.stack = append(vm.stack, Value{Type: ValueInt, Int: a.Int / b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for /")
			}
		case OpEq:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			result := false
			if a.Type == b.Type {
				switch a.Type {
				case ValueInt:
					result = a.Int == b.Int
				case ValueStr:
					result = a.Str == b.Str
				case ValueBool:
					result = a.Bool == b.Bool
				case ValueNone:
					result = true
				}
			}
			vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: result})
		case OpNe:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			result := true
			if a.Type == b.Type {
				switch a.Type {
				case ValueInt:
					result = a.Int != b.Int
				case ValueStr:
					result = a.Str != b.Str
				case ValueBool:
					result = a.Bool != b.Bool
				case ValueNone:
					result = false
				}
			}
			vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: result})
		case OpLt:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: a.Int < b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for <")
			}
		case OpGt:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: a.Int > b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for >")
			}
		case OpLe:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: a.Int <= b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for <=")
			}
		case OpGe:
			if len(vm.stack) < 2 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			b := vm.stack[len(vm.stack)-1]
			a := vm.stack[len(vm.stack)-2]
			vm.stack = vm.stack[:len(vm.stack)-2]

			if a.Type == ValueInt && b.Type == ValueInt {
				vm.stack = append(vm.stack, Value{Type: ValueBool, Bool: a.Int >= b.Int})
			} else {
				return Value{}, fmt.Errorf("invalid operands for >=")
			}
		case OpCall:
			// Simplified: assume no arguments for now
			// In real implementation, handle function calls
			return Value{}, fmt.Errorf("function calls not implemented")
		case OpReturn:
			if len(vm.stack) == 0 {
				return Value{Type: ValueNone}, nil
			}
			return vm.stack[len(vm.stack)-1], nil
		case OpJump:
			addr, _ := inst.Arg.(int)
			vm.pc = addr
		case OpJumpIfFalse:
			if len(vm.stack) == 0 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			cond := vm.stack[len(vm.stack)-1]
			vm.stack = vm.stack[:len(vm.stack)-1]
			if cond.Type != ValueBool || !cond.Bool {
				addr, _ := inst.Arg.(int)
				vm.pc = addr
			}
		case OpGetAttr:
			// Simplified attribute access
			attr, _ := inst.Arg.(string)
			if len(vm.stack) == 0 {
				return Value{}, fmt.Errorf("stack underflow")
			}
			vm.stack = vm.stack[:len(vm.stack)-1] // Pop object

			// Handle special attributes like self.balance
			if attr == "balance" {
				vm.stack = append(vm.stack, Value{Type: ValueInt, Int: 1000}) // Mock balance
			} else {
				return Value{}, fmt.Errorf("attribute not found: %s", attr)
			}
		case OpSetAttr:
			// Simplified
			return Value{}, fmt.Errorf("attribute assignment not implemented")
		default:
			return Value{}, fmt.Errorf("unknown opcode: %d", inst.Op)
		}
	}

	return Value{Type: ValueNone}, nil
}

// Compiler to generate bytecode from AST
type Compiler struct {
	program []Instruction
}

func NewCompiler() *Compiler {
	return &Compiler{
		program: []Instruction{},
	}
}

func (c *Compiler) Compile(node Node) []Instruction {
	c.compileNode(node)
	return c.program
}

func (c *Compiler) compileNode(node Node) {
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Statements {
			c.compileNode(stmt)
		}
	case *ContractDecl:
		for _, stmt := range n.Statements {
			c.compileNode(stmt)
		}
	case *FunctionDecl:
		// Skip for now
	case *IfStmt:
		c.compileNode(n.Condition)
		falseJump := len(c.program)
		c.program = append(c.program, Instruction{Op: OpJumpIfFalse, Arg: 0})
		for _, stmt := range n.ThenBody {
			c.compileNode(stmt)
		}
		endJump := len(c.program)
		c.program = append(c.program, Instruction{Op: OpJump, Arg: 0})
		c.program[falseJump].Arg = len(c.program)
		if len(n.ElseBody) > 0 {
			for _, stmt := range n.ElseBody {
				c.compileNode(stmt)
			}
		}
		c.program[endJump].Arg = len(c.program)
	case *AssignStmt:
		c.compileNode(n.Value)
		if ident, ok := n.Target.(*Identifier); ok {
			c.program = append(c.program, Instruction{Op: OpStoreVar, Arg: ident.Name})
		}
	case *PassStmt:
		// Do nothing
	case *ReturnStmt:
		if n.Value != nil {
			c.compileNode(n.Value)
		} else {
			c.program = append(c.program, Instruction{Op: OpPushNone})
		}
		c.program = append(c.program, Instruction{Op: OpReturn})
	case *ExprStmt:
		c.compileNode(n.Expression)
		// Pop result if not used
	case *BinaryExpr:
		c.compileNode(n.Left)
		c.compileNode(n.Right)
		switch n.Op {
		case "+":
			c.program = append(c.program, Instruction{Op: OpAdd})
		case "-":
			c.program = append(c.program, Instruction{Op: OpSub})
		case "*":
			c.program = append(c.program, Instruction{Op: OpMul})
		case "/":
			c.program = append(c.program, Instruction{Op: OpDiv})
		case "==":
			c.program = append(c.program, Instruction{Op: OpEq})
		case "!=":
			c.program = append(c.program, Instruction{Op: OpNe})
		case "<":
			c.program = append(c.program, Instruction{Op: OpLt})
		case ">":
			c.program = append(c.program, Instruction{Op: OpGt})
		case "<=":
			c.program = append(c.program, Instruction{Op: OpLe})
		case ">=":
			c.program = append(c.program, Instruction{Op: OpGe})
		}
	case *UnaryExpr:
		c.compileNode(n.Right)
		if n.Op == "-" {
			c.program = append(c.program, Instruction{Op: OpPushInt, Arg: int64(-1)})
			c.program = append(c.program, Instruction{Op: OpMul})
		}
	case *CallExpr:
		// Simplified
		c.program = append(c.program, Instruction{Op: OpCall})
	case *AttributeExpr:
		c.compileNode(n.Object)
		c.program = append(c.program, Instruction{Op: OpGetAttr, Arg: n.Attribute})
	case *Identifier:
		c.program = append(c.program, Instruction{Op: OpLoadVar, Arg: n.Name})
	case *NumberLiteral:
		c.program = append(c.program, Instruction{Op: OpPushInt, Arg: n.Value})
	case *StringLiteral:
		c.program = append(c.program, Instruction{Op: OpPushStr, Arg: n.Value})
	case *BoolLiteral:
		c.program = append(c.program, Instruction{Op: OpPushBool, Arg: n.Value})
	}
}
