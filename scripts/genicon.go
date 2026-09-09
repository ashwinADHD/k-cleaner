//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	iconset := os.Args[1]
	sizes := []int{16, 32, 128, 256, 512}

	for _, size := range sizes {
		for _, scale := range []int{1, 2} {
			px := size * scale
			name := "icon_" + strconv.Itoa(size) + "x" + strconv.Itoa(size)
			if scale == 2 {
				name += "@2x"
			}
			name += ".png"
			if err := writeIcon(filepath.Join(iconset, name), px); err != nil {
				panic(err)
			}
		}
	}
}

func writeIcon(path string, size int) error {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{248, 250, 252, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	center := size / 2
	radius := size * 38 / 100
	teal := color.RGBA{16, 185, 129, 255} // K-Cleaner accent
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := x-center, y-center
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, teal)
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
