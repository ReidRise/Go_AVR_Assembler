; Set-up Interupt Vector Table (IVT)
RCALL main
.import ivt.S

.org 0x200
; Safely past IVT start code
.import atmega328p.S
.import common.S 
;.import usart.S

.org 0x500
main:
    setup_stack 
    ;RCALL setup_timer
    
    ;LDI r17, 0x00 ; Baud Rate set to 9600 for 1Mhz
    ;LDI r16, 0x0c ; Baud Rate set to 9600 for 1Mhz
    ;RCALL USART_Init

    ; Stand up Port B as a output
    LDI R16, 0xff 
    OUT $DDRB, R16   ;PORTB IS OUTPUT
    LDI R16, 0x55 
    OUT $PORTB, R16

BACK:
    COM R16
    OUT $PORTB, r16
    ;LDI Z(HIGH), welcome(HIGH)
    ;LDI Z(LOW), welcome(LOW)
    ;RCALL ReadLoop
    ;LDI Z(HIGH), help(HIGH)
    ;LDI Z(LOW), help(LOW)
    ;RCALL ReadLoop
    RCALL sleep_s
    RJMP BACK

welcome:
    .db "WELCOME TO THE ATMEGA8515 RADIO\n"
help:
    .db "\tHelp Menu:\n\t[1] start\n\t[2] stop\n"