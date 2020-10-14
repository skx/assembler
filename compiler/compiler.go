// Package compiler is the package which is actually responsible for reading
// the user-program and generating the binary result.
//
// Internally this uses the parser, as you would expect
package compiler

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/skx/assembler/elf"
	"github.com/skx/assembler/parser"
	"github.com/skx/assembler/token"
)

// Compiler holds our state
type Compiler struct {

	// p holds the parser we use to generate AST
	p *parser.Parser

	// output holds the path to the binary we'll generate
	output string

	// code contains the code we generate
	code []byte

	// data is where we place constant-strings, etc.
	data []byte

	// map of "data-name" to "data-offset"
	dataOffsets map[string]int

	// patches we have to make, post-compilation.  Don't ask
	patches map[int]int

	// labels and the corresponding offsets we've seen.
	labels map[string]int

	// offsets which contain jumps to labels
	labelTargets map[int]string

	// offsets for relative label-jumps
	jmps map[int]string
}

// New creates a new instance of the compiler
func New(src string) *Compiler {

	c := &Compiler{p: parser.New(src), output: "a.out"}
	c.dataOffsets = make(map[string]int)
	c.patches = make(map[int]int)

	// mapping of "label -> XXX"
	c.labels = make(map[string]int)

	// fixups we need to make offset-of-code -> label
	c.labelTargets = make(map[int]string)

	// jump-fixups
	c.jmps = make(map[int]string)

	return c
}

// SetOutput sets the path to the executable we create.
//
// If no output has been specified we default to `./a.out`.
func (c *Compiler) SetOutput(path string) {
	c.output = path
}

// Compile walks over the parser-generated AST and assembles the source
// program.
//
// Once the program has been completed an ELF executable will be produced
func (c *Compiler) Compile() error {

	//
	// Walk over the parser-output
	//
	stmt := c.p.Next()
	for stmt != nil {

		switch stmt := stmt.(type) {

		case parser.Data:
			c.handleData(stmt)

		case parser.Error:
			return fmt.Errorf("error compiling - parser returned error %s", stmt.Value)

		case parser.Label:
			// So now we know the label with the given name
			// corresponds to the CURRENT position in the
			// generated binary-code.
			//
			// If anything refers to this we'll have to patch
			// it up
			c.labels[stmt.Name] = len(c.code)

		case parser.Instruction:
			err := c.compileInstruction(stmt)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unhandled node-type %v", stmt)
		}

		stmt = c.p.Next()
	}

	//
	// Apply data-patches.
	//
	// This is horrid.
	//
	for o, v := range c.patches {

		// start of virtual sectoin
		//  + offset
		//  + len of code segment
		//  + elf header
		//  + 2 * program header
		// life is hard
		v = 0x400000 + v + len(c.code) + 0x40 + (2 * 0x38)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))

		for i, x := range buf {
			c.code[i+o] = x
		}
	}

	//
	// OK now we need to patch references to labels
	//
	for o, s := range c.labelTargets {

		offset := c.labels[s]

		offset = 0x400000 + offset + 0x40 + (2 * 0x38)

		// So we have a new offset.

		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(offset))

		for i, x := range buf {
			c.code[i+o] = x
		}
	}

	// Patchup the jumps
	for o, s := range c.jmps {

		// the offset of the instruction
		offset := c.labels[s]

		fmt.Printf("Offset of label is %x\n", offset)

		// the offset of the position is a byte
		diff := uint(o - offset)
		fmt.Printf("Diff is is %x: %x\n", diff, byte(diff))

		c.code[o] = byte(0xff - byte(diff))
	}

	//
	// Write.  The.  Elf.  Output.
	//
	e := elf.New()
	err := e.WriteContent(c.output, c.code, c.data)
	if err != nil {
		return fmt.Errorf("error writing elf: %s", err.Error())
	}

	return nil

}

// handleData appends the data to the data-section of our binary,
// and stores the offset appropriately
func (c *Compiler) handleData(d parser.Data) {

	// Offset of the start of the data is the current
	// length of the existing data.
	offset := len(c.data)

	// Add
	c.data = append(c.data, d.Contents...)

	// Save
	c.dataOffsets[d.Name] = offset

	// TODO: Do we care about alignment?  We might
	// in the future.
}

