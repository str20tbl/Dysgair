package controllers

import (
	"fmt"
	"os"

	"app/app/models"
	"app/app/services"

	"github.com/str20tbl/revel"
)

const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
)

type Dysgair struct {
	AuthController
}

func (c *Dysgair) Index(entry models.Entry) revel.Result {
	// Fetch all word progress data from model layer
	wordProgress, err := models.NewWordProgress(c.Txn, c.User)
	if err != nil {
		revel.AppLog.Errorf("Could not load word progress: %+v", err)
		return c.RenderError(err)
	}

	// Pass WordProgress fields to template (maintains backward compatibility with template variables)
	entry = wordProgress.LatestEntry
	entries := wordProgress.Entries
	recordingCount := wordProgress.RecordingCount
	isComplete := wordProgress.IsComplete
	progress := wordProgress.Progress

	// Pass new overall progress fields
	CompletedWordsCount := wordProgress.CompletedWordsCount
	TotalWordsCount := wordProgress.TotalWordsCount
	OverallPercentage := wordProgress.OverallPercentage
	WordPercentage := wordProgress.WordPercentage

	return c.Render(entry, entries, recordingCount, isComplete, progress,
		CompletedWordsCount, TotalWordsCount, OverallPercentage, WordPercentage)
}

func (c *Dysgair) ResetProgress() revel.Result {
	c.User.ProgressID = 1
	_, err := c.Txn.Update(c.User)
	if err != nil {
		revel.AppLog.Errorf("Could not update ProgressID %+v", err)
		return c.JSONError("Failed to reset progress")
	}
	return c.RenderJSON(c.buildWordDataResponse(c.User))
}

func (c *Dysgair) IncrementProgressID() revel.Result {
	// Check if current word has 5 recordings before allowing navigation
	isComplete, err := models.IsWordComplete(c.Txn, c.User.ID, c.User.ProgressID)
	if err != nil {
		revel.AppLog.Errorf("Error checking word completion: %+v", err)
		return c.JSONError("Error checking word completion")
	}

	if !isComplete {
		return c.RenderJSON(map[string]interface{}{
			"success": false,
			"error":   "Complete 5 recordings before moving to the next word",
			"errorCy": "Cwblhewch 5 recordiad cyn symud ymlaen",
		})
	}

	// Get maximum word ID from database
	maxWordID, err := models.GetMaxWordID(c.Txn)
	if err != nil {
		revel.AppLog.Errorf("Error getting max word ID: %+v", err)
		return c.JSONError("Failed to check word limit")
	}

	// Only increment if there are more words available
	if c.User.ProgressID < maxWordID {
		c.User.ProgressID += 1
		_, err := c.Txn.Update(c.User)
		if err != nil {
			revel.AppLog.Errorf("Could not update ProgressID %+v", err)
			return c.JSONError("Failed to update progress")
		}
	}
	return c.RenderJSON(c.buildWordDataResponse(c.User))
}

func (c *Dysgair) DecrementProgressID() revel.Result {
	if c.User.ProgressID > 1 {
		c.User.ProgressID -= 1
		_, err := c.Txn.Update(c.User)
		if err != nil {
			revel.AppLog.Errorf("Could not update ProgressID %+v", err)
			return c.JSONError("Failed to update progress")
		}
	}
	return c.RenderJSON(c.buildWordDataResponse(c.User))
}

func (c *Dysgair) JumpToNextIncomplete() revel.Result {
	nextWordID, err := models.GetNextIncompleteWord(c.Txn, c.User.ID, c.User.ProgressID)
	if err != nil {
		revel.AppLog.Errorf("Error finding next incomplete word: %+v", err)
		return c.JSONError("Failed to find next incomplete word")
	}

	if nextWordID == 0 {
		// No incomplete words found - stay at current position
		return c.RenderJSON(map[string]interface{}{
			"success": false,
			"error":   "No more incomplete words found",
			"errorCy": "Dim geiriau anghyflawn yn weddill",
		})
	}

	// Update user's progress to the next incomplete word
	c.User.ProgressID = nextWordID
	_, err = c.Txn.Update(c.User)
	if err != nil {
		revel.AppLog.Errorf("Could not update ProgressID %+v", err)
		return c.JSONError("Failed to update progress")
	}

	return c.RenderJSON(c.buildWordDataResponse(c.User))
}

func (c *Dysgair) Upload(file []byte, word string, wordID int64) revel.Result {
	// 1. HTTP layer: Validate file constraints
	c.Validation.Required(file)
	c.Validation.MinSize(file, 2*KB).
		Message(`Minimum file size 2KB`)
	c.Validation.MaxSize(file, 746*MB).
		Message(`Max file size 746MB`)
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.JSONError("Invalid file size")
	}

	// 2. Model layer: Delegate all business logic
	// Inject transcription service to avoid import cycle
	filename := c.Params.Files["file"][0].Filename
	result, err := models.SaveRecording(c.Txn, file, filename, word, wordID, c.User.ID, services.Transcribe)
	if err != nil {
		revel.AppLog.Errorf("Upload: Failed to save recording for word %d by user %d: %v", wordID, c.User.ID, err)
		return c.JSONError("Upload failed")
	}

	// 3. HTTP layer: Return complete word data for AJAX UI update
	// Merge SaveRecording result with full word data
	fullData := c.buildWordDataResponse(c.User)
	// Override with SaveRecording's specific fields (they're already in fullData, but ensure latest values)
	fullData["recordingCount"] = result["recordingCount"]
	fullData["isComplete"] = result["isComplete"]
	fullData["progress"] = result["progress"]
	return c.RenderJSON(fullData)
}

// GetWordData returns current word and recording data as JSON for AJAX updates
func (c *Dysgair) GetWordData() revel.Result {
	return c.RenderJSON(c.buildWordDataResponse(c.User))
}

// buildWordDataResponse creates a JSON response with current word data.
// Delegates to model layer via WordProgress for all data fetching and business logic.
func (c *Dysgair) buildWordDataResponse(user *models.User) map[string]interface{} {
	wordProgress, err := models.NewWordProgress(c.Txn, user)
	if err != nil {
		revel.AppLog.Errorf("Could not load word progress: %+v", err)
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to load word progress",
		}
	}

	return wordProgress.ToMap()
}

// PlayAudio from a file on disk, serve the audio file
func (c *Dysgair) PlayAudio(id int64) revel.Result {
	word, err := models.GetWordByID(c.Txn, id)
	if err != nil {
		return c.RenderError(err)
	}
	audio, err := os.Open(fmt.Sprintf("/data/audio/%s.wav", word.AudioFilename))
	if err != nil {
		revel.AppLog.Errorf("Could not open audio file for word %d: %v", id, err)
		return c.RenderError(err)
	}
	return c.RenderFile(audio, revel.Attachment)
}
