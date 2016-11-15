package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"os"
	"testing"
)

func readTestFileMap(path string) (testFiles map[string]string, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		err = err
		return
	}

	testFiles = make(map[string]string)
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		v := bytes.Split(line, []byte{' '})
		if len(v) != 3 {
			err = fmt.Errorf("Unexpected format in test file line: '%s'", line)
			return
		}
		testFiles[string(v[2])] = string(v[0])
	}
	return
}

func TestBlockHash(t *testing.T) {
	testFiles, err := readTestFileMap("testdata/images/exact-hashes.txt")
	if err != nil {
		t.Fatal(err)
	}

	for file, expectedHash := range testFiles {
		t.Run(file, func(t *testing.T) {
			f, _ := os.Open("testdata/images/" + file)
			defer f.Close()

			i, _, _ := image.Decode(f)
			result := Blockhash(i, 16)

			if result != Hash(expectedHash) {
				distance := result.Distance(Hash(expectedHash))
				t.Errorf("Hash did not match (distance=%d)\nExpected: %s\nResult:   %s\n", distance, expectedHash, result)
			}
		})
	}
}

func BenchmarkBlockHash(b *testing.B) {
	testFiles, err := readTestFileMap("testdata/images/exact-hashes.txt")
	if err != nil {
		b.Fatal(err)
	}

	for file, _ := range testFiles {
		b.Run(file, func(b *testing.B) {
			f, _ := os.Open("testdata/images/" + file)
			defer f.Close()
			image, _, _ := image.Decode(f)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Blockhash(image, 16)
			}
		})
	}
}