// compileInstruction handles the instruction generation
func (c *Compiler) compileInstruction(i parser.Instruction) error {

	switch i.Instruction {

	case "add":
		err := c.assembleADD(i)
		if err != nil {
			return err
		}
		return nil

	case "clc":
		c.code = append(c.code, 0xf8)
		return nil

	case "cld":
		c.code = append(c.code, 0xfc)
		return nil

	case "cli":
		c.code = append(c.code, 0xfa)
		return nil

	case "cmc":
		c.code = append(c.code, 0xf5)
		return nil

	case "dec":
		err := c.assembleDEC(i)
		if err != nil {
			return err
		}
		return nil

	case "inc":
		err := c.assembleINC(i)
		if err != nil {
			return err
		}
		return nil

	case "int":
		n, err := c.argToByte(i.Operands[0].Token)
		if err != nil {
			return err
		}
		c.code = append(c.code, 0xcd)
		c.code = append(c.code, n)
		return nil

	case "jmp", "jne", "je", "jz", "jnz":
		err := c.assembleJMP(i)
		if err != nil {
			return err
		}
		return nil

	case "mov":
		err := c.assembleMov(i, false)
		if err != nil {
			return err
		}
		return nil

	case "nop":
		c.code = append(c.code, 0x90)
		return nil

	case "pop":
		err := c.assemblePop(i)
		if err != nil {
			return err
		}
		return nil

	case "push":
		err := c.assemblePush(i)
		if err != nil {
			return err
		}
		return nil

	case "ret":
		c.code = append(c.code, 0xc3)
		return nil

	case "stc":
		c.code = append(c.code, 0xf9)
		return nil

	case "std":
		c.code = append(c.code, 0xfd)
		return nil

	case "sti":
		c.code = append(c.code, 0xfb)
		return nil

	case "sub":
		err := c.assembleSUB(i)
		if err != nil {
			return err
		}
		return nil
	case "xor":
		err := c.assembleXOR(i)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown instruction %v", i)
}

// return register number - used for `dec`, `inc`, and `mov`.
func (c *Compiler) getreg(reg string) int {

	// registers
	registers := []string{
		"rax",
		"rcx",
		"rdx",
		"rbx",
		"rsp",
		"rbp",
		"rsi",
		"rdi"}

	for i, name := range registers {
		if reg == name {
			return i
		}
	}

	panic(fmt.Sprintf("failed to lookup register: %s", reg))
}

// get magic value for two-register operations (`add`, `sub`, `xor`).
func (c *Compiler) calcRM(dest string, src string) byte {

	// registers
	registers := []string{
		"rax",
		"rcx",
		"rdx",
		"rbx",
		"rsp",
		"rbp",
		"rsi",
		"rdi"}

	dN := -1
	sN := -1

	for i, reg := range registers {
		if reg == dest {
			dN = i
		}
		if reg == src {
			sN = i

		}
	}

	if dN < 0 || sN < 0 {
		panic(fmt.Sprintf("failed to lookup registers: %s %s", src, dest))
	}

	out := 0xc0 + (8 * sN) + dN
	if out > 255 {
		panic("calcRM received out of bounds value")
	}
	return byte(out)
}

// used by `int`
func (c *Compiler) argToByte(t token.Token) (byte, error) {

	num, err := strconv.ParseInt(t.Literal, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to convert %s to number %s", t.Literal, err)
	}

	return byte(num), nil
}

// used by `mov`
func (c *Compiler) argToByteArray(t token.Token) ([]byte, error) {

	// Store the result here
	buf := make([]byte, 4)

	num, err := strconv.ParseInt(t.Literal, 0, 64)
	if err != nil {
		return buf, fmt.Errorf("unable to convert %s to number for register %s", t.Literal, err)
	}

	binary.LittleEndian.PutUint32(buf, uint32(num))
	return buf, nil
}

// assembleADD handles addition.
func (c *Compiler) assembleADD(i parser.Instruction) error {

	// Two registers added?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.REGISTER {
		c.code = append(c.code, []byte{0x48, 0x01}...)
		out := c.calcRM(i.Operands[0].Literal, i.Operands[1].Literal)
		c.code = append(c.code, out)
		return nil
	}

	// OK number added to a register?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.NUMBER {

		// Convert the integer to a four-byte/64-bit value
		n, err := c.argToByteArray(i.Operands[1].Token)
		if err != nil {
			return err
		}

		// Work out the register
		switch i.Operands[0].Literal {
		case "rax":
			c.code = append(c.code, []byte{0x48, 0x05}...)
		case "rbx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xc3}...)
		case "rcx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xc1}...)
		case "rdx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xc2}...)
		default:
			return fmt.Errorf("add %s, number not implemented", i.Operands[0].Literal)
		}

		// Now append the value
		c.code = append(c.code, n...)
		return nil
	}

	return fmt.Errorf("unhandled ADD instruction %v", i)
}

