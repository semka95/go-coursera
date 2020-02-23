package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func SingleHash(in, out chan interface{}) {
	data := <-in
	str := fmt.Sprintf("%v", data)
	result := DataSignerCrc32(str)
	result += "~"
	result += DataSignerCrc32(DataSignerMd5(str))
	out <- result
}

func MultiHash(in, out chan interface{}) {
	data := <-in
	str := fmt.Sprintf("%v", data)
	var result string
	for i := 0; i < 6; i++ {
		result += DataSignerCrc32(strconv.Itoa(i) + str)
	}
	out <- result
}

func CombineResults(in, out chan interface{}) {
	var data []string
	for i := range in {
		data = append(data, fmt.Sprintf("%v", i))
	}
	sort.Strings(data)
	result := strings.Join(data, "_")
	out <- result
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)
}

func main() {
	// 	data := []int{0, 1}
	// 	start := time.Now()
	// 	ExecutePipeline(data)
	// 	end := time.Since(start)
	// 	fmt.Println("time: ", end)
	inputData := []int{0, 1}
	result := "123"

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("error")
			}
			result = data
		}),
	}

	ExecutePipeline(hashSignJobs...)
	fmt.Println(result)
}
