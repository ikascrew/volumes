package main

import (
	"fmt"
	"os"
	"time"

	. "github.com/mattn/go-colorable"

	"golang.org/x/crypto/ssh/terminal"
)

//端末の横幅を設定
func getTerminalWidth() (int, error) {
	termID := int(os.Stdout.Fd())
	width, _, err := terminal.GetSize(termID)
	if err != nil {
		return -1, err
	}
	return width, nil
}

type Monitor struct {
	vols   []*Volume
	cursor int
	end    chan error
}

func NewMonitor(vols []*Volume) *Monitor {
	m := Monitor{}
	m.vols = vols
	m.cursor = 0
	m.end = make(chan error)
	return &m
}

func (m *Monitor) SetCursor(c int) {
	m.cursor = c

}

func (m *Monitor) Wait() error {
	return <-m.end
}

func (m *Monitor) Start() {

	out := NewColorableStdout()

	//->[VOLUME][ -30.00][---------------=========|-------------------------]
	//  [LIGHT ][  40.00][------------------------|============-------------]
	//  [WAIT  ][ -80.00][-----===================|-------------------------]

	max := 10
	ticker := time.Tick(time.Second)

	nameNum := 0
	for _, elm := range m.vols {
		if len(elm.name) > nameNum {
			nameNum = len(elm.name)
		}
	}
	width, err := getTerminalWidth()
	if err != nil {
		m.end <- err
		return
	}
	f := "%2s" + "[%" + fmt.Sprintf("%d", nameNum) + "s][%7.2f][%s]%s"
	remain := width - (2 + nameNum + 2 + 7 + 2 + 2)

	for i := 1; i <= max; i++ {

		m.SetCursor(i % 3)

		select {
		case <-ticker:
			for idx, elm := range m.vols {
				line := elm.format(f, remain, idx == m.cursor)
				fmt.Fprintf(out, "\033[2K\r%s", line)
			}
		}
	}

	m.end <- nil
	return
}