// accembleDEC handles dec rax, rbx, etc.
func (c *Compiler) assembleDEC(i parser.Instruction) error {

	// Decrement the contents of a register
	if i.Operands[0].Indirection == false {
		// prefix
		c.code = append(c.code, []byte{0x48, 0xff}...)

		// register name
		reg := 0xc0 + c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: byte
	if i.Operands[0].Size == 8 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0xfe}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		reg += 0x08
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: word
	if i.Operands[0].Size == 16 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0x66, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		reg += 0x08
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: double word
	if i.Operands[0].Size == 32 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		reg += 0x08
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: quad word
	if i.Operands[0].Size == 64 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0x48, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		reg += 0x08
		c.code = append(c.code, byte(reg))

		return nil
	}

	return fmt.Errorf("unknown argument for DEC %v", i)
}

// assembleINC handles inc rax, rbx, etc.
func (c *Compiler) assembleINC(i parser.Instruction) error {

	// Increment the contents of a register
	if i.Operands[0].Indirection == false {
		// prefix
		c.code = append(c.code, []byte{0x48, 0xff}...)

		// register name
		reg := 0xc0 + c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: byte
	if i.Operands[0].Size == 8 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0xfe}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: word
	if i.Operands[0].Size == 16 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0x66, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: double word
	if i.Operands[0].Size == 32 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	// indirect: quad word
	if i.Operands[0].Size == 64 {
		// prefix
		c.code = append(c.code, []byte{0x67, 0x48, 0xff}...)

		// register name
		reg := c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		return nil
	}

	return fmt.Errorf("unknown argument for INC %v", i)
}

// assembleJMP handles all the jump instructions
//
// NOTE We have to fixup the offsets here.
func (c *Compiler) assembleJMP(i parser.Instruction) error {

	var byte byte

	switch i.Instruction {
	case "jmp":
		byte = 0xeb
	case "je", "jz":
		byte = 0x74
	case "jne", "jnz":
		byte = 0x75
	default:
		return fmt.Errorf("unknown jmp type")
	}

	// Ensure we're jumping to a label
	if i.Operands[0].Type != token.IDENTIFIER {
		return fmt.Errorf("we only support jumps to labels at the moment")
	}

	// emit the instruction and make a note of the fixup to make
	c.code = append(c.code, byte)
	c.jmps[len(c.code)] = i.Operands[0].Literal
	c.code = append(c.code, 0x00) // empty displacement

	return nil
}

func (c *Compiler) assembleMov(i parser.Instruction, label bool) error {

	//
	// Are we moving a register to another register?
	//
	if i.Operands[0].Type == token.REGISTER && i.Operands[1].Type == token.REGISTER {

		c.code = append(c.code, []byte{0x48, 0x89}...)
		out := c.calcRM(i.Operands[0].Literal, i.Operands[1].Literal)
		c.code = append(c.code, out)
		return nil

	}

	//
	// Are we moving a number to a register ?
	//
	if i.Operands[0].Type == token.REGISTER && i.Operands[1].Type == token.NUMBER {

		// prefix
		c.code = append(c.code, []byte{0x48, 0xc7}...)

		// register name
		reg := 0xc0 + c.getreg(i.Operands[0].Literal)
		c.code = append(c.code, byte(reg))

		// value
		n, err := c.argToByteArray(i.Operands[1].Token)
		if err != nil {
			return err
		}

		// hack
		if label {
			c.patches[len(c.code)], _ = strconv.Atoi(i.Operands[1].Literal)
		}
		c.code = append(c.code, n...)
		return nil
	}

	// mov $reg, $id
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.IDENTIFIER {

		//
		// Lookup the identifier, and if we can find it
		// then we will treat it as a constant
		//
		name := i.Operands[1].Literal
		val, ok := c.dataOffsets[name]
		if ok {

			i.Operands[1].Type = token.NUMBER
			i.Operands[1].Literal = fmt.Sprintf("%d", val)
			return c.assembleMov(i, true)
		}
		return fmt.Errorf("reference to unknown label/data: %v", i.Operands[1])
	}

	return fmt.Errorf("unknown MOV instruction: %v", i)

}

