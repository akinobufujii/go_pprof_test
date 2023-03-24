package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

func main() {
	root := ""
	flag.StringVar(&root, "root", os.ExpandEnv("$GOPATH"), "root")

	if err := calcFilehashSingle(root, "result_single.json"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := calcFilehashParallels(root, "result_parallels.json"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func calcFilehashSingle(root, filename string) error {
	hasher := md5.New()
	filehashMap := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		hasher.Reset()
		if err := calcFilehash(path, hasher); err != nil {
			return err
		}

		filehashMap[filepath.ToSlash(path)] = hex.EncodeToString(hasher.Sum(nil))

		return nil
	})

	if err != nil {
		return err
	}

	if err := writeJson(filename, filehashMap); err != nil {
		return err
	}

	return err
}

func calcFilehashParallels(root, filename string) error {
	eg, ctx := errgroup.WithContext(context.Background())
	parallels := runtime.NumCPU()

	chPath := make(chan string, parallels)
	eg.Go(func() error {
		defer close(chPath)
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			select {
			case chPath <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	})

	type calcResult struct {
		path     string
		hashData []byte
	}
	chCalcResult := make(chan calcResult, parallels)
	for i := 0; i < parallels; i++ {
		eg.Go(func() error {
			hasher := md5.New()
			for path := range chPath {
				hasher.Reset()
				if err := calcFilehash(path, hasher); err != nil {
					return err
				}

				calcResult := calcResult{
					path:     path,
					hashData: hasher.Sum(nil),
				}

				select {
				case chCalcResult <- calcResult:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		eg.Wait()
		close(chCalcResult)
	}()

	filehashMap := map[string]string{}
	for calcResult := range chCalcResult {
		filehashMap[filepath.ToSlash(calcResult.path)] = hex.EncodeToString(calcResult.hashData)
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := writeJson(filename, filehashMap); err != nil {
		return err
	}

	return nil
}

func calcFilehash(path string, hasher hash.Hash) error {
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

func writeJson(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", " ")

	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}
