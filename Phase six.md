We now implement **Phase 6 – Device Drivers**, providing essential hardware abstraction for serial, keyboard, HPET, PCI, and a storage skeleton. All drivers are deterministic, integrate with the interrupt system, and are RTOS‑safe.

---

## Phase 6: Device Drivers

### 1. Code Output

New files:

```
/lib/port.go
/lib/io_amd64.s
/drivers/serial/serial.go
/drivers/keyboard/ps2.go
/drivers/keyboard/scancodes.go
/drivers/hpet/hpet.go
/drivers/pci/pci.go
/drivers/storage/ahci.go
/interrupts/handlers.go (updated to route IRQ1)
/kernel/main.go (updated to test drivers)
```

We also add `kernel/cpu/pause_amd64.s` for the `PAUSE` instruction (used in spinlocks).

---

### File: `/lib/port.go`

Port I/O functions (in/out).

```go
package lib

// Port I/O functions implemented in assembly.
func Inb(port uint16) uint8
func Outb(port uint16, data uint8)
func Inw(port uint16) uint16
func Outw(port uint16, data uint16)
func Inl(port uint16) uint32
func Outl(port uint16, data uint32)
```

---

### File: `/lib/io_amd64.s`

Assembly implementation of port I/O.

```assembly
// io_amd64.s - x86 port I/O
#include "textflag.h"

// uint8 Inb(uint16 port)
TEXT ·Inb(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    INB DX, AX
    MOVB AL, ret+8(FP)
    RET

// void Outb(uint16 port, uint8 data)
TEXT ·Outb(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    MOVB data+8(FP), AX
    OUTB AX, DX
    RET

// uint16 Inw(uint16 port)
TEXT ·Inw(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    INW DX, AX
    MOVW AX, ret+8(FP)
    RET

// void Outw(uint16 port, uint16 data)
TEXT ·Outw(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    MOVW data+8(FP), AX
    OUTW AX, DX
    RET

// uint32 Inl(uint16 port)
TEXT ·Inl(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    INL DX, AX
    MOVL AX, ret+8(FP)
    RET

// void Outl(uint16 port, uint32 data)
TEXT ·Outl(SB), NOSPLIT, $0-12
    MOVL port+0(FP), DX
    MOVL data+8(FP), AX
    OUTL AX, DX
    RET
```

---

### File: `/drivers/serial/serial.go`

16550 UART serial driver (polling and interrupt‑capable).

