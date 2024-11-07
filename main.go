package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	serverURL                 = "http://srv.msk01.gigacorp.local/_stats"
	loadAverageThreshold      = 30.0
	memoryUsageThreshold      = 0.8
	diskUsageThreshold        = 0.9
	networkBandwidthThreshold = 0.9
	checkInterval             = 10 * time.Second
	errorThreshold            = 3
)

var errorCount int

func fetchServerStats() ([]string, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		errorCount++
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorCount++
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	data, err := reader.Read()
	if err != nil {
		errorCount++
		return nil, err
	}

	if len(data) != 6 {
		errorCount++
		return nil, fmt.Errorf("unexpected data length: %d", len(data))
	}

	return data, nil
}

func checkServerStats(stats []string) {
	loadAverage, _ := strconv.ParseFloat(stats[0], 64)
	totalMemory, _ := strconv.Atoi(stats[1])
	usedMemory, _ := strconv.Atoi(stats[2])
	totalDiskSpace, _ := strconv.Atoi(stats[3])
	usedDiskSpace, _ := strconv.Atoi(stats[4])
	networkBandwidth, _ := strconv.Atoi(stats[5])

	// Проверка Load Average
	if loadAverage > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %.2fn", loadAverage)
	}

	// Проверка использования памяти
	memoryUsagePercentage := float64(usedMemory) / float64(totalMemory) * 100
	if memoryUsagePercentage > memoryUsageThreshold*100 {
		fmt.Printf("Memory usage too high: %.2f%%n", memoryUsagePercentage)
	}

	// Проверка использования дискового пространства
	freeDiskSpaceMB := float64(totalDiskSpace-usedDiskSpace) / (1024 * 1024)
	if freeDiskSpaceMB < (float64(totalDiskSpace)*(1-diskUsageThreshold))/(1024*1024) {
		fmt.Printf("Free disk space is too low: %.2f Mb leftn", freeDiskSpaceMB)
	}

	// Проверка сетевой загрузки
	networkUsagePercentage := float64(usedMemory) / float64(networkBandwidth) * 100
	freeNetworkBandwidthMbit := float64(networkBandwidth-usedMemory) / (1024 * 1024) * 8
	if networkUsagePercentage > networkBandwidthThreshold*100 {
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s availablen", freeNetworkBandwidthMbit)
	}
}

func main() {
	for {
		stats, err := fetchServerStats()
		if err != nil {
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistics.")
			}
			time.Sleep(checkInterval)
			continue
		}

		checkServerStats(stats)

		// Сброс счетчика ошибок при успешном запросе
		errorCount = 0

		time.Sleep(checkInterval)
	}
}
