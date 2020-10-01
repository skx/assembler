;;
;;  So this is a simple example assembly program
;;
;;  It calls `int 0x80` with rax set to 0x01
;;
;;       mov eax,1        ; system call number (sys_exit)
;;       mov ebx, 0xNN      ; return code
;;       int 0x80             ; syscall
;;


;;
;;  mov eax, 0x01 works
;;
;;  However we can test our assembly by setting the register to zero,
;; then incrementing it.
;;
xor rax, rax
inc rax

;;
;; The exit-coe will be stored in rbx.
;;
;; We could set `mov rbx, 0x42`, however it is another test  of our handling
;; to allow some maths to be carried out
;;
mov rbx, 0x0000

mov rcx, 0x0007
add rbx, rcx

mov rcx, 0x0002
add rbx, rcx

;;
;; So we've said :
;;
;;    rbx = 0
;;    rbx += 7
;;    rbx += 2
;;
;; -> rbx thus contains 9.
;;
;; Now call the kernel.
;;
int 0x80
