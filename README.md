# Assembler

This repository contains a VERY VERY BASIC assembler, which is capable of
reading simple assembly-language programs, and generating an ELF binary
from them.

Specifically this will generate a binary which looks like this:

```
$ file a.out
a.out: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, no section header
```

Why?  I've written a couple of toy projects that generate assembly language programs, then pass them through an assembler:

* [brainfuck compiler](https://github.com/skx/bfcc/)
* [math compiler](https://github.com/skx/math-compiler/)

The code in this repository was born out of the process of experimenting with generating an ELF binary directly.  A necessary learning-process.



## Limitations

We don't support anywhere near the complete instruction-set which an assembly language programmer would expect.  Currently we support only things like this:

* `mov $REG, $NUM`
* `mov $REG, $REG`
  * Move a number into the specified register.
* `xor $REG, $REG`
  * Set the given register to be zero.
* `inc $REG`
  * Increment the contents of the given register.
* `add $REG, $REG`
* `int 0x80`
  * Call the kernel

Supported registers are limited to the 64-bit registers:

* `rax`
* `rbx`
* `rcx`
* `rdx`.

No other registers are supported.

There is support for storing fixed-data within our program, and locating that.  See [hello.asm](hello.asm) for an example of that.


## Example

Build the assembler:

     $ go build .

Compile the [sample program](test.asm), and execute it showing the return-code:

     $ ./assembler  test.in  && ./a.out  ; echo $?
     9

Or run the [hello.asm](hello.asm) example:

     $ ./assembler  hello.in  && ./a.out
     Hello, world\nGGoodbye, world\n

Meh, close enough..


## Internals

I'm slowly moving towards a better structure, although this is in-flux.  You
can see various tools beneath [cmd/](cmd/) for example:

* `cmd/lexer`
  * Show the output of lexing a program.
* `cmd/parser`
  * Show the output of parsing a program.

Both of those operate the same way, so for example:

    cd cmd/parser
    go build .
    ./parser ../../test.in
    $ ./parser ../../test.asm
    &{{INSTRUCTION xor} [{REGISTER rax} {REGISTER rax}]}
    &{{INSTRUCTION inc} [{REGISTER rax}]}
    &{{INSTRUCTION mov} [{REGISTER rbx} {NUMBER 0x0000}]}
    &{{INSTRUCTION mov} [{REGISTER rcx} {NUMBER 0x0007}]}
    &{{INSTRUCTION add} [{REGISTER rbx} {REGISTER rcx}]}
    &{{INSTRUCTION mov} [{REGISTER rcx} {NUMBER 0x0002}]}
    &{{INSTRUCTION add} [{REGISTER rbx} {REGISTER rcx}]}
    &{{INSTRUCTION int} [{NUMBER 0x80}]}

(The lexer would give a simple stream of tokens, instead of the parsed instructions and their operands.  But the same basic usage is present.)

In the future we'll have `cmd/compiler` to run the compilation process, but that is still work in progress.



## Debugging

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

Steve
