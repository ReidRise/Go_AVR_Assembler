; Setup stack
LDI R16, 0x08   ; RAMEND = 0x08FF
OUT 0x3E, R16   ; SPH (Stack Pointer High) = 0x08
LDI R16, 0xff
OUT 0x3D, R16   ; SPL (Stack Pointer Low) = 0xFF

; Stand up Port B as a output
LDI R16, 0xff 
OUT 0x04, R16   ;PORTB IS OUTPUT
LDI R16, 0x00 
OUT 0x05, R16

; Main loop
BACK:
COM R16
OUT 0x05, R16
RCALL delay
RJMP BACK

; Delay Function
delay: LDI r22, 1
la: LDI r23, 1
l0: LDI r24, 1
l1: LDI r25, 10
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
