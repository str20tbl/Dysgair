package services

import (
	"app/app/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/str20tbl/revel"
)

// WordKey identifies a unique word-user combination
type WordKey struct {
	UserID int64
	WordID int64
}

// FilterToCompleteWords filters entries to only include words with exactly 5 recordings
func FilterToCompleteWords(entries []models.Entry) []models.Entry {
	wordGroups := make(map[WordKey][]models.Entry)
	for _, entry := range entries {
		key := WordKey{UserID: entry.UserID, WordID: entry.WordID}
		wordGroups[key] = append(wordGroups[key], entry)
	}

	var completeEntries []models.Entry
	for _, group := range wordGroups {
		if len(group) >= 5 {
			// Sort by ID (chronological order)
			sortedGroup := sortEntriesByID(group)
			completeEntries = append(completeEntries, sortedGroup[:5]...)
		}
	}

	return completeEntries
}

// EnrichWithMetrics adds per-word aggregated metrics to entries
func EnrichWithMetrics(entries []models.Entry) []models.EnrichedEntry {
	wordGroups := make(map[WordKey][]models.Entry)
	for _, entry := range entries {
		key := WordKey{UserID: entry.UserID, WordID: entry.WordID}
		wordGroups[key] = append(wordGroups[key], entry)
	}

	var enrichedEntries []models.EnrichedEntry
	for _, group := range wordGroups {
		metrics := calculateWordMetrics(group)

		for i, entry := range group {
			enrichedEntries = append(enrichedEntries, models.EnrichedEntry{
				Entry:         entry,
				AttemptNumber: i + 1,

				// Strict metrics
				AvgWERWhisper:           metrics.AvgWERWhisper,
				AvgCERWhisper:           metrics.AvgCERWhisper,
				AvgWERWav2Vec2:          metrics.AvgWERWav2Vec2,
				AvgCERWav2Vec2:          metrics.AvgCERWav2Vec2,
				ImprovementWERWhisper:   metrics.ImprovementWERWhisper,
				ImprovementWERWav2Vec2:  metrics.ImprovementWERWav2Vec2,
				BestWERWhisper:          metrics.BestWERWhisper,
				BestWERWav2Vec2:         metrics.BestWERWav2Vec2,
				FirstAttemptWERWhisper:  group[0].WERWhisper,
				FirstAttemptWERWav2Vec2: group[0].WERWav2Vec2,
				LastAttemptWERWhisper:   group[len(group)-1].WERWhisper,
				LastAttemptWERWav2Vec2:  group[len(group)-1].WERWav2Vec2,

				// Lenient metrics
				AvgWERWhisperLenient:           metrics.AvgWERWhisperLenient,
				AvgCERWhisperLenient:           metrics.AvgCERWhisperLenient,
				AvgWERWav2Vec2Lenient:          metrics.AvgWERWav2Vec2Lenient,
				AvgCERWav2Vec2Lenient:          metrics.AvgCERWav2Vec2Lenient,
				ImprovementWERWhisperLenient:   metrics.ImprovementWERWhisperLenient,
				ImprovementWERWav2Vec2Lenient:  metrics.ImprovementWERWav2Vec2Lenient,
				BestWERWhisperLenient:          metrics.BestWERWhisperLenient,
				BestWERWav2Vec2Lenient:         metrics.BestWERWav2Vec2Lenient,
				FirstAttemptWERWhisperLenient:  group[0].WERWhisperLenient,
				FirstAttemptWERWav2Vec2Lenient: group[0].WERWav2Vec2Lenient,
				LastAttemptWERWhisperLenient:   group[len(group)-1].WERWhisperLenient,
				LastAttemptWERWav2Vec2Lenient:  group[len(group)-1].WERWav2Vec2Lenient,

				WordCompletionCount: len(group),
			})
		}
	}

	return enrichedEntries
}

