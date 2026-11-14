package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

const (
	CRCSize       uint32 = 4
	TimestampSize uint32 = 4
	KeyLenSize    uint32 = 4
	ValLenSize    uint32 = 4
)

type logEntry struct {
	timestamp uint32
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

func newTombstone(timestamp int, key string) logEntry {
	return newLogEntry(timestamp, key, "")
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

	crcHash := crc32.ChecksumIEEE(buf[CRCSize:])
	binary.BigEndian.PutUint32(buf, crcHash)

	return buf
}

func decode(buf []byte) (le logEntry, err error) {

	headerLength := CRCSize + TimestampSize + KeyLenSize + ValLenSize
	if len(buf) < int(headerLength) {
		return le, fmt.Errorf("Invalid data, header bytes insufficient")
	}

	offset := 0
	var expectedCrc uint32
	_, err = binary.Decode(buf[offset:offset+int(CRCSize)], binary.BigEndian, &expectedCrc)
	if err != nil {
		fmt.Printf("Error decoding crc %v\n", err)
		return
	}
	offset += int(CRCSize)

	_, err = binary.Decode(buf[offset:offset+int(TimestampSize)], binary.BigEndian, &le.timestamp)
	if err != nil {
		fmt.Printf("Error decoding Timestamp %v\n", err)
		return
	}
	offset += int(TimestampSize)

	var keyLen uint32
	_, err = binary.Decode(buf[offset:offset+int(KeyLenSize)], binary.BigEndian, &keyLen)
	if err != nil {
		fmt.Printf("Error decoding key length %v\n", err)
		return
	}
	offset += int(KeyLenSize)

	var valLen uint32
	_, err = binary.Decode(buf[offset:offset+int(ValLenSize)], binary.BigEndian, &valLen)
	if err != nil {
		fmt.Printf("Error decoding value length %v\n", err)
		return
	}
	offset += int(ValLenSize)

	totalExpectedBufSize := headerLength + keyLen + valLen
	if len(buf) < int(totalExpectedBufSize) {
		return le, fmt.Errorf("Invalid data, buffer data incorrect or partially read")
	}

	crcHash := crc32.ChecksumIEEE(buf[CRCSize:])
	if crcHash != expectedCrc {
		return le, fmt.Errorf("CRC hash of data does not match the expected value, some data is corrupted!")
	}

	le.key = string(buf[offset : offset+int(keyLen)])
	offset += int(keyLen)

	le.value = string(buf[offset : offset+int(valLen)])

	return
}

func main() {
	fmt.Println("hello world")

	testLogEntry := newTombstone(1794596697, "hello")
	encodedLogEntry := testLogEntry.encode()
	fmt.Printf("Encoded log data: 0x% X\n", encodedLogEntry)

	decodedLogEntry, err := decode(encodedLogEntry)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decoded entry is %+v\n", decodedLogEntry)
}
