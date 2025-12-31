package controllers

import (
	"fmt"
	"os"
	"time"

	"app/app/models"
	"app/app/services"

	"github.com/str20tbl/revel"
)

type TranscriptionReview struct {
	AuthController
}

// Index displays the transcription review page
func (t *TranscriptionReview) Index(word, cerRange, errorAttribution, modelWinner string) revel.Result {
	users, err := models.GetNormalUsers(t.Txn)
	if err != nil {
		revel.AppLog.Error(err.Error())
		users = []models.User{} // Fallback to empty slice
	}

	// Pass URL parameters to template for initial filter state
	return t.Render(users, word, cerRange, errorAttribution, modelWinner)
}

// GetTranscriptions fetches entries with filters and pagination for server-side DataTable
func (t *TranscriptionReview) GetTranscriptions(userID, wordText, errorAttribution, reviewStatus, cerRange, modelWinner, startDate, endDate string, start, length int) revel.Result {
	// Default pagination values if not provided
	if length <= 0 {
		length = 25 // Default page size
	}
	if start < 0 {
		start = 0
	}

	// Build filter from request parameters
	filter := models.EntryFilter{
		UserID:           userID,
		WordText:         wordText,
		ErrorAttribution: errorAttribution,
		ReviewStatus:     reviewStatus,
		CERRange:         cerRange,
		ModelWinner:      modelWinner,
		StartDate:        startDate,
		EndDate:          endDate,
		Limit:            length,
		Offset:           start,
	}

	// Get total count matching filters (for DataTable pagination info)
	totalFiltered, err := models.GetFilteredEntryCount(t.Txn, filter)
	if err != nil {
		revel.AppLog.Errorf("GetTranscriptions: Failed to count entries: %v", err)
		return t.JSONError("Failed to count entries")
	}

	// Delegate to model layer for query execution
	entries, err := models.FilteredEntryQuery(t.Txn, filter)
	if err != nil {
		revel.AppLog.Errorf("GetTranscriptions: Failed to fetch entries: %v", err)
		return t.JSONError("Failed to fetch entries")
	}

	// Return DataTable server-side format
	return t.JSONSuccess(map[string]interface{}{
		"data":            entries,
		"recordsTotal":    totalFiltered, // Total records matching filter
		"recordsFiltered": totalFiltered, // Same as recordsTotal (no search applied)
	})
}

// UpdateHumanTranscription saves manual transcription and recalculates metrics
func (t *TranscriptionReview) UpdateHumanTranscription(entryID int64, humanTranscription string) revel.Result {
	var entry models.Entry
	if err := t.Txn.SelectOne(&entry, "SELECT * FROM Entry WHERE id = ?", entryID); err != nil {
		revel.AppLog.Errorf("UpdateHumanTranscription: Entry %d not found: %v", entryID, err)
		return t.JSONError("Entry not found")
	}

	entry.HumanTranscription = humanTranscription
	entry.RecalculateAllMetrics()

	if _, err := t.Txn.Update(&entry); err != nil {
		revel.AppLog.Errorf("UpdateHumanTranscription: Failed to update entry %d: %v", entryID, err)
		return t.JSONError("Failed to update entry")
	}

	return t.JSONSuccess(map[string]interface{}{
		"werWhisper":                    entry.WERWhisper,
		"cerWhisper":                    entry.CERWhisper,
		"transcriptionAccuracyWhisper":  entry.TranscriptionAccuracyWhisper,
		"errorAttributionWhisper":       entry.ErrorAttributionWhisper,
		"werWav2Vec2":                   entry.WERWav2Vec2,
		"cerWav2Vec2":                   entry.CERWav2Vec2,
		"transcriptionAccuracyWav2Vec2": entry.TranscriptionAccuracyWav2Vec2,
		"errorAttributionWav2Vec2":      entry.ErrorAttributionWav2Vec2,
	})
}

