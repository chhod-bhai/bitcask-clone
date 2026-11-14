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
	crc       uint32
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
	offset := 0
	_, err = binary.Decode(buf[offset:offset+int(CRCSize)], binary.BigEndian, &le.crc)
	if err != nil {
		fmt.Printf("Error decoding crc %v\n", err)
		return
	}
	offset += int(CRCSize)

	crcHash := crc32.ChecksumIEEE(buf[CRCSize:])
	if crcHash != le.crc {
		return le, fmt.Errorf("CRC hash of data does not match the expected value, some bits are corrupted!")
	}

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
	binary.Decode(buf[offset:offset+int(ValLenSize)], binary.BigEndian, &valLen)
	_, err = binary.Decode(buf[offset:offset+int(KeyLenSize)], binary.BigEndian, &keyLen)
	if err != nil {
		fmt.Printf("Error decoding value length %v\n", err)
		return
	}
	offset += int(ValLenSize)

	le.key = string(buf[offset : offset+int(keyLen)])
	offset += int(keyLen)

	le.value = string(buf[offset : offset+int(valLen)])

	return
}

func main() {
	fmt.Println("hello world")

	testLogEntry := newLogEntry(1794596697, "hello", "world")
	encodedLogEntry := testLogEntry.encode()
	fmt.Printf("Encoded log data: 0x% X\n", encodedLogEntry)

	decodedLogEntry, err := decode(encodedLogEntry[2:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decoded entry is %+v\n", decodedLogEntry)
}
