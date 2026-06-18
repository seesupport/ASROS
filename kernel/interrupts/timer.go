package interrupts

import (
	"rtos/cpu"
	"rtos/lib"
)

// TimerInit sets up the APIC timer to generate periodic interrupts.
func TimerInit() {
	// Set the timer to divide by 1 (no divider)
	cpu.writeLapic(cpu.APIC_TIMER_DCR, 0x03) // divide by 16? Actually values: 0x0B = divide by 1, 0x03 = divide by 16 etc.
	// We'll use divide by 16 for reasonable frequency.
	cpu.writeLapic(cpu.APIC_TIMER_DCR, 0x03)

	// Set initial count to generate ~1000 Hz (we need to calibrate)
	// For now, set a fixed count.
	// We'll use a large count to get a low frequency.
	// We'll calibrate later.
	cpu.writeLapic(cpu.APIC_TIMER_ICR, 100000)

	// Set timer mode to periodic, vector 32 (IRQ0)
	cpu.writeLapic(cpu.APIC_TIMER, 0x20 | 32) // bit 17 = periodic? Actually bit 17 is for TSC deadline? For periodic, we set bit 17 (0x20000)
	// Actually: bit 17 = 1 for periodic, bit 16 = 1 for mask? Let's use: mode = 0x20000 | vector.
	cpu.writeLapic(cpu.APIC_TIMER, 0x20000 | 32)

	lib.PrintString("APIC timer initialized.\n")
}

// TimerHandler is called from the IRQ0 handler (vector 32).
// We'll connect it to the scheduler later.
func TimerHandler() {
	// Send EOI
	cpu.writeLapic(cpu.APIC_EOI, 0)
	// Schedule tick
	// To be implemented in Phase 3 (scheduler)
	lib.PrintString("Timer tick\n") // for debugging
}
