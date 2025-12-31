package appJobs

import (
	"fmt"

	"app/app/services"

	"github.com/go-gorp/gorp"
	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/cache"
)

// RecoveryProgress tracks the progress of the recovery job
type RecoveryProgress struct {
	Total        int  `json:"total"`
	Processed    int  `json:"processed"`
	AutoMatched  int  `json:"auto_matched"`
	NeedsManual  int  `json:"needs_manual"`
	Complete     bool `json:"complete"`
}

// ManualReviewItem represents a recording that needs manual review
type ManualReviewItem struct {
	FilePath    string `json:"file_path"`
	FileName    string `json:"file_name"`
	WhisperText string `json:"whisper_text"`
	Wav2Vec2Text string `json:"wav2vec2_text"`
}

// RecordingRecovery is a background job that processes orphaned recordings for a user
type RecordingRecovery struct {
	UserID int64
	Dbm    *gorp.DbMap
}

// Run executes the recording recovery job
func (r RecordingRecovery) Run() {
	revel.AppLog.Infof("RecordingRecovery job started for user %d", r.UserID)

	// Redis keys
	progressKey := fmt.Sprintf("recovery:progress:%d", r.UserID)
	manualListKey := fmt.Sprintf("recovery:manual:%d", r.UserID)

	// Clear any existing manual review list
	cache.Delete(manualListKey)

	// Get database transaction
	txn, err := r.Dbm.Begin()
	if err != nil {
		revel.AppLog.Errorf("RecordingRecovery: Failed to begin transaction: %v", err)
		return
	}
	defer txn.Rollback()

	// Get all orphaned recordings for this user
	orphanedRecordings, err := services.FindOrphanedRecordingsForUser(txn, r.UserID)
	if err != nil {
		revel.AppLog.Errorf("RecordingRecovery: Failed to find orphaned recordings: %v", err)
		return
	}

	totalRecordings := len(orphanedRecordings)
	revel.AppLog.Infof("RecordingRecovery: Processing %d orphaned recordings for user %d", totalRecordings, r.UserID)

	// Initialize progress
	progress := RecoveryProgress{
		Total:       totalRecordings,
		Processed:   0,
		AutoMatched: 0,
		NeedsManual: 0,
		Complete:    false,
	}
	updateProgress(progressKey, progress)

	// Storage for manual review items
	manualReviewItems := []ManualReviewItem{}

	// Process each recording
	for i, recording := range orphanedRecordings {
		// Transcribe the recording
		transcriptions := services.Transcribe(recording.FilePath)
		whisperText := transcriptions["whisper"]
		wav2vec2Text := transcriptions["wav2vec2"]

		if whisperText == "" && wav2vec2Text == "" {
			// No transcription - skip this recording
			revel.AppLog.Warnf("RecordingRecovery: No transcription for %s, skipping", recording.FileName)
			progress.Processed++
			updateProgress(progressKey, progress)
			continue
		}

		// Check for exact match
		exactMatch, matchedWord, err := services.CheckExactMatch(txn, whisperText, wav2vec2Text)
		if err != nil {
			revel.AppLog.Errorf("RecordingRecovery: Error checking exact match for %s: %v", recording.FileName, err)
		}

		if exactMatch && matchedWord != nil {
			// Auto-verify - create Entry record
			err := services.AutoVerifyRecording(txn, recording.FilePath, matchedWord.ID, r.UserID, whisperText, wav2vec2Text)
			if err != nil {
				revel.AppLog.Errorf("RecordingRecovery: Failed to auto-verify %s: %v", recording.FileName, err)
				// Add to manual review on error
				manualReviewItems = append(manualReviewItems, ManualReviewItem{
					FilePath:     recording.FilePath,
					FileName:     recording.FileName,
					WhisperText:  whisperText,
					Wav2Vec2Text: wav2vec2Text,
				})
				progress.NeedsManual++
			} else {
				revel.AppLog.Infof("RecordingRecovery: Auto-matched %s → word '%s' (ID=%d)", recording.FileName, matchedWord.Text, matchedWord.ID)
				progress.AutoMatched++
			}
		} else {
			// No exact match - add to manual review list
			manualReviewItems = append(manualReviewItems, ManualReviewItem{
				FilePath:     recording.FilePath,
				FileName:     recording.FileName,
				WhisperText:  whisperText,
				Wav2Vec2Text: wav2vec2Text,
			})
			progress.NeedsManual++
		}

		progress.Processed++
		updateProgress(progressKey, progress)

		// Log progress every 50 recordings
		if (i+1)%50 == 0 {
			revel.AppLog.Infof("RecordingRecovery: Progress %d/%d (Auto: %d, Manual: %d)", progress.Processed, totalRecordings, progress.AutoMatched, progress.NeedsManual)
		}
	}

	// Commit transaction
	err = txn.Commit()
	if err != nil {
		revel.AppLog.Errorf("RecordingRecovery: Failed to commit transaction: %v", err)
		return
	}

	// Store manual review list in Redis
	if len(manualReviewItems) > 0 {
		err = cache.Set(manualListKey, manualReviewItems, 0) // No expiration
		if err != nil {
			revel.AppLog.Errorf("RecordingRecovery: Failed to store manual review list: %v", err)
		}
	}

	// Mark as complete
	progress.Complete = true
	updateProgress(progressKey, progress)

	revel.AppLog.Infof("RecordingRecovery: Job completed for user %d - Total: %d, Auto-matched: %d, Needs manual review: %d",
		r.UserID, totalRecordings, progress.AutoMatched, progress.NeedsManual)
}

// updateProgress stores progress in Redis
func updateProgress(key string, progress RecoveryProgress) {
	err := cache.Set(key, progress, 0) // No expiration
	if err != nil {
		revel.AppLog.Errorf("Failed to update recovery progress: %v", err)
	}
}
