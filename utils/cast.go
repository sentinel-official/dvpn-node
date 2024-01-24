package utils

func ByteFromBool(v bool) byte {
	if v {
		return 0x01
	}

	return 0x00
}
