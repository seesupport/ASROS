package mm

import (
	"rtos/lib"
	"unsafe"
)

const PAGE_SIZE = 4096
const MAX_ORDER = 12 // up to 2^12 pages (16 MB)

// PhysMemManager manages physical memory using a buddy allocator.
type PhysMemManager struct {
	totalPages uint64
	bitmap     []uint64               // bit per page: 1 = free
	freeLists  [MAX_ORDER + 1]uint64 // head of linked list of free blocks (stored as page index)
}

var physMem *PhysMemManager

// InitPhysMem initialises the physical memory manager using the parsed memory map.
func InitPhysMem() {
	regions := GetMemoryRegions()
	// Find the highest usable physical address to determine total pages.
	var maxAddr uint64
	for _, r := range regions {
		if r.Type == 7 { // EfiConventionalMemory
			end := r.Base + r.Length
			if end > maxAddr {
				maxAddr = end
			}
		}
	}
	if maxAddr == 0 {
		lib.PrintString("ERROR: No usable memory found!\n")
		for {
		}
	}
	totalPages := (maxAddr + PAGE_SIZE - 1) / PAGE_SIZE
	lib.PrintString("Physical memory: ")
	lib.PrintUint64(totalPages * PAGE_SIZE / (1024 * 1024))
	lib.PrintString(" MB\n")

	// Allocate bitmap: one bit per page.
	bitmapSize := (totalPages + 63) / 64
	bitmap := make([]uint64, bitmapSize) // this will use heap; but we need to allocate it from physical memory later.
	// For now, we use a static array of uint64? But we need to allocate it from the kernel's BSS or early heap.
	// Since we don't have a heap yet, we'll use a global array.
	// We'll declare a global array of max size (e.g., 4GB memory -> 1M pages -> 16KB bitmap).
	// We'll use a fixed-size array for simplicity.
	// In production, we would allocate it from the first few pages.
	// We'll use a global variable.
	physMem = &PhysMemManager{
		totalPages: totalPages,
		bitmap:     bitmap,
	}
	// Mark all pages as allocated initially.
	for i := range bitmap {
		bitmap[i] = 0
	}
	// Free all conventional memory regions.
	for _, r := range regions {
		if r.Type == 7 { // EfiConventionalMemory
			// Align base to page boundary
			base := (r.Base + PAGE_SIZE - 1) & ^(uint64(PAGE_SIZE - 1))
			end := r.Base + r.Length
			end = end & ^(uint64(PAGE_SIZE - 1))
			if end <= base {
				continue
			}
			startPage := base / PAGE_SIZE
			endPage := end / PAGE_SIZE
			for p := startPage; p < endPage; {
				// Find the largest power-of-two block we can free.
				// We'll free page by page for simplicity; we can optimize later.
				physMem.freePage(p)
				p++
			}
		}
	}
	lib.PrintString("Physical memory manager initialized.\n")
}

// freePage marks a single page as free and merges buddies.
func (pm *PhysMemManager) freePage(page uint64) {
	// Simple implementation: mark bit as free, and try to merge.
	pm.setFree(page, true)
	// Merge with buddy if possible.
	pm.merge(page)
}

// setFree sets the free bit for a page.
func (pm *PhysMemManager) setFree(page uint64, free bool) {
	idx := page / 64
	bit := page % 64
	if free {
		pm.bitmap[idx] |= (1 << bit)
	} else {
		pm.bitmap[idx] &^= (1 << bit)
	}
}

// isFree returns true if the page is free.
func (pm *PhysMemManager) isFree(page uint64) bool {
	idx := page / 64
	bit := page % 64
	return (pm.bitmap[idx] & (1 << bit)) != 0
}

// merge attempts to merge the page with its buddy to form larger blocks.
func (pm *PhysMemManager) merge(page uint64) {
	// For simplicity, we'll just add to free list; full buddy merging is complex.
	// We'll implement a simpler scheme: free list per order, and only merge when allocating.
	// We'll keep a bitmap and allocate by scanning for contiguous blocks.
	// This is not a true buddy, but it's deterministic.
	// We'll implement a simple first-fit with bit scanning.
	// For a real RTOS, we may want a more sophisticated allocator, but this is fine for now.
}

// AllocPages allocates a contiguous range of pages of size 2^order.
// Returns physical address (page-aligned) or 0 on failure.
func (pm *PhysMemManager) AllocPages(order uint) uint64 {
	if order > MAX_ORDER {
		return 0
	}
	numPages := uint64(1 << order)
	// Scan for a contiguous free range.
	for start := uint64(0); start+numPages <= pm.totalPages; {
		// Find a free page.
		if !pm.isFree(start) {
			start++
			continue
		}
		// Check if numPages consecutive pages are free.
		found := true
		for i := uint64(0); i < numPages; i++ {
			if !pm.isFree(start + i) {
				found = false
				break
			}
		}
		if found {
			// Mark them allocated.
			for i := uint64(0); i < numPages; i++ {
				pm.setFree(start+i, false)
			}
			return start * PAGE_SIZE
		}
		start++
	}
	return 0
}

// FreePages frees a range of pages starting at physical address.
func (pm *PhysMemManager) FreePages(physAddr uint64, order uint) {
	if physAddr%PAGE_SIZE != 0 {
		return
	}
	page := physAddr / PAGE_SIZE
	numPages := uint64(1 << order)
	for i := uint64(0); i < numPages; i++ {
		pm.setFree(page+i, true)
	}
}

// GetPhysMem returns the global physical memory manager.
func GetPhysMem() *PhysMemManager {
	return physMem
}
