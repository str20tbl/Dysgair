package controllers

import (
	"app/app/models"

	"github.com/str20tbl/revel"
)

type UserManagement struct {
	AuthController
}

// Index displays the home page
func (u *UserManagement) Index() revel.Result {
	return u.Render()
}

// Admin displays the admin page
func (u *UserManagement) Admin() revel.Result {
	users, err := models.GetNormalUsers(u.Txn)
	if err != nil {
		revel.AppLog.Error(err.Error())
		u.Flash.Error("Methu nôl rhestr o defnyddiwr.\n\nFailed to fetch list of users.")
		users = []models.User{} // Fallback to empty slice
	}
	return u.Render(users)
}

// Profile displays the profile page
func (u *UserManagement) Profile() revel.Result {
	return u.Render()
}

// NewUser creates a new user
func (u *UserManagement) NewUser(email string) revel.Result {
	count, err := u.Txn.SelectInt("SELECT COUNT(*) FROM User WHERE Email = ?", email)
	if count == 0 {
		user, err := models.NewUser(email)
		if err == nil {
			err = u.Txn.Insert(&user)
		}
	}
	if err != nil {
		u.Flash.Error("Nad yw'r defnyddiwr newydd wedi greu.\n\nThe new user has not been created.")
	}
	return u.Redirect(u.Request.Referer())
}

// DeleteUser deletes a user
func (u *UserManagement) DeleteUser(id int64) revel.Result {
	if err := models.DeleteUser(u.Txn, id); err != nil {
		u.Flash.Error("Nad yw'r cyfrif wedi cael ei dileu. Cysylltwch gyda'r tîm os nad ydych yn medru dileu eich cyfrif.\n\nThe account has not been deleted. Contact the team if you are not able to delete your account.")
	}
	return u.Redirect(u.Request.Referer())
}

// UpdatePassword updates user password
func (u *UserManagement) UpdatePassword(id int64, oldPassword, newPassword string) revel.Result {
	if err := models.UpdatePassword(u.Txn, id, oldPassword, newPassword); err != nil {
		u.Flash.Error("Gwall wrth ddiweddaru cyfrinair.\n\nError updating password.")
	}
	return u.Redirect(u.Request.Referer())
}

// UpdateEmail updates user email
func (u *UserManagement) UpdateEmail(id int64, password, email string) revel.Result {
	if err := models.UpdateEmail(u.Txn, id, password, email); err != nil {
		u.Flash.Error("Gwall wrth ddiweddaru e-bost.\n\nError updating email.")
	}
	return u.Redirect(u.Request.Referer())
}

// UpdateUserNames updates user names
func (u *UserManagement) UpdateUserNames(id int64, first, second, username string) revel.Result {
	if err := models.UpdateUserNames(u.Txn, id, first, second, username); err != nil {
		u.Flash.Error("Gwall wrth ddiweddaru enwau.\n\nError updating names.")
	}
	return u.Redirect(u.Request.Referer())
}