// CallPythonAnalysis makes HTTP request to Python API for analysis
// Returns the analysis data or error. The analysis data is returned as map[string]interface{}
// since different endpoints return different structures - conversion to typed structs happens in caller.
func CallPythonAnalysis(analysisType string, entriesJSON []byte) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("http://api:8000/analysis/%s", analysisType)
	requestBody := fmt.Sprintf(`{"entries":%s}`, string(entriesJSON))

	resp, err := http.Post(apiURL, "application/json", strings.NewReader(requestBody))
	if err != nil {
		revel.AppLog.Errorf("Error calling Python API for %s: %v", analysisType, err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	// Note: Body.Close() error intentionally ignored (standard practice)
	// Body is fully consumed before close, errors are rare and non-critical
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		revel.AppLog.Errorf("Error reading response for %s: %v", analysisType, err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		revel.AppLog.Errorf("Error unmarshalling response for %s: %v", analysisType, err)
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	// Extract analysis field from response
	if analysis, ok := result["analysis"].(map[string]interface{}); ok {
		return analysis, nil
	}

	// If no "analysis" field, return the whole result (might be an error response)
	return result, nil
}

// Helper functions

func sortEntriesByID(entries []models.Entry) []models.Entry {
	sorted := make([]models.Entry, len(entries))
	copy(sorted, entries)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].ID > sorted[j].ID {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func calculateWordMetrics(group []models.Entry) models.WordMetrics {
	// Collect strict (raw) metrics
	var whisperWERs, whisperCERs, wav2vec2WERs, wav2vec2CERs []float64
	// Collect lenient (normalized) metrics
	var whisperWERsLenient, whisperCERsLenient, wav2vec2WERsLenient, wav2vec2CERsLenient []float64

	for _, entry := range group {
		// Strict metrics
		whisperWERs = append(whisperWERs, entry.WERWhisper)
		whisperCERs = append(whisperCERs, entry.CERWhisper)
		wav2vec2WERs = append(wav2vec2WERs, entry.WERWav2Vec2)
		wav2vec2CERs = append(wav2vec2CERs, entry.CERWav2Vec2)

		// Lenient metrics
		whisperWERsLenient = append(whisperWERsLenient, entry.WERWhisperLenient)
		whisperCERsLenient = append(whisperCERsLenient, entry.CERWhisperLenient)
		wav2vec2WERsLenient = append(wav2vec2WERsLenient, entry.WERWav2Vec2Lenient)
		wav2vec2CERsLenient = append(wav2vec2CERsLenient, entry.CERWav2Vec2Lenient)
	}

	metrics := models.WordMetrics{
		// Strict metrics
		AvgWERWhisper:   Average(whisperWERs),
		AvgCERWhisper:   Average(whisperCERs),
		AvgWERWav2Vec2:  Average(wav2vec2WERs),
		AvgCERWav2Vec2:  Average(wav2vec2CERs),
		BestWERWhisper:  Minimum(whisperWERs),
		BestWERWav2Vec2: Minimum(wav2vec2WERs),

		// Lenient metrics
		AvgWERWhisperLenient:   Average(whisperWERsLenient),
		AvgCERWhisperLenient:   Average(whisperCERsLenient),
		AvgWERWav2Vec2Lenient:  Average(wav2vec2WERsLenient),
		AvgCERWav2Vec2Lenient:  Average(wav2vec2CERsLenient),
		BestWERWhisperLenient:  Minimum(whisperWERsLenient),
		BestWERWav2Vec2Lenient: Minimum(wav2vec2WERsLenient),
	}

	if len(group) == 5 {
		// Strict improvements
		metrics.ImprovementWERWhisper = group[0].WERWhisper - group[4].WERWhisper
		metrics.ImprovementWERWav2Vec2 = group[0].WERWav2Vec2 - group[4].WERWav2Vec2

		// Lenient improvements
		metrics.ImprovementWERWhisperLenient = group[0].WERWhisperLenient - group[4].WERWhisperLenient
		metrics.ImprovementWERWav2Vec2Lenient = group[0].WERWav2Vec2Lenient - group[4].WERWav2Vec2Lenient
	}

	return metrics
}

// Average calculates the average of a slice of float64
func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Minimum finds the minimum value in a slice of float64
func Minimum(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	minVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}
