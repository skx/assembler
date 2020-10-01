package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/skx/assembler/elf"
)

// Assemble reads the given input, line by line, and assemble
// the instructions.
func Assemble(path string) error {

	//
	// Open the file to read it.
	//
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	//
	// Mapping of labels to data.
	//
	label := make(map[string]int)
	patches := make(map[int]int)

	//
	// This is where we assemble our text.
	//
	text := []byte{}
	data := []byte{}

	//
	// "Assemble"
	//
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		orig := line

		line = strings.TrimSpace(line)

		// Empty line?
		if line == "" {
			continue
		}

		// Comment?
		if strings.HasPrefix(line, "//") ||
			strings.HasPrefix(line, ";") {
			continue
		}

		// Data storage of a string
		if strings.HasPrefix(line, ".") {

			//
			// We assume
			//
			///  .NAME DB "STRING"
			//
			fields := strings.Split(line, " ")
			if fields[1] != "DB" {
				return fmt.Errorf("unknown label-type in %s", orig)
			}

			//
			// String might have spaces so we go back and strip the
			// prefix
			//
			line = orig
			line = strings.TrimPrefix(line, fields[0])
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, fields[1])
			line = strings.TrimSpace(line)

			// Add the string
			line = strings.TrimPrefix(line, "\"")
			line = strings.TrimSuffix(line, "\"")

			name := fields[0]
			name = strings.TrimPrefix(name, ".")
			label[name] = len(data)
			data = append(data, []byte(line)...)

		} else if strings.HasPrefix(line, "mov ") {

			// assume "mov REG, NUM" for the moment
			line = strings.TrimPrefix(line, "mov ")
			line = strings.TrimSpace(line)

			fields := strings.Split(line, ",")

			if fields[0] == "rax" {
				text = append(text, []byte{0x48, 0xc7, 0xc0}...)
			} else if fields[0] == "rbx" {
				text = append(text, []byte{0x48, 0xc7, 0xc3}...)
			} else if fields[0] == "rcx" {
				text = append(text, []byte{0x48, 0xc7, 0xc1}...)
			} else if fields[0] == "rdx" {
				text = append(text, []byte{0x48, 0xc7, 0xc2}...)
			} else {
				return fmt.Errorf("unknown register %s in %s", fields[0], orig)
			}

			// NUM / label
			fields[1] = strings.TrimSpace(fields[1])

			offset, ok := label[fields[1]]
			var num int64
			if ok {
				num = int64(offset)

				patches[len(text)] = offset
			} else {
				num, err = strconv.ParseInt(fields[1], 0, 64)
				if err != nil {
					return fmt.Errorf("unable to convert %s to number for register %s", fields[1], err)
				}
			}

			// Add the value
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, uint32(num))
			text = append(text, buf...)
		} else if strings.HasPrefix(line, "inc ") {
			// assume "INC REG"
			line = strings.TrimPrefix(line, "inc ")
			line = strings.TrimSpace(line)

			if line == "rax" {
				text = append(text, []byte{0x48, 0xff, 0xc0}...)
			} else if line == "rbx" {
				text = append(text, []byte{0x48, 0xff, 0xc3}...)
			} else if line == "rcx" {
				text = append(text, []byte{0x48, 0xff, 0xc1}...)
			} else if line == "rdx" {
				text = append(text, []byte{0x48, 0xff, 0xc2}...)
			} else {
				return fmt.Errorf("unknown register %s in %s", line, orig)
			}

		} else if strings.HasPrefix(line, "xor ") {
			// Register XOR

			// assume "xor REG, REG" for the moment
			line = strings.TrimPrefix(line, "xor ")
			line = strings.TrimSpace(line)

			fields := strings.Split(line, ",")

			if fields[0] == "rax" {
				text = append(text, []byte{0x48, 0x31, 0xc0}...)
			} else if fields[0] == "rbx" {
				text = append(text, []byte{0x48, 0x31, 0xdb}...)
			} else if fields[0] == "rcx" {
				text = append(text, []byte{0x48, 0x31, 0xc9}...)
			} else if fields[0] == "rdx" {
				text = append(text, []byte{0x48, 0x31, 0xd2}...)
			} else {
				return fmt.Errorf("unknown register %s in %s", fields[0], orig)
			}

		} else if line == "int 0x80" {

			// syscall
			text = append(text, []byte{0xcd, 0x80}...)
		} else if line == "add rbx, rcx" {
			// add
			text = append(text, []byte{0x48, 0x01, 0xcb}...)
		} else {
			fmt.Printf("UNKNOWN LINE: %s\n", line)
		}
	}

	//
	// Error reading lines?
	//
	if err := scanner.Err(); err != nil {
		return err
	}

	for o, v := range patches {

        // start of virtual sectoin
        //  + offset
        //  + len of code segment
        //  + elf header
        //  + 2 * program header
        // life is hard
		v = 0x400000 + v + len(text) + 0x40 + (2 * 0x38)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))

		for i, x := range buf {
			text[i+o] = x
		}
	}
	//
	// No error.
	//
	// Write.  The.  Elf.  Output.
	//
	e := elf.New()
	err = e.WriteContent("a.out", text, data)

	if err != nil {
		return fmt.Errorf("error writing elf: %s", err.Error())
	}

	return nil
}

func main() {

	//
	// Ensure we have an argument
	//
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: compiler input.asm\n")
		return
	}

	//
	// Process
	//
	err := Assemble(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
