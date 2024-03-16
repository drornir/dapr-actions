package main

import (
	"fmt"
	"io"

	"github.com/spf13/afero"

)

var FS afero.Fs

func main() {
	FS = afero.NewMemMapFs()
	f, _ := FS.Create("f.txt")
	f.WriteString("hello")
	f.Close()

	ff, _ := FS.Open("f.txt")
	bytes, _ := io.ReadAll(ff)
	fmt.Printf("%s", bytes)
}
