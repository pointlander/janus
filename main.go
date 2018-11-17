package main

import "fmt"

type GateType int

const (
	GateTypeNot GateType = iota
	GateTypeCNot
	GateTypeCCNot
)

type Wire struct {
	Name    string
	Nominal bool
	Index   uint32
}

type Gate struct {
	Type GateType
	Taps [3]uint32
}

type Circuit struct {
	Wires map[string]Wire
	Gates []Gate
}

type Device struct {
	*Circuit
	Memory []bool
}

func NewCircuit() Circuit {
	return Circuit{
		Wires: make(map[string]Wire),
	}
}

func (c *Circuit) AddBus(prefix string, count int, nominal bool) {
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		c.Wires[name] = Wire{
			Name:    name,
			Nominal: nominal,
			Index:   uint32(len(c.Wires)),
		}
	}
}

func (c *Circuit) AddGateNot(a string) {
	gate := Gate{
		Type: GateTypeNot,
		Taps: [3]uint32{c.Wires[a].Index},
	}
	c.Gates = append(c.Gates, gate)
}

func (c *Circuit) AddGateCNot(a, b string) {
	gate := Gate{
		Type: GateTypeCNot,
		Taps: [3]uint32{c.Wires[a].Index, c.Wires[b].Index},
	}
	c.Gates = append(c.Gates, gate)
}

func (cc *Circuit) AddGateCCNot(a, b, c string) {
	gate := Gate{
		Type: GateTypeCCNot,
		Taps: [3]uint32{cc.Wires[a].Index, cc.Wires[b].Index, cc.Wires[c].Index},
	}
	cc.Gates = append(cc.Gates, gate)
}

func (c *Circuit) NewDevice() Device {
	memory := make([]bool, len(c.Wires))
	for _, value := range c.Wires {
		memory[value.Index] = value.Nominal
	}
	return Device{
		Circuit: c,
		Memory:  memory,
	}
}

func (d *Device) Set(prefix string, values ...bool) {
	memory := d.Memory
	for i, value := range values {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		memory[s.Index] = value
	}
}

func (d *Device) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		s, bit := d.Wires[name], 0
		if memory[s.Index] {
			bit = 1
		}
		fmt.Printf("%s=%d\n", name, bit)
	}
}

func (d *Device) Execute() {
	memory := d.Memory
	for _, gate := range d.Gates {
		switch gate.Type {
		case GateTypeNot:
			a := memory[gate.Taps[0]]
			a = !a
			memory[gate.Taps[0]] = a
		case GateTypeCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			b = a != b
			memory[gate.Taps[1]] = b
		case GateTypeCCNot:
			a := memory[gate.Taps[0]]
			b := memory[gate.Taps[1]]
			c := memory[gate.Taps[2]]
			c = (a && b) != c
			memory[gate.Taps[2]] = c
		}
	}
}

func main() {
	circuit := NewCircuit()

	circuit.AddBus("Y", 4, false)
	circuit.AddBus("X", 4, false)
	circuit.AddBus("A", 16, false)

	circuit.AddGateCCNot("Y0", "X0", "A0")
	circuit.AddGateCCNot("Y1", "X0", "A1")
	circuit.AddGateCCNot("Y2", "X0", "A2")
	circuit.AddGateCCNot("Y3", "X0", "A3")

	circuit.AddGateCCNot("Y0", "X1", "A4")
	circuit.AddGateCCNot("Y1", "X1", "A5")
	circuit.AddGateCCNot("Y2", "X1", "A6")
	circuit.AddGateCCNot("Y3", "X1", "A7")

	circuit.AddGateCCNot("Y1", "X2", "A8")
	circuit.AddGateCCNot("Y2", "X2", "A9")
	circuit.AddGateCCNot("Y3", "X2", "A10")
	circuit.AddGateCCNot("X0", "X2", "A11")

	circuit.AddGateCCNot("Y2", "X3", "A12")
	circuit.AddGateCCNot("Y3", "X3", "A13")
	circuit.AddGateCCNot("X0", "X3", "A14")
	circuit.AddGateCCNot("X1", "X3", "A15")

	device := circuit.NewDevice()
	device.Set("Y", true, false, true, false)
	device.Set("X", true, false, true, false)
	device.Execute()
	device.Print("A", 16)
}
