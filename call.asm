.hello   DB "Hello, world\n"
.goodbye DB "Goodbye, world\n"

        mov rcx, hello
        mov rdx, 13
        call print_string

        mov rcx, goodbye
        mov rdx, 15
        call print_string

        mov rbx, 2
        call _exit

        ;; Routine to print a string.
        ;;
        ;; Assumes string address is in RCX
        ;;
        ;; Assumes string length is in RDX
        ;;
        ;; Traches: RAX, RBX, RCX, RDX
:print_string
        mov rbx, 1         ;; output is STDOUT
        mov rax, 4         ;; sys_write
        int 0x80           ;; syscall

        ret

        ;; Exit
        ;;
        ;; Assumes RBX has exit-code
:_exit
        mov rax, 1      ; SYS_exit
        int 0x80        ; syscall
        ret             ; Never reached
