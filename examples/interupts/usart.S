USART_Init:
; Set baud Rate
    LDI X, $UBRR0H
    ST X, r17 
    LDI X, $UBRR0L
    ST X, r16 

; Store current value to avoid bashing
    PUSH r16

; Enable receiver and transmitter
    LDI r16, 0b00011000
    LDI X, $UCSR0B
    ST X, r16 

; Set frame format: 8data, 1stop bit
    LDI r16, 0b00001110
    LDI X, $UCSR0B
    POP r16
    RET

USART_Transmit:
    PUSH r17

; Wait for empty transmit buffer
    LDI X, $UCSR0B
    LD r17, X
    SBRS r17, 5
    RJMP USART_Transmit

; Put data (r16) into buffer, sends the data
    LDI X, $UDR0
    ST X, r16
    RET

ReadLoop:
    PUSH r16
    PUSH r17

; Load next byte from flash to r16
ReadLoop1:
    LPM r16, Z+
    LDI r17, 0x00
    SUB r16, r17
    BREQ Done
    RCALL sleep_ms

; e.g., send to UART
    RCALL USART_Transmit      
    RJMP ReadLoop1

Done:
    POP r17
    POP r16
    RET

USART_Receive:

    ; Wait for data to be received
    LDI X, $UCSR0A
    LD r17, X
    SBRS r17, 7
    RJMP USART_Receive

    ; Get and return received data from buffer
    LDI X, $UDR0
    LD r16, X
    RET

;USART_RX_ISR:
;    PUSH temp
;    IN temp, SREG
;
;    ; Read received data
;    IN rxdata, UDR
;    sts rx_buffer, rxdata       ; store into buffer
;
;    OUT SREG, temp
;    POP temp
;    RETI