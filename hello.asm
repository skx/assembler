        ;; Output some text to the console.
        ;;
        ;; This example demonstrates using sys_write, and sys_exit
        ;;
        ;; For less duplication see the code in `call.asm`.
        ;;

.hello   DB "Hello, world\n"
.goodbye DB "Goodbye, world\n"

        mov rdx, 13        ;; write this many characters
        mov rcx, hello     ;; starting at the string
        mov rbx, 1         ;; write to STDOUT
        mov rax, 4         ;; sys_write
        int 0x80           ;; syscall

        mov rdx, 15        ;; write this many characters
        mov rcx, goodbye   ;; starting at the string
        mov rax, 4         ;; sys_write
        mov rbx, 1         ;; write to STDOUT
        int 0x80           ;; syscall

        xor rbx, rbx       ;; exit-code is 0
        mov rax, 0x01      ;; sys_exit
        int 0x80           ;; syscall
