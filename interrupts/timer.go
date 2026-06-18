package interrupts

import (
	"rtos/cpu"
	"rtos/lib"
)

var tickFunc func()

// RegisterTimerTick sets the function to call on each timer tick.
func RegisterTimerTick(fn func()) {
	tickFunc = fn
}

// TimerInit sets up the APIC timer.
func TimerInit() {
	cpu.WriteLapic(cpu.APIC_TIMER_DCR, 0x03)
	cpu.WriteLapic(cpu.APIC_TIMER_ICR, 100000)
	cpu.WriteLapic(cpu.APIC_TIMER, 0x20000|32)
	lib.PrintString("APIC timer initialized.\n")
}

// TimerHandler is called from IRQ0 (vector 32).
func TimerHandler() {
	// Send EOI
	cpu.WriteLapic(cpu.APIC_EOI, 0)
	// Call scheduler tick if registered
	if tickFunc != nil {
		tickFunc()
	}
}
