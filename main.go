package main

import (
	"bytes"
	"fmt"
	//"github.com/rainycape/magick"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"sort"
)

type Hash string

type Image struct {
	i      image.Image
	Width  int
	Height int
	Pixels []color
}

type color struct {
	R int
	G int
	B int
	A int
}

func (col *color) RGBA() (r, g, b, a int) {
	return col.R, col.G, col.B, col.A
}

func (h1 Hash) HammingDistance(h2 Hash) int {
	diffs := 0
	for i, _ := range h1 {
		if h1[i] != h2[i] {
			diffs += 1
		}
	}
	return diffs
}

func bitsToHexhash(bits []int) Hash {
	var bu bytes.Buffer
	for _, v := range bits {
		bu.WriteString(fmt.Sprintf("%d", v))
	}
	fmt.Println(bu.String())

	var buf bytes.Buffer
	for i := 0; i < len(bits)/4; i++ {
		tmp := uint(0)
		for j := 0; j < 4; j++ {
			b := i*4 + j
			tmp = tmp | (uint(bits[b]) << uint(3) >> uint(j))
		}

		buf.WriteString(fmt.Sprintf("%1x", tmp))
	}

	s := buf.String()
	var h Hash
	h = Hash(s)
	return h
}

func (img *Image) totalValue(x, y int) int {
	pixel := img.Pixels[y*img.Width+x]
	r, g, b, a := pixel.RGBA()
	//fmt.Printf("(%d, %d, %d)\n", r, g, b)

	if a == 0 {
		return 765
	}

	return int(r) + int(g) + int(b)

}

func median(blocks []int) int {
	sort.Ints(blocks)
	fmt.Println(blocks)
	length := len(blocks)

	if length%2 == 0 {
		return (blocks[length/2] + blocks[length/(2+1)]) / 2
	}

	return blocks[length/2]
}

func abs(i int) int {
	if i < 0 {
		return -i
	}

	return i
}

func blocksToBits(blocks []int, pixels_per_block int) []int {
	half_block_value := pixels_per_block * 256 * 3 / 2
	bandsize := len(blocks) / 4

	for i := 0; i < 4; i++ {
		//fmt.Println(blocks[i*bandsize : (i+1)*bandsize])
		mblocks := make([]int, ((i+1)*bandsize)-(i*bandsize))
		copy(mblocks, blocks[i*bandsize:(i+1)*bandsize])
		m := median(mblocks)
		fmt.Println(m)

		for j := i * bandsize; j < (i+1)*bandsize; j++ {

			v := blocks[j]

			if v > m {
				blocks[j] = 1
			} else if abs(v-m) < 1 && m > half_block_value {
				blocks[j] = 0
			} else {
				blocks[j] = 0
			}

			fmt.Println("J:", j, "V:", v, "M:", m, "res:", blocks[j])
		}
	}
	return blocks
}

func (img *Image) blockhashEven(bits int) Hash {
	blocksize_x := img.Width / bits
	blocksize_y := img.Height / bits

	var result []int
	for y := 0; y < bits; y++ {
		for x := 0; x < bits; x++ {
			value := 0

			for iy := 0; iy < blocksize_y; iy++ {
				for ix := 0; ix < blocksize_x; ix++ {
					cx := x*blocksize_x + ix
					cy := y*blocksize_y + iy
					value += img.totalValue(cx, cy)
				}
			}
			result = append(result, value)
		}
	}
	fmt.Println(result)

	res := blocksToBits(result, blocksize_x*blocksize_y)
	return bitsToHexhash(res)
}

func (img *Image) Blockhash(bits int) Hash {

	even_x := img.Width%bits == 0
	even_y := img.Height%bits == 0

	if even_x && even_y {
		return img.blockhashEven(bits)
	} else {
		return "examplehash"
	}
}

func createPixelArray(img image.Image) []color {
	bounds := img.Bounds()
	min := bounds.Min
	max := bounds.Max
	pixels := make([]color, (max.X-min.X)*(max.Y-min.Y))
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixel := color{R: int(r / 256), G: int(g / 256), B: int(b / 256), A: int(a / 256)}
			pixels[x+(y*(max.X-min.X))] = pixel
		}
	}
	return pixels
}

func openImage(path string) Image {
	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	//pixels, _ := img.Pixels(magick.Rect{X: 0, Y: 0, Width: uint(img.Width()), Height: uint(img.Height())})
	pixels := createPixelArray(img)

	return Image{
		i:      img,
		Width:  bounds.Max.X - bounds.Min.X,
		Height: bounds.Max.Y - bounds.Min.Y,
		Pixels: pixels,
	}
}

func main() {
	path := os.Args[1]
	img := openImage(path)
	bits := 16
	hash := img.Blockhash(bits)
	fmt.Println(hash)
}
