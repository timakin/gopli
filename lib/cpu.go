package lib

import (
	"os"
	"runtime"
)

func SetupMultiCore() {
	maxProcs := os.Getenv("GOMAXPROCS")

	if maxProcs == "" {
		cpus := runtime.NumCPU()
		runtime.GOMAXPROCS(cpus)
	}
}
