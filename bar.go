package main

import "fmt"

type bar struct {
	state int
	max   int
	text  string
}

func newBar(max int, text string) *bar {
	return &bar{
		state: 0,
		max:   max,
		text:  text,
	}
}

func (b *bar) add(n int) {
	b.state += n
	if b.state >= b.max {
		fmt.Printf("\033[2K")
		fmt.Printf("\r" + b.text + " Done. ")
		fmt.Print("\n")
	} else {
		fmt.Printf("\r"+b.text+" %%"+"%3.1f ", float64(b.state)*100/float64(b.max))
	}
}
