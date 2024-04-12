package checksum_test

import (
	"checksum"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplement(t *testing.T) {
	t1 := uint16(65535)
	assert.Equal(t, uint16(0), checksum.Complement(t1))
	t2 := uint16(0)
	assert.Equal(t, uint16(65535), checksum.Complement(t2))
	t3 := uint16(128 + 8 + 4)
	assert.Equal(t, uint16(65535-128-8-4), checksum.Complement(t3))
}

func TestCheckSum(t *testing.T) {
	t1 := []byte{0, 0}
	assert.Equal(t, uint16(65535), checksum.CheckSum(t1))
	t2 := []byte{255, 0}
	assert.Equal(t, uint16(255), checksum.CheckSum(t2))
	odd := []byte{255, 0, 1}
	assert.Equal(t, uint16(511), checksum.CheckSum(odd))
	bigEven := []byte{255, 255, 1, 1, 2, 2, 4, 4}
	assert.Equal(t, uint16(1799), checksum.CheckSum(bigEven))
}

func TestValidate(t *testing.T) {
	odd := []byte{255, 0, 1}
	assert.True(t, checksum.Validate(odd, uint16(511)))
	bigEven := []byte{255, 255, 1, 1, 2, 2, 4, 4}
	assert.True(t, checksum.Validate(bigEven, uint16(1799)))
}

func TestValidatePacket(t *testing.T) {
	odd := []byte{255, 0, 1}
	oddCS := uint16(511)
	oddPacket := append([]byte{byte(oddCS >> 8), byte(oddCS)}, odd...)
	assert.True(t, checksum.ValidatePacket(oddPacket))

	bigEven := []byte{255, 255, 1, 1, 2, 2, 4, 4}
	bigEvenCS := uint16(1799)
	bigEvenPacket := append([]byte{byte(bigEvenCS >> 8), byte(bigEvenCS)}, bigEven...)
	assert.True(t, checksum.ValidatePacket(bigEvenPacket))
}
