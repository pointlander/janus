// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
)

var (
	help   = flag.Bool("help", false, "prints help")
	graph  = flag.Bool("graph", false, "graph the search space")
	factor = flag.Uint("factor", 77, "number to factor")
	all    = flag.Bool("all", false, "factor all numbers")
	mode   = flag.String("mode", "forward", "factoring algorithm")
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
			device.SetUint64("Y", y)
			device.SetUint64("X", x)
			device.Execute(false)
			device.SetUint64("P", uint64(target))
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

func factorForward(size int, factor uint, limit int, log bool) (y, x uint64, factored bool) {
	type Hill struct {
		Y, X uint64
	}
	max := uint64(1)
	for i := 0; i < size; i++ {
		max *= 2
	}

	rand.Seed(1)
	iterations := 0
	//circuit := Multiplier4()
	circuit := Multiplier(size, FullAdderA1, HalfAdderA1)
	device := circuit.NewDeviceDual()
	//root := uint64(math.Sqrt(float64(factor)))
	device.SetUint64("Y", max-1)
	device.SetUint64("X", max-1)
	hills := []Hill{}
	inputs := device.AllocateSlice("I")
	device.GetSlice("I", inputs)
	lastX, lastY, stuck := uint64(0), uint64(0), 0
	for {
		iterations++
		if limit != 0 && iterations > limit {
			break
		}

		input := rand.Intn(len(inputs))
		inputs[input].Der = 1
		device.SetSlice("I", inputs)
		inputs[input].Der = 0
		device.Execute(false)

		var cost Dual
		target := factor
		for i := 0; i < 2*size; i++ {
			var a Dual
			if target&1 == 1 {
				a.Val = 1.0
			}
			b := device.Get(fmt.Sprintf("P%d", i))
			cost = Add(cost, Pow(Sub(a, b), 2))
			target >>= 1
		}

		for _, hill := range hills {
			acc := Dual{Val: 1.0}
			for i := 0; i < size; i++ {
				value := device.Get(fmt.Sprintf("Y%d", i))
				bit := hill.Y & 1
				if bit == 1 {
					acc = Mul(acc, value)
				} else {
					acc = Mul(acc, Sub(One, value))
				}
				hill.Y >>= 1
			}
			for i := 0; i < size; i++ {
				value := device.Get(fmt.Sprintf("X%d", i))
				bit := hill.X & 1
				if bit == 1 {
					acc = Mul(acc, value)
				} else {
					acc = Mul(acc, Sub(One, value))
				}
				hill.X >>= 1
			}
			cost = Add(cost, acc)
		}

		// Y != 1
		hill := 1
		acc := Dual{Val: 1.0}
		for i := 0; i < size; i++ {
			value := device.Get(fmt.Sprintf("Y%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		cost = Add(cost, acc)

		// X != 1
		hill = 1
		acc = Dual{Val: 1.0}
		for i := 0; i < size; i++ {
			value := device.Get(fmt.Sprintf("X%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		cost = Add(cost, acc)

		// Y != 0
		hill = 0
		acc = Dual{Val: 1.0}
		for i := 0; i < size; i++ {
			value := device.Get(fmt.Sprintf("Y%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		cost = Add(cost, acc)

		// X != 0
		hill = 0
		acc = Dual{Val: 1.0}
		for i := 0; i < size; i++ {
			value := device.Get(fmt.Sprintf("X%d", i))
			bit := hill & 1
			if bit == 1 {
				acc = Mul(acc, value)
			} else {
				acc = Mul(acc, Sub(One, value))
			}
			hill >>= 1
		}
		cost = Add(cost, acc)

		if log {
			fmt.Printf("%d Val: %f, Der: %f\n", input, cost.Val, cost.Der)
			fmt.Printf("P: %d, Y: %d, X: %d\n", device.Uint64("P"), device.Uint64("Y"), device.Uint64("X"))
		}
		if math.IsNaN(float64(cost.Der)) {
			break
		} else if cost.Val == 0 {
			y = device.Uint64("Y")
			x = device.Uint64("X")
			factored = true
			break
		}

		if cost.Der > 0 {
			inputs[input].Val = 0
		} else if cost.Der < 0 {
			inputs[input].Val = 1
		}

		yy := device.Uint64("Y")
		xx := device.Uint64("X")
		if yy == lastY && xx == lastX {
			stuck++
		} else {
			stuck = 0
		}
		lastY, lastX = yy, xx
		if stuck > 16 {
			stuck = 0
			hills = append(hills, Hill{Y: yy, X: xx})
		}

		device.Reset()
	}
	if log {
		fmt.Printf("iterations=%d\n", iterations)
		fmt.Printf("hills=%d\n", hills)
	}
	return y, x, factored
}

func factorForwardProbabilistic(size int, factor uint, limit int, log bool) (y, x uint64, factored bool) {
	type Hill struct {
		Y, X uint64
	}

	rand.Seed(1)
	iterations := 0
	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	//root := uint64(math.Sqrt(float64(factor)))
	device.SetUint64("Y", 15)
	device.SetUint64("X", 15)
	hills := []Hill{}
	inputs := device.AllocateSlice("I")
	device.GetSlice("I", inputs)
	lastX, lastY, stuck := uint64(0), uint64(0), 0
	der := make([]float32, len(inputs))
search:
	for {
		iterations++
		if limit != 0 && iterations > limit {
			break
		}

		for input := range inputs {
			inputs[input].Der = 1
			device.SetSlice("I", inputs)
			inputs[input].Der = 0
			device.Execute(false)

			var cost Dual
			target := factor
			for i := 0; i < 8; i++ {
				var a Dual
				if target&1 == 1 {
					a.Val = 1.0
				}
				b := device.Get(fmt.Sprintf("P%d", i))
				cost = Add(cost, Pow(Sub(a, b), 2))
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
				cost = Add(cost, acc)
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
			cost = Add(cost, acc)

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
			cost = Add(cost, acc)

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
			cost = Add(cost, acc)

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
			cost = Add(cost, acc)

			if log {
				fmt.Printf("%d Val: %f, Der: %f\n", input, cost.Val, cost.Der)
				fmt.Printf("P: %d, Y: %d, X: %d\n", device.Uint64("P"), device.Uint64("Y"), device.Uint64("X"))
			}
			if math.IsNaN(float64(cost.Der)) {
				break search
			} else if cost.Val == 0 {
				y = device.Uint64("Y")
				x = device.Uint64("X")
				factored = true
				break search
			}

			der[input] = float32(math.Abs(float64(cost.Der)))
			device.Reset()
		}

		var sum float32
		for i, d := range der {
			d = float32(math.Exp(float64(d)))
			sum += d
			der[i] = d
		}
		r, s, mutate := float32(0.0), rand.Float32(), 0
		for i, d := range der {
			r += d / sum
			//fmt.Printf("%f ", d/sum)
			if s < r {
				mutate = i
				break
			}
		}
		//fmt.Printf("\n")
		inputs[mutate].Val = 1 - inputs[mutate].Val

		device.SetSlice("I", inputs)
		yy := device.Uint64("Y")
		xx := device.Uint64("X")
		if yy == lastY && xx == lastX {
			stuck++
		} else {
			stuck = 0
		}
		lastY, lastX = yy, xx
		if stuck > 16 {
			stuck = 0
			//hills = append(hills, Hill{Y: yy, X: xx})
		}
	}
	if log {
		fmt.Printf("iterations=%d\n", iterations)
		fmt.Printf("hills=%d\n", hills)
	}
	return y, x, factored
}

func factorReverse(size int, factor uint, limit int, log bool) (y, x uint64, factored bool) {
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
			device.SetUint64("P", uint64(factor))
			device.Execute(true)
			var cost Dual
			for i := 0; i < 16; i++ {
				a := device.Get(fmt.Sprintf("A%d", i))
				cost = Add(cost, Pow(a, 2))
			}
			for i := 0; i < 12; i++ {
				a := device.Get(fmt.Sprintf("Z%d", i))
				cost = Add(cost, Pow(a, 2))
			}

			if log {
				fmt.Printf("%d Val: %f, Der: %f\n", name, cost.Val, cost.Der)
				fmt.Printf("Y: %d, X: %d\n", device.Uint64("Y"), device.Uint64("X"))
			}
			if math.IsNaN(float64(cost.Der)) {
				break search
			} else if cost.Val == 0 {
				y = device.Uint64("Y")
				x = device.Uint64("X")
				factored = true
				break search
			}

			values[name].Der = 0.0
			if cost.Der > 0 {
				values[name].Val = 0.0
			} else if cost.Der < 0 {
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

	if *help {
		flag.PrintDefaults()
		return
	}

	if *graph {
		searchSpace()
		return
	}

	if *factor > 15*15 {
		panic(fmt.Errorf("factor must be [0,%d]", 15*15))
	}

	var f func(size int, factor uint, limit int, log bool) (y, x uint64, factored bool)
	var iterations int
	switch *mode {
	case "forward":
		f, iterations = factorForward, 1000
	case "reverse":
		f, iterations = factorReverse, 100
	case "prob":
		f, iterations = factorForwardProbabilistic, 1000
	default:
		panic("invalid mode; valid modes: [forward, reverse, prob]")
	}

	size := 5
	max := uint64(1)
	for i := 0; i < size; i++ {
		max *= 2
	}
	space := (max - 1) * (max - 1)

	if *all {
		primes := []uint{2, 3}
		for i := uint(4); i <= uint(space); i++ {
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
		for i := uint(2); i <= uint(space); i++ {
			factors := 0
			for _, prime := range primes {
				if i%prime == 0 {
					factors++
				}
			}
			fmt.Printf("%d (%d)", i, factors)
			if primeMap[i] {
				fmt.Printf(" is prime\n")
			} else if y, x, ok := f(size, uint(i), iterations, false); ok {
				fmt.Printf(" factored %d %d\n", y, x)
				factored++
				total++
			} else {
				total++
				fmt.Printf("\n")
			}
			if i == 225 {
				break
			}
		}
		fmt.Printf("factored=%d/%d %f\n", factored, total, float64(factored)/float64(total))
		return
	}

	f(size, *factor, 0, true)
}
