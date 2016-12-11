package lib

import (
	"github.com/k0kubun/pp"
	"os"
)

func Isnil(x interface{}) bool {
	return x == nil || x == 0
}

func DeleteTmpDir(dirPath string) {
	err := os.RemoveAll(dirPath)
	if err != nil {
		pp.Print(err)
	}
}
