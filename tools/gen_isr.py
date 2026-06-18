#!/usr/bin/env python3
# Generate isr_stubs.asm for all 256 interrupt vectors.

vectors_with_errors = {8, 10, 11, 12, 13, 14, 17, 30}

with open("isr_stubs.asm", "w") as f:
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
