#!/usr/bin/env python3
import sys

vectors_with_errors = {8, 10, 11, 12, 13, 14, 17, 30}

# Generate assembly stubs
with open("kernel/cpu/isr_stubs.asm", "w") as f:
    f.write("// Generated ISR stubs\n")
    f.write("#include \"textflag.h\"\n\n")
    for v in range(256):
        if v in vectors_with_errors:
            f.write(f"TEXT ·isr_{v}(SB), NOSPLIT, $0\n")
            f.write(f"    PUSHQ ${v}\n")
        else:
            f.write(f"TEXT ·isr_{v}(SB), NOSPLIT, $0\n")
            f.write(f"    PUSHQ $0\n")
            f.write(f"    PUSHQ ${v}\n")
        f.write("    JMP common_isr_handler\n\n")

# Generate Go declarations
with open("kernel/cpu/isr_stubs.go", "w") as f:
    f.write("// Generated ISR stub declarations\n")
    f.write("package cpu\n\n")
    for v in range(256):
        f.write(f"//go:noescape\n")
        f.write(f"func isr_{v}()\n")
    f.write("\n// ISRTable maps vector to function pointer\n")
    f.write("var ISRTable = [256]uintptr{\n")
    for v in range(256):
        f.write(f"    uintptr(unsafe.Pointer(&isr_{v})),\n")
    f.write("}\n")
