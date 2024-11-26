package http2

import (
	"encoding/binary"
	"errors"
	"io"
)

// convert an unsigned 24-bit integer to an unsigned 32-bit integer
// example: 1027 represented in uint24 is
// 0000_0000 0000_0100 0000_0011
// As a result, uint32(bytes[0])<<16 gives
// 0000_0000 0000_0100 0000_0011
//
//		    \
//	         \
//	          \
//
// 0000_0000 0000_0000 0000_0000 0000_0000
//
// Then, uint32(bytes[1])<<8 gives
// 0000_0000 0000_0100 0000_0011
//
//		              \
//	                   \
//	                    \
//
// 0000_0000 0000_0000 0000_0100 0000_0000
//
// And uint32(bytes[2]) gives: 0000_0000 0000_0000 0000_0000 0000_0011
// Finally, using the OR operator results in:
// 0000_0000 0000_0000 0000_0000 0000_0000
// 0000_0000 0000_0000 0000_0100 0000_0000
// 0000_0000 0000_0000 0000_0000 0000_0011
// =======================================
// 0000_0000 0000_0000 0000_0100 0000_0011
func uint24ToUint32(bytes []byte) uint32 {
	return uint32(bytes[0])<<16 | uint32(bytes[1])<<8 | uint32(bytes[2])
}

// Parse the frame stream identifier, which is 4 octets.
// +-+-------------+---------------+-------------------------------+
// |R|                 Stream Identifier (31)                      |
// +-+-------------+---------------+-------------------------------+
//
//	↑
//
// A reserved 1-bit field which must be ignored.
// Then, the 31 next bits represent an unsigned 31-bit integer.
//
// Here's the trick with (1<<31 -1). 1<<31 is represented in binary as :
// 1000_0000 0000_0000 0000_0000 0000_0000
// And 1<<31 -1 is:
// 0111_1111 1111_1111 1111_1111 1111_1111
// Using the AND operator, it sets the first bit to 0, and remains the rest unchanged.
func parseStreamIdentifier(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes) & (1<<31 - 1)
}

func encodeStreamIdentifier(streamId uint32) []byte {
	bytes := make([]byte, 0, 4)
	bytes = append(bytes, byte(streamId>>24), byte(streamId>>16), byte(streamId>>8), byte(streamId))

	return bytes
}

func readBytes(reader io.Reader, nBytes int) ([]byte, error) {
	buffer := make([]byte, nBytes)
	_, err := io.ReadFull(reader, buffer)

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, errors.New("unexpected end of buffer")
	}

	//if err != nil && !strings.Contains(err.Error(), "unknown certificate") {
	//	return buffer, err
	//}

	return buffer, nil
}
