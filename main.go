package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
)

var (
	graph  = flag.Bool("graph", false, "graph the search space")
	factor = flag.Uint("factor", 77, "number to factor")
)

func searchSpace() {
	circuit := Multiplier4()
	circuit.ComputeRanks()
	//circuit.PrintRanked()
	//circuit.PrintConnections("A12")

	device := circuit.NewDeviceBool()
	target := 225
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

	if *factor > 15*15 {
		panic(fmt.Errorf("factor must be [0,%d]", 15*15))
	}

	rand.Seed(1)
	iterations := 0
	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	device.SetUint64("Y", 4, 1)
	device.SetUint64("X", 4, 1)
	for {
		iterations++
		var name string
		if rand.Intn(2) == 0 {
			name = fmt.Sprintf("Y%d", rand.Intn(4))
			bit := device.Get(name)
			bit.Der = 1
			device.Set(name, bit)
		} else {
			name = fmt.Sprintf("X%d", rand.Intn(4))
			bit := device.Get(name)
			bit.Der = 1
			device.Set(name, bit)
		}
		device.Execute(false)
		var total Dual
		target := *factor
		for i := 0; i < 8; i++ {
			var a Dual
			if target&1 == 1 {
				a.Val = 1.0
			}
			b := device.Get(fmt.Sprintf("P%d", i))
			total = Add(total, Pow(Sub(a, b), 2))
			target >>= 1
		}

		fmt.Printf("%s Val: %f, Der: %f\n", name, total.Val, total.Der)
		fmt.Printf("P: %d, Y: %d, X: %d\n", device.Uint64("P", 8), device.Uint64("Y", 4), device.Uint64("X", 4))
		if math.IsNaN(float64(total.Der)) {
			break
		} else if total.Val == 0 {
			break
		}

		if total.Der > 0 {
			device.Set(name, Dual{0, 0})
		} else if total.Der < 0 {
			device.Set(name, Dual{1, 0})
		} else {
			bit := device.Get(name)
			bit.Der = 0
			device.Set(name, bit)
		}
		y := device.Uint64("Y", 4)
		x := device.Uint64("X", 4)
		device.Reset()
		device.SetUint64("Y", 4, y)
		device.SetUint64("X", 4, x)
	}
	fmt.Printf("iterations=%d\n", iterations)
}
