package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gorp/gorp"
	"github.com/google/uuid"
	"github.com/str20tbl/revel"
)

type Entry struct {
	ID              int64  `db:"id"`
	UserID          int64  `db:"UserID"`
	WordID          int64  `db:"WordID"`
	Text            string `db:"Text"`
	English         string `db:"English"`
	AttemptWhisper  string `db:"AttemptWhisper,size:2000"`
	AttemptWav2Vec2 string `db:"AttemptWav2Vec2"`
	// Lenient (normalized) versions of transcriptions - stored in DB for analytics
	AttemptWhisperLenient         string              `db:"AttemptWhisperLenient,size:2000"`
	AttemptWav2Vec2Lenient        string              `db:"AttemptWav2Vec2Lenient"`
	Recording                     string              `db:"Recording"`
	Coloured                      []string            `db:"-"`
	HumanTranscription            string              `db:"HumanTranscription"`
	IsReviewed                    bool                `db:"IsReviewed"`
	ReviewedBy                    int64               `db:"ReviewedBy"`
	ReviewedAt                    string              `db:"ReviewedAt"`
	WERWhisper                    float64             `db:"WERWhisper"`
	CERWhisper                    float64             `db:"CERWhisper"`
	WERWav2Vec2                   float64             `db:"WERWav2Vec2"`
	CERWav2Vec2                   float64             `db:"CERWav2Vec2"`
	TranscriptionAccuracyWhisper  float64             `db:"TranscriptionAccuracyWhisper"`
	TranscriptionAccuracyWav2Vec2 float64             `db:"TranscriptionAccuracyWav2Vec2"`
	ErrorAttributionWhisper       ErrorClassification `db:"ErrorAttributionWhisper"`
	ErrorAttributionWav2Vec2      ErrorClassification `db:"ErrorAttributionWav2Vec2"`
	EditOperations                string              `db:"EditOperations"`
	ErrorNotes                    string              `db:"ErrorNotes"`
	// Lenient versions of metrics (with space-insensitive matching)
	WERWhisperLenient                    float64             `db:"WERWhisperLenient"`
	CERWhisperLenient                    float64             `db:"CERWhisperLenient"`
	TranscriptionAccuracyWhisperLenient  float64             `db:"TranscriptionAccuracyWhisperLenient"`
	ErrorAttributionWhisperLenient       ErrorClassification `db:"ErrorAttributionWhisperLenient"`
	WERWav2Vec2Lenient                   float64             `db:"WERWav2Vec2Lenient"`
	CERWav2Vec2Lenient                   float64             `db:"CERWav2Vec2Lenient"`
	TranscriptionAccuracyWav2Vec2Lenient float64             `db:"TranscriptionAccuracyWav2Vec2Lenient"`
	ErrorAttributionWav2Vec2Lenient      ErrorClassification `db:"ErrorAttributionWav2Vec2Lenient"`

	// Backwards compatibility aliases (computed fields for display)
	// These should reference AttemptWhisperLenient/AttemptWav2Vec2Lenient where needed
	NormalizedWhisper  string `db:"-"`
	NormalizedWav2Vec2 string `db:"-"`
}

func (entry *Entry) Init(txn *gorp.Transaction, user *User) {
	if user.ProgressID == 0 {
		user.ProgressID += 1
		_, err := txn.Update(user)
		if err != nil {
			revel.AppLog.Errorf("Could not update ProgressID %+v", err)
		}
	}
	err := txn.SelectOne(&entry, "SELECT * FROM Entry WHERE UserID = ? AND WordID = ? ORDER BY id DESC LIMIT 1", user.ID, user.ProgressID)
	if err != nil {
		revel.AppLog.Errorf("could not select Entry %+v", err)
		var word Word
		// Intentionally ignore error: Fallback query in error handler
		// If this fails too, word will be zero-value which is handled below
		_ = txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", user.ProgressID)
		revel.AppLog.Infof("Word id %d %+v", user.ProgressID, word)
		entry.Text = word.Text
		entry.English = word.English
		entry.WordID = word.ID
		entry.UserID = user.ID
	}
	entry.Coloured = entry.longestCommon()
}

