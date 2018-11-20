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
	Buses   map[string]uint32
	Wires   map[string]Wire
	Aliases map[string]string
	Gates   []Gate
}

type Device struct {
	*Circuit
	Memory []bool
}

func NewCircuit() Circuit {
	return Circuit{
		Buses:   make(map[string]uint32),
		Wires:   make(map[string]Wire),
		Aliases: make(map[string]string),
	}
}

func (c *Circuit) AddBus(prefix string, count int, nominal bool) {
	_, ok := c.Buses[prefix]
	if ok {
		panic(fmt.Errorf("bus %s already exists", prefix))
	}
	c.Buses[prefix] = uint32(count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		c.Wires[name] = Wire{
			Name:    name,
			Nominal: nominal,
			Index:   uint32(len(c.Wires)),
		}
	}
}

func (c *Circuit) AddWire(name string, nominal bool) {
	_, ok := c.Wires[name]
	if ok {
		panic(fmt.Errorf("wire %s already exists", name))
	}
	c.Wires[name] = Wire{
		Name:    name,
		Nominal: nominal,
		Index:   uint32(len(c.Wires)),
	}
}

func (c *Circuit) AddAlias(name, alias string) {
	_, ok := c.Aliases[alias]
	if ok {
		panic(fmt.Errorf("alias %s already exists", alias))
	}
	c.Aliases[alias] = name
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

func (d *Device) Reset() {
	memory := d.Memory
	for _, value := range d.Wires {
		memory[value.Index] = value.Nominal
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

func (d *Device) SetUint64(prefix string, count int, value uint64) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		s := d.Wires[name]
		if value&1 == 0 {
			memory[s.Index] = false
		} else {
			memory[s.Index] = true
		}
		value >>= 1
	}
}

func (d *Device) Print(prefix string, count int) {
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), 0
		if alias, ok := d.Aliases[name]; ok {
			s := d.Wires[alias]
			if memory[s.Index] {
				bit = 1
			}
		} else {
			s := d.Wires[name]
			if memory[s.Index] {
				bit = 1
			}
		}
		fmt.Printf("%s=%d\n", name, bit)
	}
}

func (d *Device) Uint64(prefix string, count int) uint64 {
	var value uint64
	memory := d.Memory
	for i := 0; i < count; i++ {
		name, bit := fmt.Sprintf("%s%d", prefix, i), uint64(0)
		if alias, ok := d.Aliases[name]; ok {
			s := d.Wires[alias]
			if memory[s.Index] {
				bit = 1
			}
		} else {
			s := d.Wires[name]
			if memory[s.Index] {
				bit = 1
			}
		}
		value = value | (bit << uint(i))
	}
	return value
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
	}
	halfAdder := func(a, b, d string) {
		circuit.AddGateCCNot(a, b, d)
		circuit.AddGateCNot(b, a)
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

	device := circuit.NewDevice()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", 4, y)
			device.SetUint64("X", 4, x)
			device.Execute()
			r := device.Uint64("P", 8)
			if r != x*y {
				panic(fmt.Errorf("%d * %d != %d (%d)", x, y, r, x*y))
			}
			fmt.Printf("%d * %d == %d\n", x, y, r)
			device.Reset()
		}
	}
}
