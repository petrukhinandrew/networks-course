package checksum

const one = uint16(65535)

func rawSum(src []byte) uint16 {
	acc := uint16(0)
	var lb, gb byte
	handled := true
	for idx, b := range src {
		if idx%2 == 0 {
			gb = b
			handled = false
		} else {
			lb = b
			tmp := uint16(gb)<<8 + uint16(lb)
			acc ^= tmp
			handled = true
		}
	}
	if !handled {
		tmp := uint16(gb) << 8
		acc ^= tmp
	}
	return acc
}
func CheckSum(src []byte) uint16 {
	return Complement(rawSum(src))
}

func Validate(src []byte, cs uint16) bool {
	s := rawSum(src)
	return cs^s == one
}
func ValidatePacket(src []byte) bool {
	return rawSum(src) == one
}
func Complement(src uint16) uint16 {
	return uint16(src ^ one)
}
