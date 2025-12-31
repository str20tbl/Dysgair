package services

import (
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"

	"app/app/models"

	"github.com/go-gorp/gorp"
	"github.com/str20tbl/revel"
)

// OrphanedRecording represents an audio file without a database entry
type OrphanedRecording struct {
	FilePath string
	FileName string
}

// FindAllRecordings scans /data/recordings and returns ALL audio files found
func FindAllRecordings() ([]OrphanedRecording, error) {
	// Scan filesystem for all audio files
	var recordings []OrphanedRecording
	recordingsDir := "/data/recordings"

	err := filepath.WalkDir(recordingsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			revel.AppLog.Warnf("FindAllRecordings: Error accessing path %s: %v", path, err)
			return nil // Continue walking despite errors
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if it's an audio file
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".webm" && ext != ".wav" && ext != ".mp3" && ext != ".ogg" && ext != ".m4a" {
			return nil
		}

		// Add ALL audio files
		recordings = append(recordings, OrphanedRecording{
			FilePath: path,
			FileName: filepath.Base(path),
		})

		return nil
	})

	if err != nil {
		revel.AppLog.Errorf("FindAllRecordings: Failed to walk directory %s: %v", recordingsDir, err)
		return nil, err
	}

	revel.AppLog.Infof("FindAllRecordings: Found %d total recordings", len(recordings))
	return recordings, nil
}

// FindOrphanedRecordingsForUser scans /data/recordings and returns only recordings
// that do NOT have an Entry record for the specified userID
func FindOrphanedRecordingsForUser(txn *gorp.Transaction, userID int64) ([]OrphanedRecording, error) {
	// Get all recordings from filesystem
	allRecordings, err := FindAllRecordings()
	if err != nil {
		return nil, err
	}

	// Get all recording paths that already have entries for this user
	var existingRecordings []string
	_, err = txn.Select(&existingRecordings,
		"SELECT Recording FROM Entry WHERE UserID = ?", userID)
	if err != nil {
		revel.AppLog.Errorf("FindOrphanedRecordingsForUser: Failed to query existing entries: %v", err)
		return nil, err
	}

	// Create a map for O(1) lookup
	existingMap := make(map[string]bool)
	for _, recording := range existingRecordings {
		existingMap[recording] = true
	}

	// Filter out recordings that already have entries for this user
	orphanedRecordings := make([]OrphanedRecording, 0)
	for _, recording := range allRecordings {
		if !existingMap[recording.FilePath] {
			orphanedRecordings = append(orphanedRecordings, recording)
		}
	}

	revel.AppLog.Infof("FindOrphanedRecordingsForUser: Found %d orphaned recordings for user %d (out of %d total)",
		len(orphanedRecordings), userID, len(allRecordings))

	return orphanedRecordings, nil
}

// WordMatch represents a word match with its confidence score
type WordMatch struct {
	WordID   int64
	Word     string
	English  string
	Distance int     // Levenshtein distance
	Score    float64 // Confidence score (0-1, higher is better)
}

// FindTopNWordMatches finds the top N closest matching words using Levenshtein distance
// Returns N best matches sorted by distance (closest first)
func FindTopNWordMatches(txn *gorp.Transaction, transcription string, n int) ([]*WordMatch, error) {
	// Normalize transcription for comparison
	normalized := strings.ToLower(strings.TrimSpace(transcription))

	if normalized == "" {
		return nil, nil // No transcription to match
	}

	// Get all words from Word table
	var words []models.Word
	_, err := txn.Select(&words, "SELECT ID, Text, English FROM Word")
	if err != nil {
		revel.AppLog.Errorf("FindTopNWordMatches: Failed to query words: %v", err)
		return nil, err
	}

	// Calculate Levenshtein distance to each word
	matches := make([]*WordMatch, 0, len(words))
	for _, word := range words {
		normalizedWord := strings.ToLower(strings.TrimSpace(word.Text))

		// Calculate distance using runes for proper Unicode support
		distance := levenshteinDistanceString(normalized, normalizedWord)

		// Calculate confidence score (1.0 = perfect match, decreases with distance)
		maxLen := max(len(normalized), len(normalizedWord))
		score := 0.0
		if maxLen > 0 {
			score = 1.0 - (float64(distance) / float64(maxLen))
		}

		matches = append(matches, &WordMatch{
			WordID:   word.ID,
			Word:     word.Text,
			English:  word.English,
			Distance: distance,
			Score:    score,
		})
	}

	// Sort by distance (lowest first)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Distance < matches[i].Distance {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Take top N
	if len(matches) > n {
		matches = matches[:n]
	}

	return matches, nil
}

