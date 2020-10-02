.hello   DB "Hello, world\n"
.goodbye DB "Goodbye, world\n"

        mov rdx, 13        ;; write this many characters
        mov rcx, hello     ;; starting at the string
        mov rbx, 1         ;; output is STDOUT
        mov rax, 4         ;; sys_write
        int 0x80           ;; syscall

        mov rdx, 15        ;; write this many characters
        mov rcx, goodbye   ;; starting at the string
        mov rax, 4         ;; sys_write
        mov rbx, 1         ;; output is STDOUT
        int 0x80           ;; syscall


        xor rbx, rbx       ;; exit-code is 0
        xor rax, rax       ;; syscall will be 1 - so set to xero, then increase
        inc rax            ;;
        int 0x80           ;; syscall
