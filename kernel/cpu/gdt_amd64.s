// gdt_amd64.s - GDT loading routines
#include "textflag.h"

// void loadGDT(GDTR* gdtr)
TEXT ·loadGDT(SB), NOSPLIT, $0-8
    MOVQ gdtr+0(FP), AX
    LGDT 0(AX)
    RET

// void reloadSegments()
TEXT ·reloadSegments(SB), NOSPLIT, $0
    // Reload data segments
    MOVQ $0x10, AX   // Kernel data selector
    MOVW AX, DS
    MOVW AX, ES
    MOVW AX, FS
    MOVW AX, GS
    MOVW AX, SS

    // Far jump to reload CS
    // Push return address and new CS
    MOVQ $0x08, AX   // Kernel code selector
    PUSHQ AX
    LEAQ .Lreloop(SB), BX
    PUSHQ BX
    RETFQ
.Lreloop:
    RET
