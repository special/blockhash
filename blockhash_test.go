package main

import (
	"image"
	_ "image/jpeg"
	"os"
	"testing"
)

func BenchmarkBlockHash(t *testing.B) {
	f, _ := os.Open("test_data/1.jpg")
	defer f.Close()

	i, _, _ := image.Decode(f)
	_ = Blockhash(i, 16)
}
