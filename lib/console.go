package lib

func PrintHex64(val uint64) {
	const hex = "0123456789ABCDEF"
	var buf [16]byte
	for i := 15; i >= 0; i-- {
		buf[i] = hex[val&0xF]
		val >>= 4
	}
	PrintString("0x")
	for i := 0; i < 16; i++ {
		PrintString(string(buf[i]))
	}
}

func PrintUint64(val uint64) {
	var buf [20]byte
	idx := 19
	if val == 0 {
		PrintString("0")
		return
	}
	for val > 0 {
		buf[idx] = byte('0' + val%10)
		val /= 10
		idx--
	}
	PrintString(string(buf[idx+1:]))
}
