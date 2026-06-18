import "unsafe"

func InitIDT() {
	for i := 0; i < IDT_SIZE; i++ {
		handler := cpu.ISRTable[i]
		setIDTEntry(i, handler, 0x08, 0x8E)
	}
	// Override specific exceptions with custom handlers? We can override by replacing the entry.
	// We'll set custom handlers for exceptions (like divide error) by setting the entry to a different function.
	// But we already have generic ones that call handleInterrupt with vector, which prints and halts.
	// That's sufficient for now.
}
