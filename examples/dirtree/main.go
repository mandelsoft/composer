package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/composer/filesystem"
	"github.com/mandelsoft/composer/utils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
)

func ExitOnError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {

	// --- begin environment ---
	env, err := filesystem.New()
	// --- end environment ---
	ExitOnError(err)

	// --- begin structure code ---
	fs := memoryfs.New()
	err = env.Compose(func() {
		env.FileSystem(fs, func() {
			env.Directory("female", 0770, func() {
				env.File("alice", 0660, func() {
					env.StringContent("age: 24")
				})
				env.File("carol", 0660, func() {
					env.StringContent("age: 26")
				})
			})
			env.Directory("male", 0770, func() {
				env.File("bob", 0660, func() {
					env.StringContent("age: 25")
				})
				env.File("dave", 0660, func() {
					env.StringContent("age: 26")
				})
			})
		})
	})
	ExitOnError(err)
	// --- end structure code ---

	root := utils.MapFSTree(fs, "/")
	fmt.Printf("// --- begin tree output ---")
	fmt.Printf("%s\n", utils.MapNodetoASCII(root))
	fmt.Printf("// --- end tree output ---")
}
