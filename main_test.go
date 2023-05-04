package main

import (
	"crypto/md5"
	"hash"
	"io"
	"os"
	"testing"
)

func BenchmarkCalcFileHashCopy(b *testing.B) {
	filepath := "README.md"
	hasher := md5.New()
	calcHash := func(path string, hasher hash.Hash) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hasher, file); err != nil {
			return err
		}

		return nil
	}

	for i := 0; i < b.N; i++ {
		hasher.Reset()
		if err := calcHash(filepath, hasher); err != nil {
			b.Fatal(err)
		}
		hasher.Sum(nil)
	}
}

func BenchmarkCalcFileHashCopyBuf(b *testing.B) {
	filepath := "README.md"
	hasher := md5.New()
	buf := make([]byte, 32*1024)
	calcHash := func(path string, hasher hash.Hash, buf []byte) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.CopyBuffer(hasher, file, buf); err != nil {
			return err
		}

		return nil
	}

	for i := 0; i < b.N; i++ {
		hasher.Reset()
		if err := calcHash(filepath, hasher, buf); err != nil {
			b.Fatal(err)
		}
		hasher.Sum(nil)
	}
}
