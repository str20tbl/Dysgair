package controllers

import (
	"fmt"
	"path/filepath"
	"strings"

	"app/app/models"
	"app/app/services"
	"app/appJobs"

	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/cache"
	"github.com/str20tbl/modules/jobs/app/jobs"
)

type RecoveryDashboard struct {
	AuthController
}

// Index displays the recovery dashboard
func (r *RecoveryDashboard) Index() revel.Result {
	users, err := models.GetNormalUsers(r.Txn)
	if err != nil {
		revel.AppLog.Error(err.Error())
		users = []models.User{}
	}

	// Check if there's an active session
	sessionUserID := r.Session["RecoveryUserID"]
	processedCount := 0
	if processedRecordings, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedRecordings.(string); ok && processedStr != "" {
			processedCount = len(strings.Split(processedStr, ","))
		}
	}

	return r.Render(users, sessionUserID, processedCount)
}

// StartSession initializes a recovery session with a target user and starts background processing
func (r *RecoveryDashboard) StartSession(userID int64) revel.Result {
	// Validate userID
	var user models.User
	if err := r.Txn.SelectOne(&user, "SELECT * FROM User WHERE id = ?", userID); err != nil {
		return r.JSONError("Invalid user ID")
	}

	// Initialize session
	r.Session["RecoveryUserID"] = fmt.Sprintf("%d", userID)
	r.Session["ProcessedRecordings"] = ""

	// Trigger background job to process all recordings
	jobs.Now(appJobs.RecordingRecovery{
		UserID: userID,
		Dbm:    Dbm,
	})

	revel.AppLog.Infof("Recovery session started for user %d (%s %s) - background job triggered",
		userID, user.FirstName, user.LastName)

	return r.JSONSuccess(map[string]interface{}{
		"userID":     userID,
		"userName":   fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		"message":    "Recovery session started",
		"processing": true,
	})
}

// GetProgress returns the progress of the background recovery job
func (r *RecoveryDashboard) GetProgress() revel.Result {
	// Check session
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if !ok {
		return r.JSONError("No active recovery session")
	}

	sessionUserIDStr, ok := sessionUserIDVal.(string)
	if !ok || sessionUserIDStr == "" {
		return r.JSONError("No active recovery session")
	}

	// Parse userID
	var sessionUserID int64
	fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

	// Get progress from Redis
	progressKey := fmt.Sprintf("recovery:progress:%d", sessionUserID)
	var progress appJobs.RecoveryProgress
	err := cache.Get(progressKey, &progress)
	if err != nil {
		// Progress not found - check if manual review list exists (resume scenario)
		manualListKey := fmt.Sprintf("recovery:manual:%d", sessionUserID)
		var manualItems []appJobs.ManualReviewItem
		if cache.Get(manualListKey, &manualItems) == nil {
			// Manual review list exists - job completed, return complete status
			revel.AppLog.Infof("GetProgress: Progress cache missing but manual review list exists for user %d - job completed", sessionUserID)
			return r.JSONSuccess(map[string]interface{}{
				"total":        len(manualItems),
				"processed":    0,
				"auto_matched": 0,
				"needs_manual": len(manualItems),
				"complete":     true,
				"status":       "manual_review",
			})
		}
		// No progress yet - job might not have started
		return r.JSONSuccess(map[string]interface{}{
			"total":        0,
			"processed":    0,
			"auto_matched": 0,
			"needs_manual": 0,
			"complete":     false,
			"status":       "initializing",
		})
	}

	return r.JSONSuccess(map[string]interface{}{
		"total":        progress.Total,
		"processed":    progress.Processed,
		"auto_matched": progress.AutoMatched,
		"needs_manual": progress.NeedsManual,
		"complete":     progress.Complete,
		"status":       "processing",
	})
}

// GetManualReviewList returns the list of recordings that need manual review
func (r *RecoveryDashboard) GetManualReviewList() revel.Result {
	// Check session
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if !ok {
		return r.JSONError("No active recovery session")
	}

	sessionUserIDStr, ok := sessionUserIDVal.(string)
	if !ok || sessionUserIDStr == "" {
		return r.JSONError("No active recovery session")
	}

	// Parse userID
	var sessionUserID int64
	fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

	// Get manual review list from Redis
	manualListKey := fmt.Sprintf("recovery:manual:%d", sessionUserID)
	var manualItems []appJobs.ManualReviewItem
	err := cache.Get(manualListKey, &manualItems)
	if err != nil {
		// No manual items yet
		manualItems = []appJobs.ManualReviewItem{}
	}

	return r.JSONSuccess(map[string]interface{}{
		"items": manualItems,
		"count": len(manualItems),
	})
}

