

.org 0x02
;RJMP EXT_INT0 ; IRQ0 Handler

.org 0x04
;RJMP EXT_INT1 ; IRQ1 Handler

.org 0x06
;RJMP TIM1_CAPT ; Timer1 Capture Handler

.org 0x08
;RJMP TIM1_COMPA ; Timer1 Compare A Handler

.org 0x0a
;RJMP TIM1_COMPB ; Timer1 Compare B Handler

.org 0x0c
;RJMP TIM1_OVF ; Timer1 Overflow Handler

.org 0x0e
;RJMP TIM0_OVF ; Timer0 Overflow Handler

.org 0x10
;RJMP SPI_STC ; SPI Transfer Complete Handler

.org 0x12
;RJMP USART_RXC ; USART RX Complete Handler

.org 0x14
;RJMP USART_UDRE ; USART Transmit Buffer Ready Handler

.org 0x16
;RJMP USART_TXC ; USART TX Complete Handler

.org 0x18
;RJMP ANA_COMP ; Analog Comparator Handler

.org 0x1a
;RJMP EXT_INT2 ; IRQ2 Handler

.org 0x1c
;RJMP TIM0_COMP ; Timer0 Compare Handler

.org 0x1e
;RJMP EE_RDY ; EEPROM Ready Handler

.org 0x20
;RJMP SPM_RDY