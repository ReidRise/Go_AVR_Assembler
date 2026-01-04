; Loads the following data structure into memory at the set location
; [head:u8][tail:u8][...:u8]

; address X [in]
create_ring_buffer_16:
    PUSH r17
    PUSH r18
    LDI r17, 0x00
    ST X+, r17 ; HEAD
    LDI r17, 0x00
    ST X+, r17 ; TAIL
    LDI r18, 0x10
clear_ring_buffer_16:
    ST X+, r17
    DEC r18
    BRNE clear_ring_buffer_16
    ; clean-up
    POP r18
    POP r17
    RET

; address      X   [in]
; value_to_add r16 [in]
; error        r17 [out]
ring_buffer_queue_16:
    PUSH r19
    PUSH r20
    PUSH r21
    LD r19, X+ ; HEAD

    PUSH r26 ; PUSH TAIL value to stack
    PUSH r27 
    LD r20, X+ ; TAIL

    SUB r19, r20 ; Check if head == tail 
    BREQ ring_buffer_queue_error_16

    ADD r26, r19  ; Add head loc to X
    IN r21, $SREG ; Handle Carry
    SBRC r21, 0
    INC r27

    ; Load Value to HEAD
    ST X, r16

    ; TAIL+1 % SIZE
    INC r20
    ANDI r20, 0x0f

    ; Restore X and store new tail
    POP r27
    POP r26
    ST X, r20 
    LDI r17, 0x00
    RJMP ring_buffer_queue_end_16


ring_buffer_queue_error_16:
    LDI r17, 0xff ; -1
ring_buffer_queue_end_16:
    PUSH r21
    PUSH r20
    PUSH r19
    RET

; address          X   [in]
; value_to_dequeue r16 [out]
; error            r17 [out]
ring_buffer_dequeue:
    PUSH r19
    PUSH r20
    PUSH r21

    PUSH r26 ; PUSH X to stack
    PUSH r27 
    LD r19, X+ ; HEAD
    LD r20, X+ ; TAIL


    SUB r19, r20 ; Check if head == tail 
    BREQ ring_buffer_dequeue_error_16

    ADD r26, r19  ; Add head loc to X
    IN r21, $SREG ; Handle Carry
    SBRC r21, 0
    INC r27

    ; Load Value to HEAD
    LD r16, X

    ; HEAD+1 % SIZE
    INC r19
    ANDI r19, 0x0f

    ; Restore X and store new tail
    POP r27
    POP r26
    ST X, r19 
    LDI r17, 0x00
    RJMP ring_buffer_dequeue_end_16

ring_buffer_dequeue_error_16:
    LDI r17, 0xff ; -1
ring_buffer_dequeue_end_16:
    PUSH r21
    PUSH r20
    PUSH r19
    PUSH r18
    RET

; sleep Function
sleep_ms: 
    LDI r22, 1
    ms_la: LDI r23, 1
    ms_l0: LDI r24, 1
    ms_l1: LDI r25, 1
    ms_l2: LDI r26, 3
    ms_l3: LDI r27, 100
    ms_l4: DEC r27
    NOP
    BRNE ms_l4
    DEC r26
    BRNE ms_l3
    DEC r25
    BRNE ms_l2
    DEC r24
    BRNE ms_l1
    DEC r23
    BRNE ms_l0
    DEC r22
    BREQ ms_la
    RET

; sleep Function
sleep_s: 
    LDI r22, 1
    la: LDI r23, 1
    l0: LDI r24, 1
    l1: LDI r25, 1
    l2: LDI r26, 255
    l3: LDI r27, 255
    l4: DEC r27
    NOP
    BRNE l4
    DEC r26
    BRNE l3
    DEC r25
    BRNE l2
    DEC r24
    BRNE l1
    DEC r23
    BRNE l0
    DEC r22
    BREQ la
    RET