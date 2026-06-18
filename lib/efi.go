//go:build baremetal
package lib

import "unsafe"

// EFI types from UEFI specification
type EFIStatus uintptr
type EFIHandle uintptr
type EFISystemTable struct {
	Hdr          EFITableHeader
	FirmwareVendor *uint16
	FirmwareRevision uint32
	ConsoleIn    *EFISimpleTextInputProtocol
	ConsoleOut   *EFISimpleTextOutputProtocol
	StdErr       *EFISimpleTextOutputProtocol
	Runtime      *EFIRuntimeServices
	Boot         *EFIBootServices
	NumTables    uintptr
	ConfigTables *EFIConfigurationTable
}

type EFITableHeader struct {
	Signature uint64
	Revision  uint32
	HeaderSize uint32
	CRC32     uint32
	Reserved  uint32
}

type EFISimpleTextOutputProtocol struct {
	Reset          uintptr
	OutputString   uintptr
	TestString     uintptr
	QueryMode      uintptr
	SetMode        uintptr
	SetAttribute   uintptr
	ClearScreen    uintptr
	SetCursorPosition uintptr
	EnableCursor   uintptr
	Mode           *EFISimpleTextOutputMode
}

type EFISimpleTextOutputMode struct {
	MaxMode       int32
	Mode          int32
	Attribute     int32
	CursorColumn  int32
	CursorRow     int32
	CurrentMode   int32
}

type EFIBootServices struct {
	Hdr               EFITableHeader
	RaiseTPL          uintptr
	RestoreTPL        uintptr
	AllocatePages     uintptr
	FreePages         uintptr
	GetMemoryMap      uintptr
	AllocatePool      uintptr
	FreePool          uintptr
	CreateEvent       uintptr
	SetTimer          uintptr
	WaitForEvent      uintptr
	SignalEvent       uintptr
	CloseEvent        uintptr
	CheckEvent        uintptr
	InstallProtocolInterface uintptr
	ReinstallProtocolInterface uintptr
	UninstallProtocolInterface uintptr
	HandleProtocol    uintptr
	Reserved          uintptr
	RegisterProtocolNotify uintptr
	LocateHandle      uintptr
	LocateDevicePath  uintptr
	InstallConfigurationTable uintptr
	LoadImage         uintptr
	StartImage        uintptr
	Exit              uintptr
	UnloadImage       uintptr
	ExitBootServices  uintptr
	GetNextMonotonicCount uintptr
	Stall             uintptr
	SetWatchdogTimer  uintptr
	ConnectController uintptr
	DisconnectController uintptr
	OpenProtocol      uintptr
	CloseProtocol     uintptr
	OpenProtocolInformation uintptr
	ProtocolsPerHandle uintptr
	LocateHandleBuffer uintptr
	LocateProtocol    uintptr
	InstallMultipleProtocolInterfaces uintptr
	UninstallMultipleProtocolInterfaces uintptr
	CalculateCrc32    uintptr
	CopyMem           uintptr
	SetMem            uintptr
	CreateEventEx     uintptr
}

type EFIRuntimeServices struct {
	Hdr               EFITableHeader
	GetTime           uintptr
	SetTime           uintptr
	GetWakeupTime     uintptr
	SetWakeupTime     uintptr
	SetVirtualAddressMap uintptr
	ConvertPointer    uintptr
	GetVariable       uintptr
	SetVariable       uintptr
	GetNextVariableName uintptr
	QueryVariableInfo uintptr
	UpdateCapsule     uintptr
	QueryCapsuleCapabilities uintptr
	ResetSystem       uintptr
}

type EFIConfigurationTable struct {
	VendorGuid [16]byte
	VendorTable uintptr
}

type EFIMemoryDescriptor struct {
	Type          uint32
	PhysicalStart uint64
	VirtualStart  uint64
	NumPages      uint64
	Attribute     uint64
}

const (
	EFI_SUCCESS EFIStatus = 0
	EFI_ERROR EFIStatus = 0x8000000000000000
)

// GetSystemTable returns the EFI system table pointer passed from entry.
var gSystemTable *EFISystemTable

func SetSystemTable(st *EFISystemTable) {
	gSystemTable = st
}

func GetSystemTable() *EFISystemTable {
	return gSystemTable
}

// Console output functions
func (con *EFISimpleTextOutputProtocol) OutputString(str *uint16) EFIStatus {
	fn := *(*func(*EFISimpleTextOutputProtocol, *uint16) EFIStatus)(unsafe.Pointer(&con.OutputString))
	return fn(con, str)
}

// Convert Go string to UTF-16 and output
func PrintString(s string) {
	if gSystemTable == nil || gSystemTable.ConsoleOut == nil {
		return
	}
	// Convert ASCII to UTF-16 (simplified)
	utf16 := make([]uint16, len(s)+1)
	for i, c := range s {
		utf16[i] = uint16(c)
	}
	utf16[len(s)] = 0
	gSystemTable.ConsoleOut.OutputString(&utf16[0])
}