// MatchWordByLevenshtein finds the closest matching word using Levenshtein distance
// Returns the best match based on comparing transcription to all words in Word table
func MatchWordByLevenshtein(txn *gorp.Transaction, transcription string) (*WordMatch, error) {
	// Normalize transcription for comparison
	normalized := strings.ToLower(strings.TrimSpace(transcription))

	if normalized == "" {
		return nil, nil // No transcription to match
	}

	// Get all words from Word table
	var words []models.Word
	_, err := txn.Select(&words, "SELECT ID, Text, English FROM Word")
	if err != nil {
		revel.AppLog.Errorf("MatchWordByLevenshtein: Failed to query words: %v", err)
		return nil, err
	}

	var bestMatch *WordMatch
	bestDistance := -1

	// Calculate Levenshtein distance to each word
	for _, word := range words {
		normalizedWord := strings.ToLower(strings.TrimSpace(word.Text))

		// Calculate distance using runes for proper Unicode support
		distance := levenshteinDistanceString(normalized, normalizedWord)

		// Track best match (shortest distance)
		if bestDistance == -1 || distance < bestDistance {
			bestDistance = distance

			// Calculate confidence score (1.0 = perfect match, decreases with distance)
			maxLen := max(len(normalized), len(normalizedWord))
			score := 0.0
			if maxLen > 0 {
				score = 1.0 - (float64(distance) / float64(maxLen))
			}

			bestMatch = &WordMatch{
				WordID:   word.ID,
				Word:     word.Text,
				English:  word.English,
				Distance: distance,
				Score:    score,
			}
		}
	}

	if bestMatch != nil {
		revel.AppLog.Infof("MatchWordByLevenshtein: Best match for '%s' → '%s' (distance=%d, score=%.2f)",
			transcription, bestMatch.Word, bestMatch.Distance, bestMatch.Score)
	}

	return bestMatch, nil
}

// levenshteinDistanceString calculates Levenshtein distance between two strings
// This is a simplified version for word matching (character-level)
func levenshteinDistanceString(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Create distance matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first column and row
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Calculate distances
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len1][len2]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RecoveryResult holds the results of recovering orphaned recordings
type RecoveryResult struct {
	OrphanedCount  int
	ProcessedCount int
	FailedCount    int
	CreatedEntries []int64  // Entry IDs created
	Failures       []string // Error messages for failed recoveries
}

// RecoverOrphanedRecordings processes all audio files and creates Entry records
// userID: The user to assign recovered entries to (typically admin or target user)
func RecoverOrphanedRecordings(txn *gorp.Transaction, userID int64) (*RecoveryResult, error) {
	result := &RecoveryResult{
		CreatedEntries: []int64{},
		Failures:       []string{},
	}

	// Find all recordings
	orphaned, err := FindAllRecordings()
	if err != nil {
		return nil, err
	}

	result.OrphanedCount = len(orphaned)
	revel.AppLog.Infof("RecoverOrphanedRecordings: Processing %d orphaned recordings", result.OrphanedCount)

	// Process each orphaned file
	for _, orphan := range orphaned {
		revel.AppLog.Infof("RecoverOrphanedRecordings: Processing %s", orphan.FilePath)

		// Transcribe the audio file
		transcriptions := Transcribe(orphan.FilePath)
		whisperText := transcriptions["whisper"]
		wav2vec2Text := transcriptions["wav2vec2"]

		if whisperText == "" && wav2vec2Text == "" {
			errMsg := "No transcription returned for " + orphan.FileName
			revel.AppLog.Warnf("RecoverOrphanedRecordings: %s", errMsg)
			result.Failures = append(result.Failures, errMsg)
			result.FailedCount++
			continue
		}

		// Use Whisper as primary, Wav2Vec2 as fallback
		primaryTranscription := whisperText
		if primaryTranscription == "" {
			primaryTranscription = wav2vec2Text
		}

		// Match to closest word
		match, err := MatchWordByLevenshtein(txn, primaryTranscription)
		if err != nil {
			errMsg := "Failed to match word for " + orphan.FileName + ": " + err.Error()
			revel.AppLog.Errorf("RecoverOrphanedRecordings: %s", errMsg)
			result.Failures = append(result.Failures, errMsg)
			result.FailedCount++
			continue
		}

		if match == nil {
			errMsg := "No word match found for " + orphan.FileName
			revel.AppLog.Warnf("RecoverOrphanedRecordings: %s", errMsg)
			result.Failures = append(result.Failures, errMsg)
			result.FailedCount++
			continue
		}

		// Create Entry record
		entry := models.Entry{
			UserID:             userID,
			WordID:             match.WordID,
			Text:               match.Word,
			English:            match.English,
			Recording:          orphan.FilePath,
			AttemptWhisper:     whisperText,
			AttemptWav2Vec2:    wav2vec2Text,
			IsReviewed:         false, // Flag for manual review
			HumanTranscription: "",    // Will be filled during review
		}

		// Note: entry.UploadEntry() is not called because:
		// 1. We already have transcriptions
		// 2. Metrics will be calculated during review via RecalculateAllMetrics()

		// Insert into database
		err = txn.Insert(&entry)
		if err != nil {
			errMsg := "Failed to insert entry for " + orphan.FileName + ": " + err.Error()
			revel.AppLog.Errorf("RecoverOrphanedRecordings: %s", errMsg)
			result.Failures = append(result.Failures, errMsg)
			result.FailedCount++
			continue
		}

		revel.AppLog.Infof("RecoverOrphanedRecordings: Created Entry ID=%d for %s → matched to '%s' (score=%.2f)",
			entry.ID, orphan.FileName, match.Word, match.Score)

		result.CreatedEntries = append(result.CreatedEntries, entry.ID)
		result.ProcessedCount++
	}

	revel.AppLog.Infof("RecoverOrphanedRecordings: Completed. Processed=%d, Failed=%d, Created=%d entries",
		result.ProcessedCount, result.FailedCount, len(result.CreatedEntries))

	return result, nil
}

