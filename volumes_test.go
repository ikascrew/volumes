package volumes

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

// TestNew は New() が空の Volumes を返すことを確認する。
func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() should not return nil")
	}
	if len(v.vols) != 0 {
		t.Fatalf("New() vols should be empty, got %d", len(v.vols))
	}
	if v.cursor != 0 {
		t.Fatalf("New() cursor should be 0, got %d", v.cursor)
	}
}

// TestAddAndGet は Add で追加した Volume に Set/Get できることを確認する。
func TestAddAndGet(t *testing.T) {
	v := New()
	v.Add("vol1", 10)
	v.Add("vol2", 20)

	if len(v.vols) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(v.vols))
	}

	// cursor はデフォルトで0番目を指す
	v.Set(5.0)
	if got := v.Get(); got != 5.0 {
		t.Fatalf("Get() = %v, want 5.0", got)
	}

	// 2番目には影響しない
	v.SetCursor(1)
	if got := v.Get(); got != 0.0 {
		t.Fatalf("Get() for vol2 = %v, want 0.0 (untouched)", got)
	}
}

// TestSetCursor はカーソル移動が範囲内でのみ有効なことを確認する。
func TestSetCursor(t *testing.T) {
	v := New()
	v.Add("vol1", 10)
	v.Add("vol2", 10)
	v.Add("vol3", 10)

	cases := []struct {
		name    string
		idx     int
		wantIdx int
	}{
		{"範囲内へ移動", 2, 2},
		{"負値は無視", -1, 2},
		{"範囲外(上)は無視", 3, 2},
		{"先頭へ移動", 0, 0},
		{"範囲外(大きく超過)は無視", 100, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v.SetCursor(c.idx)
			if got := v.GetCursor(); got != c.wantIdx {
				t.Fatalf("SetCursor(%d): GetCursor() = %d, want %d", c.idx, got, c.wantIdx)
			}
		})
	}
}

// TestSetValue はカーソル位置に関係なく idx のボリュームへ値を設定できること、
// 範囲外は無視されることを確認する。
func TestSetValue(t *testing.T) {
	v := New()
	v.Add("vol1", 10)
	v.Add("vol2", 10)

	// カーソルは0のまま、idx=1に値を設定
	v.SetValue(1, 7.5)
	v.SetCursor(1)
	if got := v.Get(); got != 7.5 {
		t.Fatalf("SetValue(1, 7.5): Get() = %v, want 7.5", got)
	}

	// 範囲外は無視される
	v.SetValue(-1, 99.0)
	v.SetValue(5, 99.0)
	if got := v.Get(); got != 7.5 {
		t.Fatalf("out-of-range SetValue should be ignored, Get() = %v, want 7.5", got)
	}
}

// TestInfo はスナップショットが内部状態を正しく写し、
// 返り値を書き換えても内部に影響しないことを確認する。
func TestInfo(t *testing.T) {
	v := New()
	v.Add("vol1", 10)
	v.Add("vol2", 20)
	v.SetValue(0, 3.0)
	v.SetValue(1, 4.0)

	info := v.Info()
	if len(info) != 2 {
		t.Fatalf("Info() length = %d, want 2", len(info))
	}

	want := []Info{
		{Name: "vol1", Value: 3.0, Max: 10},
		{Name: "vol2", Value: 4.0, Max: 20},
	}
	for i, w := range want {
		if info[i] != w {
			t.Fatalf("Info()[%d] = %+v, want %+v", i, info[i], w)
		}
	}

	// 返り値を書き換えても内部に影響しないこと
	info[0].Name = "changed"
	info[0].Value = 999.0
	info[0].Max = 999

	info2 := v.Info()
	if info2[0] != want[0] {
		t.Fatalf("mutating returned Info affected internal state: %+v", info2[0])
	}
}

// TestNotifyNonBlocking は Start() (モニタgoroutine) を起動していない状態で
// Set/SetCursor/SetValue を呼んでもデッドロックしないことを確認する。
// client リポジトリが依存する重要仕様。
func TestNotifyNonBlocking(t *testing.T) {
	v := New()
	v.Add("vol1", 10)
	v.Add("vol2", 10)

	done := make(chan struct{})
	go func() {
		defer close(done)
		v.Set(1.0)
		v.SetCursor(1)
		v.SetValue(0, 2.0)
		v.Set(3.0)
	}()

	select {
	case <-done:
		// ok: non-blocking
	case <-time.After(2 * time.Second):
		t.Fatal("notify blocked: Set/SetCursor/SetValue did not return without a running monitor")
	}
}

