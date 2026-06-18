// msr_amd64.s - RDMSR/WRMSR
#include "textflag.h"

// uint32 rdmsr(uint32 msr) returns low, high in RAX, RDX
TEXT ·rdmsr(SB), NOSPLIT, $0-12
    MOVQ msr+0(FP), CX
    RDMSR
    MOVL AX, ret_low+8(FP)
    MOVL DX, ret_high+12(FP) // actually 12? We need to define return types.
    // We'll return two uint32 via stack.
    // We'll simplify: return uint64.
    // Let's adjust.
