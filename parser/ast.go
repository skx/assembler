package parser

import (
	"fmt"

	"github.com/skx/assembler/token"
)

// Node is something we return from our parser.
type Node interface {
	// Output this as a readable string
	String() string
}

// Error contains an error-message
type Error struct {
	Node
	Value string
}

// String outputs this Error structure as a string.
func (e Error) String() string {
	return fmt.Sprintf("<ERROR:%s>", e.Value)
}

// Data holds a data-statement, which might look like either of these:
//
//   .foo DB "Steve"
//   .bar DB 0x030, 0x40, 0x90
//
type Data struct {
	Node

	// Name is the name of the data-section
	Name string

	// Contents holds the string/byte data for the reference
	Contents []byte
}

// String outputs this Data structure as a string.
func (d Data) String() string {
	return fmt.Sprintf("<DATA: name:%s data:%v>", d.Name, d.Contents)
}

// Operand is used to hold the operand for an instruction.
//
// Some instructions have zero operands (e.g. `nop`), others have
// one (e.g. `inc rax`), and finally we have several which take two
// operands (e.g. `mov rax, rbx`).
//
type Operand struct {
	// Token contains our parent token.
	token.Token

	// If we're operating upon memory-addresses we need to be
	// able to understand the size of the thing we're operating
	// upon.
	//
	// For example `inc byte ptr [rax]` will increment a byte,
	// or 8 bits.  We have different define sizes available to us:
	//
	//   byte -> 8 bits.
	//   word -> 16 bits.
	//  dword -> 32 bites.
	//  qword -> 64 bites.
	Size int

	// Is indirection used?
	//
	// i.e. `rax` has no indirection, but `[rax]` does.
	Indirection bool
}

// Instruction holds a parsed instruction.
//
// For example "mov rax, rax".
//
type Instruction struct {
	Node

	// Instruction holds the instruction we've found, as a string.
	Instruction string

	// Operands holds the operands for this instruction.
	//
	// Operands will include numbers, registers, and indrected registers.
	Operands []Operand
}

// String outputs this Error structure as a string
func (d Instruction) String() string {
	return fmt.Sprintf("<INSTRUCTION: %s args:%v>", d.Instruction, d.Operands)
}

// Label holds a label, as seen when it is defined.
//
// For example ":foo" will define a label with name "foo".
type Label struct {
	Node

	// Name has the name of the instruction
	Name string
}

// String outputs this Label structure as a string.
func (l Label) String() string {
	return fmt.Sprintf("<LABEL: %s>", l.Name)
}
