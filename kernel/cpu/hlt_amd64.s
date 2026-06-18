// hlt_amd64.s - HLT instruction
TEXT ·halt(SB), NOSPLIT, $0
    HLT
    RET
