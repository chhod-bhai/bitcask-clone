package model

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
	HeaderSize    uint32 = CRCSize + TimestampSize + KeyLenSize + ValLenSize
)

type LogEntry struct {
	timestamp uint32
	key       string
	value     string
}

func NewLogEntry(timestamp uint32, key, value string) LogEntry {
	return LogEntry{
		timestamp: timestamp,
		key:       key,
		value:     value,
	}
}

func NewTombstone(timestamp uint32, key string) LogEntry {
	return NewLogEntry(timestamp, key, "")
}

func (le *LogEntry) Encode() []byte {
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

func DecodeLogEntry(buf []byte) (le LogEntry, err error) {

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
