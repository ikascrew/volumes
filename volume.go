package main

import (
	"fmt"
	"math"
	"strings"
)

const CURSOR = "->"

type Volume struct {
	value int
	name  string

	current float64
}

func NewVolume(name string, val int) *Volume {

	v := Volume{}

	v.value = val
	v.name = name

	v.current = 0.0

	return &v
}

func (v *Volume) Set(val float64) {
	v.current = val
}

func (v *Volume) format(f string, remain int, current bool) string {

	margin := ""
	barNum := remain / 2
	if (remain % 2) == 0 {
		barNum = barNum - 1
		margin = " "
	}

	absV := int(math.Abs(v.current) / float64(v.value) * float64(barNum))
	if absV > barNum {
		absV = barNum
	}

	val := strings.Repeat("=", absV)
	sp := strings.Repeat("-", barNum-absV)
	rev := strings.Repeat("-", barNum)

	bar := rev + "|" + val + sp
	if v.current < -1 {
		bar = sp + val + "|" + rev
	}

	cur := ""
	if current {
		cur = CURSOR
	}

	return fmt.Sprintf(f, cur, v.name, v.current, bar, margin)
}