// assemblePop would compile "pop offset", and "push 0x1234"
func (c *Compiler) assemblePop(i parser.Instruction) error {

	// known pop-types
	table := make(map[string][]byte)
	table["rax"] = []byte{0x58}
	table["rbx"] = []byte{0x5b}
	table["rcx"] = []byte{0x59}
	table["rdx"] = []byte{0x5a}
	table["rbp"] = []byte{0x5d}
	table["rsp"] = []byte{0x5c}
	table["rsi"] = []byte{0x5e}
	table["rdi"] = []byte{0x5f}
	table["r8"] = []byte{0x41, 0x58}
	table["r9"] = []byte{0x41, 0x59}
	table["r10"] = []byte{0x41, 0x5a}
	table["r11"] = []byte{0x41, 0x5b}
	table["r12"] = []byte{0x41, 0x5c}
	table["r13"] = []byte{0x41, 0x5d}
	table["r14"] = []byte{0x41, 0x5e}
	table["r15"] = []byte{0x41, 0x5f}

	// Is this "pop rax|rbx..|rdx", or something in the table?
	if i.Operands[0].Type == token.REGISTER {
		bytes, ok := table[i.Operands[0].Literal]
		if ok {
			c.code = append(c.code, bytes...)
			return nil
		}
		return fmt.Errorf("unknown register in 'pop'")
	}

	return fmt.Errorf("unknown pop-type: %v", i)

}

// assemblePush would compile "push offset", and "push 0x1234"
func (c *Compiler) assemblePush(i parser.Instruction) error {

	// Is this a number?  Just output it
	if i.Operands[0].Type == token.NUMBER {
		n, err := c.argToByteArray(i.Operands[1].Token)
		if err != nil {
			return err
		}
		c.code = append(c.code, 0x68)
		c.code = append(c.code, n...)
		return nil
	}

	// Is this a label?
	if i.Operands[0].Type == token.IDENTIFIER {

		c.code = append(c.code, 0x68)

		c.labelTargets[len(c.code)] = i.Operands[0].Literal

		c.code = append(c.code, []byte{0x0, 0x0, 0x0, 0x0}...)
		return nil
	}

	// is this a register?
	table := make(map[string][]byte)
	table["rax"] = []byte{0x50}
	table["rcx"] = []byte{0x51}
	table["rdx"] = []byte{0x52}
	table["rbx"] = []byte{0x53}
	table["rsp"] = []byte{0x54}
	table["rbp"] = []byte{0x55}
	table["rsi"] = []byte{0x56}
	table["rdi"] = []byte{0x57}
	table["r8"] = []byte{0x41, 0x50}
	table["r9"] = []byte{0x41, 0x51}
	table["r10"] = []byte{0x41, 0x52}
	table["r11"] = []byte{0x41, 0x53}
	table["r12"] = []byte{0x41, 0x54}
	table["r13"] = []byte{0x41, 0x55}
	table["r14"] = []byte{0x41, 0x56}
	table["r15"] = []byte{0x41, 0x57}

	// Is this "push rax|rbx..|rdx", or something in the table?
	if i.Operands[0].Type == token.REGISTER {
		bytes, ok := table[i.Operands[0].Literal]
		if ok {
			c.code = append(c.code, bytes...)
			return nil
		}
		return fmt.Errorf("unknown register in 'push'")
	}

	return fmt.Errorf("unknown push-type: %v", i)
}

// assembleSUB handles subtraction.
func (c *Compiler) assembleSUB(i parser.Instruction) error {

	// Two registers subtracted?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.REGISTER {
		c.code = append(c.code, []byte{0x48, 0x29}...)
		out := c.calcRM(i.Operands[0].Literal, i.Operands[1].Literal)
		c.code = append(c.code, out)
		return nil
	}

	// OK number subtracted from a register?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.NUMBER {

		// Convert the integer to a four-byte/64-bit value
		n, err := c.argToByteArray(i.Operands[1].Token)
		if err != nil {
			return err
		}

		// Work out the register
		switch i.Operands[0].Literal {
		case "rax":
			c.code = append(c.code, []byte{0x48, 0x2d}...)
		case "rbx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xeb}...)
		case "rcx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xe9}...)
		case "rdx":
			c.code = append(c.code, []byte{0x48, 0x81, 0xea}...)
		default:
			return fmt.Errorf("SUB %s, number not implemented", i.Operands[0].Literal)
		}

		// Now append the value
		c.code = append(c.code, n...)
		return nil
	}

	return fmt.Errorf("unhandled SUB instruction %v", i)
}

// assembleXOR handles xor rax, rbx, etc.
func (c *Compiler) assembleXOR(i parser.Instruction) error {

	// Two registers xor'd?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.REGISTER {
		c.code = append(c.code, []byte{0x48, 0x31}...)
		out := c.calcRM(i.Operands[0].Literal, i.Operands[1].Literal)
		c.code = append(c.code, out)
		return nil
	}

	return fmt.Errorf("unknown argument for XOR %v", i)
}
