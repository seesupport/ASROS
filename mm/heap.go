package mm

import (
	"rtos/lib"
	"unsafe"
)

// Slab allocator: allocates fixed-size blocks from pre-allocated pages.
// This is O(1) and deterministic (no GC, no fragmentation within slab).

type Slab struct {
	objectSize uint64
	numObjects uint64
	freeList   unsafe.Pointer // linked list of free objects
	mem        unsafe.Pointer // start of slab memory
	next       *Slab
}

// SlabAllocator manages multiple slabs for different sizes.
type SlabAllocator struct {
	slabs [16]*Slab // for sizes 8,16,24,...,128 bytes (we can expand)
}

var slabAlloc *SlabAllocator

// InitHeap initialises the slab allocator.
func InitHeap() {
	slabAlloc = &SlabAllocator{}
	// Pre-create slabs for common sizes: 8,16,24,32,40,48,56,64,72,80,88,96,104,112,120,128.
	// But we'll lazily create them on demand.
	lib.PrintString("Heap (slab allocator) initialised\n")
}

// Alloc allocates a block of at least size bytes.
// Returns pointer to the block, or nil if allocation fails.
func Alloc(size uint64) unsafe.Pointer {
	// Round up size to nearest power of 2? Or use fixed sizes.
	// For simplicity, we use a slab per size class.
	// Find the appropriate slab.
	if slabAlloc == nil {
		lib.PrintString("Heap not initialised!\n")
		return nil
	}
	// Determine size class (we'll just round up to multiple of 8).
	rounded := (size + 7) & ^uint64(7)
	idx := int(rounded/8) - 1
	if idx < 0 || idx >= len(slabAlloc.slabs) {
		// Allocate from larger slab? For now, return nil.
		return nil
	}
	slab := slabAlloc.slabs[idx]
	if slab == nil {
		// Create a new slab of this size.
		slab = createSlab(rounded)
		if slab == nil {
			return nil
		}
		slabAlloc.slabs[idx] = slab
	}
	// Pop from free list.
	if slab.freeList == nil {
		// Out of objects in this slab; create a new slab.
		newSlab := createSlab(rounded)
		if newSlab == nil {
			return nil
		}
		newSlab.next = slabAlloc.slabs[idx]
		slabAlloc.slabs[idx] = newSlab
		slab = newSlab
	}
	// Get free object.
	ptr := slab.freeList
	// Update free list to next.
	slab.freeList = *(*unsafe.Pointer)(ptr)
	return ptr
}

// Free frees a block previously allocated by Alloc.
func Free(ptr unsafe.Pointer, size uint64) {
	// Determine slab size class.
	rounded := (size + 7) & ^uint64(7)
	idx := int(rounded/8) - 1
	if idx < 0 || idx >= len(slabAlloc.slabs) {
		return
	}
	slab := slabAlloc.slabs[idx]
	// We need to find which slab this ptr belongs to.
	// We'll simplify: we assume the caller knows the size and it's from the correct slab.
	// For correctness, we can store a magic in each block, but we'll keep it simple.
	// We'll put the freed block at the head of free list.
	*(*unsafe.Pointer)(ptr) = slab.freeList
	slab.freeList = ptr
}

// createSlab allocates a new slab for a given object size.
func createSlab(objSize uint64) *Slab {
	// Allocate a page of memory from physical manager.
	pmm := GetPhysMem()
	phys := pmm.AllocPages(0) // one page
	if phys == 0 {
		return nil
	}
	// Map it to virtual address (for now, identity map).
	virt := unsafe.Pointer(uintptr(phys))
	// Calculate number of objects that fit.
	numObjects := PAGE_SIZE / objSize
	if numObjects == 0 {
		return nil
	}
	// Initialize free list: each object points to the next.
	freeList := unsafe.Pointer(uintptr(phys))
	prev := freeList
	for i := uint64(1); i < numObjects; i++ {
		next := unsafe.Pointer(uintptr(phys) + uintptr(i*objSize))
		*(*unsafe.Pointer)(prev) = next
		prev = next
	}
	*(*unsafe.Pointer)(prev) = nil // last points to nil

	return &Slab{
		objectSize: objSize,
		numObjects: numObjects,
		freeList:   freeList,
		mem:        virt,
	}
}
