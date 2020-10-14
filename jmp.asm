;; This is an example of control-flow.
;;


:start
        jmp foo

:bar
        nop          ; Nothing happens
        mov rbx,33   ; first syscall argument: exit code
        mov rax,1    ; system call number (sys_exit)
        int 0x80     ; call kernel

:foo
        jmp bar
