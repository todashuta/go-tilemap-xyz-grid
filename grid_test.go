package grid

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// update を付けると testdata/ の golden を現在の出力で上書きする。
//
// 比較をバイト一致ではなくピクセル一致にしているのは、PNG の deflate 出力が
// Go のバージョンごとに変わるため。以前はバイト一致を見ていて、描画が完全に
// 同じでもテストが落ちていた。「絵が変わった」のか「圧縮が変わった」のかを
// 区別したいので、バイト列そのものを正にはしない。
var update = flag.Bool("update", false, "regenerate golden files under testdata/")

// 負の z/x/y は XYZ タイルの仕様外なので意図的にテストしていない。New は現在
// 範囲を見ておらず、画像自体は生成できてしまう。ここで挙動を golden に固定
// すると、将来バリデーションを追加したときにすぐ壊れる。
func TestNew(t *testing.T) {
	tests := []struct {
		name string
		z    int
		x    int
		y    int
	}{
		{name: "two digits", z: 42, x: 44, y: 46},
		{name: "top left origin", z: 1, x: 0, y: 0},
		{name: "four digits", z: 100, x: 999, y: 1023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustNew(t, tt.z, tt.x, tt.y)
			want := golden(t, tt.z, tt.x, tt.y, got)
			assertSameImage(t, want, got)
		})
	}
}

// 枠は「任意のタイル画像の上に重ねたとき視認できること」が要件なので、golden
// に依存せず色そのものを確かめる。
func TestNewBorder(t *testing.T) {
	const size = 256
	black := color.RGBA64{0, 0, 0, 0xffff}
	white := color.RGBA64{0xffff, 0xffff, 0xffff, 0xffff}

	img := decode(t, mustNew(t, 8, 100, 150))

	for i := 0; i < size; i++ {
		for _, p := range []struct{ x, y int }{{i, 0}, {i, size - 1}, {0, i}, {size - 1, i}} {
			if got := img.At(p.x, p.y); !sameColor(got, black) {
				t.Errorf("outer pixel (%d, %d) = %v, want black", p.x, p.y, got)
			}
		}
		// 角の 2px は黒枠の垂直描画が y=0..255 を塗りつぶすため黒のまま。
		// 「角は黒で締まる」のを仕様とみなして i が端のときは白を確認しない。
		if i == 0 || i == size-1 {
			continue
		}
		for _, p := range []struct{ x, y int }{{i, 1}, {i, size - 2}, {1, i}, {size - 2, i}} {
			if got := img.At(p.x, p.y); !sameColor(got, white) {
				t.Errorf("inner pixel (%d, %d) = %v, want white", p.x, p.y, got)
			}
		}
	}
}

// 「タイル番号が読めること」の最低限の保証。枠の内側に黒い画素が実際に描かれて
// いることだけを見る。座標ごとの正確な字形は golden のほうが見てくれるので、
// ここではフォントの実装に依存する検証はしない。
func TestNewLabel(t *testing.T) {
	const size = 256
	black := color.RGBA64{0, 0, 0, 0xffff}

	img := decode(t, mustNew(t, 12, 3, 1))

	// 枠の黒を数えないよう、外周 2px を除いた内側だけを走査する。
	found := 0
	for y := 2; y < size-2; y++ {
		for x := 2; x < size-2; x++ {
			if sameColor(img.At(x, y), black) {
				found++
			}
		}
	}
	if found == 0 {
		t.Error("no label was drawn inside the border")
	}
}

func TestNewIsPNG(t *testing.T) {
	got := mustNew(t, 3, 5, 2)

	if !bytes.HasPrefix(got, []byte("\x89PNG\r\n\x1a\n")) {
		t.Error("output does not start with the PNG signature")
	}
	img := decode(t, got)
	if want := image.Rect(0, 0, 256, 256); !img.Bounds().Eq(want) {
		t.Errorf("bounds = %v, want %v", img.Bounds(), want)
	}
}

func mustNew(t *testing.T, z, x, y int) []byte {
	t.Helper()
	got, err := New(z, x, y)
	if err != nil {
		t.Fatalf("New(%d, %d, %d) returned error: %v", z, x, y, err)
	}
	return got
}

func golden(t *testing.T, z, x, y int, got []byte) []byte {
	t.Helper()
	path := filepath.Join("testdata", fmt.Sprintf("z%d-x%d-y%d.png", z, x, y))
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run `go test -update` to create it)", path, err)
	}
	return want
}

// png.Decode が返す image.NRGBA の At は color.NRGBA を返す。color.Color は
// インターフェースなので == では比較できず、color.Equal は非推奨になっていて
// 標準ライブラリに代替がない。RGBA() で 16bit に正規化して自前で比べる。
func sameColor(a, b color.Color) bool {
	return toRGBA64(a) == toRGBA64(b)
}

func toRGBA64(c color.Color) color.RGBA64 {
	r, g, b, a := c.RGBA()
	return color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)}
}

func decode(t *testing.T, b []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("output is not a decodable PNG: %v", err)
	}
	return img
}

func assertSameImage(t *testing.T, want, got []byte) {
	t.Helper()

	w := decode(t, want)
	g := decode(t, got)
	if !g.Bounds().Eq(w.Bounds()) {
		t.Fatalf("bounds = %v, want %v", g.Bounds(), w.Bounds())
	}

	const maxReported = 5
	diff := 0
	for y := w.Bounds().Min.Y; y < w.Bounds().Max.Y; y++ {
		for x := w.Bounds().Min.X; x < w.Bounds().Max.X; x++ {
			if !sameColor(g.At(x, y), w.At(x, y)) {
				if diff < maxReported {
					t.Errorf("pixel (%d, %d) = %v, want %v", x, y, g.At(x, y), w.At(x, y))
				}
				diff++
			}
		}
	}
	if diff > maxReported {
		t.Errorf("... and %d more differing pixels", diff-maxReported)
	}
	if diff > 0 {
		t.Errorf("%d pixels differ in total", diff)
	}
}
