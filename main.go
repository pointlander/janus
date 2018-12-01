package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
)

var (
	graph   = flag.Bool("graph", false, "graph the search space")
	factor  = flag.Uint("factor", 77, "number to factor")
	reverse = flag.Bool("reverse", false, "factor in reverse")
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

func factorForward(factor uint, limit int, log bool) (y, x uint64, factored bool) {
	type Hill struct {
		Y, X uint64
	}

	rand.Seed(1)
	iterations := 0
	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	//root := uint64(math.Sqrt(float64(factor)))
	device.SetUint64("Y", 4, 15)
	device.SetUint64("X", 4, 15)
	hills := []Hill{}
	lastX, lastY, count := uint64(0), uint64(0), 0
	yy, xx := device.AllocateSlice("Y"), device.AllocateSlice("X")
	for {
		iterations++
		if limit != 0 && iterations > limit {
			break
		}
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
		target := factor
		for i := 0; i < 8; i++ {
			var a Dual
			if target&1 == 1 {
				a.Val = 1.0
			}
			b := device.Get(fmt.Sprintf("P%d", i))
			total = Add(total, Pow(Sub(a, b), 2))
			target >>= 1
		}

		for _, hill := range hills {
			acc := Dual{Val: 1.0}
			for i := 0; i < 4; i++ {
				value := device.Get(fmt.Sprintf("Y%d", i))
				bit := hill.Y & 1
				if bit == 1 {
					acc = Mul(acc, value)
				} else {
					acc = Mul(acc, Sub(One, value))
				}
				hill.Y >>= 1
			}
			for i := 0; i < 4; i++ {
				value := device.Get(fmt.Sprintf("X%d", i))
				bit := hill.X & 1
				if bit == 1 {
					acc = Mul(acc, value)
				} else {
					acc = Mul(acc, Sub(One, value))
				}
				hill.X >>= 1
			}
			total = Add(total, acc)
		}

		// Y != 1
		hill := 1
		acc := Dual{Val: 1.0}
		for i := 0; i < 4; i++ {
			value := device.Get(fmt.Sprintf("Y%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		total = Add(total, acc)

		// X != 1
		hill = 1
		acc = Dual{Val: 1.0}
		for i := 0; i < 4; i++ {
			value := device.Get(fmt.Sprintf("X%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		total = Add(total, acc)

		// Y != 0
		hill = 0
		acc = Dual{Val: 1.0}
		for i := 0; i < 4; i++ {
			value := device.Get(fmt.Sprintf("Y%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		total = Add(total, acc)

		// X != 0
		hill = 0
		acc = Dual{Val: 1.0}
		for i := 0; i < 4; i++ {
			value := device.Get(fmt.Sprintf("X%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		total = Add(total, acc)

		if log {
			fmt.Printf("%s Val: %f, Der: %f\n", name, total.Val, total.Der)
			fmt.Printf("P: %d, Y: %d, X: %d\n", device.Uint64("P", 8), device.Uint64("Y", 4), device.Uint64("X", 4))
		}
		if math.IsNaN(float64(total.Der)) {
			break
		} else if total.Val == 0 {
			y = device.Uint64("Y", 4)
			x = device.Uint64("X", 4)
			factored = true
			break
		}

		bit := device.Get(name)
		bit.Der = 0
		if total.Der > 0 {
			bit.Val = 0
		} else if total.Der < 0 {
			bit.Val = 1
		}
		device.Set(name, bit)

		y := device.Uint64("Y", 4)
		x := device.Uint64("X", 4)
		if y == lastY && x == lastX {
			count++
		} else {
			count = 0
		}
		if count > 16 {
			count = 0
			hills = append(hills, Hill{Y: y, X: x})
		}
		lastY, lastX = y, x
		device.GetSlice("Y", yy)
		device.GetSlice("X", xx)
		device.Reset()
		device.SetSlice("Y", yy)
		device.SetSlice("X", xx)
	}
	if log {
		fmt.Printf("iterations=%d\n", iterations)
		fmt.Printf("hills=%d\n", hills)
	}
	return y, x, factored
}

func factorReverse(factor uint, limit int, log bool) (y, x uint64, factored bool) {
	rand.Seed(1)
	iterations := 0
	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	values := device.AllocateSlice("G")
	for i := range values {
		if rand.Intn(2) == 0 {
			values[i].Val = 1.0
		}
	}
search:
	for {
		iterations++
		if limit != 0 && iterations > limit {
			break
		}
		for name := range values {
			//name := rand.Intn(len(values))
			values[name].Der = 1.0
			device.SetSlice("G", values)
			device.SetUint64("P", 8, uint64(factor))
			device.Execute(true)
			var total Dual
			for i := 0; i < 16; i++ {
				a := device.Get(fmt.Sprintf("A%d", i))
				total = Add(total, Pow(a, 2))
			}
			for i := 0; i < 12; i++ {
				a := device.Get(fmt.Sprintf("Z%d", i))
				total = Add(total, Pow(a, 2))
			}

			if log {
				fmt.Printf("%d Val: %f, Der: %f\n", name, total.Val, total.Der)
				fmt.Printf("Y: %d, X: %d\n", device.Uint64("Y", 4), device.Uint64("X", 4))
			}
			if math.IsNaN(float64(total.Der)) {
				break search
			} else if total.Val == 0 {
				y = device.Uint64("Y", 4)
				x = device.Uint64("X", 4)
				factored = true
				break search
			}

			values[name].Der = 0.0
			if total.Der > 0 {
				values[name].Val = 0.0
			} else if total.Der < 0 {
				values[name].Val = 1.0
			}
			device.Reset()
		}
	}
	if log {
		fmt.Printf("iterations=%d\n", iterations)
	}
	return y, x, factored
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

	if *reverse {
		factorReverse(*factor, 0, true)
		return
	}

	primes := []uint{2, 3}
	for i := uint(4); i < uint(226); i++ {
		isPrime := true
		for _, prime := range primes {
			if i%prime == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primes = append(primes, i)
		}
	}
	primeMap := make(map[uint]bool)
	for _, prime := range primes {
		primeMap[prime] = true
	}

	factored, total := 0, 0
	for i := uint(2); i < uint(226); i++ {
		factors := 0
		for _, prime := range primes {
			if i%prime == 0 {
				factors++
			}
		}
		fmt.Printf("%d (%d)", i, factors)
		if primeMap[i] {
			fmt.Printf(" is prime\n")
		} else if y, x, ok := factorForward(uint(i), 1000, false); ok {
			fmt.Printf(" factored %d %d\n", y, x)
			factored++
			total++
		} else {
			total++
			fmt.Printf("\n")
		}
	}
	fmt.Printf("factored=%d/%d\n", factored, total)

	/*factored, total = 0, 0
	for i := uint(2); i < uint(226); i++ {
		fmt.Printf("%d", i)
		if primeMap[i] {
			fmt.Printf(" is prime\n")
		} else if y, x, ok := factorReverse(uint(i), 100, false); ok {
			fmt.Printf(" factored %d %d\n", y, x)
			factored++
			total++
		} else {
			total++
			fmt.Printf("\n")
		}
	}
	fmt.Printf("factored=%d/%d\n", factored, total)*/
}
