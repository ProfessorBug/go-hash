package main

import (
	"fmt"
	"flag"
	"runtime"
	"time"
	"io"
	"os"
	"encoding/hex"
	"crypto/sha256"
	"strconv"
)

var (
	reqs int
	size int
	cpu int
	max int
)

type Request struct {
	min int
	max int
}

func init() {
	flag.IntVar(&reqs, "b", 1000, "# bundles of requests")
	flag.IntVar(&size, "s", 10000, "Request size (split of min/max")
	flag.IntVar(&cpu, "cpu", runtime.NumCPU(), "CPU")
	flag.IntVar(&max, "w", 4, "Maximum number of workers")
}

type Response struct {
	respTime time.Duration
	hash     string
	nonce    int
}

func dispatcher(reqChan chan Request) {
	defer close(reqChan)
	min := 0
	max := size

	for i := 0; i < reqs; i++ {
		var r Request
		r.min = min
		r.max = max

		reqChan <- r

		min += 10000
		max += size
	}
}

func workerPool(reqChan chan Request, respChan chan Response) {
	for i := 0; i < max; i++ {
		go worker(reqChan, respChan)
	}
}

func worker(reqChan chan Request, respChan chan Response) {
	for req := range reqChan {
		for i := req.min; i < req.max; i++ {
			hash256 := sha256.New()
			io.WriteString(hash256, "SOME RANDOM STRING" + strconv.Itoa(i))

			hash256Str := hex.EncodeToString(hash256.Sum(nil))
			if hash256Str[0:5] == "00000" {
				fmt.Println("-----XXXXX-----")
				fmt.Println(hash256Str)

				startReq := time.Now()
				r := Response{time.Since(startReq), hash256Str, i}
				respChan <- r
			}
		}
	}
}

func consumer(respChan chan Response) bool {
	fmt.Println("TA DA:")
	fmt.Println("------")

	select {
	case r, ok := <-respChan:
		if ok {
			fmt.Println(r.nonce)
			fmt.Println("----------")
			os.Exit(0)
		}
	}
	return true
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(cpu)
	reqChan := make(chan Request, 100000000)
	respChan := make(chan Response, 1)

	go dispatcher(reqChan)
	go workerPool(reqChan, respChan)
	consumer(respChan)
}