// GetNextRecording returns the next recording that hasn't been processed in this session
// and doesn't already have an Entry record for the specified user
// txn: Database transaction
// userID: User ID to check for existing entries
// processedIDs: List of file paths already processed in this session
func GetNextRecording(txn *gorp.Transaction, userID int64, processedIDs []string) (*OrphanedRecording, error) {
	// Find orphaned recordings for this user only
	orphanedRecordings, err := FindOrphanedRecordingsForUser(txn, userID)
	if err != nil {
		return nil, err
	}

	// Create map of processed IDs for O(1) lookup
	processedMap := make(map[string]bool)
	for _, id := range processedIDs {
		processedMap[id] = true
	}

	// Find first unprocessed recording, working backwards (newest first)
	// Iterate in reverse order to go from newest to oldest
	for i := len(orphanedRecordings) - 1; i >= 0; i-- {
		if !processedMap[orphanedRecordings[i].FilePath] {
			return &orphanedRecordings[i], nil
		}
	}

	// No more unprocessed recordings
	return nil, nil
}

// normalizeForMatching removes all punctuation, symbols, and extra whitespace
// keeping only letters and single spaces for fuzzy matching
func normalizeForMatching(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Keep only letters and spaces
	var result []rune
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			result = append(result, r)
		}
	}

	// Convert back to string and normalize whitespace
	normalized := string(result)

	// Collapse multiple spaces to single space
	normalized = strings.Join(strings.Fields(normalized), " ")

	return strings.TrimSpace(normalized)
}

// CheckExactMatch checks if either ASR transcription exactly matches any word (normalized)
// Returns true if exact match found, along with the matched Word
func CheckExactMatch(txn *gorp.Transaction, whisperText, wav2vec2Text string) (bool, *models.Word, error) {
	// Get all words
	var words []models.Word
	_, err := txn.Select(&words, "SELECT ID, Text, English FROM Word")
	if err != nil {
		return false, nil, err
	}

	// Normalize transcriptions - remove punctuation, symbols, extra whitespace
	normalizedWhisper := normalizeForMatching(whisperText)
	normalizedWav2Vec2 := normalizeForMatching(wav2vec2Text)

	// Check each word
	for _, word := range words {
		normalizedWord := normalizeForMatching(word.Text)

		if normalizedWord != "" && (normalizedWord == normalizedWhisper || normalizedWord == normalizedWav2Vec2) {
			revel.AppLog.Infof("Exact match found: '%s' (normalized: '%s') matches word '%s' (ID=%d)",
				whisperText, normalizedWhisper, word.Text, word.ID)
			return true, &word, nil
		}
	}

	return false, nil, nil
}

// AutoVerifyRecording creates an Entry record for an auto-verified exact match
func AutoVerifyRecording(txn *gorp.Transaction, recordingPath string, wordID, userID int64, whisperText, wav2vec2Text string) error {
	// Get word details
	var word models.Word
	if err := txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", wordID); err != nil {
		return err
	}

	// Create Entry record
	entry := models.Entry{
		UserID:             userID,
		WordID:             wordID,
		Text:               word.Text,
		English:            word.English,
		Recording:          recordingPath,
		AttemptWhisper:     whisperText,
		AttemptWav2Vec2:    wav2vec2Text,
		IsReviewed:         false,
		HumanTranscription: "",
	}

	// Calculate metrics
	entry.RecalculateAllMetrics()

	// Insert into database
	err := txn.Insert(&entry)
	if err != nil {
		return err
	}

	revel.AppLog.Infof("AutoVerifyRecording: Created Entry ID=%d for recording → word '%s' (ID=%d)", entry.ID, word.Text, wordID)
	return nil
}

// CreateRecoveryEntry creates an Entry record for a manually verified match
func CreateRecoveryEntry(txn *gorp.Transaction, recordingPath string, wordID, userID int64, whisperText, wav2vec2Text string) error {
	// Get word details
	var word models.Word
	if err := txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", wordID); err != nil {
		return err
	}

	// Create Entry record
	entry := models.Entry{
		UserID:             userID,
		WordID:             wordID,
		Text:               word.Text,
		English:            word.English,
		Recording:          recordingPath,
		AttemptWhisper:     whisperText,
		AttemptWav2Vec2:    wav2vec2Text,
		IsReviewed:         false,
		HumanTranscription: "",
	}

	// Calculate metrics
	entry.RecalculateAllMetrics()

	// Insert into database
	err := txn.Insert(&entry)
	if err != nil {
		return err
	}

	revel.AppLog.Infof("CreateRecoveryEntry: Created Entry ID=%d for recording → word '%s' (ID=%d)", entry.ID, word.Text, wordID)
	return nil
}
