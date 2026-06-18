// cr_amd64.s - CR register access
#include "textflag.h"

TEXT ·readCR0(SB), NOSPLIT, $0-8
    MOVQ CR0, AX
    MOVQ AX, ret+0(FP)
    RET

TEXT ·writeCR0(SB), NOSPLIT, $0-8
    MOVQ val+0(FP), AX
    MOVQ AX, CR0
    RET

TEXT ·readCR4(SB), NOSPLIT, $0-8
    MOVQ CR4, AX
    MOVQ AX, ret+0(FP)
    RET

TEXT ·writeCR4(SB), NOSPLIT, $0-8
    MOVQ val+0(FP), AX
    MOVQ AX, CR4
    RET

TEXT ·writeCR3(SB), NOSPLIT, $0-8
    MOVQ val+0(FP), AX
    MOVQ AX, CR3
    RET