// MarkAsReviewed marks an entry as reviewed
func (t *TranscriptionReview) MarkAsReviewed(entryID int64, notes string) revel.Result {
	user, ok := t.connected()
	if !ok {
		return t.JSONError("Not authenticated")
	}

	var entry models.Entry
	if err := t.Txn.SelectOne(&entry, "SELECT * FROM Entry WHERE id = ?", entryID); err != nil {
		revel.AppLog.Errorf("MarkAsReviewed: Entry %d not found: %v", entryID, err)
		return t.JSONError("Entry not found")
	}

	entry.IsReviewed = true
	entry.ReviewedBy = user.ID
	entry.ReviewedAt = time.Now().Format("2006-01-02 15:04:05")
	if notes != "" {
		entry.ErrorNotes = notes
	}

	if _, err := t.Txn.Update(&entry); err != nil {
		revel.AppLog.Errorf("MarkAsReviewed: Failed to update entry %d: %v", entryID, err)
		return t.JSONError("Failed to mark as reviewed")
	}

	return t.JSONSuccess(map[string]interface{}{
		"reviewedAt": entry.ReviewedAt,
	})
}

// DeleteEntry deletes an entry from the database
func (t *TranscriptionReview) DeleteEntry(entryID int64) revel.Result {
	// Check if entry exists
	var entry models.Entry
	if err := t.Txn.SelectOne(&entry, "SELECT * FROM Entry WHERE id = ?", entryID); err != nil {
		revel.AppLog.Errorf("DeleteEntry: Entry %d not found: %v", entryID, err)
		return t.JSONError("Entry not found")
	}

	// Delete from database
	if _, err := t.Txn.Delete(&entry); err != nil {
		revel.AppLog.Errorf("DeleteEntry: Failed to delete entry %d: %v", entryID, err)
		return t.JSONError("Failed to delete entry")
	}

	revel.AppLog.Infof("DeleteEntry: Successfully deleted entry %d", entryID)
	return t.JSONSuccess(map[string]interface{}{
		"message": "Entry deleted successfully",
	})
}

// ExportTranscriptions exports transcriptions as LaTeX table
func (t *TranscriptionReview) ExportTranscriptions(userID, wordText, errorAttribution, reviewStatus, cerRange, modelWinner string) revel.Result {
	// Build filter from request parameters (no limit for export)
	filter := models.EntryFilter{
		UserID:           userID,
		WordText:         wordText,
		ErrorAttribution: errorAttribution,
		ReviewStatus:     reviewStatus,
		CERRange:         cerRange,
		ModelWinner:      modelWinner,
		Limit:            0, // No limit for export
	}

	// Delegate to model layer for query execution
	entries, err := models.FilteredEntryQuery(t.Txn, filter)
	if err != nil {
		return t.RenderText("Error fetching data")
	}

	// Delegate to service layer for LaTeX document generation
	latexDocument := services.GenerateTranscriptionDocument(entries)

	// Set response headers for file download
	t.Response.Out.Header().Set("Content-Type", "text/plain; charset=utf-8")
	t.Response.Out.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=transcriptions_%s.tex", time.Now().Format("20060102_150405")))

	return t.RenderText(latexDocument)
}

// PlayRecording serves a recording file
func (t *TranscriptionReview) PlayRecording(filename string) revel.Result {
	audio, err := os.Open(fmt.Sprintf("/data/recordings/%s", filename))
	if err != nil {
		return t.RenderText("Recording not found")
	}
	return t.RenderFile(audio, revel.Inline)
}

// RecalculateAllMetrics recalculates WER/CER for ALL entries (including reviewed ones)
func (t *TranscriptionReview) RecalculateAllMetrics() revel.Result {
	var entries []models.Entry
	_, err := t.Txn.Select(&entries, "SELECT * FROM Entry WHERE Text != '' AND (AttemptWhisper != '' OR AttemptWav2Vec2 != '')")
	if err != nil {
		revel.AppLog.Errorf("RecalculateAllMetrics: Failed to fetch entries: %v", err)
		return t.JSONError("Failed to fetch entries")
	}

	updatedCount := 0
	for i := range entries {
		entries[i].RecalculateAllMetrics()
		if _, err = t.Txn.Update(&entries[i]); err != nil {
			revel.AppLog.Errorf("Failed to update entry ID %d: %v", entries[i].ID, err)
			continue
		}
		updatedCount++
	}

	return t.JSONSuccess(map[string]interface{}{
		"updatedCount": updatedCount,
		"totalEntries": len(entries),
	})
}
