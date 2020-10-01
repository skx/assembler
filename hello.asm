.hello-world DB   "Hello, world\n"
.goodbye-world DB "Goodbye, world\n"

        xor rax, rax
        xor rbx, rbx

        ;; write 15 characters
        mov rdx, 15
        ;; starting at the string
        mov rcx, hello-world
        ;; output is STDOUT
        mov rbx, 1
        ;; sys_write
        mov rax, 4
        ;; syscall
        int 0x80

        ;; Goodbye-time
        mov rdx, 18
        mov rcx, goodbye-world
        mov rax, 4
        mov rbx, 1
        int 0x80

        ;; exit-code is 0
        mov rbx, 0
        ;;  sys_exit
        mov rax, 1
        ;; syscall
        int 0x80
