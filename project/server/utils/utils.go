package utils

func AppendLittleEndianUint16(buffer []byte, v uint16) []byte {
	buffer = append(buffer, uint8(v))
	buffer = append(buffer, uint8(v>>8))
	return buffer
}
