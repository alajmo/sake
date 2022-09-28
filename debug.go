package main

import (
	"log"
	"runtime/pprof"
	"os"
	_ "net/http/pprof"

	"github.com/alajmo/sake/cmd"
)

func main() {
	fc, err := os.Create("./benchmarks/cpu.prof")
	if err != nil {
	 log.Fatal(err)
	}

	fh, err := os.Create("./benchmarks/heap.prof")
	if err != nil {
	 log.Fatal(err)
	}

	fg, err := os.Create("./benchmarks/goroutine.prof")
	if err != nil {
	 log.Fatal(err)
	}

	pprof.StartCPUProfile(fc)
	cmd.Execute()
	pprof.Lookup("heap").WriteTo(fh, 0)
	pprof.Lookup("goroutine").WriteTo(fg, 0)
	defer pprof.StopCPUProfile()

	// go tool pprof -http="localhost:8000" ./benchmarks/cpu.prof
	// go tool pprof -http="localhost:8000" ./benchmarks/heap.prof
	// go tool pprof -http="localhost:8000" ./benchmarks/goroutine.prof
}
