package volumes

import (
	"fmt"
	"math"
	"os"
	"strings"

	. "github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh/terminal"
)

var Cursor = "->"

type Volumes struct {
	vols []*Volume

	cursor int

	//更新
	update chan error
	//終了
	finish chan error
	end    chan error
}

type Volume struct {
	max   int
	name  string
	value float64
}

//新規作成
func New() *Volumes {
	v := Volumes{}
	v.vols = make([]*Volume, 0)
	v.update = make(chan error)
	v.finish = make(chan error)
	v.end = make(chan error)
	return &v
}

func (v *Volumes) Add(name string, vol int) {
	obj := NewVolume(name, vol)
	v.vols = append(v.vols, obj)
	return
}

func (v *Volumes) Set(val float64) {
	v.vols[v.cursor].Set(val)
	v.update <- nil
	return
}

func (v *Volumes) Start() {

	out := NewColorableStdout()
	//->[VOLUME][ -30.00][---------------=========|-------------------------]

	for {
		select {
		case <-v.update:

			nameNum := v.getNameMax()

			width, err := getTerminalWidth()
			if err != nil {
				//エラーの為終了
				v.end <- err
				return
			}

			clen := len(Cursor)
			f := "%" + fmt.Sprintf("%d", clen) + "s" + "[%" + fmt.Sprintf("%d", nameNum) + "s][%7.2f][%s]%s"
			remain := width - (2 + nameNum + 2 + 7 + 2 + 2)

			for idx, elm := range v.vols {
				line := elm.format(f, remain, idx == v.cursor)
				fmt.Fprintf(out, "%s\n", line)
			}

			fmt.Fprintf(out, "\033[%dA", len(v.vols))
		case <-v.end:

			fmt.Fprintln(out)
			return
		}
	}

	return
}

func (v *Volumes) SetCursor(idx int) {

	if idx < 0 || idx > len(v.vols)-1 {
		return
	}

	v.cursor = idx
	v.update <- nil
}

func (v *Volumes) GetCursor() int {
	return v.cursor
}

func (v *Volumes) Get() float64 {
	return v.vols[v.cursor].Get()
}

func (v *Volumes) Wait() error {
	return <-v.finish
}

func (v *Volumes) Finish(err error) {
	v.finish <- err
	v.end <- err
}

func (v *Volumes) getNameMax() int {
	nameNum := 0
	for _, elm := range v.vols {
		if len(elm.name) > nameNum {
			nameNum = len(elm.name)
		}
	}
	return nameNum
}

func NewVolume(name string, val int) *Volume {

	v := Volume{}

	v.max = val
	v.name = name

	v.value = 0.0

	return &v
}

func (v *Volume) Set(val float64) {
	v.value = val
}

func (v *Volume) Get() float64 {
	return v.value
}

func (v *Volume) format(f string, remain int, current bool) string {

	margin := ""
	barNum := remain / 2
	if (remain % 2) == 0 {
		barNum = barNum - 1
		margin = " "
	}

	absV := int(math.Abs(v.value) / float64(v.max) * float64(barNum))
	if absV > barNum {
		absV = barNum
	}

	val := strings.Repeat("=", absV)
	sp := strings.Repeat("-", barNum-absV)
	rev := strings.Repeat("-", barNum)

	bar := rev + "|" + val + sp
	if v.value < -1 {
		bar = sp + val + "|" + rev
	}

	cur := ""
	if current {
		cur = Cursor
	}

	return fmt.Sprintf(f, cur, v.name, v.value, bar, margin)
}

//端末の横幅を設定
func getTerminalWidth() (int, error) {
	termID := int(os.Stdout.Fd())
	width, _, err := terminal.GetSize(termID)
	if err != nil {
		return -1, err
	}
	return width, nil
}
