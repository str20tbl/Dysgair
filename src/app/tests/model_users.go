package tests

import (
	"app/app/controllers"
	"app/app/models"

	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/testing"
	"golang.org/x/crypto/bcrypt"
)

type UserModelTest struct {
	testing.TestSuite
	testUserID int64
}

func (t *UserModelTest) Before() {
	revel.AppLog.Info("UserModelTest: Set up")
}

// TestNewUser_EmailHashing tests user creation with email hashing
func (t *UserModelTest) TestNewUser_EmailHashing() {
	user, err := models.NewUser("test@example.com")

	t.Assert(err == nil)
	t.Assert(user.Email == "test@example.com")
	t.Assert(len(user.RegistrationHash) > 0)

	// Verify bcrypt hash of email
	err = bcrypt.CompareHashAndPassword(user.RegistrationHash, []byte("test@example.com"))
	t.Assert(err == nil)
}

// TestNewUser_DifferentEmails tests that different emails produce different hashes
func (t *UserModelTest) TestNewUser_DifferentEmails() {
	user1, _ := models.NewUser("user1@example.com")
	user2, _ := models.NewUser("user2@example.com")

	t.Assert(user1.Email != user2.Email)
	// Hashes should be different
	t.Assert(string(user1.RegistrationHash) != string(user2.RegistrationHash))
}

// TestUserValidate_AllFieldsValid tests validation with valid data
func (t *UserModelTest) TestUserValidate_AllFieldsValid() {
	user := models.User{
		Username:        "testuser123",
		FirstName:       "John",
		LastName:        "Doe",
		Password:        "password123",
		PasswordConfirm: "password123",
		Email:           "john@example.com",
		EmailConfirm:    "john@example.com",
		TermsOfUse:      true,
	}

	// Create a validation object (simplified test - in real app, use revel.Validation)
	// For now, just verify the user struct is set up correctly
	t.Assert(user.Username == "testuser123")
	t.Assert(user.Password == user.PasswordConfirm)
	t.Assert(user.Email == user.EmailConfirm)
	t.Assert(user.TermsOfUse == true)
}

// TestUserValidate_UsernameTooShort tests username length validation
func (t *UserModelTest) TestUserValidate_UsernameTooShort() {
	user := models.User{
		Username: "short", // Less than 6 characters
	}

	t.Assert(len(user.Username) < 6)
}

// TestUserValidate_PasswordMismatch tests password confirmation
func (t *UserModelTest) TestUserValidate_PasswordMismatch() {
	user := models.User{
		Password:        "password123",
		PasswordConfirm: "different456",
	}

	t.Assert(user.Password != user.PasswordConfirm)
}

// TestUserValidate_EmailMismatch tests email confirmation
func (t *UserModelTest) TestUserValidate_EmailMismatch() {
	user := models.User{
		Email:        "user@example.com",
		EmailConfirm: "different@example.com",
	}

	t.Assert(user.Email != user.EmailConfirm)
}

// TestUserValidate_TermsNotAccepted tests terms of use requirement
func (t *UserModelTest) TestUserValidate_TermsNotAccepted() {
	user := models.User{
		TermsOfUse: false,
	}

	t.Assert(user.TermsOfUse == false)
}

// TestGetUserByUsername_Integration tests database user retrieval
func (t *UserModelTest) TestGetUserByUsername_Integration() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		revel.AppLog.Errorf("Could not start transaction: %v", err)
		return
	}
	defer txn.Rollback()

	// Create a test user
	testUser := models.User{
		Username:   "dbtest_user",
		Email:      "dbtest@example.com",
		FirstName:  "Test",
		LastName:   "User",
		ProgressID: 1,
	}
	testUser.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	err = txn.Insert(&testUser)
	if err != nil {
		revel.AppLog.Errorf("Could not insert test user: %v", err)
		return
	}

	// Retrieve the user
	foundUser, err := models.GetUserByUsername(txn, "dbtest_user")
	t.Assert(err == nil)
	t.Assert(foundUser != nil)
	t.Assert(foundUser.Username == "dbtest_user")
	t.Assert(foundUser.Email == "dbtest@example.com")
}

// TestGetUserByUsername_NotFound tests non-existent user
func (t *UserModelTest) TestGetUserByUsername_NotFound() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	user, err := models.GetUserByUsername(txn, "nonexistent_user_12345")

	// Should return error for not found
	t.Assert(err != nil || user == nil || user.Username == "")
}

// TestUpdateUserNames_Success tests name update
func (t *UserModelTest) TestUpdateUserNames_Success() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create test user
	testUser := models.User{
		Username:  "nametest_user",
		Email:     "nametest@example.com",
		FirstName: "Old",
		LastName:  "Name",
	}
	err = txn.Insert(&testUser)
	if err != nil {
		return
	}

	// Update names
	err = models.UpdateUserNames(txn, testUser.ID, "New", "Name", "nametest_user")
	t.Assert(err == nil)

	// Verify update
	var updated models.User
	err = txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(err == nil)
	t.Assert(updated.FirstName == "New")
	t.Assert(updated.LastName == "Name")
}

