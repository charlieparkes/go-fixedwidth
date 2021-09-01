package fixedwidth

// PadRight appends rune 'r' to the right side of byte string 'bs' until length equals 'l'.
func PadRight(bs []byte, l int, r rune) []byte {
	for len(bs) < l {
		bs = append(bs, byte(r))
	}
	return bs
}

// PadLeft appends rune 'r' to the left side of byte string 'bs' until length equals 'l'.
func PadLeft(bs []byte, l int, r rune) []byte {
	for len(bs) < l {
		bs = append([]byte{byte(r)}, bs...)
	}
	return bs
}