// TestNewVolume は NewVolume の初期状態(value=0, max/nameが保持される)を確認する。
func TestNewVolume(t *testing.T) {
	cases := []struct {
		name string
		max  int
	}{
		{"vol1", 10},
		{"", 0},
		{"negative-max", -5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vol := NewVolume(c.name, c.max)
			if vol == nil {
				t.Fatal("NewVolume() should not return nil")
			}
			if got := vol.Get(); got != 0.0 {
				t.Fatalf("NewVolume(%q, %d).Get() = %v, want 0.0", c.name, c.max, got)
			}
			if vol.name != c.name {
				t.Fatalf("NewVolume(%q, %d).name = %q, want %q", c.name, c.max, vol.name, c.name)
			}
			if vol.max != c.max {
				t.Fatalf("NewVolume(%q, %d).max = %d, want %d", c.name, c.max, vol.max, c.max)
			}
		})
	}
}

// TestVolumeSetGet は Set/Get が任意の値(境界値含む)をそのまま往復させることを確認する。
func TestVolumeSetGet(t *testing.T) {
	cases := []float64{0, 1, -1, 0.5, -0.5, 1e9, -1e9, math.MaxFloat64, -math.MaxFloat64}

	for _, val := range cases {
		vol := NewVolume("v", 10)
		vol.Set(val)
		if got := vol.Get(); got != val {
			t.Fatalf("Set(%v) then Get() = %v, want %v", val, got, val)
		}
	}
}

// TestGetNameMax は getNameMax が最長の name の文字数を返し、
// vols が空の場合は 0 を返すことを確認する。
func TestGetNameMax(t *testing.T) {
	t.Run("空のVolumesは0", func(t *testing.T) {
		v := New()
		if got := v.getNameMax(); got != 0 {
			t.Fatalf("getNameMax() on empty Volumes = %d, want 0", got)
		}
	})

	t.Run("最長のnameの長さを返す", func(t *testing.T) {
		v := New()
		v.Add("a", 10)
		v.Add("bbbbb", 10)
		v.Add("cc", 10)
		if got := v.getNameMax(); got != 5 {
			t.Fatalf("getNameMax() = %d, want 5", got)
		}
	})

	t.Run("同数の名前でも最大値を返す", func(t *testing.T) {
		v := New()
		v.Add("aaa", 10)
		v.Add("bbb", 10)
		if got := v.getNameMax(); got != 3 {
			t.Fatalf("getNameMax() = %d, want 3", got)
		}
	})
}

// TestSetCursorOnEmptyVolumes は要素が0個のときに SetCursor が
// どんな idx を渡してもカーソルを動かさない(範囲外として無視される)ことを確認する。
func TestSetCursorOnEmptyVolumes(t *testing.T) {
	v := New()

	cases := []int{-1, 0, 1, 100}
	for _, idx := range cases {
		v.SetCursor(idx)
		if got := v.GetCursor(); got != 0 {
			t.Fatalf("SetCursor(%d) on empty Volumes: GetCursor() = %d, want 0", idx, got)
		}
	}
}

// TestSetValueOnEmptyVolumes は要素が0個のときに SetValue が
// パニックせず何もしないことを確認する。
func TestSetValueOnEmptyVolumes(t *testing.T) {
	v := New()

	done := make(chan struct{})
	go func() {
		defer close(done)
		v.SetValue(0, 1.0)
		v.SetValue(-1, 1.0)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("SetValue on empty Volumes blocked or panicked")
	}
}

// TestAddCursorDefault は Add で要素を追加してもカーソル位置(デフォルト0)は
// 変化しないことを確認する。
func TestAddCursorDefault(t *testing.T) {
	v := New()
	if got := v.GetCursor(); got != 0 {
		t.Fatalf("GetCursor() before Add = %d, want 0", got)
	}

	v.Add("vol1", 10)
	v.Add("vol2", 20)
	v.Add("vol3", 30)

	if got := v.GetCursor(); got != 0 {
		t.Fatalf("GetCursor() after Add = %d, want 0 (unchanged)", got)
	}
}

