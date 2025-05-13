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

RCALL USART_Init

; Main loop
BACK:
COM R16
OUT 0x18, R16
RCALL USART_Transmit
RCALL delay
RJMP BACK

USART_Init:
; Set baud rate
PUSH r16
PUSH r17
OUT 0x20, r17
OUT 0x09, r16
; Enable receiver and transmitter
LDI r16, 0b00011000
OUT 0x0a, r16
; Set frame format: 8data, 2stop bit
LDI r16, 0b10001110
OUT 0x20, r16
POP r17
POP r16
RET

USART_Transmit:
; Wait for empty transmit buffer
SBIS 0x0B, 5 ; UCSRA, UDRE
RJMP USART_Transmit
; Put data (r16) into buffer, sends the data
OUT 0x0C, r16
RET

; Delay Function
delay: LDI r22, 1
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
