package main

import (
	"fmt"
	"os"

	vol "github.com/ikascrew/volumes"
	"github.com/nsf/termbox-go"
)

func main() {

	v := vol.New()

	v.Add("Volume", 300)
	v.Add("Light", 100)
	v.Add("Wait", 100)

	go v.Start()
	go controller(v)

	err := v.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(2)
	}

	fmt.Println("End")
}

func controller(v *vol.Volumes) error {

	//termboxの初期化
	err := termbox.Init()
	if err != nil {
		return err
	}
	//プログラム終了時termboxを閉じる
	defer termbox.Close()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc: //ESCキーで終了
				v.Finish(nil)
				break
			case termbox.KeyArrowUp:
				c := v.GetCursor() - 1
				if c < 0 {
					c = 0
				}
				v.SetCursor(c)
			case termbox.KeyArrowDown:
				c := v.GetCursor() + 1
				if c > 2 {
					c = 2
				}
				v.SetCursor(c)
			case termbox.KeyArrowLeft:
				v.Set(v.Get() - 1)
			case termbox.KeyArrowRight:
				v.Set(v.Get() + 1)
			case termbox.KeySpace:
			default: //その他のキー
			}
		default:
		}
	}

	return nil
}
