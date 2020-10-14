        ;;
        ;; This file demonstrates using `call` to invoke subroutines.
        ;;
        ;; Here we have three subroutines of interest:
        ;;
        ;;  print_string - prints a string with explicit address & size.
        ;;
        ;;  print_asciiz_string - Prints a null-terminated string
        ;;
        ;;  _exit - Exits the program
        ;;

.hello   DB "Hello, world\n\0"
.message DB "This string has its size calculated dynamically!\n\0"
.goodbye DB "Goodbye, world\n\0"

        ;; print a string, with a size
        mov rcx, hello
        mov rdx, 13
        call print_string

        ;; print a string with ZERO size calculation
        mov rcx, message
        call print_asciiz_string

        ;; print a string with an explicit size
        mov rcx, goodbye
        mov rdx, 15
        call print_string

        ;; exit this script
        mov rbx, 2
        call _exit

        ;; Routine to print a string.
        ;;
        ;; Assumes string address is in RCX
        ;; Assumes string length is in RDX
        ;;
        ;; Traches: RAX, RBX, RCX, RDX
:print_string
        mov rbx, 1         ;; output is STDOUT
        mov rax, 4         ;; sys_write
        int 0x80           ;; syscall

        ret

        ;; Routing to print a '0x00'-terminated string
        ;;
        ;; Assumes string address is in RCX
:print_asciiz_string
        xor rdx, rdx            ; zero the length
        push rcx                ; save string
:len_loop
        cmp byte ptr [rcx], 0x00
        je len_loop_over
        inc rdx
        inc rcx
        jmp len_loop
:len_loop_over
        pop rcx                 ; restore string-pointer
                                ; rdx has the mesage
        call print_string       ; call the print routine
        ret                     ; and return from here


        ;; Exit
        ;;
        ;; Assumes RBX has exit-code
:_exit
        mov rax, 1      ; SYS_exit
        int 0x80        ; syscall
        ret             ; Never reached
