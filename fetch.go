package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func fetchAudio(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch %v: %v", url, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch with status %s", res.Status)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %v", err)
	}
	return b, nil
}

func fetchTranscript(audio []byte, googleAPIKey string) (string, error) {
	speechURL := fmt.Sprintf("https://speech.googleapis.com/v1/speech:recognize?key=%s", googleAPIKey)

	req := map[string]interface{}{
		"config": map[string]interface{}{
			"encoding":        "LINEAR16",
			"sampleRateHertz": 8000,
			"languageCode":    "en-US",
		},
		"audio": map[string]interface{}{
			"content": base64.StdEncoding.EncodeToString(audio),
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("could not encode speech request: %v", err)
	}
	res, err := http.Post(speechURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("could not transcribe: %v", err)
	}
	defer res.Body.Close()

	var data struct {
		Error struct {
			Code    int
			Message string
			Status  string
		}
		Results []struct {
			Alternatives []struct {
				Transcript string
				Confidence float64
			}
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("could not decode speech response: %v", err)
	}
	if data.Error.Code != 0 {
		return "", fmt.Errorf("speech API error: %d %s %s",
			data.Error.Code, data.Error.Status, data.Error.Message)
	}
	if len(data.Results) == 0 || len(data.Results[0].Alternatives) == 0 {
		return "", fmt.Errorf("no transcriptions found")
	}
	text := data.Results[0].Alternatives[0].Transcript
	text = strings.ToLower(text)
	return text, nil
}
