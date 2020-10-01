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

As you'll discover we do not have a proper parser, so we have considered only a couple of instructions:

* `mov $REG, $NUM`
  * Move a number into the specified register.
* `xor $REG, [$REG]`
  * Set the given register to be zero.
* `inc $REG`
  * Increment the contents of the given register.
* `add rbx, rcx`
  * `rbx` = `rbx` + `rcx`
* `int 0x800`
  * Call the kernel

Supported registers only include `rax`, `rbx`, `rcx`, and `rdx`.  Zero other registers are supported.  No other instructions are supported.

We have zero support for control-flow, or compiler directives.

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
