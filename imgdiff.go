package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"math"
	"os"

	"image/color"
	"image/png"

	_ "image/jpeg"
	_ "image/png"
)

var (
	black   = color.RGBA{A: 255}
	magenta = color.RGBA{255, 0, 255, 255}
)

func main() {
	min := flag.Uint("min", 1, "minimum difference for each RGBA component")
	mask := flag.Bool("mask", true, "mask difference with magenta on black")
	outName := flag.String("o", "diff.png", "output filename")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("usage: imgdiff [flags] file1 file2")
		os.Exit(1)
	}
	if *min > math.MaxUint8 {
		log.Fatalf("-min > %d", math.MaxUint8)
	}
	*min |= *min << 8 // match [0, 0xffff] RGBA distribution

	img1, img2, err := open(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}

	d, err := difference(img1, img2, uint32(*min), *mask)
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.OpenFile(*outName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()
	if err = png.Encode(output, d); err != nil {
		log.Fatal(err)
	}
}

func open(in1, in2 string) (img1, img2 image.Image, err error) {
	file1, err := os.Open(in1)
	if err != nil {
		return
	}
	defer file1.Close()

	file2, err := os.Open(in2)
	if err != nil {
		return
	}
	defer file2.Close()

	img1, _, err = image.Decode(file1)
	if err != nil {
		return
	}

	img2, _, err = image.Decode(file2)
	if err != nil {
		return
	}
	return
}

func difference(img1, img2 image.Image, min uint32, mask bool) (image.Image, error) {
	if img1.Bounds() != img2.Bounds() {
		return nil, fmt.Errorf("bounds %v != %v", img1.Bounds(), img2.Bounds())
	}
	bounds := img1.Bounds()
	out := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if colordiff(img1.At(x, y), img2.At(x, y), min) {
				if mask {
					out.SetRGBA(x, y, magenta)
				} else {
					out.Set(x, y, img2.At(x, y))
				}
			} else if mask {
				out.SetRGBA(x, y, black)
			}
		}
	}
	return out, nil
}

func colordiff(c1, c2 color.Color, min uint32) bool {
	r1, g1, b1, a1 := color.RGBAModel.Convert(c1).RGBA()
	r2, g2, b2, a2 := color.RGBAModel.Convert(c2).RGBA()
	return udiff(r1, r2) > min || udiff(g1, g2) > min || udiff(b1, b2) > min || udiff(a1, a2) > min
}

func udiff(x, y uint32) uint32 {
	if x > y {
		return x - y
	}
	return y - x
}
