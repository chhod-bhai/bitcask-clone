package store

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/chhod-bhai/bitcask-clone/model"
)

type store struct {
	keydir       map[string]model.KeydirEntry
	files        map[uint32]*os.File
	activeFileID uint32
	writeOffset  int64
	storePath    string
}

// Merge implements [Store].
func (s *store) Merge(directoryPath string, options any) error {
	panic("unimplemented")
}

// Close implements [Store].
func (s *store) Close() error {
	mainFile, found := s.files[s.activeFileID]
	if !found {
		// file not open
		panic("file not found for closing")
	}
	return mainFile.Close()
}

// Delete implements [Store].
func (s *store) Delete(key string) error {
	now := time.Now().Unix()
	_, found := s.keydir[key]
	if !found {
		return errors.New("Cannot delete key that does not exist")
	}

	logEntry := model.NewTombstone(uint32(now), key)
	if err := s.putEntry(key, logEntry); err != nil {
		return err
	}
	delete(s.keydir, key)
	return nil
}

// Get implements [Store].
func (s *store) Get(key string) (string, error) {
	entryPtr, found := s.keydir[key]
	if !found {
		return "", nil
	}
	f, _ := s.files[entryPtr.FileID]
	// currently assume there is only 1 file and its always present in open files
	val := make([]byte, entryPtr.RecordSize)
	_, err := f.ReadAt(val, entryPtr.Offset)
	if err != nil {
		// error reading from file
		return "", err
	}
	le, err := model.DecodeLogEntry(val)
	if err != nil {
		// error decoding logEntry
		return "", err
	}
	return le.Value, nil
}

func (s *store) putEntry(key string, le model.LogEntry) error {
	file := s.files[s.activeFileID]
	bytesWritten, err := file.Write(le.Encode())
	if err != nil {
		return err
	}
	if bytesWritten < len(le.Value) {
		// TODO: handle partial writes later
		return io.ErrShortWrite
	}
	keyDirEntry := model.KeydirEntry{
		FileID:     s.activeFileID,
		Offset:     s.writeOffset,
		RecordSize: uint32(int(model.HeaderSize) + len(key) + len(le.Value)),
		Timestamp:  uint32(le.Timestamp),
	}
	s.writeOffset += int64(bytesWritten)
	s.keydir[key] = keyDirEntry
	return nil
}

// Put implements [Store].
func (s *store) Put(key string, value string) error {
	if value == "" {
		return errors.New("Value cannot be empty while putting a record")
	}
	now := time.Now().Unix()
	logEntry := model.NewLogEntry(uint32(now), key, value)
	return s.putEntry(key, logEntry)
}

type Store interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Close() error
	Merge(directoryPath string, options any) error
}

func Open(storePath string) Store {
	file, err := os.OpenFile(storePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		// TODO: proper error handling
		panic(err)
	}
	activeFileId := uint32(file.Fd())
	storeInstance := store{
		keydir: map[string]model.KeydirEntry{},
		files:  map[uint32]*os.File{},
	}
	storeInstance.activeFileID = activeFileId
	storeInstance.files[activeFileId] = file
	fileInfo, _ := file.Stat()
	storeInstance.writeOffset = fileInfo.Size()
	storeInstance.storePath = storePath
	// populate keyDir using the store files here
	return &storeInstance
}
