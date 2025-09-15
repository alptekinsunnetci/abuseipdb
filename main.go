package main

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type WorkerPool struct {
	apiClient *APIClient
	config    *Config
}

func NewWorkerPool(apiClient *APIClient, config *Config) *WorkerPool {
	return &WorkerPool{
		apiClient: apiClient,
		config:    config,
	}
}

func (wp *WorkerPool) ProcessNetworks(networks []string) []ReportRow {
	networkCh := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	rows := make([]ReportRow, 0, 1000) 

	worker := func() {
		defer wg.Done()
		for network := range networkCh {
			networkRows, err := wp.apiClient.CheckBlock(network)
			if err != nil {
				log.Printf("Network %s sorgusunda hata: %v", network, err)
				time.Sleep(wp.config.GetRetryDelay())
				continue
			}
			if len(networkRows) > 0 {
				mu.Lock()
				rows = append(rows, networkRows...)
				mu.Unlock()
			}
			time.Sleep(wp.config.GetRetryDelay())
		}
	}

	wg.Add(wp.config.Concurrency)
	for i := 0; i < wp.config.Concurrency; i++ {
		go worker()
	}

	for _, network := range networks {
		networkCh <- network
	}
	close(networkCh)
	wg.Wait()

	return rows
}

func main() {
	rand.Seed(time.Now().UnixNano())

	config := LoadConfig()

	keys := config.APIKeys
	if len(keys) == 0 {
		log.Fatalf("config.yaml'da hiçbir API anahtarı tanımlanmamış")
	}

	networks := config.Prefixes
	if len(networks) == 0 {
		log.Fatalf("config.yaml'da hiçbir network tanımlanmamış")
	}

	uniqueNetworks := make([]string, 0, len(networks))
	seen := make(map[string]bool)
	for _, network := range networks {
		network = strings.TrimSpace(network) 
		if !seen[network] {
			seen[network] = true
			uniqueNetworks = append(uniqueNetworks, network)
		}
	}
	networks = uniqueNetworks

	log.Printf("Sorgulanacak network sayısı: %d", len(networks))

	apiClient := NewAPIClient(keys, config)

	workerPool := NewWorkerPool(apiClient, config)
	rows := workerPool.ProcessNetworks(networks)

	for i := 0; i < len(rows)-1; i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[i].AbuseConfidenceScore < rows[j].AbuseConfidenceScore {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}

	outputPath := config.GetOutputFileName()
	if err := RenderReport(rows, outputPath); err != nil {
		log.Fatalf("Rapor yazılamadı: %v", err)
	}

	log.Printf("Rapor oluşturuldu: %s (%d rapor bulundu)", outputPath, len(rows))
}
