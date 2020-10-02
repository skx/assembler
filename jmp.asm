;; This is an example of control-flow.
;;
;; Ordinarily you'd write something like this:
;;
;;       jmp foo
;;
;; However we don't yet support the use of `jmp` instructions, so
;; instead we abuse the instructions we do support, which means
;; we push an address upon the stack, and then return into it.
;;



        push foo     ; "jmp foo" - indirectly.
        ret

:bar
        nop          ; Nothing happens
        mov rbx,33   ; first syscall argument: exit code
        mov rax,1    ; system call number (sys_exit)
        int 0x80     ; call kernel

:foo
        push bar     ; "jmp bar" - indirectly.
        ret
