;oops
RJMP main
.import atmega8515.S

main:
setup_stack
RCALL USART_Init

; Main loop
BACK:
COM R16
OUT 0x18, r16
LDI r31, 0x05
LDI r30, 0x00
RCALL ReadLoop
LDI r31, 0x04
LDI r30, 0x00
RCALL ReadLoop
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

ReadLoop:
    PUSH r16
    PUSH r17
ReadLoop1:
    LPM r16, Z+         ; Load next byte from flash to r16
    LDI r17, 0x00
    SUB r16, r17
    BREQ Done
    RCALL delay
    RCALL USART_Transmit      ; e.g., send to UART
    RJMP ReadLoop1
Done:
    POP r16
    POP r17
    RET

.org 0x400
mystring:
.db " THIS IS AN EXAMPLE LOG "
NOP

.org 0x500
debugLabel:
.db " [DEBUG]"
NOP