// GetNextRecording fetches the next orphaned recording and auto-verifies if exact match found
func (r *RecoveryDashboard) GetNextRecording() revel.Result {
	// Check session
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if !ok {
		return r.JSONError("No active recovery session")
	}

	sessionUserIDStr, ok := sessionUserIDVal.(string)
	if !ok || sessionUserIDStr == "" {
		return r.JSONError("No active recovery session")
	}

	// Parse userID
	var sessionUserID int64
	fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

	// Get orphaned recordings from Redis cache
	redisKey := fmt.Sprintf("recovery:orphaned:%d", sessionUserID)
	var orphanedRecordings []string
	err := cache.Get(redisKey, &orphanedRecordings)
	if err != nil {
		revel.AppLog.Errorf("GetNextRecording: Failed to retrieve orphaned recordings from Redis: %v", err)
		return r.JSONError("Recovery session expired or not found. Please start a new session.")
	}

	// Get processed recordings from session
	processedRecordings := []string{}
	if processedVal, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedVal.(string); ok && processedStr != "" {
			processedRecordings = strings.Split(processedStr, ",")
		}
	}

	// Create map of processed IDs for O(1) lookup
	processedMap := make(map[string]bool)
	for _, id := range processedRecordings {
		processedMap[id] = true
	}

	// Counter for auto-verified recordings in this request
	autoVerifiedCount := 0
	maxAutoVerify := 10 // Prevent timeout

	// Try to auto-verify up to maxAutoVerify recordings
	for autoVerifiedCount < maxAutoVerify {
		// Get next recording from cached orphaned list (iterate backwards for newest first)
		var recording *services.OrphanedRecording
		for i := len(orphanedRecordings) - 1; i >= 0; i-- {
			if !processedMap[orphanedRecordings[i]] {
				recording = &services.OrphanedRecording{
					FilePath: orphanedRecordings[i],
					FileName: filepath.Base(orphanedRecordings[i]),
				}
				break
			}
		}

		if recording == nil {
			// No more unprocessed recordings
			return r.JSONSuccess(map[string]interface{}{
				"completed":         true,
				"autoVerifiedCount": autoVerifiedCount,
				"message":           fmt.Sprintf("All recordings processed! %d auto-verified in this batch.", autoVerifiedCount),
			})
		}

		// Transcribe if needed
		transcriptions := services.Transcribe(recording.FilePath)
		whisperText := transcriptions["whisper"]
		wav2vec2Text := transcriptions["wav2vec2"]

		if whisperText == "" && wav2vec2Text == "" {
			// No transcription available - mark as processed and skip
			processedRecordings = append(processedRecordings, recording.FilePath)
			processedMap[recording.FilePath] = true
			r.Session["ProcessedRecordings"] = strings.Join(processedRecordings, ",")
			continue
		}

		// Check for exact match
		exactMatch, matchedWord, err := services.CheckExactMatch(r.Txn, whisperText, wav2vec2Text)
		if err != nil {
			revel.AppLog.Errorf("GetNextRecording: Error checking exact match: %v", err)
		}

		if exactMatch && matchedWord != nil {
			// Auto-verify!
			err := services.AutoVerifyRecording(r.Txn, recording.FilePath, matchedWord.ID, sessionUserID, whisperText, wav2vec2Text)
			if err != nil {
				revel.AppLog.Errorf("GetNextRecording: Failed to auto-verify: %v", err)
				// Don't fail - just stop auto-verifying and return this recording for manual review
				break
			}

			// Mark as processed
			processedRecordings = append(processedRecordings, recording.FilePath)
			processedMap[recording.FilePath] = true
			r.Session["ProcessedRecordings"] = strings.Join(processedRecordings, ",")
			autoVerifiedCount++

			revel.AppLog.Infof("Auto-verified recording %s → word '%s' (ID=%d)", recording.FileName, matchedWord.Text, matchedWord.ID)

			// Continue to next recording
			continue
		}

		// No exact match - return this recording for manual review
		// Find top 10 word matches
		matches, err := services.FindTopNWordMatches(r.Txn, whisperText, 10)
		if err != nil {
			// Try wav2vec2 as fallback
			if wav2vec2Text != "" {
				matches, err = services.FindTopNWordMatches(r.Txn, wav2vec2Text, 10)
			}
		}

		if err != nil {
			revel.AppLog.Errorf("GetNextRecording: Failed to find matches: %v", err)
			matches = []*services.WordMatch{}
		}

		// Get orphaned recording count from session cache
		totalCount := len(orphanedRecordings)
		processedCount := len(processedRecordings)

		return r.JSONSuccess(map[string]interface{}{
			"completed": false,
			"recording": map[string]interface{}{
				"filePath":     recording.FilePath,
				"fileName":     recording.FileName,
				"whisperText":  whisperText,
				"wav2vec2Text": wav2vec2Text,
			},
			"matches":           matches,
			"autoVerifiedCount": autoVerifiedCount,
			"progress": map[string]interface{}{
				"processed": processedCount,
				"total":     totalCount,
				"remaining": totalCount - processedCount,
			},
		})
	}

	// Reached max auto-verify limit - return status
	return r.JSONSuccess(map[string]interface{}{
		"autoVerifyLimitReached": true,
		"autoVerifiedCount":      autoVerifiedCount,
		"message":                fmt.Sprintf("Auto-verified %d recordings. Click 'Next' to continue.", autoVerifiedCount),
	})
}

