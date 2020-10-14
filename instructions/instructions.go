// Package instructions contains the comment instruction-definitions
// for the instructions that we understand.
//
// These are abstracted here, so that you only don't need to touch
// the parser/lexer to add new instructions.
//
// Just add the instructions here, and update the compiler to emit the
// appropriate code.
package instructions

var (
	// InstructionLengths is a map that returns the number of operands
	// the given assembly-language operation will accept.
	//
	// For example a `nop` argument requires zero arguments so the
	// entry for that will be `0`.
	InstructionLengths map[string]int

	// Instructions is automatically generated from the InstructionLengths
	// map, and contains the known instruction-types we can lex, parse, and
	// compile.
	Instructions []string
)

func init() {

	// Setup our instruction-lengths
	InstructionLengths = make(map[string]int)

	InstructionLengths["add"] = 2
	InstructionLengths["dec"] = 1
	InstructionLengths["inc"] = 1
	InstructionLengths["int"] = 1
	InstructionLengths["mov"] = 2
	InstructionLengths["nop"] = 0
	InstructionLengths["pop"] = 1
	InstructionLengths["push"] = 1
	InstructionLengths["ret"] = 0
	InstructionLengths["sub"] = 2
	InstructionLengths["xor"] = 2

	// jump
	InstructionLengths["je"] = 1
	InstructionLengths["jmp"] = 1
	InstructionLengths["jne"] = 1
	InstructionLengths["jnz"] = 1
	InstructionLengths["jz"] = 1

	// Processor control instructions
	InstructionLengths["clc"] = 0
	InstructionLengths["cld"] = 0
	InstructionLengths["cli"] = 0
	InstructionLengths["cmc"] = 0
	InstructionLengths["stc"] = 0
	InstructionLengths["std"] = 0
	InstructionLengths["sti"] = 0

	// Now record the known-instructions
	for k := range InstructionLengths {
		Instructions = append(Instructions, k)
	}
}
