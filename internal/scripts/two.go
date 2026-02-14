package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	TargetCityCode = "118973"
	API_User       = "hackathontest"
	API_Pass       = "Hac@98147521"
	URL_Codes      = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/TBOHotelCodeList"
	URL_Details    = "http://api.tbotechnology.in/TBOHolidays_HotelAPI/Hoteldetails"
)

func main() {
	startTime := time.Now()
	fmt.Printf("🚀 Starting Scrape for City: %s\n", TargetCityCode)

	// 1. Get Hotel Codes
	codes, err := getHotelCodes(TargetCityCode)
	if err != nil {
		log.Fatalf("❌ Error fetching codes: %v", err)
	}
	total := len(codes)
	fmt.Printf("✅ Found %d hotels. Fetching details...\n", total)

	var allData []interface{}
	successCount := 0

	for i, code := range codes {
		fmt.Printf("\r⏳ Progress: %d/%d | Success: %d", i+1, total, successCount)

		// Use a map to handle the dynamic response
		data, err := fetchRawDetails(code)
		if err == nil {
			allData = append(allData, data)
			successCount++
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 2. Save everything to file
	fileName := fmt.Sprintf("Final_Hotels_City_%s.json", TargetCityCode)
	saveToFile(fileName, allData)

	fmt.Printf("\n\n✨ DONE! Saved %d hotels to %s\n", successCount, fileName)
	fmt.Printf("⏱️ Total Time: %v\n", time.Since(startTime))
}

func fetchRawDetails(code string) (interface{}, error) {
	// API is sensitive to exact field names - using 'Hotelcodes' as confirmed in diagnostic
	payload := map[string]string{"Hotelcodes": code, "Language": "en"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", URL_Details, bytes.NewBuffer(body))
	req.SetBasicAuth(API_User, API_Pass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Check if "Status" says 200
	if status, ok := result["Status"].(map[string]interface{}); ok {
		if codeVal, ok := status["Code"].(float64); ok && codeVal == 200 {
			// Return the "HotelDetails" part of the response
			return result["HotelDetails"], nil
		}
	}

	return nil, fmt.Errorf("api error")
}

func getHotelCodes(cityCode string) ([]string, error) {
	payload := map[string]string{"CityCode": cityCode, "IsDetailedResponse": "false"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", URL_Codes, bytes.NewBuffer(body))
	req.SetBasicAuth(API_User, API_Pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	var codes []string
	if hotels, ok := res["Hotels"].([]interface{}); ok {
		for _, h := range hotels {
			if hotelMap, ok := h.(map[string]interface{}); ok {
				if c, ok := hotelMap["Code"].(string); ok {
					codes = append(codes, c)
				}
			}
		}
	}
	return codes, nil
}

func saveToFile(name string, data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	_ = ioutil.WriteFile(name, b, 0644)
}