```go
package serial

import (
	"rtos/lib"
	"rtos/sync"
)

// COM port base addresses
const (
	COM1 = 0x3F8
	COM2 = 0x2F8
	COM3 = 0x3E8
	COM4 = 0x2E8
)

// Registers offsets (relative to base)
const (
	THR = 0 // Transmit Holding Register (write)
	RBR = 0 // Receive Buffer Register (read)
	IER = 1 // Interrupt Enable Register
	IIR = 2 // Interrupt Identification Register (read)
	FCR = 2 // FIFO Control Register (write)
	LCR = 3 // Line Control Register
	MCR = 4 // Modem Control Register
	LSR = 5 // Line Status Register
	MSR = 6 // Modem Status Register
	SCR = 7 // Scratch Register
)

// Line Status Register bits
const (
	LSR_DR  = 1 << 0 // Data Ready
	LSR_OE  = 1 << 1 // Overrun Error
	LSR_PE  = 1 << 2 // Parity Error
	LSR_FE  = 1 << 3 // Framing Error
	LSR_BI  = 1 << 4 // Break Interrupt
	LSR_THR = 1 << 5 // Transmitter Holding Register Empty
	LSR_TEM = 1 << 6 // Transmitter Empty
	LSR_RFE = 1 << 7 // Receiver FIFO Error
)

// SerialPort represents a UART.
type SerialPort struct {
	base      uint16
	mu        sync.Mutex
	irq       uint8
	callback  func(byte) // optional interrupt callback
}

// Init initialises the serial port at given base address with baud rate 115200.
func Init(base uint16) *SerialPort {
	port := &SerialPort{base: base}

	// Disable interrupts
	lib.Outb(base+IER, 0)

	// Set baud rate: 115200 (divisor = 1, since 115200 = 1.8432 MHz / 16)
	// Enable DLAB
	lib.Outb(base+LCR, 0x80)
	// Divisor low byte
	lib.Outb(base+THR, 0x01) // 1 for 115200
	// Divisor high byte
	lib.Outb(base+IER, 0x00)
	// Clear DLAB, set 8N1
	lib.Outb(base+LCR, 0x03)

	// Enable FIFO, clear RX/TX
	lib.Outb(base+FCR, 0x07)

	// Set modem control: DTR, RTS, and enable interrupts
	lib.Outb(base+MCR, 0x0B)

	// Enable receive interrupt
	lib.Outb(base+IER, 0x01) // RDI enabled

	return port
}

// PutChar sends a single byte.
func (p *SerialPort) PutChar(c byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Wait until THR is empty
	for (lib.Inb(p.base+LSR) & LSR_THR) == 0 {
		// spin
	}
	lib.Outb(p.base+THR, c)
}

// PutString sends a null-terminated string (or Go string).
func (p *SerialPort) PutString(s string) {
	for i := 0; i < len(s); i++ {
		p.PutChar(s[i])
	}
}

// GetChar reads a character if available; returns (byte, bool).
func (p *SerialPort) GetChar() (byte, bool) {
	if (lib.Inb(p.base+LSR) & LSR_DR) == 0 {
		return 0, false
	}
	return lib.Inb(p.base + RBR), true
}

// GetCharBlocking waits for a character.
func (p *SerialPort) GetCharBlocking() byte {
	for {
		if c, ok := p.GetChar(); ok {
			return c
		}
		// We can yield to scheduler to avoid busy loop
		// but we'll spin for now; in future we can block on interrupt.
	}
}

// IRQHandler handles serial interrupts (called from IRQ).
func (p *SerialPort) IRQHandler() {
	// Read IIR to see why interrupt
	iir := lib.Inb(p.base + IIR)
	if (iir & 0x01) != 0 { // no interrupt pending
		return
	}
	// Check receiver line status
	lsr := lib.Inb(p.base + LSR)
	if lsr&LSR_DR != 0 {
		c := lib.Inb(p.base + RBR)
		if p.callback != nil {
			p.callback(c)
		}
	}
}

// SetCallback sets a callback for received characters.
func (p *SerialPort) SetCallback(fn func(byte)) {
	p.callback = fn
}
```

We'll register the serial IRQ (IRQ4 for COM1) in the main kernel. We'll need to connect it to vector 36 (IRQ4 = vector 32+4). We'll update the IDT entry for vector 36 to point to a handler that calls `serial.IRQHandler()`. We'll use the generic handler and check vector 36.

---

### File: `/drivers/keyboard/scancodes.go`

Scancode to ASCII mapping (simplified for US keyboard).

```go
package keyboard

// ScanCodeSet1 maps scancodes to ASCII (US layout).
var ScanCodeSet1 = map[uint8]byte{
	0x02: '1',
	0x03: '2',
	0x04: '3',
	0x05: '4',
	0x06: '5',
	0x07: '6',
	0x08: '7',
	0x09: '8',
	0x0A: '9',
	0x0B: '0',
	0x0C: '-',
	0x0D: '=',
	0x10: 'q',
	0x11: 'w',
	0x12: 'e',
	0x13: 'r',
	0x14: 't',
	0x15: 'y',
	0x16: 'u',
	0x17: 'i',
	0x18: 'o',
	0x19: 'p',
	0x1A: '[',
	0x1B: ']',
	0x1E: 'a',
	0x1F: 's',
	0x20: 'd',
	0x21: 'f',
	0x22: 'g',
	0x23: 'h',
	0x24: 'j',
	0x25: 'k',
	0x26: 'l',
	0x27: ';',
	0x28: '\'',
	0x2B: '\\',
	0x2C: 'z',
	0x2D: 'x',
	0x2E: 'c',
	0x2F: 'v',
	0x30: 'b',
	0x31: 'n',
	0x32: 'm',
	0x33: ',',
	0x34: '.',
	0x35: '/',
	0x39: ' ', // space
	// Add more as needed (shift handling omitted for brevity)
}
```

---

### File: `/drivers/keyboard/ps2.go`

PS/2 keyboard driver.