// VerifyMatch creates an Entry record for a manual match
func (r *RecoveryDashboard) VerifyMatch(recordingPath string, wordID int64) revel.Result {
	// Check session
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if !ok {
		return r.JSONError("No active recovery session")
	}

	sessionUserIDStr, ok := sessionUserIDVal.(string)
	if !ok || sessionUserIDStr == "" {
		return r.JSONError("No active recovery session")
	}

	// Parse userID
	var sessionUserID int64
	fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

	// Get word
	var word models.Word
	if err := r.Txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", wordID); err != nil {
		return r.JSONError("Word not found")
	}

	// Transcribe recording
	transcriptions := services.Transcribe(recordingPath)
	whisperText := transcriptions["whisper"]
	wav2vec2Text := transcriptions["wav2vec2"]

	// Create entry
	err := services.CreateRecoveryEntry(r.Txn, recordingPath, wordID, sessionUserID, whisperText, wav2vec2Text)
	if err != nil {
		revel.AppLog.Errorf("VerifyMatch: Failed to create entry: %v", err)
		return r.JSONError(fmt.Sprintf("Failed to create entry: %v", err))
	}

	// Mark as processed
	processedRecordings := []string{}
	if processedVal, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedVal.(string); ok && processedStr != "" {
			processedRecordings = strings.Split(processedStr, ",")
		}
	}
	processedRecordings = append(processedRecordings, recordingPath)
	r.Session["ProcessedRecordings"] = strings.Join(processedRecordings, ",")

	revel.AppLog.Infof("Manually verified recording %s → word '%s' (ID=%d)", recordingPath, word.Text, wordID)

	return r.JSONSuccess(map[string]interface{}{
		"message": fmt.Sprintf("Verified recording for word '%s'", word.Text),
	})
}

// SkipRecording marks a recording as processed without creating an entry
func (r *RecoveryDashboard) SkipRecording(recordingPath string) revel.Result {
	// Mark as processed
	processedRecordings := []string{}
	if processedVal, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedVal.(string); ok && processedStr != "" {
			processedRecordings = strings.Split(processedStr, ",")
		}
	}
	processedRecordings = append(processedRecordings, recordingPath)
	r.Session["ProcessedRecordings"] = strings.Join(processedRecordings, ",")

	revel.AppLog.Infof("Skipped recording %s", recordingPath)

	return r.JSONSuccess(map[string]interface{}{
		"message": "Recording skipped",
	})
}

