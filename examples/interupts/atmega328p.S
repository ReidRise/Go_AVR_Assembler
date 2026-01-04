; Register Definitions
.define PINB  0x03
.define DDRB  0x04
.define PORTB 0x05
.define PINC  0x06
.define DDRC  0x07
.define PORTC 0x08
.define PIND  0x09
.define DDRD  0x0a
.define PORTD 0x0b

.define TIFR0  0x15
.define TIFR1  0x16
.define TIFR2  0x17

.define PCIFR  0x1b
.define EIFR   0x1c
.define EIMSK  0x1d
.define GPIOR0 0x1e
.define EECR   0x1f
.define EEDR   0x20
.define EEARL  0x21
.define EEARH  0x22
.define GTCCR  0x23
.define TCCR0A 0x24
.define TCCR0B 0x25
.define TCNT0  0x26
.define OCR0A  0x27
.define OCR0B  0x28

.define GPIOR1 0x2a
.define GPIOR2 0x2b
.define SPCR   0x2c
.define SPSR   0x2d
.define SPDR   0x2e

.define ACSR  0x30

.define SMCR   0x33
.define MCUSR  0x34
.define MCUCR  0x35

.define SPMCSR 0x37

.define SPL    0x3d
.define SPH    0x3e
.define SREG   0x3f

; Only Accessible Via ST/LD
.define UCSR0A 0xC0
.define UCSR0B 0xC1
.define UBRR0L 0xC4
.define UBRR0H 0xC5
.define UDR0   0xC6

; Setup stack
.macro setup_stack
LDI R16, 0x08   ; RAMEND = 0x025F
OUT $SPH, R16   ; SPH (Stack Pointer High) = 0x02
LDI R16, 0xff
OUT $SPL, R16   ; SPL (Stack Pointer Low) = 0x5F
.endmacro
