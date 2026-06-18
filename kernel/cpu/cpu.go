package cpu

// InitCPU performs early CPU initialization (disable interrupts, etc.)
func InitCPU() {
	DisableInterrupts()
	// CPUID checks can be done here
}

// Halt executes HLT.
func Halt() // implemented in assembly
