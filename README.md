[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/skx/assembler)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/assembler)](https://goreportcard.com/report/github.com/skx/assembler)
[![license](https://img.shields.io/github/license/skx/assembler.svg)](https://github.com/skx/assembler/blob/master/LICENSE)

* [Assembler](#assembler)
  * [Limitations](#limitations)
  * [Installation](#installation)
  * [Example Usage](#example-usage)
* [Internals](#internals)
  * [Adding New Instructions](#adding-new-instructions)
  * [Debugging Generated Binaries](#debugging-generated-binaries)
* [Bugs?](#bugs)


# Assembler

This repository contains a VERY BASIC x86-64 assembler, which is capable of
reading assembly-language input, and generating a staticly linked ELF binary
output.

It is more a proof-of-concept than a useful assembler, but I hope to take it to the state where it can compile the kind of x86-64 assembly I produce in some of my other projects.

Currently the assembler will generate a binary which looks like this:

```
$ file a.out
a.out: ELF 64-bit LSB executable, x86-64, version 1 (SYSV)
       statically linked, no section header
```

Why?  I've written a couple of toy projects that generate assembly language programs, then pass them through an assembler:

* [brainfuck compiler](https://github.com/skx/bfcc/)
* [math compiler](https://github.com/skx/math-compiler/)

The code in this repository was born out of the process of experimenting with generating an ELF binary directly.  A necessary learning-process.



## Limitations

We don't support anywhere near the complete instruction-set which an assembly language programmer would expect.  Currently we support only things like this:

* `add $REG, $REG` + `add $REG, $NUMBER`
  * Add a number, or the contents of another register, to a register.
* `dec $REG`
  * Decrement the contents of the given register.
* `inc $REG`
  * Increment the contents of the given register.
* `mov $REG, $NUMBER`
* `mov $REG, $REG`
  * Move a number into the specified register.
* `nop`
  * Do nothing.
* `push $NUMBER`, or `push $IDENTIFIER`
  * See [jmp.asm](jmp.asm) for an example.
* `ret`
  * Return from call.
  * **NOTE**: We don't actually support making calls, though that can be emulated via `push` - see [jmp.asm](jmp.asm) for an example.
* `sub $REG, $REG` + `sub $REG, $NUMBER`
  * Subtract a number, or the contents of another register, from a register.
* `xor $REG, $REG`
  * Set the given register to be zero.
* `int $NUM`
  * Call the kernel.
* Processor (flag) control instructions:
  * `clc`, `cld`, `cli`, `cmc`, `stc`, `std`, and `sti`.

Note that in **all cases** we only support the following set of (four) registers:

* `rax`
* `rbx`
* `rcx`
* `rdx`

There is support for storing fixed-data within our program, and locating that.  See [hello.asm](hello.asm) for an example of that.

We also have some other (obvious) limitations:

* There is notably no support for comparison instructions, and jumping instructions.
  * We _emulate_ (unconditional) jump instructions via "`push`" and "`ret`", see [jmp.asm](jmp.asm) for an example of that.
* The entry-point is __always__ at the beginning of the source.
* You can only reference data AFTER it has been declared.
  * These are added to the `data` section of the generated binary, but must be defined first.
  * See [hello.asm](hello.asm) for an example of that.



## Installation

If you have this repository cloned locally you can build the assembler like so:

    cd cmd/assembler
    go build .
    go install .

If you wish to fetch and install via your existing toolchain:

    go get -u github.com/skx/assembler/cmd/assembler

You can repeat for the other commands if you wish:

    go get -u github.com/skx/assembler/cmd/lexer
    go get -u github.com/skx/assembler/cmd/parser

Of course these binary-names are very generic, so perhaps better to work locally!


## Example Usage

Build the assembler:

     $ cd cmd/assembler
     $ go build .

Compile the [sample program](test.asm), and execute it showing the return-code:

     $ cmd/assembler/assembler test.asm && ./a.out ; echo $?
     9

Or run the [hello.asm](hello.asm) example:

     $ cmd/assembler/assembler  hello.in && ./a.out
     Hello, world
     Goodbye, world

You'll note that the `\n` character was correctly expanded into a newline.


# Internals

The core of our code consists of a small number of simple packages:

* A simple tokenizer [lexer/lexer.go](lexer/lexer.go)
* A simple parser [parser/parser.go](parser/parser.go)
  * This populates a simple internal-form/AST [parser/ast.go](parser/ast.go).
* A simple compiler [compiler/compiler.go](compiler/compiler.go)
* A simple elf-generator [elf/elf.go](elf/elf.go)
  * Taken from [vishen/go-x64-executable](https://github.com/vishen/go-x64-executable/).


In addition to the package modules we also have a couple of binaries:

* `cmd/lexer`
  * Show the output of lexing a program.
  * This is useful for debugging and development-purposes, it isn't expected to be useful to end-users.
* `cmd/parser`
  * Show the output of parsing a program.
    * This is useful for debugging and development-purposes, it isn't expected to be useful to end-users.
* `cmd/assembler`
  * Assemble a program, producing an executable binary.

These commands located beneath `cmd` each operate the same way.  They each take a single argument which is a file containing assembly-language instructions.

For example here is how you'd build and test the parser:

    cd cmd/parser
    go build .
    $ ./parser ../../test.asm
    &{{INSTRUCTION xor} [{REGISTER rax} {REGISTER rax}]}
    &{{INSTRUCTION inc} [{REGISTER rax}]}
    &{{INSTRUCTION mov} [{REGISTER rbx} {NUMBER 0x0000}]}
    &{{INSTRUCTION mov} [{REGISTER rcx} {NUMBER 0x0007}]}
    &{{INSTRUCTION add} [{REGISTER rbx} {REGISTER rcx}]}
    &{{INSTRUCTION mov} [{REGISTER rcx} {NUMBER 0x0002}]}
    &{{INSTRUCTION add} [{REGISTER rbx} {REGISTER rcx}]}
    &{{INSTRUCTION int} [{NUMBER 0x80}]}


## Adding New Instructions

This is how you might add a new instruction to the assembler, for example you might add `jmp 0x00000` or some similar instruction:

* Add a new entry for the instruction in [instructions/instructions.go](instructions/instructions.go)
  * i.e. Update `InstructionLengths` map to add the instruction.
  * This will be used by both the tokenization process, and the parser.
* Generate the appropriate output in `compiler/compiler.go`, inside the function `compileInstruction`.
  * i.e. Emit the binary-code for the instruction.



## Debugging Generated Binaries

Launch the binary under gdb:

    $ gdb ./a.out

Start it:

    (gdb) starti
    Starting program: /home/skx/Repos/github.com/skx/assembler/a.out

    Program stopped.
    0x00000000004000b0 in ?? ()

Dissassemble:

    (gdb)  x/5i $pc

Or show string-contents at an address:

    (gdb) x/s 0x400000


# Bugs?

Feel free to report, as this is more a proof of concept rather than a robust tool they are to be expected.

Specifically I expect that we're missing support for many instructions, but I hope the code generated for those that is present is correct.


Steve
