# AVR Assembler (Go)

## Overview

This project is a pure Go assembler for the AVR family of microcontrollers (like the ATmega8515), focused on providing complete 16-bit instruction set support.

It was created to fill the gap in tooling for developers who want to work directly in AVR assembly without relying on large, complex toolchains.The goal is to enable simple, lightweight, and pure assembly programming for AVR projects.

### Features

- Full 16-bit AVR instruction set support (in progress)
- Minimal external dependencies (pure Go)
- Flexible instruction encoding engine
- Label and branching support
- Macro support (custom macros for repeated blocks)
- Multi-file compilation (for large projects)
- Definition support .db

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
| Full AVR 16-bit instruction set | In progress |
| Label support | Supported |
| Relative branching | Supported |
| Macro expansion | Supported |
| Multi-file project support | Supported |
| Helpful error reporting | Planned |

## License
Open-source under the MIT License.