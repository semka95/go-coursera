package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SingleHash evaluates hash of given data
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for data := range in {
		str := fmt.Sprintf("%v", data)
		wg.Add(1)
		go func(data string, out chan<- interface{}, wg *sync.WaitGroup) {
			first := crc(data)
			second := crc(md(data, mu))

			out <- <-first + "~" + <-second

			wg.Done()
		}(str, out, wg)
	}

	wg.Wait()
}

// MultiHash evaluates hash of given data
func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)

		go func(data string, out chan<- interface{}, wg *sync.WaitGroup) {
			var resultCh []chan string
			for i := 0; i < 6; i++ {
				resultCh = append(resultCh, crc(fmt.Sprintf("%v", i)+data))
			}

			var result string
			for _, resultCh := range resultCh {
				result += <-resultCh
			}

			out <- result

			wg.Done()
		}(data.(string), out, wg)
	}

	wg.Wait()
}

// CombineResults combines given data
func CombineResults(in, out chan interface{}) {
	var data []string

	for v := range in {
		data = append(data, v.(string))
	}

	sort.Strings(data)
	result := strings.Join(data, "_")

	fmt.Println("result: ", result)
	out <- result
}

// ExecutePipeline provides pipeline processing for worker functions
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})

	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		wg.Add(1)
		out := make(chan interface{})

		go func(j job, in, out chan interface{}, wg *sync.WaitGroup) {
			j(in, out)
			close(out)
			wg.Done()
		}(j, in, out, wg)

		in = out
	}

	wg.Wait()
}

func crc(data string) chan string {
	out := make(chan string)

	go func(data string, out chan<- string) {
		out <- DataSignerCrc32(data)
	}(data, out)

	return out
}

func md(data string, mu *sync.Mutex) string {
	mu.Lock()
	out := DataSignerMd5(data)
	mu.Unlock()

	return out
}

func main() {
	inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}

	start := time.Now()
	ExecutePipeline(hashSignJobs...)
	end := time.Since(start)

	fmt.Println("time: ", end)
}
