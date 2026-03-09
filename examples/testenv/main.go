package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/composer/utils"
	"github.com/mandelsoft/vfs/pkg/osfs"
)

func ExitOnError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {

	root := utils.MapFSTree(osfs.OsFs, "../../examples/testenv/testdata")
	fmt.Printf("// --- begin tree output ---\n")
	fmt.Printf("%s\n", utils.MapNodetoASCII(root))
	fmt.Printf("// --- end tree output ---\n")
}
