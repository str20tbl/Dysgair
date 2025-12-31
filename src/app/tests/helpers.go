package tests

import (
	"app/app/controllers"
	"app/app/models"
	"golang.org/x/crypto/bcrypt"
)

// Test Helper Functions
// These utilities help with common test setup and assertions

// CreateTestUser creates a test user in the database
// Returns the created user and any error
func CreateTestUser(username, email, password string) (*models.User, error) {
	if controllers.Dbm == nil {
		return nil, nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	user := &models.User{
		Username:   username,
		Email:      email,
		FirstName:  "Test",
		LastName:   "User",
		ProgressID: 1,
		UserType:   0,
	}

	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = txn.Insert(user)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateTestWord creates a test word in the database
func CreateTestWord(text, english string) (*models.Word, error) {
	if controllers.Dbm == nil {
		return nil, nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	word := &models.Word{
		Text:    text,
		English: english,
	}

	err = txn.Insert(word)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return word, nil
}

// CreateTestEntry creates a test entry in the database
func CreateTestEntry(userID, wordID int64, text, attempt string) (*models.Entry, error) {
	if controllers.Dbm == nil {
		return nil, nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	entry := &models.Entry{
		UserID:         userID,
		WordID:         wordID,
		Text:           text,
		AttemptWhisper: attempt,
		Recording:      "/data/recordings/test.wav",
	}

	err = txn.Insert(entry)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// CreateTestEntries creates multiple test entries for a user/word combination
// Useful for testing the 5-recording system
func CreateTestEntries(userID, wordID int64, count int) error {
	if controllers.Dbm == nil {
		return nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	for i := 0; i < count; i++ {
		entry := &models.Entry{
			UserID:         userID,
			WordID:         wordID,
			Text:           "testword",
			AttemptWhisper: "testword",
			Recording:      "/data/recordings/test.wav",
		}

		err = txn.Insert(entry)
		if err != nil {
			return err
		}
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	return nil
}

// CleanupTestUser deletes a test user and all related data
func CleanupTestUser(userID int64) error {
	if controllers.Dbm == nil {
		return nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	// Delete user's entries first (foreign key constraint)
	_, err = txn.Exec("DELETE FROM Entry WHERE UserID = ?", userID)
	if err != nil {
		return err
	}

	// Delete user
	_, err = txn.Exec("DELETE FROM User WHERE id = ?", userID)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// CleanupTestWord deletes a test word
func CleanupTestWord(wordID int64) error {
	if controllers.Dbm == nil {
		return nil
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	_, err = txn.Exec("DELETE FROM Word WHERE id = ?", wordID)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// AssertFloatEquals checks if two floats are equal within tolerance
func AssertFloatEquals(actual, expected, tolerance float64) bool {
	diff := actual - expected
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// AssertStringContains checks if a string contains a substring
func AssertStringContains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && contains(str, substr)
}

// contains is a simple substring check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test Data Generators

// GenerateUniqueEmail generates a unique email for testing
func GenerateUniqueEmail(prefix string) string {
	// In production, use timestamp or UUID
	// For simplicity, using prefix + @test.com
	return prefix + "@test.com"
}

// GenerateUniqueUsername generates a unique username for testing
func GenerateUniqueUsername(prefix string) string {
	return prefix + "_user"
}

// Mock Data Structures

// MockTranscriptionResponse simulates transcription API response
type MockTranscriptionResponse struct {
	Whisper  string
	Wav2Vec2 string
}

// GetMockTranscription returns mock transcription data
func GetMockTranscription(word string) MockTranscriptionResponse {
	return MockTranscriptionResponse{
		Whisper:  word,
		Wav2Vec2: word,
	}
}
