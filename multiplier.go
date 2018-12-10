// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type FullAdder func(circuit *Circuit, a, b, c, d string) (sum, carry string)
type HalfAdder func(circuit *Circuit, b, c, d string) (sum, carry string)

func FullAdderA1(circuit *Circuit, a, b, c, d string) (sum, carry string) {
	circuit.AddGateCCNot(a, b, d)
	circuit.AddGateCNot(b, a)
	circuit.AddGateCCNot(a, c, d)
	circuit.AddGateCNot(c, a)
	circuit.AddAlias(b, "G")
	circuit.AddAlias(c, "G")
	return a, d
}
func HalfAdderA1(circuit *Circuit, a, b, d string) (sum, carry string) {
	circuit.AddGateCCNot(a, b, d)
	circuit.AddGateCNot(b, a)
	circuit.AddAlias(b, "G")
	return a, d
}

func FullAdderA2(circuit *Circuit, a, b, c, d string) (sum, carry string) {
	circuit.AddGateCCNot(a, c, d)
	circuit.AddGateCNot(a, c)
	circuit.AddGateCCNot(b, c, d)
	circuit.AddGateCNot(c, b)
	circuit.AddAlias(a, "G")
	circuit.AddAlias(c, "G")
	return b, d
}
func HalfAdderA2(circuit *Circuit, b, c, d string) (sum, carry string) {
	circuit.AddGateCCNot(b, c, d)
	circuit.AddGateCNot(c, b)
	circuit.AddAlias(c, "G")
	return b, d
}

func FullAdderA3(circuit *Circuit, a, b, c, d string) (sum, carry string) {
	circuit.AddGateCCNot(a, c, d)
	circuit.AddGateCNot(a, c)
	circuit.AddGateCCNot(b, c, d)
	circuit.AddGateCNot(b, c)
	circuit.AddAlias(a, "G")
	circuit.AddAlias(b, "G")
	return c, d
}
func HalfAdderA3(circuit *Circuit, b, c, d string) (sum, carry string) {
	circuit.AddGateCCNot(b, c, d)
	circuit.AddGateCNot(b, c)
	circuit.AddAlias(b, "G")
	return c, d
}

func Multiplier(size int, full FullAdder, half HalfAdder) Circuit {
	circuit := NewCircuit()

	circuit.AddBus("I", 0, true)
	circuit.AddBus("Y", size, false)
	circuit.AddBus("X", size, false)
	circuit.AddBus("A", size*size, false)
	circuit.AddBus("P", 2*size, true)
	circuit.AddBus("Z", 0, false)
	circuit.AddBus("G", 0, true)

	circuit.AddAlias("Y", "I")
	circuit.AddAlias("X", "I")
	circuit.AddAlias("Y", "G")
	circuit.AddAlias("X", "G")

	a, sums := 0, make([][]string, 2*size)
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			product, column := fmt.Sprintf("A%d", a), x+y
			circuit.AddGateCCNot(fmt.Sprintf("Y%d", y), fmt.Sprintf("X%d", x), product)
			sums[column] = append(sums[column], product)
			a++
		}
	}

	for i := range sums {
		product := fmt.Sprintf("P%d", i)
		for {
			if length := len(sums[i]); length > 2 {
				z := circuit.AddWire("Z", false)
				sum, carry := full(&circuit, sums[i][0], sums[i][1], sums[i][2], z)
				sums[i][2] = sum
				sums[i] = sums[i][2:]
				sums[i+1] = append(sums[i+1], carry)
			} else if length == 2 {
				z := circuit.AddWire("Z", false)
				sum, carry := half(&circuit, sums[i][0], sums[i][1], z)
				sums[i][1] = sum
				sums[i] = sums[i][1:]
				sums[i+1] = append(sums[i+1], carry)
			} else {
				circuit.AddAlias(sums[i][0], product)
				break
			}
		}
	}

	return circuit
}
