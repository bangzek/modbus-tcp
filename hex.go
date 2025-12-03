package modbus

type hexs []byte

func (h hexs) Len() int {
	return len(h)*2 + len(h) - 1
}

func (h hexs) Append(b []byte) []byte {
	a := [16]byte{
		'0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
	}
	for i := 0; i < len(h); i++ {
		if i != 0 {
			b = append(b, ' ')
		}
		b = append(b, a[h[i]>>4])
		b = append(b, a[h[i]&0xF])
	}
	return b
}

func (h hexs) Append2(b []byte) []byte {
	a := [16]byte{
		'0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
	}
	for i := 0; i < len(h); i++ {
		b = append(b, a[h[i]>>4])
		b = append(b, a[h[i]&0xF])
	}
	return b
}
