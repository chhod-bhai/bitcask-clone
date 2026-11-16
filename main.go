package main

import (
	"os"
	"path/filepath"

	"github.com/chhod-bhai/bitcask-clone/store"
)

func main() {
	cwd, _ := os.Getwd()
	storePath := filepath.Join(cwd, "data/1.bitcask-clone.data")
	bitcaskCloneStore := store.Open(storePath)
	bitcaskCloneStore.Put("hello", "world")
	bitcaskCloneStore.Get("hello")
}
