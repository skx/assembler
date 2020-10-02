;; Basic exit-code example.
        nop                     ; Nothing happens
        mov rbx,31              ; first syscall argument: exit code
        mov rax,1               ; system call number (sys_exit)
        int 0x80                ; call kernel