```go
package keyboard

import (
	"rtos/lib"
	"rtos/sync"
)

// PS2 controller ports
const (
	PS2_DATA   = 0x60
	PS2_STATUS = 0x64
	PS2_COMMAND = 0x64
)

// Status register bits
const (
	STATUS_OUTBUF_FULL = 1 << 0
	STATUS_INBUF_FULL  = 1 << 1
)

// Keyboard represents the PS/2 keyboard.
type Keyboard struct {
	mu       sync.Mutex
	callback func(byte) // ASCII character callback
}

// Init initialises the PS/2 keyboard.
func Init() *Keyboard {
	kb := &Keyboard{}
	// Disable keyboard
	lib.Outb(PS2_COMMAND, 0xAD)
	// Flush output buffer
	for lib.Inb(PS2_STATUS)&STATUS_OUTBUF_FULL != 0 {
		lib.Inb(PS2_DATA)
	}
	// Set controller config byte to enable interrupts
	lib.Outb(PS2_COMMAND, 0x20)
	config := lib.Inb(PS2_DATA)
	config |= 1 << 0 // enable keyboard interrupt
	lib.Outb(PS2_COMMAND, 0x60)
	lib.Outb(PS2_DATA, config)
	// Enable keyboard
	lib.Outb(PS2_COMMAND, 0xAE)
	// Reset keyboard
	lib.Outb(PS2_DATA, 0xFF)
	// Wait for ACK (0xFA)
	for {
		if lib.Inb(PS2_STATUS)&STATUS_OUTBUF_FULL != 0 {
			if lib.Inb(PS2_DATA) == 0xFA {
				break
			}
		}
	}
	// Enable scanning
	lib.Outb(PS2_DATA, 0xF4)
	// Wait for ACK
	for {
		if lib.Inb(PS2_STATUS)&STATUS_OUTBUF_FULL != 0 {
			if lib.Inb(PS2_DATA) == 0xFA {
				break
			}
		}
	}
	return kb
}

// IRQHandler is called from IRQ1 (keyboard interrupt).
func (kb *Keyboard) IRQHandler() {
	// Read scancode
	scancode := lib.Inb(PS2_DATA)
	// Map to ASCII (ignore break codes for now)
	if scancode < 0x80 { // only make codes
		if ascii, ok := ScanCodeSet1[scancode]; ok {
			if kb.callback != nil {
				kb.callback(ascii)
			}
		}
	}
}

// SetCallback sets the callback for ASCII characters.
func (kb *Keyboard) SetCallback(fn func(byte)) {
	kb.callback = fn
}
```

---

### File: `/drivers/hpet/hpet.go`

High Precision Event Timer driver.

```go
package hpet

import (
	"rtos/lib"
	"unsafe"
)

// HPET registers (relative to base)
const (
	HPET_GENERAL_CAPABILITIES = 0x00
	HPET_GENERAL_CONFIG       = 0x10
	HPET_GENERAL_INT_STATUS   = 0x20
	HPET_MAIN_COUNTER         = 0xF0
	HPET_TIMER0_CONFIG        = 0x100
	HPET_TIMER0_COMPARATOR    = 0x108
)

// HPET represents the High Precision Event Timer.
type HPET struct {
	base    uintptr
	freq    uint64 // ticks per second
	period  uint64 // femtoseconds per tick
}

var globalHPET *HPET

// Init initialises the HPET.
func Init() *HPET {
	// HPET base is usually at 0xFED00000 (mapped by firmware)
	base := uintptr(0xFED00000)

	// Read capabilities
	cap := *(*uint64)(unsafe.Pointer(base + HPET_GENERAL_CAPABILITIES))
	period := (cap >> 32) & 0xFFFFFFFF // femtoseconds per tick
	freq := uint64(1000000000000000) / period // ticks per second (1e15 / period)

	hpet := &HPET{
		base:   base,
		freq:   freq,
		period: period,
	}

	// Enable HPET: set bit 0 in general config
	config := *(*uint64)(unsafe.Pointer(base + HPET_GENERAL_CONFIG))
	config |= 1
	*(*uint64)(unsafe.Pointer(base + HPET_GENERAL_CONFIG)) = config

	globalHPET = hpet
	lib.PrintString("HPET initialised, frequency: ")
	lib.PrintUint64(freq)
	lib.PrintString(" Hz\n")
	return hpet
}

// GetTicks returns the current main counter value.
func (h *HPET) GetTicks() uint64 {
	return *(*uint64)(unsafe.Pointer(h.base + HPET_MAIN_COUNTER))
}

// GetFrequency returns ticks per second.
func (h *HPET) GetFrequency() uint64 {
	return h.freq
}

// SleepBusy spins for approximately `us` microseconds using HPET.
func (h *HPET) SleepBusy(us uint64) {
	ticks := us * h.freq / 1000000 // convert microseconds to ticks
	start := h.GetTicks()
	for h.GetTicks()-start < ticks {
		// spin
	}
}
```