// GetAllWords returns all words in the database for search filtering
func (r *RecoveryDashboard) GetAllWords() revel.Result {
	var words []models.Word
	_, err := r.Txn.Select(&words, "SELECT ID, Text, English FROM Word ORDER BY Text")
	if err != nil {
		revel.AppLog.Errorf("GetAllWords: Failed to load words: %v", err)
		return r.JSONError("Failed to load words")
	}

	return r.JSONSuccess(map[string]interface{}{
		"words": words,
	})
}

// GetProcessedRecordings returns the list of processed recording paths from the session
func (r *RecoveryDashboard) GetProcessedRecordings() revel.Result {
	// Get processed recordings from session
	processedRecordings := []string{}
	if processedVal, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedVal.(string); ok && processedStr != "" {
			processedRecordings = strings.Split(processedStr, ",")
		}
	}

	return r.JSONSuccess(map[string]interface{}{
		"recordings": processedRecordings,
		"count":      len(processedRecordings),
	})
}

// UpdateWord updates the Text and English fields of a Word
func (r *RecoveryDashboard) UpdateWord(wordID int64, text, english string) revel.Result {
	var word models.Word
	if err := r.Txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", wordID); err != nil {
		return r.JSONError("Word not found")
	}

	word.Text = text
	word.English = english

	if _, err := r.Txn.Update(&word); err != nil {
		revel.AppLog.Errorf("UpdateWord: Failed to update word %d: %v", wordID, err)
		return r.JSONError("Failed to update word")
	}

	revel.AppLog.Infof("Updated word %d: Text='%s', English='%s'", wordID, text, english)

	return r.JSONSuccess(map[string]interface{}{
		"message": "Word updated successfully",
		"word": map[string]interface{}{
			"id":      wordID,
			"text":    text,
			"english": english,
		},
	})
}

// GetRecoveryStats returns progress statistics
func (r *RecoveryDashboard) GetRecoveryStats() revel.Result {
	// Check if there's an active session
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if !ok {
		return r.JSONError("No active recovery session")
	}

	sessionUserIDStr, ok := sessionUserIDVal.(string)
	if !ok || sessionUserIDStr == "" {
		return r.JSONError("No active recovery session")
	}

	// Parse userID
	var sessionUserID int64
	fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

	// Get orphaned recordings from Redis cache
	redisKey := fmt.Sprintf("recovery:orphaned:%d", sessionUserID)
	var orphanedRecordings []string
	err := cache.Get(redisKey, &orphanedRecordings)
	if err != nil {
		revel.AppLog.Errorf("GetRecoveryStats: Failed to retrieve orphaned recordings from Redis: %v", err)
		return r.JSONError("Recovery session expired or not found. Please start a new session.")
	}

	// Get processed count from session
	processedRecordings := []string{}
	if processedVal, ok := r.Session["ProcessedRecordings"]; ok {
		if processedStr, ok := processedVal.(string); ok && processedStr != "" {
			processedRecordings = strings.Split(processedStr, ",")
		}
	}

	totalCount := len(orphanedRecordings)
	processedCount := len(processedRecordings)

	return r.JSONSuccess(map[string]interface{}{
		"total":      totalCount,
		"processed":  processedCount,
		"remaining":  totalCount - processedCount,
		"percentage": float64(processedCount) / float64(totalCount) * 100,
	})
}

// ResetSession clears the recovery session
func (r *RecoveryDashboard) ResetSession() revel.Result {
	// Get userID from session before deleting
	sessionUserIDVal, ok := r.Session["RecoveryUserID"]
	if ok {
		if sessionUserIDStr, ok := sessionUserIDVal.(string); ok && sessionUserIDStr != "" {
			var sessionUserID int64
			fmt.Sscanf(sessionUserIDStr, "%d", &sessionUserID)

			// Delete Redis cache for this user's orphaned recordings
			redisKey := fmt.Sprintf("recovery:orphaned:%d", sessionUserID)
			err := cache.Delete(redisKey)
			if err != nil {
				revel.AppLog.Warnf("ResetSession: Failed to delete Redis key %s: %v", redisKey, err)
			}
		}
	}

	// Clear session data
	delete(r.Session, "RecoveryUserID")
	delete(r.Session, "ProcessedRecordings")

	revel.AppLog.Info("Recovery session reset")

	return r.JSONSuccess(map[string]interface{}{
		"message": "Session reset successfully",
	})
}
