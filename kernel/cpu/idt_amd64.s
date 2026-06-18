// idt_amd64.s - IDT loading
#include "textflag.h"

// void loadIDT(IDTR* idtr)
TEXT ·loadIDT(SB), NOSPLIT, $0-8
    MOVQ idtr+0(FP), AX
    LIDT 0(AX)
    RET

// void EnableInterrupts()
TEXT ·EnableInterrupts(SB), NOSPLIT, $0
    STI
    RET

// void DisableInterrupts()
TEXT ·DisableInterrupts(SB), NOSPLIT, $0
    CLI
    RET
