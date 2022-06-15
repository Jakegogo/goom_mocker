package bytecode

var (
	//nolint build by build tag
	// defaultFuncPrologue32 32位系统 function Prologue
	defaultFuncPrologue32 = []byte{0x65, 0x8b, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x8b, 0x89, 0xfc, 0xff, 0xff, 0xff}
	// defaultFuncPrologue64 64位系统 function Prologue
	defaultFuncPrologue64 = []byte{0x65, 0x48, 0x8b, 0x0c, 0x25, 0x30, 0x00, 0x00, 0x00, 0x48}
)
