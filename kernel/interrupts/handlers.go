package interrupts

import (
	"rtos/lib"
	"rtos/cpu"
	"unsafe"
)

// DefaultHandler is the generic handler for all interrupts.
// It is called from the assembly stub.
//go:noinline
func DefaultHandler() {
	// This should never be called directly; the assembly pushes a vector and calls handleInterrupt.
}

// handleInterrupt is called from the common ISR assembly.
// It receives a pointer to the saved context (stack pointer after saving registers) and the vector.
//go:noinline
func handleInterrupt(context *unsafe.Pointer, vector uint8) {
	// Print vector for debugging
	lib.PrintString("Interrupt vector: ")
	lib.PrintUint64(uint64(vector))
	lib.PrintString("\n")

	// For exceptions, we halt.
	if vector < 32 {
		lib.PrintString("Exception occurred, halting.\n")
		for {
			cpu.DisableInterrupts()
			// spin
		}
	}

	// For IRQs, we just acknowledge and return (for now).
	if vector >= 32 && vector < 48 {
		// Acknowledge APIC EOI later
	}
}

// Define handler functions for each vector.
// These are referenced in IDT setup.
// We'll use a single function for all, but we need separate functions to set in IDT.
// We'll generate them via a script as well.
// For simplicity, we'll just use the same function for all, but the IDT entry needs a unique address.
// We'll use a single handler that reads the vector from the stack? Not possible.
// So we'll have a function for each vector that calls handleInterrupt with the vector number.
// We can generate these in Go as well.

// But we already generate assembly stubs that call handleInterrupt with the vector.
// So we just need handleInterrupt to be exported and visible to assembly.

// We'll define specific handlers for exceptions (for better messages).

// Exception handlers (called from assembly stubs)
func HandlerDivideError()   { handleException(0, "Divide Error") }
func HandlerDebug()         { handleException(1, "Debug") }
func HandlerNMI()           { handleException(2, "NMI") }
func HandlerBreakpoint()    { handleException(3, "Breakpoint") }
func HandlerOverflow()      { handleException(4, "Overflow") }
func HandlerBoundRange()    { handleException(5, "Bound Range") }
func HandlerInvalidOpcode() { handleException(6, "Invalid Opcode") }
func HandlerDeviceNotAvailable() { handleException(7, "Device Not Available") }
func HandlerDoubleFault()   { handleException(8, "Double Fault") }
func HandlerInvalidTSS()    { handleException(10, "Invalid TSS") }
func HandlerSegmentNotPresent() { handleException(11, "Segment Not Present") }
func HandlerStackFault()    { handleException(12, "Stack Fault") }
func HandlerGPFault()       { handleException(13, "General Protection Fault") }
func HandlerPageFault()     { handleException(14, "Page Fault") }
// ... others

func handleException(vector uint8, name string) {
	lib.PrintString("Exception: ")
	lib.PrintString(name)
	lib.PrintString(" (vector ")
	lib.PrintUint64(uint64(vector))
	lib.PrintString(")\n")
	for {
		cpu.DisableInterrupts()
	}
}

// HandlerIRQ0..IRQ15 are for IRQs.
// HandlerIRQ0 calls TimerHandler
func HandlerIRQ0() {
	interrupts.TimerHandler()
}
func HandlerIRQ1()  { handleIRQ(1) }
// ... up to 15
func HandlerIRQ15() { handleIRQ(15) }

func handleIRQ(irq uint8) {
	lib.PrintString("IRQ ")
	lib.PrintUint64(uint64(irq))
	lib.PrintString("\n")
	// Send EOI to APIC (will be done later)
}
