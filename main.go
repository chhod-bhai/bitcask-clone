package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chhod-bhai/bitcask-clone/store"
)

func main() {
	cwd, _ := os.Getwd()
	storePath := filepath.Join(cwd, "data/1.bitcask-clone.data")
	bitcaskCloneStore := store.Open(storePath)
	defer bitcaskCloneStore.Close()

	bitcaskCloneStore.Put("hello", "world")
	fmt.Println(bitcaskCloneStore.Get("hello"))
	bitcaskCloneStore.Put("hello", "metal")
	fmt.Println(bitcaskCloneStore.Get("hello"))
	bitcaskCloneStore.Put("hello", "petal")
	fmt.Println(bitcaskCloneStore.Get("hello"))

}
