                 Y3   Y2   Y1   Y0
                 X3   X2   X1   X0
                 -------------------
                 X0Y3 X0Y2 X0Y1 X0Y0
            X1Y3 X1Y2 X1Y1 X1Y0
       X2Y3 X2Y2 X2Y1 X2Y0
  X3Y3 X3Y2 X3Y1 X3Y0

                 Y3   Y2   Y1   Y0
                 X3   X2   X1   X0
                 -------------------
                 A3   A2   A1   A0
            A7   A6   A5   A4
       A11  A10  A9   A8
  A15  A14  A13  A12

x = 1 - x

x = (1 - y) * x + (1 - x) * y
x = x - xy + y - xy
x = x + y - 2xy

x = (1 - .5 * y1 - .5 * y2) * x + (1 - x) * (.5 * y1 + .5 * y2)

Use sqrt as starting point
Make search more granular
Use waste heat in search
Reshape search space

Restructure adder network in random ways

4 bit
factored=81/176 0.460227
8 bit
factored=1430/58529 0.024432
