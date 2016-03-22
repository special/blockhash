package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/color"
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
	R uint8
	G uint8
	B uint8
	A uint8
}

func (col *color) RGBA() (r, g, b, a uint8) {
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
	h := Hash(s)
	return h
}

func (img *Image) totalValue(x, y int) int {
	pixel := img.Pixels[y*img.Width+x]
	r, g, b, a := pixel.RGBA()
	//fmt.Println(r, g, b)

	if a == 0 {
		return 765
	}

	return int(r) + int(g) + int(b)

}

func median(blocks []int) int {
	sort.Ints(blocks)
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
		mblocks := make([]int, ((i+1)*bandsize)-(i*bandsize))
		copy(mblocks, blocks[i*bandsize:(i+1)*bandsize])
		m := median(mblocks)

		for j := i * bandsize; j < (i+1)*bandsize; j++ {

			v := blocks[j]
			res := (v > m || ((abs(v-m) < 1) && (m > half_block_value)))

			if res == true {
				blocks[j] = 1
			} else {
				blocks[j] = 0
			}

			//fmt.Println("J:", j, "V:", v, "M:", m, "res:", blocks[j])
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

	res := blocksToBits(result, blocksize_x*blocksize_y)
	return bitsToHexhash(res)
}

func Blockhash(i image.Image, bits int) Hash {
	img := unpackImage(i)

	even_x := img.Width%bits == 0
	even_y := img.Height%bits == 0

	if even_x && even_y {
		//return img.blockhashEven(bits)
	}

	blocks := make([][]int, bits)
	for x := 0; x < bits; x++ {
		for y := 0; y < bits; y++ {
			blocks[x] = make([]int, bits)
		}
	}

	block_width := img.Width / bits
	block_height := img.Height / bits
	var block_top, block_bottom, block_left, block_right int
	var weight_top, weight_bottom, weight_left, weight_right int

	for y := 0; y < img.Height; y++ {
		if even_y {
			block_top = y / block_height
			block_bottom = y / block_height
			weight_top, weight_bottom = 1, 0
		} else {
			y_frac := (y + 1) % block_height
			y_int := (y + 1) % block_height

			weight_top = (1 - y_frac)
			weight_bottom = y_frac

			if y_int > 0 || (y+1) == img.Height {
				block_top = y / block_height
				block_bottom = y / block_height
			} else {
				block_top = y / block_height
				block_bottom = -(-y / block_height)
			}
		}

		for x := 0; x < img.Width; x++ {
			value := img.totalValue(x, y)

			if even_x {
				block_left = x / block_width
				block_right = x / block_width
				weight_left, weight_right = 1, 0
			} else {
				x_frac := (x + 1) % block_width
				x_int := (x + 1) % block_width

				weight_left = (1 - x_frac)
				weight_right = x_frac

				if x_int > 0 || (x+1) == img.Width {
					block_left = x / block_width
					block_right = x / block_width
				} else {
					block_left = x / block_width
					block_right = -(-x / block_width)
				}
			}

			blocks[block_top][block_left] += value * weight_top * weight_left
			blocks[block_top][block_right] += value * weight_top * weight_right
			blocks[block_bottom][block_left] += value * weight_bottom * weight_left
			blocks[block_bottom][block_right] += value * weight_bottom * weight_right
		}

	}

	var result []int
	for x := 0; x < bits; x++ {
		for y := 0; y < bits; y++ {
			result = append(result, blocks[x][y])
		}
	}

	res := blocksToBits(result, block_width*block_height)

	return bitsToHexhash(res)
}

func createPixelArray(img image.Image) []color {
	bounds := img.Bounds()
	min := bounds.Min
	max := bounds.Max
	pixels := make([]color, (max.X-min.X)*(max.Y-min.Y))
	for x := min.X; x < max.X; x++ {
		for y := min.Y; y < max.Y; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixel := color{
				R: uint8(((r * 0xffff) / a) >> 8),
				G: uint8(((g * 0xffff) / a) >> 8),
				B: uint8(((b * 0xffff) / a) >> 8),
				A: uint8(a),
			}
			pixels[x+(y*(max.X-min.X))] = pixel
		}
	}
	return pixels
}

func unpackImage(img image.Image) Image {

	bounds := img.Bounds()
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

	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	i, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	bits := 16
	hash := Blockhash(i, bits)
	fmt.Println(hash)

	//path2 := os.Args[2]
	//img2 := openImage(path2)
	//hash2 := img2.Blockhash(bits)
	//fmt.Println(hash2)

	//fmt.Println(hash.HammingDistance(hash2))
}