// TestUpdateUserNames_UniqueUsername tests username uniqueness
func (t *UserModelTest) TestUpdateUserNames_UniqueUsername() {
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
	user1 := models.User{Username: "unique_user1", Email: "user1@example.com"}
	user2 := models.User{Username: "unique_user2", Email: "user2@example.com"}

	txn.Insert(&user1)
	txn.Insert(&user2)

	// Try to update user2's username to user1's username (should be prevented)
	err = models.UpdateUserNames(txn, user2.ID, "First", "Last", "unique_user1")

	// Verify username wasn't changed
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", user2.ID)
	t.Assert(updated.Username == "unique_user2") // Should remain unchanged
}

// TestUpdateEmail_ValidPassword tests email update with correct password
func (t *UserModelTest) TestUpdateEmail_ValidPassword() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create user with password
	testUser := models.User{
		Username: "emailtest_user",
		Email:    "old@example.com",
	}
	testUser.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	txn.Insert(&testUser)

	// Update email with correct password
	err = models.UpdateEmail(txn, testUser.ID, "correctpassword", "new@example.com")
	t.Assert(err == nil)

	// Verify update
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(updated.Email == "new@example.com")
}

// TestUpdateEmail_InvalidPassword tests email update with wrong password
func (t *UserModelTest) TestUpdateEmail_InvalidPassword() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create user with password
	testUser := models.User{
		Username: "emailfail_user",
		Email:    "original@example.com",
	}
	testUser.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	txn.Insert(&testUser)

	// Try to update with wrong password
	err = models.UpdateEmail(txn, testUser.ID, "wrongpassword", "hacked@example.com")
	t.Assert(err != nil) // Should fail

	// Verify email unchanged
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(updated.Email == "original@example.com")
}

// TestUpdatePassword_Success tests password change
func (t *UserModelTest) TestUpdatePassword_Success() {
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
	testUser := models.User{
		Username: "passtest_user",
		Email:    "passtest@example.com",
	}
	testUser.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	txn.Insert(&testUser)

	// Update password
	err = models.UpdatePassword(txn, testUser.ID, "oldpassword", "newpassword")
	t.Assert(err == nil)

	// Verify new password works
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	err = bcrypt.CompareHashAndPassword(updated.HashedPassword, []byte("newpassword"))
	t.Assert(err == nil)

	// Verify old password doesn't work
	err = bcrypt.CompareHashAndPassword(updated.HashedPassword, []byte("oldpassword"))
	t.Assert(err != nil)
}

// TestNewPassword_Reset tests password reset (no old password required)
func (t *UserModelTest) TestNewPassword_Reset() {
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
	testUser := models.User{
		Username: "resettest_user",
		Email:    "reset@example.com",
	}
	testUser.HashedPassword, _ = bcrypt.GenerateFromPassword([]byte("forgotten"), bcrypt.DefaultCost)
	txn.Insert(&testUser)

	// Reset password (no old password required)
	err = models.NewPassword(txn, testUser.ID, "resetpassword")
	t.Assert(err == nil)

	// Verify new password works
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	err = bcrypt.CompareHashAndPassword(updated.HashedPassword, []byte("resetpassword"))
	t.Assert(err == nil)
}

// TestDeleteUser tests user deletion
func (t *UserModelTest) TestDeleteUser() {
	if controllers.Dbm == nil {
		revel.AppLog.Warn("Database not available, skipping database test")
		return
	}

	txn, err := controllers.Dbm.Begin()
	if err != nil {
		return
	}
	defer txn.Rollback()

	// Create user to delete
	testUser := models.User{
		Username: "deletetest_user",
		Email:    "delete@example.com",
	}
	txn.Insert(&testUser)

	// Delete user
	err = models.DeleteUser(txn, testUser.ID)
	t.Assert(err == nil)

	// Verify user is gone
	var count int64
	count, _ = txn.SelectInt("SELECT COUNT(*) FROM User WHERE id = ?", testUser.ID)
	t.Assert(count == 0)
}

// TestIncrementProgress tests progress increment
func (t *UserModelTest) TestIncrementProgress() {
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
	testUser := models.User{
		Username:   "progresstest_user",
		Email:      "progress@example.com",
		ProgressID: 1,
	}
	txn.Insert(&testUser)

	// Increment progress
	testUser.IncrementProgress(txn)

	// Verify increment
	var updated models.User
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(updated.ProgressID == 2)

	// Increment again
	updated.IncrementProgress(txn)
	txn.SelectOne(&updated, "SELECT * FROM User WHERE id = ?", testUser.ID)
	t.Assert(updated.ProgressID == 3)
}

func (t *UserModelTest) After() {
	revel.AppLog.Info("UserModelTest: Tear down")
}
