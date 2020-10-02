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
}

// New creates a new instance of the compiler
func New(src string) *Compiler {

	c := &Compiler{p: parser.New(src), output: "a.out"}
	c.dataOffsets = make(map[string]int)
	c.patches = make(map[int]int)

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
			//		case parser.Label:
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

	offset := len(c.data)
	c.data = append(c.data, d.Contents...)

	c.dataOffsets[d.Name] = offset
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
	case "inc":
		err := c.assembleINC(i)
		if err != nil {
			return err
		}
		return nil
	case "int":
		n, err := c.argToByte(i.Operands[0])
		if err != nil {
			return err
		}
		c.code = append(c.code, 0xcd)
		c.code = append(c.code, n)
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
	case "xor":
		err := c.assembleXOR(i)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown instruction %v", i)

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

	// Add instructions - we use a simple table for the register-
	// register-case.
	type regs struct {
		A string
		B string
	}
	// Create a simple map
	codes := make(map[regs]([]byte))

	codes[regs{A: "rax", B: "rax"}] = []byte{0x48, 0x01, 0xc0}
	codes[regs{A: "rax", B: "rbx"}] = []byte{0x48, 0x01, 0xd8}
	codes[regs{A: "rax", B: "rcx"}] = []byte{0x48, 0x01, 0xc8}
	codes[regs{A: "rax", B: "rdx"}] = []byte{0x48, 0x01, 0xd0}

	codes[regs{A: "rbx", B: "rax"}] = []byte{0x48, 0x01, 0xc3}
	codes[regs{A: "rbx", B: "rbx"}] = []byte{0x48, 0x01, 0xdb}
	codes[regs{A: "rbx", B: "rcx"}] = []byte{0x48, 0x01, 0xcb}
	codes[regs{A: "rbx", B: "rdx"}] = []byte{0x48, 0x01, 0xd3}

	codes[regs{A: "rcx", B: "rax"}] = []byte{0x48, 0x01, 0xc1}
	codes[regs{A: "rcx", B: "rbx"}] = []byte{0x48, 0x01, 0xd9}
	codes[regs{A: "rcx", B: "rcx"}] = []byte{0x48, 0x01, 0xc9}
	codes[regs{A: "rcx", B: "rdx"}] = []byte{0x48, 0x01, 0xd1}

	codes[regs{A: "rdx", B: "rax"}] = []byte{0x48, 0x01, 0xc2}
	codes[regs{A: "rdx", B: "rbx"}] = []byte{0x48, 0x01, 0xda}
	codes[regs{A: "rdx", B: "rcx"}] = []byte{0x48, 0x01, 0xca}
	codes[regs{A: "rdx", B: "rdx"}] = []byte{0x48, 0x01, 0xd2}

	// simple registers?
	bytes, ok := codes[regs{A: i.Operands[0].Literal,
		B: i.Operands[1].Literal}]

	if ok {
		c.code = append(c.code, bytes...)
		return nil
	}

	// OK number added to a register?
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.NUMBER {

		// Convert the integer to a four-byte/64-bit value
		n, err := c.argToByteArray(i.Operands[1])
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

// assembleINC handles inc rax, rbx, etc.
func (c *Compiler) assembleINC(i parser.Instruction) error {

	if i.Operands[0].Literal == "rax" {
		c.code = append(c.code, []byte{0x48, 0xff, 0xc0}...)
		return nil
	}
	if i.Operands[0].Literal == "rbx" {
		c.code = append(c.code, []byte{0x48, 0xff, 0xc3}...)
		return nil
	}
	if i.Operands[0].Literal == "rcx" {
		c.code = append(c.code, []byte{0x48, 0xff, 0xc1}...)
		return nil
	}
	if i.Operands[0].Literal == "rdx" {
		c.code = append(c.code, []byte{0x48, 0xff, 0xc2}...)
		return nil
	}
	return fmt.Errorf("unknown argument for XOR %v", i)
}

func (c *Compiler) assembleMov(i parser.Instruction, label bool) error {

	//
	// Are we moving a register to another register?
	//
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.REGISTER {
		fmt.Printf("TODO: mov reg,reg\n")
		return nil
	}

	//
	// Are we moving a number to a register ?
	//
	if i.Operands[0].Type == token.REGISTER &&
		i.Operands[1].Type == token.NUMBER {

		if i.Operands[0].Literal == "rax" {
			c.code = append(c.code, []byte{0x48, 0xc7, 0xc0}...)

			n, err := c.argToByteArray(i.Operands[1])
			if err != nil {
				return err
			}

			if label {
				c.patches[len(c.code)], _ = strconv.Atoi(i.Operands[1].Literal)
			}
			c.code = append(c.code, n...)
			return nil
		}
		if i.Operands[0].Literal == "rbx" {
			c.code = append(c.code, []byte{0x48, 0xc7, 0xc3}...)
			n, err := c.argToByteArray(i.Operands[1])
			if err != nil {
				return err
			}
			if label {
				c.patches[len(c.code)], _ = strconv.Atoi(i.Operands[1].Literal)
			}
			c.code = append(c.code, n...)
			return nil
		}
		if i.Operands[0].Literal == "rcx" {
			c.code = append(c.code, []byte{0x48, 0xc7, 0xc1}...)
			n, err := c.argToByteArray(i.Operands[1])
			if err != nil {
				return err
			}
			if label {
				c.patches[len(c.code)], _ = strconv.Atoi(i.Operands[1].Literal)
			}
			c.code = append(c.code, n...)
			return nil
		}
		if i.Operands[0].Literal == "rdx" {
			c.code = append(c.code, []byte{0x48, 0xc7, 0xc2}...)
			n, err := c.argToByteArray(i.Operands[1])
			if err != nil {
				return err
			}
			if label {
				c.patches[len(c.code)], _ = strconv.Atoi(i.Operands[1].Literal)
			}
			c.code = append(c.code, n...)
			return nil
		}

		return fmt.Errorf("moving a constant (number) to an unknown register: %v", i)
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

// assembleXOR handles xor rax, rbx, etc.
func (c *Compiler) assembleXOR(i parser.Instruction) error {

	if i.Operands[0].Literal == "rax" {
		c.code = append(c.code, []byte{0x48, 0x31, 0xc0}...)
		return nil
	}
	if i.Operands[0].Literal == "rbx" {
		c.code = append(c.code, []byte{0x48, 0x31, 0xdb}...)
		return nil
	}
	if i.Operands[0].Literal == "rcx" {
		c.code = append(c.code, []byte{0x48, 0x31, 0xc9}...)
		return nil
	}
	if i.Operands[0].Literal == "rdx" {
		c.code = append(c.code, []byte{0x48, 0x31, 0xd2}...)
		return nil
	}
	return fmt.Errorf("unknown argument for XOR %v", i)
}
