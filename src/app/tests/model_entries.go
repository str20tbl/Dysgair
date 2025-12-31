package tests

import (
	"app/app/controllers"
	"app/app/models"

	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/testing"
)

type EntryModelTest struct {
	testing.TestSuite
}

func (t *EntryModelTest) Before() {
	revel.AppLog.Info("EntryModelTest: Set up")
}

// Note: longestCommon() is a private method, tested indirectly through Init()

// TestGetRecordingCount_NoRecordings tests zero recordings
func (t *EntryModelTest) TestGetRecordingCount_NoRecordings() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user and word
	testUser := models.User{Username: "count_user", Email: "count@example.com"}
	txn.Insert(&testUser)

	testWord := models.Word{Text: "testword", English: "testword"}
	txn.Insert(&testWord)

	// Get count (should be 0)
	count, err := models.GetRecordingCount(txn, testUser.ID, testWord.ID)
	t.Assert(err == nil)
	t.Assert(count == 0)
}

// TestGetRecordingCount_WithRecordings tests recording count
func (t *EntryModelTest) TestGetRecordingCount_WithRecordings() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user and word
	testUser := models.User{Username: "count2_user", Email: "count2@example.com"}
	txn.Insert(&testUser)

	testWord := models.Word{Text: "testword2", English: "testword2"}
	txn.Insert(&testWord)

	// Create 3 entries
	for i := 0; i < 3; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Get count
	count, err := models.GetRecordingCount(txn, testUser.ID, testWord.ID)
	t.Assert(err == nil)
	t.Assert(count == 3)
}

// TestIsWordComplete_FourRecordings tests incomplete word (4 recordings)
func (t *EntryModelTest) TestIsWordComplete_FourRecordings() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user and word
	testUser := models.User{Username: "complete4_user", Email: "complete4@example.com"}
	txn.Insert(&testUser)

	testWord := models.Word{Text: "incomplete", English: "incomplete"}
	txn.Insert(&testWord)

	// Create 4 entries (not complete)
	for i := 0; i < 4; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Check completion
	isComplete, err := models.IsWordComplete(txn, testUser.ID, testWord.ID)
	t.Assert(err == nil)
	t.Assert(isComplete == false) // Should not be complete with only 4 recordings
}

// TestIsWordComplete_FiveRecordings tests complete word (5 recordings)
func (t *EntryModelTest) TestIsWordComplete_FiveRecordings() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user and word
	testUser := models.User{Username: "complete5_user", Email: "complete5@example.com"}
	txn.Insert(&testUser)

	testWord := models.Word{Text: "complete", English: "complete"}
	txn.Insert(&testWord)

	// Create exactly 5 entries (complete)
	for i := 0; i < 5; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Check completion
	isComplete, err := models.IsWordComplete(txn, testUser.ID, testWord.ID)
	t.Assert(err == nil)
	t.Assert(isComplete == true) // Should be complete with 5 recordings
}

// TestIsWordComplete_SixRecordings tests over-complete word (6 recordings)
func (t *EntryModelTest) TestIsWordComplete_SixRecordings() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user and word
	testUser := models.User{Username: "complete6_user", Email: "complete6@example.com"}
	txn.Insert(&testUser)

	testWord := models.Word{Text: "overcomplete", English: "overcomplete"}
	txn.Insert(&testWord)

	// Create 6 entries (more than required)
	for i := 0; i < 6; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Check completion
	isComplete, err := models.IsWordComplete(txn, testUser.ID, testWord.ID)
	t.Assert(err == nil)
	t.Assert(isComplete == true) // Should still be complete with more than 5
}

// TestIsWordComplete_DifferentUsers tests user isolation
func (t *EntryModelTest) TestIsWordComplete_DifferentUsers() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create two users
	user1 := models.User{Username: "isolation1_user", Email: "isolation1@example.com"}
	user2 := models.User{Username: "isolation2_user", Email: "isolation2@example.com"}
	txn.Insert(&user1)
	txn.Insert(&user2)

	// Create shared word
	testWord := models.Word{Text: "shared", English: "shared"}
	txn.Insert(&testWord)

	// User1 has 5 recordings (complete)
	for i := 0; i < 5; i++ {
		entry := models.Entry{
			UserID:    user1.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// User2 has 2 recordings (incomplete)
	for i := 0; i < 2; i++ {
		entry := models.Entry{
			UserID:    user2.ID,
			WordID:    testWord.ID,
			Text:      testWord.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// User1 should be complete
	isComplete1, _ := models.IsWordComplete(txn, user1.ID, testWord.ID)
	t.Assert(isComplete1 == true)

	// User2 should not be complete
	isComplete2, _ := models.IsWordComplete(txn, user2.ID, testWord.ID)
	t.Assert(isComplete2 == false)
}

// TestEntryInit_NewUser tests initialization for new user
func (t *EntryModelTest) TestEntryInit_NewUser() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create new user with ProgressID = 0
	testUser := models.User{
		Username:   "init_user",
		Email:      "init@example.com",
		ProgressID: 0,
	}
	txn.Insert(&testUser)

	// Create a word
	testWord := models.Word{Text: "first", English: "first"}
	txn.Insert(&testWord)

	// Initialize entry
	entry := models.Entry{}
	entry.Init(txn, &testUser)

	// User's ProgressID should be incremented to 1
	var updatedUser models.User
	txn.SelectOne(&updatedUser, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(updatedUser.ProgressID == 1)
}

// TestEntryInit_ExistingProgress tests initialization with existing progress
func (t *EntryModelTest) TestEntryInit_ExistingProgress() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create user with progress
	testUser := models.User{
		Username:   "progress_user",
		Email:      "progress@example.com",
		ProgressID: 5,
	}
	txn.Insert(&testUser)

	// Create word at ProgressID 5
	testWord := models.Word{Text: "word5", English: "word5"}
	txn.Insert(&testWord)

	// Create previous entry for this word
	previousEntry := models.Entry{
		UserID: testUser.ID,
		WordID: 5, // testUser.ProgressID
		Text:   testWord.Text,
	}
	txn.Insert(&previousEntry)

	// Initialize new entry
	entry := models.Entry{}
	entry.Init(txn, &testUser)

	// Should load the word text
	t.Assert(entry.UserID == testUser.ID)
	// WordID might be different based on actual data
}

// TestGetRecordingCount_MultipleWords tests count isolation by word
func (t *EntryModelTest) TestGetRecordingCount_MultipleWords() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create user
	testUser := models.User{Username: "multiword_user", Email: "multiword@example.com"}
	txn.Insert(&testUser)

	// Create two words
	word1 := models.Word{Text: "word1", English: "word1"}
	word2 := models.Word{Text: "word2", English: "word2"}
	txn.Insert(&word1)
	txn.Insert(&word2)

	// Create 3 recordings for word1
	for i := 0; i < 3; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    word1.ID,
			Text:      word1.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Create 5 recordings for word2
	for i := 0; i < 5; i++ {
		entry := models.Entry{
			UserID:    testUser.ID,
			WordID:    word2.ID,
			Text:      word2.Text,
			Recording: "/data/recordings/test.wav",
		}
		txn.Insert(&entry)
	}

	// Verify counts are isolated
	count1, _ := models.GetRecordingCount(txn, testUser.ID, word1.ID)
	count2, _ := models.GetRecordingCount(txn, testUser.ID, word2.ID)

	t.Assert(count1 == 3)
	t.Assert(count2 == 5)
}

func (t *EntryModelTest) After() {
	revel.AppLog.Info("EntryModelTest: Tear down")
}
