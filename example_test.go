package grid_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	grid "github.com/todashuta/go-tilemap-xyz-grid"
)

// 描画内容の回帰を見るために黒画素数を出力している。PNG のバイト列やその
// ハッシュを出力すると、Go のバージョンで deflate の出力が変わっただけで
// この Example が落ちるため。
//
// 黒画素数自体はフォントのレンダリングが変われば動く。それは検知したい
// 回帰なので、ここで構えている。
func ExampleNew() {
	pngBytes, err := grid.New(42, 44, 46)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(img.Bounds())
	fmt.Println(countBlack(img))

	// Output:
	// (0,0)-(256,256)
	// 1286
}

func countBlack(img image.Image) int {
	const size = 256

	n := 0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0xffff {
				n++
			}
		}
	}
	return n
}
