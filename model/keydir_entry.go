package model

type KeydirEntry struct {
	FileID     uint32
	Offset     int64
	RecordSize uint32
	Timestamp  uint32
}
