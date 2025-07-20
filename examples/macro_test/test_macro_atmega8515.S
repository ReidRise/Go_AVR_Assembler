.macro setup_stack
; Setup stack
LDI R16, 0x02   ; RAMEND = 0x025F
OUT 0x3E, R16   ; SPH (Stack Pointer High) = 0x02
LDI R16, 0x5f
OUT 0x3D, R16   ; SPL (Stack Pointer Low) = 0x5F

; Stand up Port B as a output
LDI R16, 0xff 
OUT 0x17, R16   ;PORTB IS OUTPUT
LDI R16, 0x55 
OUT 0x18, R16
.endmacro

.macro flip_led
COM r16
OUT 0x18, R16
RCALL delay
.endmacro

setup_stack

; Main loop
BACK: 
flip_led
RCALL delay
RJMP BACK

; Delay Function
delay: LDI r22, 1
la: LDI r23, 1
l0: LDI r24, 1
l1: LDI r25, 1
l2: LDI r26, 200
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
