package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Config holds the API details
const (
	APIURL      = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/search"
	Username    = "hackathontest"
	Password    = "Hac@98147521"
	TotalPax    = 102 // Target total people (approx 100+)
	MaxPaxBatch = 6   // API restriction per request
)

// Request Structures
type Room struct {
	Adults   int `json:"Adults"`
	Children int `json:"Children"`
}

type Filters struct {
	Refundable bool   `json:"Refundable"`
	NoOfRooms  int    `json:"NoOfRooms"`
	MealType   string `json:"MealType"`
}

type Payload struct {
	CheckIn            string  `json:"CheckIn"`
	CheckOut           string  `json:"CheckOut"`
	HotelCodes         string  `json:"HotelCodes"`
	GuestNationality   string  `json:"GuestNationality"`
	PaxRooms           []Room  `json:"PaxRooms"`
	ResponseTime       int     `json:"ResponseTime"`
	IsDetailedResponse bool    `json:"IsDetailedResponse"`
	Filters            Filters `json:"Filters"`
}

func ain() {
	// Calculate how many batches we need
	batchCount := (TotalPax + MaxPaxBatch - 1) / MaxPaxBatch

	fmt.Printf("🚀 Starting Bulk Fetch...\n")
	fmt.Printf("🎯 Target: %d Pax | 📦 Batches: %d | ⚡ Max per Batch: %d\n\n", TotalPax, batchCount, MaxPaxBatch)

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 1; i <= batchCount; i++ {
		wg.Add(1)
		go func(batchID int) {
			defer wg.Done()
			fetchAndSaveBatch(batchID)
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n✅ All batches completed in %v\n", time.Since(startTime))
}

func fetchAndSaveBatch(batchID int) {
	// 1. Construct the Payload
	requestBody := Payload{
		CheckIn:          "2026-10-01",
		CheckOut:         "2026-10-04",
		HotelCodes:       "1279415",
		GuestNationality: "AE",
		PaxRooms: []Room{
			{Adults: 3, Children: 0},
			{Adults: 3, Children: 0},
		},
		ResponseTime:       0,
		IsDetailedResponse: true,
		Filters: Filters{
			Refundable: false,
			NoOfRooms:  2,
			MealType:   "All",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[Batch %d] Error marshaling: %v", batchID, err)
		return
	}

	// 2. Create Request
	req, err := http.NewRequest("POST", APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Batch %d] Error creating request: %v", batchID, err)
		return
	}

	// 3. Set Headers & Auth
	req.SetBasicAuth(Username, Password)
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute Request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Batch %d] Request failed: %v", batchID, err)
		return
	}
	defer resp.Body.Close()

	// 5. Read Response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Batch %d] Error reading body: %v", batchID, err)
		return
	}

	// 6. Save to File
	fileName := fmt.Sprintf("batch_%d.json", batchID)
	err = ioutil.WriteFile(fileName, body, 0644)
	if err != nil {
		log.Printf("Error writing file %s: %v", fileName, err)
		return
	}

	// 7. Log Success
	fmt.Printf("✅ Saved data to %s (Size: %d bytes)\n", fileName, len(body))
}
