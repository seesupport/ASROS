// context_amd64.s - context switch for x86_64
#include "textflag.h"

// void switchContext(TaskControlBlock* current, TaskControlBlock* next)
// Saves current task's registers into its TCB, loads next task's registers.
TEXT ·switchContext(SB), NOSPLIT, $0-16
    MOVQ current+0(FP), AX
    MOVQ next+8(FP), BX

    // If current is nil, skip saving.
    CMPQ AX, $0
    JZ load_next

    // Save current task's registers.
    // We need to save all callee-saved registers and RSP, RIP.
    // We assume the TCB layout: R15, R14, R13, R12, RBP, RBX, RSP, RIP
    // (order must match TaskControlBlock struct).
    MOVQ R15, 0(AX)        // R15
    MOVQ R14, 8(AX)        // R14
    MOVQ R13, 16(AX)       // R13
    MOVQ R12, 24(AX)       // R12
    MOVQ RBP, 32(AX)       // RBP
    MOVQ RBX, 40(AX)       // RBX
    MOVQ RSP, 48(AX)       // RSP
    // Save RIP: the return address from this function.
    // We need to get the return address from the stack.
    // At this point, the stack has the return address at (RSP).
    MOVQ 0(RSP), CX
    MOVQ CX, 56(AX)        // RIP

load_next:
    // Load next task's registers.
    MOVQ 0(BX), R15
    MOVQ 8(BX), R14
    MOVQ 16(BX), R13
    MOVQ 24(BX), R12
    MOVQ 32(BX), RBP
    MOVQ 40(BX), RBX
    MOVQ 48(BX), RSP
    // Finally, jump to the saved RIP.
    MOVQ 56(BX), CX
    JMP CX

// taskExitStub: called when a task returns.
TEXT ·taskExitStub(SB), NOSPLIT, $0
    CALL ·TaskExit(SB)
    // Should never return; but if it does, halt.
    HLT
