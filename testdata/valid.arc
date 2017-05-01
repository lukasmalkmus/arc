.begin
.org 0x800
main: ld [x], %r1
ld [y], %r2
add %r1, %r2, %r3
st %r3, [z]
be main
.org 0x1000
x: 0
y: 0
z: 0
.end
