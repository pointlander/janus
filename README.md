# About

Number factoring is implemented with a [reversible multiplier logic circuit](https://arxiv.org/abs/0907.3357). Logic gates are implemented with continuous equations that are equivalent to the reversible logic gates. Gradient descent is then performed on the resulting equations in order to factor a number.

# Testing

To factor all numbers with a 4x4 multiplier circuit in forward mode:

```bash
go build
./janus -all
```
