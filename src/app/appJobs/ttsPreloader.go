package appJobs

import (
	"app/app/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/str20tbl/revel"
)

// TTSRequest for Piper (includes model parameter)
type TTSRequest struct {
	Text      string `json:"text"`
	SpeakerID string `json:"speaker_id"`
	Model     string `json:"model"`
}

// TTSBasicRequest for Orpheus and SpeechT5 (no model parameter)
type TTSBasicRequest struct {
	Text      string `json:"text"`
	SpeakerID string `json:"speaker_id"`
}

type TTSPreloader struct {
	Dbm *gorp.DbMap
}

func (t TTSPreloader) Run() {
	revel.AppLog.Info("TTS Preloader job started")

	// Ensure audio directory exists
	audioDir := "/data/audio"
	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		//revel.AppLog.Infof("Creating audio directory: %s", audioDir)
		if err := os.MkdirAll(audioDir, 0755); err != nil {
			revel.AppLog.Errorf("Failed to create audio directory: %v", err)
			return
		}
	}

	// Get TTS configuration from app.conf
	modelType, found := revel.Config.String("tts.model.type")
	if !found {
		modelType = "piper" // Default to Piper (fast)
	}

	language, found := revel.Config.String("tts.language")
	if !found {
		language = "cy" // Default to Welsh-only
	}

	voice, found := revel.Config.String("tts.voice")
	if !found {
		voice = "gwryw-gogledd-pro" // Default to male North Wales voice
	}

	//revel.AppLog.Infof("TTS Config - Model: %s, Language: %s, Voice: %s", modelType, language, voice)

	// Get all words from database
	var words []models.Word
	_, err := t.Dbm.Select(&words, "SELECT * FROM Word")
	if err != nil {
		revel.AppLog.Errorf("Failed to load words from database: %v", err)
		return
	}

	//revel.AppLog.Infof("Pre-loading TTS audio for %d words...", len(words))

	// Process each word
	successCount := 0
	failCount := 0
	for i, word := range words {
		err := fetchAudio(word.Text, word.AudioFilename, voice, language, modelType)
		if err != nil {
			revel.AppLog.Errorf("Failed to fetch audio for word '%s': %v", word.Text, err)
			failCount++
		} else {
			successCount++
		}

		// Log progress every 100 words
		if (i+1)%100 == 0 {
			revel.AppLog.Infof("Progress: %d/%d words processed", i+1, len(words))
		}
	}

	revel.AppLog.Infof("TTS Preloader completed - Success: %d, Failed: %d", successCount, failCount)
}

func fetchAudio(text, filename, voice, language, modelType string) error {
	// Determine the API endpoint based on model type
	var endpoint string
	switch modelType {
	case "orpheus":
		endpoint = "https://tts.techiaith.cymru/api/orpheus"
	case "speecht5":
		endpoint = "https://tts.techiaith.cymru/api/speecht5"
	case "piper":
		fallthrough
	default:
		endpoint = "https://tts.techiaith.cymru/api/piper"
	}

	// Create the appropriate request body based on model type
	var jsonData []byte
	var err error

	if modelType == "piper" {
		// Piper requires the model parameter
		requestBody := TTSRequest{
			Text:      text,
			SpeakerID: voice,
			Model:     language,
		}
		jsonData, err = json.Marshal(requestBody)
		//revel.AppLog.Debugf("TTS Request (Piper) - Text: '%s', Speaker: %s, Model: %s", text, voice, language)
	} else {
		// Orpheus and SpeechT5 only accept text and speaker_id
		requestBody := TTSBasicRequest{
			Text:      text,
			SpeakerID: voice,
		}
		jsonData, err = json.Marshal(requestBody)
		//revel.AppLog.Debugf("TTS Request (%s) - Text: '%s', Speaker: %s", modelType, text, voice)
	}

	if err != nil {
		return fmt.Errorf("could not marshal TTS request: %w", err)
	}

	// Make POST request to TTS API
	resp, err := http.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not fetch audio from API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Read error response body for details
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		if errorMsg == "" {
			errorMsg = "(no error details provided)"
		}
		return fmt.Errorf("TTS API returned status %d for '%s' using %s model: %s",
			resp.StatusCode, text, modelType, errorMsg)
	}

	// Validate Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "audio/wav" && contentType != "audio/x-wav" {
		revel.AppLog.Warnf("Unexpected Content-Type '%s' for word '%s' (expected audio/wav)", contentType, text)
	}

	// Create output file
	outputPath := fmt.Sprintf("/data/audio/%s.wav", filename)
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create audio file at %s: %w", outputPath, err)
	}
	defer out.Close()

	// Copy response to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("could not write audio file: %w", err)
	}

	//revel.AppLog.Debugf("Successfully saved audio for '%s' (%d bytes) to %s", text, bytesWritten, outputPath)
	return nil
}
