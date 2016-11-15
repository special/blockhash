package main

import (
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"sort"
)

type Hash string

func (h1 Hash) Distance(h2 Hash) int {
	h1raw, _ := hex.DecodeString(string(h1))
	h2raw, _ := hex.DecodeString(string(h2))

	diffs := 0
	for i, _ := range h1raw {
		// Naive popcount
		d := h1raw[i] ^ h2raw[i]
		for d > 0 {
			if d&1 == 1 {
				diffs++
			}
			d >>= 1
		}
	}
	return diffs
}

func bitsToHexhash(bits []int) Hash {
	re := make([]byte, len(bits)/8)

	for i := 0; i < len(re); i++ {
		for j := 0; j < 8; j++ {
			re[i] |= byte(bits[i*8+j]&1) << uint(7-j)
		}
	}

	return Hash(hex.EncodeToString(re))
}

func totalValue(img *image.NRGBA, x, y int) int {
	offset := img.PixOffset(x, y)

	if img.Pix[offset+3] == 0 {
		return 765
	}

	return int(img.Pix[offset]) +
		int(img.Pix[offset+1]) +
		int(img.Pix[offset+2])
}

func medianf(blocks []float64) float64 {
	// Copy blocks
	sorted := make([]float64, len(blocks))
	copy(sorted, blocks)
	sort.Float64s(sorted)
	length := len(sorted)

	if length%2 == 0 {
		return (sorted[length/2] + sorted[length/2+1]) / 2
	}

	return sorted[length/2]
}

func blocksToBitsf(blocks []float64, nblocks, pixels_per_block int) (result []int) {
	half_block_value := float64(pixels_per_block * 256 * 3 / 2)
	bandsize := nblocks / 4
	result = make([]int, nblocks)

	for i := 0; i < 4; i++ {
		m := medianf(blocks[i*bandsize : (i+1)*bandsize])
		for j := i * bandsize; j < (i+1)*bandsize; j++ {
			if (blocks[j] > m) || (math.Abs(blocks[j]-m) < 1 && m > half_block_value) {
				result[j] = 1
			} else {
				result[j] = 0
			}
		}
	}


	return
}

func Blockhash(i image.Image, bits int) Hash {
	img := unpackImage(i)
	blocks := make([]float64, bits*bits)
	bounds := img.Bounds()
	block_width := float64(bounds.Dx()) / float64(bits)
	block_height := float64(bounds.Dy()) / float64(bits)
	var block_top, block_bottom, block_left, block_right int
	var weight_top, weight_bottom, weight_left, weight_right float64

	for y := 0; y < bounds.Dy(); y++ {
		y_mod := math.Mod(float64(y+1), block_height)
		y_int, y_frac := math.Modf(y_mod)

		weight_top = (1 - y_frac)
		weight_bottom = y_frac

		// y_int will be 0 on bottom/right borders and on block boundaries
		if y_int > 0 || (y+1) == bounds.Dy() {
			block_top = int(math.Floor(float64(y) / block_height))
			block_bottom = block_top
		} else {
			block_top = int(math.Floor(float64(y) / block_height))
			block_bottom = int(math.Ceil(float64(y) / block_height))
		}

		for x := 0; x < bounds.Dx(); x++ {
			value := float64(totalValue(img, x, y))

			x_mod := math.Mod(float64(x+1), block_width)
			x_int, x_frac := math.Modf(x_mod)

			weight_left = (1 - x_frac)
			weight_right = x_frac

			if x_int > 0 || (x+1) == bounds.Dx() {
				block_left = int(math.Floor(float64(x) / block_width))
				block_right = block_left
			} else {
				block_left = int(math.Floor(float64(x) / block_width))
				block_right = int(math.Ceil(float64(x) / block_width))
			}


			blocks[block_top*bits+block_left] += value * weight_top * weight_left
			blocks[block_top*bits+block_right] += value * weight_top * weight_right
			blocks[block_bottom*bits+block_left] += value * weight_bottom * weight_left
			blocks[block_bottom*bits+block_right] += value * weight_bottom * weight_right
		}
	}

	// XXX is truncate correct? that's what C code is doing
	res := blocksToBitsf(blocks, bits*bits, int(block_width*block_height))
	return bitsToHexhash(res)
}

func unpackImage(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	re := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), bounds.Size()})

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			re.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}

	return re
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
