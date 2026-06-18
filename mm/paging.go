package mm

import (
	"rtos/lib"
	"unsafe"
)

// Page table entry flags for x86_64.
const (
	PT_PRESENT       = 1 << 0
	PT_WRITE         = 1 << 1
	PT_USER          = 1 << 2
	PT_WRITETHROUGH  = 1 << 3
	PT_CACHE_DISABLE = 1 << 4
	PT_ACCESSED      = 1 << 5
	PT_DIRTY         = 1 << 6
	PT_PAGE_SIZE     = 1 << 7  // 1 for 2MB pages in PDPTE
	PT_GLOBAL        = 1 << 8
	PT_NO_EXEC       = 1 << 63
)

// Page table levels.
const (
	PML4_SHIFT = 39
	PDPT_SHIFT = 30
	PD_SHIFT   = 21
	PAGE_SHIFT = 12
)

// Page tables are 4KB aligned.
type PageTableEntry uint64
type PML4Table [512]PageTableEntry
type PDPTable [512]PageTableEntry
type PDTable [512]PageTableEntry
type PTTable [512]PageTableEntry

// Global page table root (physical address).
var pml4Phys uint64

// InitPaging initialises paging: identity maps first 4GB, enables PAE and paging.
func InitPaging() {
	// Allocate one page for PML4, one for PDPT, one for PD, and one for PT.
	// Use physical memory allocator (must be initialised before this).
	// We'll allocate pages using the physical manager.
	pmm := GetPhysMem()
	if pmm == nil {
		lib.PrintString("ERROR: Physical memory manager not initialised before paging\n")
		for {
		}
	}

	// Allocate PML4 page.
	pml4Phys = pmm.AllocPages(0) // order 0 = 1 page
	if pml4Phys == 0 {
		lib.PrintString("ERROR: Failed to allocate PML4 page\n")
		for {
		}
	}
	pml4 := (*PML4Table)(unsafe.Pointer(uintptr(pml4Phys + KERNEL_BASE))) // We'll map later; for now, identity.

	// Allocate PDPT page.
	pdptPhys := pmm.AllocPages(0)
	pdpt := (*PDPTable)(unsafe.Pointer(uintptr(pdptPhys + KERNEL_BASE)))

	// Allocate PD page (for first 2MB identity mapping).
	pdPhys := pmm.AllocPages(0)
	pd := (*PDTable)(unsafe.Pointer(uintptr(pdPhys + KERNEL_BASE)))

	// Allocate PT page (for first 4KB identity mapping) — optional if using 2MB pages.
	// We'll use 2MB huge pages for identity mapping to reduce TLB pressure.
	// So we don't need PT for identity.

	// Set up PML4 entry 0 to point to PDPT.
	pml4[0] = PageTableEntry(pdptPhys | PT_PRESENT | PT_WRITE)

	// Set up PDPT entry 0 to point to PD.
	pdpt[0] = PageTableEntry(pdPhys | PT_PRESENT | PT_WRITE)

	// Set up PD entries for 2MB pages covering first 4GB.
	// We'll map all 4GB identity using 2MB pages (2048 entries).
	for i := 0; i < 512; i++ {
		phys := uint64(i) * (2 * 1024 * 1024) // 2MB per entry
		pd[i] = PageTableEntry(phys | PT_PRESENT | PT_WRITE | PT_PAGE_SIZE)
	}

	// Load the PML4 address into CR3.
	writeCR3(pml4Phys)

	// Enable PAE, PGE, and paging (set CR4.PAE, CR4.PGE, CR0.PG).
	// Read CR4, set PAE (bit 5), PGE (bit 7).
	cr4 := readCR4()
	cr4 |= (1 << 5) // PAE
	cr4 |= (1 << 7) // PGE
	writeCR4(cr4)

	// Set CR0.PG (bit 31) and enable WP (bit 16) for write protection.
	cr0 := readCR0()
	cr0 |= (1 << 31) // PG
	cr0 |= (1 << 16) // WP
	writeCR0(cr0)

	lib.PrintString("Paging enabled (identity mapping first 4GB)\n")
}

// readCR0, writeCR0, readCR4, writeCR4, writeCR3 are implemented in assembly.
//go:linkname readCR0
func readCR0() uint64

//go:linkname writeCR0
func writeCR0(val uint64)

//go:linkname readCR4
func readCR4() uint64

//go:linkname writeCR4
func writeCR4(val uint64)

//go:linkname writeCR3
func writeCR3(val uint64)