We'll integrate HPET for high‑precision delays used by drivers and later for deadline scheduling.

---

### File: `/drivers/pci/pci.go`

PCI bus enumeration (basic).

```go
package pci

import (
	"rtos/lib"
)

// PCI configuration space access ports.
const (
	PCI_CONFIG_ADDR = 0xCF8
	PCI_CONFIG_DATA = 0xCFC
)

// PCI functions.
func readConfig(bus, slot, funcNum uint8, offset uint8) uint32 {
	// Build address: bit 31 = enable, bits 30-24 = reserved, bits 23-16 = bus, bits 15-11 = slot, bits 10-8 = function, bits 7-2 = offset
	addr := uint32(0x80000000) | (uint32(bus) << 16) | (uint32(slot) << 11) | (uint32(funcNum) << 8) | (uint32(offset) & 0xFC)
	lib.Outl(PCI_CONFIG_ADDR, addr)
	return lib.Inl(PCI_CONFIG_DATA)
}

func readConfig16(bus, slot, funcNum uint8, offset uint8) uint16 {
	// For 16-bit reads, we can just read 32-bit and shift.
	val := readConfig(bus, slot, funcNum, offset & 0xFC)
	if offset & 0x02 != 0 {
		return uint16(val >> 16)
	}
	return uint16(val & 0xFFFF)
}

func readConfig8(bus, slot, funcNum uint8, offset uint8) uint8 {
	val := readConfig(bus, slot, funcNum, offset & 0xFC)
	shift := (offset & 0x03) * 8
	return uint8(val >> shift)
}

// Device represents a PCI device.
type Device struct {
	Bus    uint8
	Slot   uint8
	Func   uint8
	Vendor uint16
	Device uint16
	Class  uint8
	Subclass uint8
	ProgIF uint8
}

// Enumerate scans the PCI bus and returns a list of devices.
func Enumerate() []Device {
	devices := []Device{}
	for bus := uint8(0); bus < 256; bus++ {
		for slot := uint8(0); slot < 32; slot++ {
			vendor := readConfig16(bus, slot, 0, 0)
			if vendor == 0xFFFF {
				continue
			}
			dev := Device{
				Bus:  bus,
				Slot: slot,
				Func: 0,
			}
			// Read device ID
			dev.Device = readConfig16(bus, slot, 0, 2)
			dev.Vendor = vendor
			// Read class, subclass, progIF
			class := readConfig8(bus, slot, 0, 0x0B)
			subclass := readConfig8(bus, slot, 0, 0x0A)
			progIF := readConfig8(bus, slot, 0, 0x09)
			dev.Class = class
			dev.Subclass = subclass
			dev.ProgIF = progIF
			devices = append(devices, dev)
			// Check for multi-function devices
			headerType := readConfig8(bus, slot, 0, 0x0E)
			if (headerType & 0x80) != 0 {
				// Multi-function device, scan functions 1-7
				for fn := uint8(1); fn < 8; fn++ {
					vendor2 := readConfig16(bus, slot, fn, 0)
					if vendor2 != 0xFFFF {
						dev2 := Device{
							Bus:  bus,
							Slot: slot,
							Func: fn,
						}
						dev2.Device = readConfig16(bus, slot, fn, 2)
						dev2.Vendor = vendor2
						dev2.Class = readConfig8(bus, slot, fn, 0x0B)
						dev2.Subclass = readConfig8(bus, slot, fn, 0x0A)
						dev2.ProgIF = readConfig8(bus, slot, fn, 0x09)
						devices = append(devices, dev2)
					}
				}
			}
		}
	}
	return devices
}
```

---

### File: `/drivers/storage/ahci.go`

AHCI storage driver skeleton (for later).

