package models

import (
	"github.com/go-gorp/gorp"
	"github.com/str20tbl/revel"
	"golang.org/x/crypto/bcrypt"
)

// User of dysgair.nehpets.co.uk
type User struct {
	ID               int64  `db:"id"`
	Username         string `db:"Username"`
	FirstName        string `db:"FirstName"`
	LastName         string `db:"LastName"`
	HashedPassword   []byte `db:"HashedPassword"`
	Password         string `db:"-"`
	PasswordConfirm  string `db:"-"`
	Email            string `db:"Email"`
	EmailConfirm     string `db:"-"`
	TermsOfUse       bool   `db:"TermsOfUse"`
	UserType         int    `db:"UserType"`
	RegistrationHash []byte `db:"RegistrationHash"`
	ProgressID       int64  `db:"ProgressID"`
}

func NewUser(email string) (u User, err error) {
	u.Email = email
	u.RegistrationHash, err = bcrypt.GenerateFromPassword([]byte(u.Email), bcrypt.DefaultCost)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

func (user *User) IncrementProgress(txn *gorp.Transaction) {
	user.ProgressID += 1
	_, err := txn.Update(user)
	if err != nil {
		revel.AppLog.Errorf("Could not update user.ProgressID: %v", err)
		return
	}
	// Reload user from database to ensure struct is up-to-date
	err = txn.SelectOne(user, "SELECT * FROM User WHERE id = ?", user.ID)
	if err != nil {
		revel.AppLog.Errorf("Could not reload user after progress update: %v", err)
	}
}

// Validate User member fields
func (user *User) Validate(v *revel.Validation) {
	v.Required(user.Username)
	v.MinSize(user.Username, 6)
	v.Required(user.FirstName)
	v.Required(user.LastName)
	v.Required(user.Password)
	v.MinSize(user.Password, 6)
	v.Required(user.PasswordConfirm)
	v.Required(user.PasswordConfirm == user.Password).Message("The passwords do not match.")
	v.Required(user.Email)
	v.Email(user.Email)
	v.Required(user.EmailConfirm)
	v.Required(user.EmailConfirm == user.Email).Message("The email addresses do not match.")
	v.Required(user.TermsOfUse)
}

func GetUserByUsername(txn *gorp.Transaction, username string) (user *User, err error) {
	query := "SELECT * FROM User WHERE Username = ?"
	err = txn.SelectOne(&user, query, username)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

// GetUserByHash retrieves a user by their registration hash
func GetUserByHash(txn *gorp.Transaction, hash string) (*User, error) {
	var user User
	err := txn.SelectOne(&user, "SELECT * FROM User WHERE RegistrationHash = ?", hash)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(txn *gorp.Transaction, userID int64) (*User, error) {
	var user User
	err := txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return nil, err
	}
	return &user, nil
}

// GetNormalUsers retrieves all non-admin users (UserType = 0) ordered by username.
// This is used by Analytics, TranscriptionReview, and UserManagement for displaying user lists.
func GetNormalUsers(txn *gorp.Transaction) ([]User, error) {
	var users []User
	_, err := txn.Select(&users, "SELECT * FROM User WHERE UserType = 0 ORDER BY Username")
	if err != nil {
		return nil, err
	}
	return users, nil
}

const userByIDQuery = "SELECT * FROM User WHERE id = ?"

// UpdateUserNames from the database
func UpdateUserNames(txn *gorp.Transaction, userID int64, first, second, username string) (err error) {
	var user User
	err = txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	count, err := txn.SelectInt("SELECT COUNT(*) FROM User WHERE Username = ?", username)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	user.FirstName = first
	user.LastName = second
	if count == 0 {
		user.Username = username
	}
	_, err = txn.Update(&user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

// UpdateEmail from the database
func UpdateEmail(txn *gorp.Transaction, userID int64, password, email string) (err error) {
	var user User
	err = txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	user.Email = email
	_, err = txn.Update(&user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

// UpdatePassword from the database
func UpdatePassword(txn *gorp.Transaction, userID int64, oldPassword, newPassword string) (err error) {
	var user User
	err = txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(oldPassword))
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	_, err = txn.Update(&user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

// NewPassword from the database
func NewPassword(txn *gorp.Transaction, userID int64, password string) (err error) {
	var user User
	err = txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	user.HashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	_, err = txn.Update(&user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}

// DeleteUser from the database
func DeleteUser(txn *gorp.Transaction, userID int64) (err error) {
	var user User
	err = txn.SelectOne(&user, userByIDQuery, userID)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	_, err = txn.Delete(&user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return
}
