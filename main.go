package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RequestTimout     = 2 * time.Second
	RequestsFrequency = 300 * time.Millisecond
	ErrorThreshold    = 3
	ServerURL         = "http://srv.msk01.gigacorp.local/_stats"
	LoadThreshold     = 30
	MemoryThreshold   = 80
	DiskThreshold     = 90
	NetworkThreshold  = 90
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
	poll := CreateServerPoller(ServerURL, RequestTimout, RequestsFrequency, ErrorThreshold)
	analyze := CreateStatsAnalyzer(LoadThreshold, MemoryThreshold, DiskThreshold, NetworkThreshold)

	for response := range poll() {
		stats := ParseStats(response)
		analyze(stats)
	}
}

func CreateStatsAnalyzer(loadThreshold, memoryThreshold, diskThreshold, networkThreshold int) func(serverStats ServerStats) {
	return func(stats ServerStats) {
		memoryUsagePercent := int(float64(stats.MemoryUsage) / float64(stats.MemoryCapacity) * 100)
		diskUsagePercent := int(float64(stats.DiskUsage) / float64(stats.DiskCapacity) * 100)
		networkUsagePercent := int(float64(stats.NetworkUsage) / float64(stats.NetworkCapacity) * 100)

		if stats.LoadAverage > loadThreshold {
			fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
		}
		if memoryUsagePercent > memoryThreshold {
			fmt.Printf("Memory usage too high: %d%%\n", memoryUsagePercent)
		}
		if diskUsagePercent > diskThreshold {
			availableSpace := (stats.DiskCapacity - stats.DiskUsage) / 1024 / 1024
			fmt.Printf("Free disk space is too low: %d Mb left\n", availableSpace)
		}
		if networkUsagePercent > networkThreshold {
			availableBandwidth := (stats.NetworkCapacity - stats.NetworkUsage) / 1000 / 1000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", availableBandwidth)
		}
	}
}

func ParseStats(rawStats []byte) ServerStats {
	stats := [7]int{}
	for index, value := range strings.Split(strings.Trim(string(rawStats), "\n"), ",") {
		number, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		stats[index] = number
	}
	return ServerStats{stats[0], stats[1], stats[2], stats[3], stats[4], stats[5], stats[6]}
}

func CreateServerPoller(url string, reqTimeout time.Duration, reqFreq time.Duration, errorThreshold int) func() chan []byte {
	return func() chan []byte {
		responsesChan := make(chan []byte, 3)
		client := http.Client{Timeout: reqTimeout}
		errorCounter := 0

		go func() {
			defer close(responsesChan)
			for {
				time.Sleep(reqFreq)
				if errorCounter >= errorThreshold {
					fmt.Printf("Unable to fetch server statistic")
					break
				}

				response, err := client.Get(url)
				if err != nil {
					errorCounter++
					fmt.Printf("failed to send request %s\n", err)
					continue
				}
				if response.StatusCode != http.StatusOK {
					errorCounter++
					continue
				}
				body, err := io.ReadAll(response.Body)
				if err != nil {
					errorCounter++
					fmt.Printf("failed to parse response %s\n", err)
					continue
				}
				_ = response.Body.Close()
				responsesChan <- body
			}
		}()
		return responsesChan
	}
}