// TestWaitFinish は Finish が finish/end 双方のチャンネルへ送信し、
// Wait がその値を受け取れることを確認する。
func TestWaitFinish(t *testing.T) {
	v := New()

	wantErr := fmt.Errorf("boom")

	// Finish は finish と end の両方へ送るため、end 側も受信しておかないと
	// goroutine がブロックしたままになる。
	endDone := make(chan struct{})
	go func() {
		defer close(endDone)
		<-v.end
	}()

	finishDone := make(chan error, 1)
	go func() {
		finishDone <- v.Wait()
	}()

	go v.Finish(wantErr)

	select {
	case got := <-finishDone:
		if got != wantErr {
			t.Fatalf("Wait() = %v, want %v", got, wantErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() did not return in time")
	}

	select {
	case <-endDone:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Finish() did not signal end channel in time")
	}
}

// TestVolumeFormat は Volume.format のバー文字列生成ロジックを検証する。
// name は "vol1" (4文字) 固定にして f の %4s 幅に一致させている。
func TestVolumeFormat(t *testing.T) {
	const f = "%2s[%4s][%7.2f][%s]%s"

	cases := []struct {
		name       string
		value      float64
		max        int
		remain     int
		current    bool
		wantBar    string
		wantMargin string
	}{
		{
			name:       "正値・remain奇数",
			value:      5.0,
			max:        10,
			remain:     11,
			wantBar:    "-----|==---",
			wantMargin: "",
		},
		{
			name:       "正値・remain偶数(marginあり)",
			value:      5.0,
			max:        10,
			remain:     10,
			wantBar:    "----|==--",
			wantMargin: " ",
		},
		{
			name:       "負値だが-1以上は正値と同じ表現",
			value:      -0.5,
			max:        10,
			remain:     11,
			wantBar:    "-----|-----",
			wantMargin: "",
		},
		{
			name:       "0値",
			value:      0.0,
			max:        10,
			remain:     11,
			wantBar:    "-----|-----",
			wantMargin: "",
		},
		{
			name:       "-1未満は反転表現",
			value:      -5.0,
			max:        10,
			remain:     11,
			wantBar:    "---==|-----",
			wantMargin: "",
		},
		{
			name:       "max超過はbarNumにクランプ",
			value:      20.0,
			max:        10,
			remain:     11,
			wantBar:    "-----|=====",
			wantMargin: "",
		},
		{
			name:       "カーソル位置はCursorを表示",
			value:      5.0,
			max:        10,
			remain:     11,
			current:    true,
			wantBar:    "-----|==---",
			wantMargin: "",
		},
		{
			name:       "値がちょうど-1は反転しない境界値",
			value:      -1.0,
			max:        10,
			remain:     11,
			wantBar:    "-----|-----",
			wantMargin: "",
		},
		{
			name:       "remainが最小(1)でもパニックしない",
			value:      5.0,
			max:        10,
			remain:     1,
			wantBar:    "|",
			wantMargin: "",
		},
		{
			name:       "負のmax超過(絶対値がbarNumを超える)もbarNumにクランプ",
			value:      -50.0,
			max:        10,
			remain:     11,
			wantBar:    "=====|-----",
			wantMargin: "",
		},
		{
			name:       "remainが0(偶数でbarNumが負)でもパニックしない",
			value:      5.0,
			max:        10,
			remain:     0,
			wantBar:    "|",
			wantMargin: " ",
		},
		{
			name:       "remainが負でもパニックしない",
			value:      5.0,
			max:        10,
			remain:     -3,
			wantBar:    "|",
			wantMargin: "",
		},
		{
			name:       "maxが0でもNaNにならずバー0扱い",
			value:      5.0,
			max:        0,
			remain:     11,
			wantBar:    "-----|-----",
			wantMargin: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vol := NewVolume("vol1", c.max)
			vol.Set(c.value)

			got := vol.format(f, c.remain, c.current)

			cur := ""
			if c.current {
				cur = Cursor
			}
			want := fmt.Sprintf(f, cur, "vol1", c.value, c.wantBar, c.wantMargin)

			if got != want {
				t.Fatalf("format() = %q, want %q", got, want)
			}
			if c.current && !strings.Contains(got, Cursor) {
				t.Fatalf("format() with current=true should contain cursor %q, got %q", Cursor, got)
			}
		})
	}
}
