package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/str20tbl/revel"
)

// TranscriptionResult represents the API response structure
type TranscriptionResult struct {
	Success bool              `json:"success"`
	Results map[string]string `json:"results"`
}

// Transcribe calls the Python API to get ASR transcriptions from both Whisper and Wav2Vec2 models.
// It takes an audio filename and returns a map with "whisper" and "wav2vec2" transcription results.
// The transcriptions are cleaned (trimmed and periods removed) before being returned.
func Transcribe(filename string) map[string]string {
	revel.AppLog.Infof("Transcribing file: %s", filename)

	// Call Python API for transcription
	apiURL := fmt.Sprintf("http://api:8000/transcribe?filename=%s", filename)
	resp, err := http.Get(apiURL)
	if err != nil {
		revel.AppLog.Errorf("Error calling transcription API: %v", err)
		return map[string]string{"whisper": "", "wav2vec2": ""}
	}
	// Note: Body.Close() error intentionally ignored (standard practice)
	// Body is fully consumed before close, errors are rare and non-critical
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		revel.AppLog.Errorf("Error reading transcription response: %v", err)
		return map[string]string{"whisper": "", "wav2vec2": ""}
	}

	revel.AppLog.Infof("Transcription API response: %s", string(bodyBytes))

	// Parse JSON response
	var result TranscriptionResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		revel.AppLog.Errorf("Error unmarshalling transcription response: %v", err)
		return map[string]string{"whisper": "", "wav2vec2": ""}
	}

	// Clean up transcriptions (trim whitespace and remove periods)
	whisper := cleanTranscription(result.Results["whisper"])
	wav2vec2 := cleanTranscription(result.Results["wav2vec2"])

	return map[string]string{
		"whisper":  whisper,
		"wav2vec2": wav2vec2,
	}
}

// cleanTranscription normalizes transcription text by:
//   - Trimming whitespace
//   - Replacing hyphens with spaces (e.g., "well-known" → "well known")
//   - Removing all punctuation (.,!?;:"')
//   - Normalizing whitespace (collapse multiple spaces)
//
// This ensures consistent text format for metric calculations, particularly important
// since Whisper produces punctuation while Wav2Vec2 does not.
func cleanTranscription(text string) string {
	text = strings.TrimSpace(text)

	// Replace hyphens with spaces to split hyphenated words
	text = strings.ReplaceAll(text, "-", " ")

	// Remove all punctuation
	text = strings.Map(func(r rune) rune {
		if strings.ContainsRune(".,!?;:\"'", r) {
			return -1 // remove character
		}
		return r
	}, text)

	// Normalize whitespace: collapse multiple spaces
	text = strings.Join(strings.Fields(text), " ")

	return text
}