func (entry *Entry) UploadEntry(wordID, userID int64, transcriptions map[string]string) {
	// Set transcriptions from both models (provided by caller)
	entry.AttemptWhisper = transcriptions["whisper"]
	entry.AttemptWav2Vec2 = transcriptions["wav2vec2"]

	entry.WordID = wordID
	entry.UserID = userID

	// Note: All statistical analysis (WER/CER, auto-population, error attribution)
	// is now performed during the review stage via RecalculateAllMetrics()
}

func (entry *Entry) longestCommon() (longest []string) {
	longest = []string{"", "", ""}
	// Use Whisper ASR output for visual diff
	str1 := strings.ToLower(entry.AttemptWhisper)
	str2 := strings.ToLower(entry.Text)
	for i := 0; i < len(str1); i++ {
		for j := 0; j < len(str2); j++ {
			k := 0
			for i+k < len(str1) && j+k < len(str2) && str1[i+k] == str2[j+k] {
				k++
			}
			if k > len(longest[1]) {
				longest = []string{str1[0:i], str1[i : i+k], str1[i+k:]}
			}
		}
	}
	revel.AppLog.Infof("%+v", longest)
	return
}

// GetRecordingCount returns the number of recordings for a specific user and word
func GetRecordingCount(txn *gorp.Transaction, userID, wordID int64) (int, error) {
	count, err := txn.SelectInt("SELECT COUNT(*) FROM Entry WHERE UserID = ? AND WordID = ?", userID, wordID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetCompletedWordsCount returns the number of words with 5+ recordings for a user
func GetCompletedWordsCount(txn *gorp.Transaction, userID int64) (int, error) {
	query := `
		SELECT COUNT(DISTINCT WordID)
		FROM Entry
		WHERE UserID = ?
		AND WordID IN (
			SELECT WordID
			FROM Entry
			WHERE UserID = ?
			GROUP BY WordID
			HAVING COUNT(*) >= 5
		)`
	count, err := txn.SelectInt(query, userID, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetTotalWordsCount returns the total number of words in the dataset
func GetTotalWordsCount(txn *gorp.Transaction) (int, error) {
	count, err := txn.SelectInt("SELECT COUNT(*) FROM Word")
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// IsWordComplete checks if a user has completed 5 recordings for a specific word
func IsWordComplete(txn *gorp.Transaction, userID, wordID int64) (bool, error) {
	count, err := GetRecordingCount(txn, userID, wordID)
	if err != nil {
		return false, err
	}
	return count >= 5, nil
}

// GetNextIncompleteWord finds the first word (starting from the beginning) that has < 5 recordings for the user.
// Returns the WordID of the next incomplete word, or 0 if all words are complete.
// This ensures users fill the word list from the front rather than jumping to higher IDs.
func GetNextIncompleteWord(txn *gorp.Transaction, userID, currentProgressID int64) (int64, error) {
	// Find the minimum WordID that has fewer than 5 recordings
	// Uses LEFT JOIN to include words with zero recordings
	query := `
		SELECT w.id
		FROM Word w
		LEFT JOIN (
			SELECT WordID, COUNT(*) as count
			FROM Entry
			WHERE UserID = ?
			GROUP BY WordID
		) e ON w.id = e.WordID
		WHERE COALESCE(e.count, 0) < 5
		ORDER BY w.id ASC
		LIMIT 1`

	nextWordID, err := txn.SelectNullInt(query, userID)
	if err != nil {
		return 0, err
	}

	if !nextWordID.Valid {
		// All words complete
		return 0, nil
	}

	return nextWordID.Int64, nil
}

// SaveRecording encapsulates the complete recording persistence workflow.
// This method consolidates all business logic for saving a recording:
// - File processing via UploadEntry()
// - Transcription via provided transcribe function
// - Database persistence
// - Recording count calculation
// - Word completion status determination
//
// The transcribe function is injected to avoid import cycles with services package.
// Returns a map suitable for JSON response and any error encountered.
func SaveRecording(txn *gorp.Transaction, file []byte, filename, word string, wordID, userID int64, transcribe func(string) map[string]string) (map[string]interface{}, error) {
	// Create entry and save file (generates recording path)
	entry := Entry{Text: word}

	// Get the Word record to populate English translation
	wordRecord, err := GetWordByID(txn, wordID)
	if err != nil {
		revel.AppLog.Warnf("Could not fetch word %d for English translation: %v", wordID, err)
		// Non-fatal: continue without English
	} else {
		entry.English = wordRecord.English
	}

	fileExt := filepath.Ext(filename)
	entry.Recording = fmt.Sprintf("/data/recordings/%s%s", uuid.New().String(), fileExt)
	err = os.WriteFile(entry.Recording, file, 0644)
	if err != nil {
		return nil, err
	}

	// Get transcriptions using injected function
	transcriptions := transcribe(entry.Recording)

	// Populate entry with transcriptions and metrics
	entry.UploadEntry(wordID, userID, transcriptions)

	// Persist entry to database
	err = txn.Insert(&entry)
	if err != nil {
		return nil, err
	}

	// Calculate recording count using existing model function
	count, err := GetRecordingCount(txn, userID, wordID)
	if err != nil {
		// Non-fatal: return 0 if count fails
		count = 0
	}

	// Determine completion status using existing model function
	isComplete, err := IsWordComplete(txn, userID, wordID)
	if err != nil {
		// Non-fatal: assume incomplete if check fails
		isComplete = false
	}

	// Return structured result for controller to marshal as JSON
	return map[string]interface{}{
		"success":        true,
		"recordingCount": count,
		"isComplete":     isComplete,
		"progress":       fmt.Sprintf("%d/5", count),
	}, nil
}

// WordProgress encapsulates all data about a user's progress on a specific word.
// This consolidates data fetching logic that was previously duplicated in controllers.
type WordProgress struct {
	LatestEntry    Entry
	Entries        []Entry
	RecordingCount int
	IsComplete     bool
	Progress       string
	Word           string
	English        string
	WordID         int64
	ProgressID     int64

	// Overall dataset progress
	CompletedWordsCount int     // Words with 5+ recordings
	TotalWordsCount     int     // Total words in dataset (979)
	OverallPercentage   float64 // (CompletedWords / TotalWords) * 100
	WordPercentage      int     // (RecordingCount / 5) * 100
}

// GetEntriesForWord retrieves all entries for a specific user and word,
// ordered by most recent first.
func GetEntriesForWord(txn *gorp.Transaction, userID, wordID int64) ([]Entry, error) {
	var entries []Entry
	_, err := txn.Select(&entries, "SELECT * FROM Entry WHERE UserID = ? AND WordID = ? ORDER BY id DESC", userID, wordID)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// NewWordProgress creates a WordProgress instance by fetching all relevant data
// for the user's current word. This consolidates logic previously duplicated
// in Index() and buildWordDataResponse().
func NewWordProgress(txn *gorp.Transaction, user *User) (*WordProgress, error) {
	// Initialize latest entry for current word
	var latestEntry Entry
	latestEntry.Init(txn, user)

	// Get all entries for current word
	entries, err := GetEntriesForWord(txn, user.ID, user.ProgressID)
	if err != nil {
		revel.AppLog.Errorf("Could not select Entries: %+v", err)
		// Non-fatal: continue with empty entries
		entries = []Entry{}
	}

	// Get recording count
	recordingCount, err := GetRecordingCount(txn, user.ID, user.ProgressID)
	if err != nil {
		revel.AppLog.Errorf("Could not get recording count: %+v", err)
		recordingCount = 0
	}

	// Calculate completion status (business logic)
	isComplete := recordingCount >= 5

	// Get overall dataset progress
	completedWordsCount, err := GetCompletedWordsCount(txn, user.ID)
	if err != nil {
		revel.AppLog.Errorf("Could not get completed words count: %+v", err)
		completedWordsCount = 0
	}

	totalWordsCount, err := GetTotalWordsCount(txn)
	if err != nil {
		revel.AppLog.Errorf("Could not get total words count: %+v", err)
		totalWordsCount = 979 // Fallback to known dataset size
	}

	// Calculate percentages
	overallPercentage := 0.0
	if totalWordsCount > 0 {
		overallPercentage = (float64(completedWordsCount) / float64(totalWordsCount)) * 100.0
	}

	wordPercentage := (recordingCount * 100) / 5

	return &WordProgress{
		LatestEntry:         latestEntry,
		Entries:             entries,
		RecordingCount:      recordingCount,
		IsComplete:          isComplete,
		Progress:            fmt.Sprintf("%d/5", recordingCount),
		Word:                latestEntry.Text,
		English:             latestEntry.English,
		WordID:              latestEntry.WordID,
		ProgressID:          user.ProgressID,
		CompletedWordsCount: completedWordsCount,
		TotalWordsCount:     totalWordsCount,
		OverallPercentage:   overallPercentage,
		WordPercentage:      wordPercentage,
	}, nil
}

// ToMap converts WordProgress to a map suitable for JSON responses.
// Used by AJAX endpoints that need to return complete word data.
func (wp *WordProgress) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"success":             true,
		"word":                wp.Word,
		"english":             wp.English,
		"wordID":              wp.WordID,
		"progressID":          wp.ProgressID,
		"entries":             wp.Entries,
		"recordingCount":      wp.RecordingCount,
		"isComplete":          wp.IsComplete,
		"progress":            wp.Progress,
		"latestEntry":         wp.LatestEntry,
		"completedWordsCount": wp.CompletedWordsCount,
		"totalWordsCount":     wp.TotalWordsCount,
		"overallPercentage":   wp.OverallPercentage,
		"wordPercentage":      wp.WordPercentage,
	}
}

// EntryWithUser combines Entry data with associated User information.
// Used for queries that join Entry and User tables (e.g., transcription review, export).
type EntryWithUser struct {
	Entry
	Username string `db:"Username"`
	Email    string `db:"Email"`
}

// EntryFilter contains criteria for filtering entries.
// Consolidates filtering logic previously duplicated in TranscriptionReview controller.
type EntryFilter struct {
	UserID           string
	WordText         string
	ErrorAttribution string
	ReviewStatus     string // "REVIEWED", "UNREVIEWED", or ""
	CERRange         string // For drill-down from Analytics (e.g., "high", "medium", "low")
	ModelWinner      string // For drill-down from Analytics (e.g., "whisper", "wav2vec2")
	StartDate        string
	EndDate          string
	Limit            int // 0 = no limit
	Offset           int // For pagination, 0 = start from beginning
}

// FilteredEntryQuery builds and executes a filtered query for entries with user information.
// This consolidates SQL-building logic that was previously duplicated in
// GetTranscriptions() and ExportTranscriptions() in TranscriptionReview controller.
func FilteredEntryQuery(txn *gorp.Transaction, filter EntryFilter) ([]EntryWithUser, error) {
	query := `SELECT
		e.ID, e.UserID, e.WordID, e.Text, e.English,
		e.AttemptWhisper, e.AttemptWav2Vec2,
		e.AttemptWhisperLenient, e.AttemptWav2Vec2Lenient,
		e.Recording,
		e.HumanTranscription, e.IsReviewed,
		e.WERWhisper, e.CERWhisper, e.WERWav2Vec2, e.CERWav2Vec2,
		e.TranscriptionAccuracyWhisper, e.TranscriptionAccuracyWav2Vec2,
		e.ErrorAttributionWhisper, e.ErrorAttributionWav2Vec2,
		e.WERWhisperLenient, e.CERWhisperLenient,
		e.WERWav2Vec2Lenient, e.CERWav2Vec2Lenient,
		e.TranscriptionAccuracyWhisperLenient, e.TranscriptionAccuracyWav2Vec2Lenient,
		e.ErrorAttributionWhisperLenient, e.ErrorAttributionWav2Vec2Lenient,
		e.ErrorNotes,
		usr.Username, usr.Email
	FROM Entry e INNER JOIN User usr ON e.UserID = usr.id WHERE 1=1`
	args := []interface{}{}

	// Apply UserID filter
	if filter.UserID != "" && filter.UserID != "0" {
		query += " AND e.UserID = ?"
		var userID int64
		if _, err := fmt.Sscanf(filter.UserID, "%d", &userID); err != nil {
			revel.AppLog.Warnf("Failed to parse UserID filter '%s': %v (treating as 0)", filter.UserID, err)
		}
		args = append(args, userID)
	}

	// Apply WordText filter
	if filter.WordText != "" {
		query += " AND e.Text LIKE ?"
		args = append(args, "%"+filter.WordText+"%")
	}

	// Apply ErrorAttribution filter
	// When ModelWinner is specified (drill-down from error costs), filter by that model's attribution
	// Otherwise, default to Whisper (backward compatibility)
	if filter.ErrorAttribution != "" && filter.ErrorAttribution != "ALL" {
		if filter.ModelWinner == "wav2vec2" {
			query += " AND e.ErrorAttributionWav2Vec2 = ?"
		} else {
			// Default to Whisper if no model specified or if whisper specified
			query += " AND e.ErrorAttributionWhisper = ?"
		}
		args = append(args, filter.ErrorAttribution)
	}

	// Apply ReviewStatus filter
	if filter.ReviewStatus == "REVIEWED" {
		query += " AND e.IsReviewed = TRUE"
	} else if filter.ReviewStatus == "UNREVIEWED" {
		query += " AND e.IsReviewed = FALSE"
	}

	// Apply date range filters
	if filter.StartDate != "" {
		query += " AND e.ReviewedAt >= ?"
		args = append(args, filter.StartDate)
	}

	if filter.EndDate != "" {
		query += " AND e.ReviewedAt <= ?"
		args = append(args, filter.EndDate)
	}

	// Add ordering: entries needing review (empty HumanTranscription) first, then by ID desc
	query += " ORDER BY (e.HumanTranscription IS NULL OR e.HumanTranscription = '') DESC, e.id DESC"

	// Add limit and offset for pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	// Execute query
	var entries []EntryWithUser
	_, err := txn.Select(&entries, query, args...)
	if err != nil {
		return nil, err
	}

	// Populate normalized text fields for display
	for i := range entries {
		entries[i].PopulateNormalizedText()
	}

	return entries, nil
}

// GetFilteredEntryCount returns the total count of entries matching the filter criteria.
// Used for server-side pagination to show total records.
func GetFilteredEntryCount(txn *gorp.Transaction, filter EntryFilter) (int64, error) {
	query := `SELECT COUNT(*)
	FROM Entry e INNER JOIN User usr ON e.UserID = usr.id WHERE 1=1`
	args := []interface{}{}

	// Apply UserID filter
	if filter.UserID != "" && filter.UserID != "0" {
		query += " AND e.UserID = ?"
		var userID int64
		if _, err := fmt.Sscanf(filter.UserID, "%d", &userID); err != nil {
			revel.AppLog.Warnf("Failed to parse UserID filter '%s': %v (treating as 0)", filter.UserID, err)
		}
		args = append(args, userID)
	}

	// Apply WordText filter
	if filter.WordText != "" {
		query += " AND e.Text LIKE ?"
		args = append(args, "%"+filter.WordText+"%")
	}

	// Apply ErrorAttribution filter
	// When ModelWinner is specified (drill-down from error costs), filter by that model's attribution
	// Otherwise, default to Whisper (backward compatibility)
	if filter.ErrorAttribution != "" && filter.ErrorAttribution != "ALL" {
		if filter.ModelWinner == "wav2vec2" {
			query += " AND e.ErrorAttributionWav2Vec2 = ?"
		} else {
			// Default to Whisper if no model specified or if whisper specified
			query += " AND e.ErrorAttributionWhisper = ?"
		}
		args = append(args, filter.ErrorAttribution)
	}

	// Apply ReviewStatus filter
	if filter.ReviewStatus == "REVIEWED" {
		query += " AND e.IsReviewed = TRUE"
	} else if filter.ReviewStatus == "UNREVIEWED" {
		query += " AND e.IsReviewed = FALSE"
	}

	// Apply date range filters
	if filter.StartDate != "" {
		query += " AND e.ReviewedAt >= ?"
		args = append(args, filter.StartDate)
	}

	if filter.EndDate != "" {
		query += " AND e.ReviewedAt <= ?"
		args = append(args, filter.EndDate)
	}

	// Execute count query
	count, err := txn.SelectInt(query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ModelsAgree checks if Whisper and Wav2Vec2 models produced identical transcriptions.
// This business logic was previously in the controller's CSV export.
func (e *Entry) ModelsAgree() bool {
	return strings.ToLower(strings.TrimSpace(e.AttemptWhisper)) == strings.ToLower(strings.TrimSpace(e.AttemptWav2Vec2))
}

// GetEntriesForAnalysis retrieves entries for statistical analysis.
// Optionally filters by userID. Returns all entries ordered by most recent.
func GetEntriesForAnalysis(txn *gorp.Transaction, userID string) ([]Entry, error) {
	query := "SELECT * FROM Entry WHERE 1=1"
	args := []interface{}{}

	if userID != "" && userID != "0" {
		query += " AND UserID = ?"
		var id int64
		if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
			revel.AppLog.Warnf("Failed to parse userID for analysis '%s': %v (treating as 0)", userID, err)
		}
		args = append(args, id)
	}

	query += " ORDER BY id DESC"

	var entries []Entry
	_, err := txn.Select(&entries, query, args...)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