```go
package storage

import (
	"rtos/lib"
)

// AHCI controller stub.
type AHCI struct {
	// Will be filled in later with port registers, DMA, etc.
}

func InitAHCI() *AHCI {
	// Scan for AHCI devices using PCI.
	// For now, just return stub.
	lib.PrintString("AHCI skeleton initialised.\n")
	return &AHCI{}
}

// Read would read a sector.
func (a *AHCI) Read(sector uint64, buffer []byte) error {
	// stub
	return nil
}
```

---

### File: `/interrupts/handlers.go` (updated to handle IRQ1 and IRQ4)

We need to route IRQ1 (keyboard, vector 33) and IRQ4 (serial COM1, vector 36) to the appropriate driver handlers.

We'll add a table of handler functions for IRQs.

In `handlers.go`:

```go
var irqHandlers [16]func()

// RegisterIRQHandler registers a handler for an IRQ (0-15).
func RegisterIRQHandler(irq uint8, handler func()) {
	if irq < 16 {
		irqHandlers[irq] = handler
	}
}

func handleInterrupt(context *unsafe.Pointer, vector uint8) {
	if vector >= 32 && vector < 48 {
		irq := vector - 32
		// Call registered handler
		if irq < 16 && irqHandlers[irq] != nil {
			irqHandlers[irq]()
		}
		// Send EOI (done inside timer handler for IRQ0, but for others we need to send EOI)
		// We'll send EOI in the handler itself, or we can do it here.
		// For simplicity, we'll have handlers send EOI.
		return
	}
	// ... rest
}
```

We'll also need to add `cpu.writeLapic(cpu.APIC_EOI, 0)` in the handlers.

---

### File: `/kernel/main.go` (updated to test drivers)

We'll initialise serial, keyboard, HPET, PCI, and AHCI. We'll create a test task that prints "Hello" via serial and echoes keyboard input.

```go
package kernel

import (
	"rtos/cpu"
	"rtos/drivers/hpet"
	"rtos/drivers/keyboard"
	"rtos/drivers/pci"
	"rtos/drivers/serial"
	"rtos/drivers/storage"
	"rtos/interrupts"
	"rtos/lib"
	"rtos/mm"
	"rtos/scheduler"
	"rtos/sync"
)

var (
	serialPort *serial.SerialPort
	keyboardDev *keyboard.Keyboard
)

func initDrivers() {
	// HPET
	hpet.Init()

	// Serial (COM1)
	serialPort = serial.Init(serial.COM1)
	interrupts.RegisterIRQHandler(4, func() {
		serialPort.IRQHandler()
		// Send EOI
		cpu.WriteLapic(cpu.APIC_EOI, 0)
	})
	// Redirect IRQ4 to vector 36 (IRQ4 = vector 32+4)
	cpu.IOAPICRedirect(4, 36, 0)

	// Keyboard
	keyboardDev = keyboard.Init()
	interrupts.RegisterIRQHandler(1, func() {
		keyboardDev.IRQHandler()
		cpu.WriteLapic(cpu.APIC_EOI, 0)
	})
	cpu.IOAPICRedirect(1, 33, 0)

	// PCI
	devices := pci.Enumerate()
	lib.PrintString("PCI devices found: ")
	lib.PrintUint64(uint64(len(devices)))
	lib.PrintString("\n")
	for _, dev := range devices {
		lib.PrintString("  Bus: ")
		lib.PrintUint64(uint64(dev.Bus))
		lib.PrintString(" Slot: ")
		lib.PrintUint64(uint64(dev.Slot))
		lib.PrintString(" Func: ")
		lib.PrintUint64(uint64(dev.Func))
		lib.PrintString(" Vendor: 0x")
		lib.PrintHex64(uint64(dev.Vendor))
		lib.PrintString(" Device: 0x")
		lib.PrintHex64(uint64(dev.Device))
		lib.PrintString(" Class: 0x")
		lib.PrintHex64(uint64(dev.Class))
		lib.PrintString("\n")
	}

	// Storage (AHCI) - just a stub
	storage.InitAHCI()
}

func testTask() {
	serialPort.PutString("Hello from serial test task!\n")
	for {
		// Echo keyboard input
		// We'll poll keyboard via callback; but we'll use a simple key echo in the callback.
		scheduler.Yield()
	}
}

func main() {
	// ... previous init steps up to scheduler init
	initDrivers()

	// Create test task
	scheduler.CreateTask(testTask, 2)

	// Set keyboard callback to echo
	keyboardDev.SetCallback(func(c byte) {
		serialPort.PutChar(c)
	})

	// ... scheduler init and start
}
```

