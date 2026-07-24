package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type result struct {
	path string
	hash string
}

func hashKarunga(path string, ans chan<- result) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("can't open:", path)
		return
	}
	defer file.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		ans <- result{path: path, hash: ""}
		return
	}

	ans <- result{
		path: path,
		hash: hex.EncodeToString(hasher.Sum(nil)),
	}
}

func worker(job <-chan string, ans chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range job {
		hashKarunga(path, ans)
	}
}

func walking(rootPath string, job chan<- string) {
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		job <- path
		return nil
	})

	if err != nil {
		fmt.Println("Walk error:", err)
	}

	close(job)
}

func main() {
	rootPath := "C:/Users/aecr/Desktop/project"

	noOfWorker := 200

	job := make(chan string, 1000)
	ans := make(chan result, 1000)

	var wg sync.WaitGroup

	fmt.Println("Start")
	start := time.Now()

	// Producer
	go walking(rootPath, job)

	// Workers
	wg.Add(noOfWorker)
	for i := 0; i < noOfWorker; i++ {
		go worker(job, ans, &wg)
	}

	// Close result channel when all workers finish
	go func() {
		wg.Wait()
		close(ans)
	}()

	// Collector
	res := make(map[string]int)

	for r := range ans {
		res[r.hash]++
	}

	// Print statistics
	unique := len(res)

	for hash, count := range res {
		fmt.Println(hash, count)
	}

	fmt.Println("\nUnique hashes:", unique)
	fmt.Println("Time:", time.Since(start))
}
