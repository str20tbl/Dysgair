package controllers

import (
	"strings"

	"app/app/models"
	"app/appJobs"

	"github.com/str20tbl/modules/jobs/app/jobs"
	"github.com/str20tbl/revel"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	GorpController
}

// Before loads the user if authenticated, or sets an empty user for template safety.
// Does not redirect - allows anonymous access to public pages.
func (c *App) Before() (revel.Result, *App) {
	if user, ok := c.connected(); ok {
		c.addUser(user)
	} else {
		// Public pages: set empty user for template safety
		c.ViewArgs["app_user"] = &models.User{}
	}
	return nil, c
}

// Index page for app
func (c *App) Index() revel.Result {
	return c.Render()
}

// Register a user for voice building
func (c *App) Register(reg bool) revel.Result {
	return c.Render(reg)
}

// Reg
// Tudalen cofrestru o'r ap.
// Registration page of the app.
func (c *App) Reg(hash string) revel.Result {
	if len(hash) > 0 {
		user, err := models.GetUserByHash(c.Txn, hash)
		if err == nil && user != nil {
			return c.Render(hash)
		}
	}
	return c.Redirect("/")
}

func (c *App) RegUser(name, lastName, email, password, confirm, terms string) revel.Result {
	if terms != "t" {
		c.FlashErrorBilingual(
			"Cadarnhau eich bod chi'n derbyn y telerau ac amodau.",
			"Please confirm acceptance of the Terms & Conditions.",
		)
	} else {
		if password == confirm {
			count, err := c.Txn.SelectInt("SELECT COUNT(*) FROM User WHERE Username = ?", email)
			if count == 0 {
				var user models.User
				user, err = models.NewUser(email)
				if err == nil {
					err = c.Txn.Insert(&user)
					err = models.NewPassword(c.Txn, user.ID, password)
					if err == nil {
						err = models.UpdateUserNames(c.Txn, user.ID, name, lastName, email)
						if err == nil {
							userPtr, err := models.GetUserByUsername(c.Txn, email)
							if err == nil && userPtr != nil {
								revel.AppLog.Info(userPtr.FirstName + "<<<< >>>>" + userPtr.LastName)
								appJob := appJobs.Mailer{User: *userPtr}
								jobs.Now(appJob)
								c.FlashSuccessBilingual(
									"Gwiriwch eich e-bost a gwiriwch eich cyfeiriad.",
									"Please check your email and verify your address.",
								)
								return c.Redirect("/Register")
							}
						}
					}
				} else {
					c.FlashErrorBilingual(
						"Nad yw'r defnyddiwr newydd wedi greu.",
						"The new user has not been created.",
					)
					return c.RedirectBack()
				}
			} else {
				c.FlashErrorBilingual(
					"Ebost mewn defnydd yn barod.",
					"Email already in use.",
				)
				return c.RedirectBack()
			}
		}
	}
	return c.RedirectBack()
}

// Login to the app with a username & password
func (c *App) Login(username, password string) revel.Result {
	user, err := models.GetUserByUsername(c.Txn, username)
	if user != nil && len(user.Username) > 0 && err == nil {
		err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
		if err == nil {
			setSessionInfo(c.GorpController, *user)
			if user.UserType > 0 {
				return c.Redirect("/Admin")
			}
			return c.Redirect("/Dysgair")
		}
	}
	return c.Redirect("/")
}

// Logout from app
func (c *App) Logout() revel.Result {
	// Delete each session var
	for k := range c.Session {
		delete(c.Session, k)
	}
	// Load login page
	return c.Redirect("/")
}

func (c *App) ResetInvite(resetEmail string) revel.Result {
	user, err := models.GetUserByUsername(c.Txn, resetEmail)
	if err != nil || user == nil {
		return c.Redirect("/")
	}
	appJob := appJobs.ResetMailer{User: *user}
	jobs.Now(appJob)
	return c.Redirect("/")
}

func (c *App) Reset(hash string) revel.Result {
	user, err := models.GetUserByHash(c.Txn, hash)
	if err != nil || user == nil {
		return c.Redirect("/")
	}
	return c.Render(*user)
}

func (c *App) PasswordReset(userID int64, password, confirm string) revel.Result {
	user, err := models.GetUserByID(c.Txn, userID)
	if err != nil || user == nil {
		c.FlashErrorBilingual(
			"Cyfrinair heb ei ailosod.",
			"Password not reset.",
		)
		return c.RedirectBack()
	}
	if strings.EqualFold(password, confirm) {
		err = models.NewPassword(c.Txn, user.ID, password)
		if err != nil {
			revel.AppLog.Error(err.Error())
			return c.RedirectBack()
		}
		c.FlashSuccessBilingual(
			"Cyfrinair wedi ei ailosod.",
			"Password reset.",
		)
		return c.Redirect("/Register")
	}
	c.FlashErrorBilingual(
		"Cyfrinair heb ei ailosod.",
		"Password not reset.",
	)
	return c.RedirectBack()
}

func (c *App) Confirm(hash string) revel.Result {
	user, err := models.GetUserByHash(c.Txn, hash)
	if err != nil || user == nil {
		return c.Redirect("/")
	}
	user.TermsOfUse = true
	_, err = c.Txn.Update(user)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	setSessionInfo(c.GorpController, *user)
	if user.UserType == 1 {
		return c.Redirect("/Admin")
	}
	c.FlashSuccessBilingual(
		"E-bost wedi ei wirio.",
		"E-mail confirmed.",
	)
	return c.Redirect("/Dysgair")
}

// Set session info from user struct
func setSessionInfo(c GorpController, user models.User) {
	c.Session["app_user"] = user.Username
}