We need to export `cpu.IOAPICRedirect` from `apic.go` and add `cpu.WriteLapic`.

Add to `apic.go`:

```go
// IOAPICRedirect is exported.
func IOAPICRedirect(irq uint8, vector uint8, flags uint32) {
	// as before
}

// WriteLapic writes to a local APIC register.
func WriteLapic(offset uint32, val uint32) {
	*(*uint32)(unsafe.Pointer(lapicBase + uintptr(offset))) = val
}
```

We also need to adjust the APIC timer init to use `WriteLapic`.

---

### 2. File Structure (Phase 6 additions)

```
/lib/
    port.go
    io_amd64.s
/drivers/
    serial/
        serial.go
    keyboard/
        ps2.go
        scancodes.go
    hpet/
        hpet.go
    pci/
        pci.go
    storage/
        ahci.go
/interrupts/
    handlers.go (updated)
/kernel/
    main.go (updated)
/kernel/cpu/
    apic.go (updated)
```

---

### 3. Design Explanation

- **Serial**: Provides basic UART functions with blocking I/O. Interrupt handling can be registered; the callback is called on each received byte.
- **Keyboard**: Uses PS/2 controller; reads scancodes and maps to ASCII using a simple table. Interrupt driven; callback returns ASCII.
- **HPET**: Provides high‑resolution timer for busy delays (used for calibration and later for deadline scheduling).
- **PCI**: Enumerates all devices by scanning the bus, reading vendor/device IDs and class codes.
- **AHCI**: Skeleton for mass storage; will be expanded in later phases.
- **Interrupt Integration**: IRQ1 (keyboard) and IRQ4 (serial) are redirected via I/O APIC to vectors 33 and 36. The generic interrupt handler calls registered functions which then process the interrupts and send EOI.

---

### 4. Interfaces (Public APIs)

- `serial.Init(base uint16) *SerialPort`
- `serial.PutChar(c byte)`, `PutString(s string)`
- `serial.GetChar() (byte, bool)`, `GetCharBlocking() byte`
- `serial.SetCallback(fn func(byte))`
- `keyboard.Init() *Keyboard`
- `keyboard.SetCallback(fn func(byte))`
- `hpet.Init() *HPET`, `GetTicks() uint64`, `SleepBusy(us uint64)`
- `pci.Enumerate() []Device`
- `storage.InitAHCI() *AHCI` (stub)
- `interrupts.RegisterIRQHandler(irq uint8, handler func())`

---

### 5. Initialization Order (Phase 6)

- After APIC and timer init, before scheduler start:
  1. HPET init.
  2. Serial init and IRQ registration.
  3. Keyboard init and IRQ registration.
  4. PCI enumeration.
  5. AHCI stub.
- Then create test tasks and start scheduler.

---

### 6. Testing Instructions

Build and run as before.

Expected output:
- The serial test task should print "Hello from serial test task!".
- Keyboard input (pressing keys) should echo characters on the serial console.
- PCI enumeration will print a list of detected devices.

If you press keys, you should see the corresponding characters echoed.

---

## What Was Implemented

- Serial driver (16550) with polling and interrupt callback.
- PS/2 keyboard driver with scancode to ASCII mapping.
- HPET driver for high‑precision timing.
- PCI bus enumeration.
- AHCI storage driver skeleton.
- Integration of IRQ handlers for serial and keyboard.
- Test task to demonstrate serial output and keyboard echo.

---

## What Is Pending

- Phase 7: Filesystem (VFS, RAMFS, FAT32 read/write).
- Phase 8: Networking (Ethernet, IPv4, UDP, TCP).
- Phase 9: RTOS Runtime Layer (Go runtime adjustments).
- Phase 10: Testing and benchmarks.

---

## Next Step

Proceed to **Phase 7 – Filesystem**. We will implement:
- Virtual File System (VFS) layer with file operations.
- RAM filesystem for early boot.
- FAT32 support (read‑only initially) to access files on disk.

We will integrate with the AHCI driver to read from disk.

---

Proceed to Phase 7? If so, say "Continue to Phase 7".
