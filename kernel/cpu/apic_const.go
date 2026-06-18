package cpu

// Local APIC register offsets (relative to base address)
const (
	APIC_ID       = 0x20
	APIC_VERSION  = 0x30
	APIC_TASKPRI  = 0x80
	APIC_EOI      = 0xB0
	APIC_LDR      = 0xD0
	APIC_DFR      = 0xE0
	APIC_SPURIOUS = 0xF0
	APIC_ESR      = 0x280
	APIC_ICR_LOW  = 0x300
	APIC_ICR_HIGH = 0x310
	APIC_TIMER    = 0x320
	APIC_THERMAL  = 0x330
	APIC_PERF     = 0x340
	APIC_LINT0    = 0x350
	APIC_LINT1    = 0x360
	APIC_ERROR    = 0x370
	APIC_TIMER_ICR = 0x380
	APIC_TIMER_CCR = 0x390
	APIC_TIMER_DCR = 0x3E0
)

// Spurious vector
const SPURIOUS_VECTOR = 0xFF

// Timer modes
const (
	TIMER_MODE_ONESHOT = 0
	TIMER_MODE_PERIODIC = 1
	TIMER_MODE_TSC_DEADLINE = 2
)
