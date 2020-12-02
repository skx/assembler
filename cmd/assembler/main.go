package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/assembler/compiler"
)

func main() {

	//
	// Ensure we have an argument
	//
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: compiler input.asm\n")
		return
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("error:%s\n", err.Error())
		return
	}

	// Create the compiler
	c := compiler.New(string(data))

	c.SetOutput("./a.out")

	err = c.Compile()
	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		return
	}
}
