package main

import (
	"github.com/matryer/is"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseIDL(t *testing.T) {
	i := is.New(t)
	err := ParseIDL(filepath.Join("test", "WebView2.idl"), "test")
	if err != nil {
		println(err.Error())
	}
	i.NoErr(err)
	err = os.Chdir("test")
	i.NoErr(err)
	command := exec.Command("go", "fmt", "./...")
	err = command.Run()
	i.NoErr(err)
}
