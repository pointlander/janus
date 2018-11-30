package main

import (
	"flag"
	"fmt"
	"os"
)

var graph = flag.Bool("graph", false, "graph the search space")

func searchSpace() {
	circuit := Multiplier4()
	circuit.ComputeRanks()
	//circuit.PrintRanked()
	//circuit.PrintConnections("A12")

	device := circuit.NewDeviceBool()
	target := 121
	file, err := os.Create("points.dat")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fileSimple, err := os.Create("simple.dat")
	if err != nil {
		panic(err)
	}
	defer fileSimple.Close()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", 4, y)
			device.SetUint64("X", 4, x)
			device.Execute(false)
			device.SetUint64("P", 8, uint64(target))
			device.Execute(true)
			count := 0.0
			for i := 0; i < 16; i++ {
				name := fmt.Sprintf("A%d", i)
				wire := device.Wires[device.Resolve(name)]
				if device.Get(name) {
					count += wire.Rank
				}
			}
			for i := 0; i < 12; i++ {
				name := fmt.Sprintf("Z%d", i)
				wire := device.Wires[device.Resolve(name)]
				if device.Get(name) {
					count += wire.Rank
				}
			}
			device.Reset()
			fmt.Fprintf(file, "%d %d %f\n", x, y, count)
			fitness := target - int(x*y)
			if fitness < 0 {
				fitness = -fitness
			}
			fmt.Fprintf(fileSimple, "%d %d %d\n", x, y, fitness)
		}
	}
}

func main() {
	flag.Parse()

	if *graph {
		searchSpace()
		return
	}

	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	device.SetUint64("Y", 4, 4)
	device.Set("Y0", Dual{Val: 1, Der: 1})
	//device.Set("Y1", Dual{Val: .75, Der: 1})
	device.SetUint64("X", 4, 5)
	device.Execute(false)
	target := 25
	var total Dual
	for i := 0; i < 8; i++ {
		var a Dual
		if target&1 == 1 {
			a.Val = 1.0
		}
		b := device.Get(fmt.Sprintf("P%d", i))
		total = Add(total, Pow(Sub(a, b), 2))
		target >>= 1
	}
	fmt.Printf("Val: %f, Der: %f\n", total.Val, total.Der)
	device.Reset()
}
