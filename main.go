package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ServerStats struct {
	LoadAverage     int
	MemoryCapacity  int
	MemoryUsage     int
	DiskCapacity    int
	DiskUsage       int
	NetworkCapacity int
	NetworkUsage    int
}

func main() {
	const (
		RequestTimout     = 3 * time.Second
		RequestsFrequency = 2 * time.Second
		ServerUrl         = "http://localhost:8080/_stats"
	)
	poller := serverPoller(ServerUrl, RequestTimout, RequestsFrequency)
	parser := statsParser(poller())
	analyzer := statsAnalyzer(parser())

	analyzer()
}

func statsAnalyzer(channel chan ServerStats) func() {
	analyze := func(stats ServerStats) {
		fmt.Printf("Load Average: %d\n", stats.LoadAverage)
		fmt.Printf("Memory Usage: %d/%d\n", stats.MemoryUsage, stats.MemoryCapacity)
		fmt.Printf("Disk Usage: %d/%d\n", stats.DiskUsage, stats.DiskCapacity)
		fmt.Printf("Network Usage: %d/%d\n", stats.NetworkUsage, stats.NetworkCapacity)
	}

	analyzer := func() {
		for stats := range channel {
			analyze(stats)
		}
	}
	return analyzer
}

func statsParser(channel chan []byte) func() chan ServerStats {
	parseStats := func(statsRaw []byte) ServerStats {
		stats := [7]int{}
		for index, value := range strings.Split(strings.Trim(string(statsRaw), "\n"), ",") {
			number, err := strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			stats[index] = number
		}
		return ServerStats{stats[0], stats[1], stats[2], stats[3], stats[4], stats[5], stats[6]}
	}

	parser := func() chan ServerStats {
		statsChan := make(chan ServerStats, 3)
		go func() {
			defer close(statsChan)
			for response := range channel {
				statsChan <- parseStats(response)
			}
		}()
		return statsChan
	}

	return parser
}

func serverPoller(url string, reqTimeout time.Duration, reqFreq time.Duration) func() chan []byte {
	poller := func() chan []byte {
		responsesChan := make(chan []byte, 3)
		client := http.Client{Timeout: reqTimeout}
		go func() {
			defer close(responsesChan)
			for {
				log.Printf("Sending request to %s...", url)
				response, err := client.Get(url)
				if err != nil {
					log.Printf("failed to send request %s\n", err)
				}
				body, err := io.ReadAll(response.Body)
				if err != nil {
					log.Printf("failed to parse response %s\n", err)
				}
				responsesChan <- body
				time.Sleep(reqFreq)
			}
		}()
		return responsesChan
	}
	return poller
}
