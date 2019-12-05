package main

import (
	"sort"
	"strconv"
	"strings"
)

func SingleHash(in, out chan string) {
	data := <-in
	result := DataSignerCrc32(data)
	result += "~"
	result += DataSignerCrc32(DataSignerMd5(data))
	out <- result
}

func MultiHash(in, out chan string) {
	data := <-in
	var result string
	for i := 0; i < 6; i++ {
		result += DataSignerCrc32(strconv.Itoa(i) + data)
	}
	out <- result
}

func CombineResults(in, out chan string) {
	var data []string
	for i := range in {
		data = append(data, i)
	}
	sort.Strings(data)
	result := strings.Join(data, "_")
	out <- result
}

func ExecutePipeline(jobs []job) {

}

func main() {
	// 	data := []int{0, 1}
	// 	start := time.Now()
	// 	ExecutePipeline(data)
	// 	end := time.Since(start)
	// 	fmt.Println("time: ", end)
}
