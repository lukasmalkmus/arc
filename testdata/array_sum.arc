! ------------------------------------------------------- !
! This program sums the elements from array that is       !
! located starting with 3000.                             !
! ------------------------------------------------------- !
! Used registers                                          !
! ==============                                          !
! r1: length                                              !
! r2: start (3000)                                        !
! r3: sum of the elements (is initialized with zero)      !
! r4: the current element                                 !
! ==============                                          !
! r1, r2 and r4 are set back to 0 after the loop is done  !
! ------------------------------------------------------- !

        .begin
        .org 2048
        call init_r
        call loop

init_r: ld [length], %r1
        ld [start], %r2
        ld [zero], %r3
        jmpl %r15+4, %r0

loop:   ld %r2, %r4
        addcc %r2, 4, %r2
        addcc %r3, %r4, %r3
        addcc %r1, -1, %r1
        be done
        ba loop

done:   ld [zero], %r1
        ld [zero], %r2
        ld [zero], %r4
        jmpl %r15+4, %r0

start:  3000
length: 4
zero:   0

        .org 3000
        10
        20
        -0xa
        aH
        .end