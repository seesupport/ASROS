package cpu

import (
	"rtos/lib"
	"unsafe"
)

// Local APIC base address (will be mapped later)
var lapicBase uintptr

// InitAPIC initializes the local APIC and I/O APIC.
func InitAPIC() {
	// Enable local APIC via MSR
	// Read IA32_APIC_BASE MSR (0x1B)
	msrLow, msrHigh := rdmsr(0x1B)
	if (msrLow & (1 << 11)) == 0 {
		// APIC not enabled, enable it
		msrLow |= (1 << 11) // enable
		wrmsr(0x1B, msrLow, msrHigh)
	}
	// Get APIC base address (physical)
	basePhys := uint64(msrLow & 0xFFFFF000)
	// Map it to virtual address (we assume identity mapping for now, but later we'll use VMM)
	// For now, use identity mapping.
	lapicBase = uintptr(basePhys)

	lib.PrintString("Local APIC base: 0x")
	lib.PrintHex64(basePhys)
	lib.PrintString("\n")

	// Set spurious interrupt vector and enable APIC
	spurious := readLapic(APIC_SPURIOUS)
	spurious |= 0x100 // enable
	spurious |= SPURIOUS_VECTOR
	writeLapic(APIC_SPURIOUS, spurious)

	// Set task priority to 0 (accept all interrupts)
	writeLapic(APIC_TASKPRI, 0)

	// Configure LINT0 and LINT1 as disabled
	writeLapic(APIC_LINT0, 0x10000)
	writeLapic(APIC_LINT1, 0x10000)

	// Set error handling
	writeLapic(APIC_ERROR, 0xFE) // vector 0xFE

	// Initialize I/O APIC (we'll need to find it via ACPI or PCI)
	// For now, we assume the default I/O APIC at FEC00000.
	initIOAPIC()

	// Map IRQ0 to vector 32 with delivery mode fixed
	ioapicRedirect(0, 32, 0) // IRQ0 -> vector 32, edge triggered
}

// Local APIC read/write helpers
func readLapic(offset uint32) uint32 {
	return *(*uint32)(unsafe.Pointer(lapicBase + uintptr(offset)))
}

func writeLapic(offset uint32, val uint32) {
	*(*uint32)(unsafe.Pointer(lapicBase + uintptr(offset))) = val
}

// I/O APIC base (default)
var ioapicBase uintptr = 0xFEC00000

func initIOAPIC() {
	// Read I/O APIC version
	ver := *(*uint32)(unsafe.Pointer(ioapicBase))
	lib.PrintString("I/O APIC version: ")
	lib.PrintUint64(uint64(ver & 0xFF))
	lib.PrintString("\n")
}

// ioapicRedirect redirects an IRQ to a vector.
func ioapicRedirect(irq uint8, vector uint8, flags uint32) {
	// Select register
	reg := uint32(0x10 + irq*2)
	// Low dword
	low := uint32(vector) | (flags & 0xFFFF)
	*(*uint32)(unsafe.Pointer(ioapicBase + 0x00)) = reg
	*(*uint32)(unsafe.Pointer(ioapicBase + 0x10)) = low
	// High dword (destination)
	*(*uint32)(unsafe.Pointer(ioapicBase + 0x00)) = reg + 1
	*(*uint32)(unsafe.Pointer(ioapicBase + 0x10)) = 0 // destination: APIC ID 0
}

// rdmsr/wrmsr helpers (assembly)
//go:linkname rdmsr
func rdmsr(msr uint32) (low uint32, high uint32)

// rdmsr returns the 64-bit value of the MSR.
func rdmsr(msr uint32) uint64

//go:linkname wrmsr
func wrmsr(msr uint32, low uint32, high uint32)
