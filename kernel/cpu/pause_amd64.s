// pause_amd64.s - PAUSE instruction
#include "textflag.h"

TEXT ·Pause(SB), NOSPLIT, $0
    PAUSE
    RET
