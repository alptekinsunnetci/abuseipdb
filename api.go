package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const abuseCheckBlockURL = "https://api.abuseipdb.com/api/v2/check-block"

type AbuseCheckBlockResponse struct {
	Data struct {
		NetworkAddress   string `json:"networkAddress"`
		Netmask          string `json:"netmask"`
		MinAddress       string `json:"minAddress"`
		MaxAddress       string `json:"maxAddress"`
		NumPossibleHosts int    `json:"numPossibleHosts"`
		AddressSpaceDesc string `json:"addressSpaceDesc"`
		ReportedAddress  []struct {
			IPAddress            string  `json:"ipAddress"`
			NumReports           int     `json:"numReports"`
			MostRecentReport     string  `json:"mostRecentReport"`
			AbuseConfidenceScore int     `json:"abuseConfidenceScore"`
			CountryCode          *string `json:"countryCode"`
		} `json:"reportedAddress"`
	} `json:"data"`
}

type APIClient struct {
	client *http.Client
	keys   []string
	config *Config
}

func NewAPIClient(keys []string, config *Config) *APIClient {
	return &APIClient{
		client: &http.Client{Timeout: config.GetRequestTimeout()},
		keys:   keys,
		config: config,
	}
}

func (c *APIClient) getRandomAPIKey() string {
	if len(c.keys) == 0 {
		return ""
	}
	return c.keys[rand.Intn(len(c.keys))]
}

func (c *APIClient) CheckBlock(network string) ([]ReportRow, error) {
	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		apiKey := c.getRandomAPIKey()
		if apiKey == "" {
			return nil, fmt.Errorf("API anahtarı bulunamadı")
		}

		req, err := http.NewRequest("GET", abuseCheckBlockURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("network", network)
		q.Add("maxAgeInDays", "7")
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Key", apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			if attempt == c.config.MaxRetries {
				return nil, fmt.Errorf("son deneme başarısız: %v", err)
			}
			log.Printf("Network %s deneme %d/%d başarısız: %v", network, attempt, c.config.MaxRetries, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 {
			if attempt == c.config.MaxRetries {
				return nil, fmt.Errorf("Network %s için %d deneme sonunda 401 hatası (yetkisiz erişim)", network, c.config.MaxRetries)
			}
			log.Printf("Network %s deneme %d/%d: 401 hatası, farklı anahtar deneniyor", network, attempt, c.config.MaxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		var parsed AbuseCheckBlockResponse
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&parsed); err != nil {
			return nil, err
		}

		// Son 7 gün içindeki raporları filtrele
		var rows []ReportRow
		for _, addr := range parsed.Data.ReportedAddress {
			if addr.MostRecentReport == "" {
				continue
			}

			t, err := time.Parse(time.RFC3339, addr.MostRecentReport)
			if err != nil {
				if t2, err2 := time.Parse("2006-01-02 15:04:05 MST", addr.MostRecentReport); err2 == nil {
					t = t2
				} else {
					continue
				}
			}

			if time.Since(t) > 7*24*time.Hour {
				continue
			}

			countryCode := ""
			if addr.CountryCode != nil {
				countryCode = *addr.CountryCode
			}

			row := ReportRow{
				IPAddress:            addr.IPAddress,
				CountryName:          countryCode,
				NumReports:           addr.NumReports,
				AbuseConfidenceScore: addr.AbuseConfidenceScore,
				LastReportedAt:       t.Format(time.RFC3339),
			}
			rows = append(rows, row)
		}

		return rows, nil
	}

	return nil, fmt.Errorf("Network %s için %d deneme sonunda başarısız", network, c.config.MaxRetries)
}
