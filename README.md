# AVR Assembler (Go)

## Overview

This project is a pure Go assembler for the AVR family of microcontrollers (like the ATmega8515), focused on providing complete 16-bit instruction set support.

It was created to fill the gap in tooling for developers who want to work directly in AVR assembly without relying on large, complex toolchains.The goal is to enable simple, lightweight, and pure assembly programming for AVR projects.

### Features

- ✅ Full 16-bit AVR instruction set support (in progress)
- ✅ Minimal external dependencies (pure Go)
- ✅ Flexible instruction encoding engine
- ✅ Label and branching support
- 🔜 Planned macro support (custom macros for repeated blocks)
- 🔜 Planned multi-file compilation (for large projects)
- 🔜 Planned definition/includes support (like .equ, .include)

Why Build This?

Existing toolchains (like AVR-GCC and Atmel Studio) are often heavyweight for pure assembly projects.

Most assemblers focus on C development first — this project is focused purely on assembly.

Hobbyists, educators, and low-level developers benefit from having a clean, simple assembler they can fully understand and extend.

## Quick Start


### Build the assembler
`go build ./utilities/assemble/main.go`

### Assemble your program
`./main path/to/program.S`

The output is a Intel HEX file (future configurable).

## Roadmap

| Feature | Status |
| ---------------- | ------ |
| Full AVR 16-bit instruction set | 🚧 In progress |
| Label support | ✅ |
| Relative branching | ✅ |
| Macro expansion | 🔜 Planned |
| Multi-file project support | 🔜 Planned |
| .equ / .include definitions | 🔜 Planned |
| Helpful error reporting | 🔜 Planned |

## License
Open-source under the MIT License.

Happy hacking 🤖!