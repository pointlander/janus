// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func Multiplier4() Circuit {
	circuit := NewCircuit()

	circuit.AddBus("I", 0, true)
	circuit.AddBus("Y", 4, false)
	circuit.AddBus("X", 4, false)
	circuit.AddBus("A", 16, false)
	circuit.AddBus("P", 8, true)
	circuit.AddBus("G", 0, true)

	circuit.AddAlias("Y", "I")
	circuit.AddAlias("X", "I")
	circuit.AddAlias("Y", "G")
	circuit.AddAlias("X", "G")

	circuit.AddGateCCNot("Y0", "X0", "A0")
	circuit.AddGateCCNot("Y1", "X0", "A1")
	circuit.AddGateCCNot("Y2", "X0", "A2")
	circuit.AddGateCCNot("Y3", "X0", "A3")

	circuit.AddGateCCNot("Y0", "X1", "A4")
	circuit.AddGateCCNot("Y1", "X1", "A5")
	circuit.AddGateCCNot("Y2", "X1", "A6")
	circuit.AddGateCCNot("Y3", "X1", "A7")

	circuit.AddGateCCNot("Y0", "X2", "A8")
	circuit.AddGateCCNot("Y1", "X2", "A9")
	circuit.AddGateCCNot("Y2", "X2", "A10")
	circuit.AddGateCCNot("Y3", "X2", "A11")

	circuit.AddGateCCNot("Y0", "X3", "A12")
	circuit.AddGateCCNot("Y1", "X3", "A13")
	circuit.AddGateCCNot("Y2", "X3", "A14")
	circuit.AddGateCCNot("Y3", "X3", "A15")

	fullAdder := func(a, b, c, d string) {
		circuit.AddGateCCNot(a, b, d)
		circuit.AddGateCNot(b, a)
		circuit.AddGateCCNot(a, c, d)
		circuit.AddGateCNot(c, a)
		circuit.AddAlias(b, "G")
		circuit.AddAlias(c, "G")
	}
	halfAdder := func(a, b, d string) {
		circuit.AddGateCCNot(a, b, d)
		circuit.AddGateCNot(b, a)
		circuit.AddAlias(b, "G")
	}

	circuit.AddAlias("A0", "P0")

	circuit.AddWire("Z0", false)
	halfAdder("A1", "A4", "Z0")
	circuit.AddAlias("A1", "P1")
	circuit.AddWire("Z1", false)
	fullAdder("A8", "A2", "Z0", "Z1")
	circuit.AddWire("Z2", false)
	fullAdder("A12", "A3", "Z1", "Z2")
	circuit.AddWire("Z3", false)
	halfAdder("A7", "Z2", "Z3")

	circuit.AddWire("Z4", false)
	halfAdder("A9", "A6", "Z4")
	circuit.AddWire("Z5", false)
	fullAdder("A10", "A13", "Z4", "Z5")
	circuit.AddWire("Z6", false)
	fullAdder("A14", "A11", "Z5", "Z6")

	circuit.AddWire("Z7", false)
	halfAdder("A8", "A5", "Z7")
	circuit.AddAlias("A8", "P2")
	circuit.AddWire("Z8", false)
	fullAdder("A12", "A9", "Z7", "Z8")
	circuit.AddAlias("A12", "P3")
	circuit.AddWire("Z9", false)
	fullAdder("A7", "A10", "Z8", "Z9")
	circuit.AddAlias("A7", "P4")
	circuit.AddWire("Z10", false)
	fullAdder("Z3", "A14", "Z9", "Z10")
	circuit.AddAlias("Z3", "P5")
	circuit.AddWire("Z11", false)
	fullAdder("Z6", "A15", "Z10", "Z11")
	circuit.AddAlias("Z6", "P6")
	circuit.AddAlias("Z11", "P7")

	return circuit
}
