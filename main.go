package main

import (
	"fmt"
	"os"
)

func main() {

	vols := make([]*Volume, 3)

	vols[0] = NewVolume("Volume", 200)
	vols[1] = NewVolume("Light", 300)
	vols[2] = NewVolume("Wait", 100)

	vols[0].Set(100.0)
	vols[1].Set(-100.0)
	vols[2].Set(200.0)

	w := NewMonitor(vols)
	go w.Start()

	err := w.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(2)
	}

	fmt.Println("")
}
