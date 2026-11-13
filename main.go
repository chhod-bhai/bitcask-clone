package main

import (
	"encoding/binary"
	"fmt"
)

const (
	CRCSize       uint32 = 4
	TimestampSize uint32 = 4
	KeyLenSize    uint32 = 4
	ValLenSize    uint32 = 4
)

type logEntry struct {
	crc       uint32
	timestamp uint32
	keyLen    uint32
	valLen    uint32
	key       string
	value     string
}

func newLogEntry(timestamp int, key, value string) logEntry {
	return logEntry{
		timestamp: uint32(timestamp),
		key:       key,
		value:     value,
	}
}

func (le *logEntry) encode() []byte {
	// buffer length: CRC(4) + TS(4) + KeyLenVar(4) + ValLenVar(4) + keyLen + valLen
	keyLen := uint32(len(le.key))
	valLen := uint32(len(le.value))
	bufLen := CRCSize + TimestampSize + KeyLenSize + ValLenSize + keyLen + valLen
	buf := make([]byte, bufLen)
	offset := CRCSize
	// append an empty crc on the initial run, skipping initially
	binary.BigEndian.PutUint32(buf[offset:], le.timestamp)
	offset += TimestampSize
	binary.BigEndian.PutUint32(buf[offset:], keyLen)
	offset += KeyLenSize
	binary.BigEndian.PutUint32(buf[offset:], valLen)
	offset += ValLenSize
	offset += uint32(copy(buf[offset:], le.key))
	copy(buf[offset:], le.value)
	return buf
}
func main() {
	fmt.Println("hello world")
	testLogEntry := newLogEntry(1794596697, "hello", "world")
	fmt.Printf("0x% X\n", testLogEntry.encode())
}
