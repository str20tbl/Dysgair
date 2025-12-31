package controllers

import (
	"database/sql"
	"errors"
	"fmt"

	"app/app/models"

	"github.com/go-gorp/gorp"
	"github.com/str20tbl/revel"
)

// Dbm The GLOBAL DB Object
var Dbm *gorp.DbMap

// GorpController base app controller so that they all have access to the
// Global Transaction
type GorpController struct {
	*revel.Controller
	Txn *gorp.Transaction
}

// Begin Transaction function
func (c *GorpController) Begin() revel.Result {
	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

// Commit Transaction function
func (c *GorpController) Commit() revel.Result {
	if c.Txn != nil {
		if err := c.Txn.Commit(); err != nil && errors.Is(err, sql.ErrTxDone) {
			panic(err)
		}
	}
	c.Txn = nil
	return nil
}

// Rollback Transaction function
func (c *GorpController) Rollback() revel.Result {
	if c.Txn != nil {
		if err := c.Txn.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
			panic(err)
		}
	}
	c.Txn = nil
	return nil
}

// connected checks if a user is authenticated and returns the user
func (c *GorpController) connected() (*models.User, bool) {
	if c.ViewArgs["app_user"] != nil {
		user := c.ViewArgs["app_user"].(*models.User)
		return user, true
	}
	if val, ok := c.Session["app_user"]; ok {
		user, err := models.GetUserByUsername(c.Txn, val.(string))
		if err != nil {
			revel.AppLog.Error("Auth error:", err)
			return nil, false
		}
		if user != nil && len(user.Username) > 0 {
			return user, true
		}
	}
	return nil, false
}

// addUser adds the user to ViewArgs for template rendering
func (c *GorpController) addUser(user *models.User) {
	c.ViewArgs["app_user"] = user
}

// currentUser retrieves the authenticated user from ViewArgs.
// This user is guaranteed to exist for AuthController-based controllers
// because AuthController.Before() either sets it or redirects.
// For App controller (public pages), this may return an empty User struct.
func (c *GorpController) currentUser() *models.User {
	return c.ViewArgs["app_user"].(*models.User)
}

// JSONError returns a standardized JSON error response
func (c *GorpController) JSONError(message string) revel.Result {
	return c.RenderJSON(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// JSONErrorf returns a formatted JSON error response
func (c *GorpController) JSONErrorf(format string, args ...interface{}) revel.Result {
	return c.RenderJSON(map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf(format, args...),
	})
}

// JSONSuccess returns a standardized JSON success response with data
func (c *GorpController) JSONSuccess(data map[string]interface{}) revel.Result {
	data["success"] = true
	return c.RenderJSON(data)
}

// FlashErrorBilingual sets a bilingual error flash message (Welsh/English)
func (c *GorpController) FlashErrorBilingual(cy, en string) {
	messageFormat := `<span class="cy">%s</span><span class="en">%s</span>`
	msg := fmt.Sprintf(messageFormat, cy, en)
	c.Flash.Error(msg)
}

// FlashSuccessBilingual sets a bilingual success flash message (Welsh/English)
func (c *GorpController) FlashSuccessBilingual(cy, en string) {
	messageFormat := `<span class="cy">%s</span><span class="en">%s</span>`
	msg := fmt.Sprintf(messageFormat, cy, en)
	c.Flash.Success(msg)
}

// RedirectBack redirects to the previous page (HTTP referer)
func (c *GorpController) RedirectBack() revel.Result {
	return c.Redirect(c.Request.Referer())
}

// AuthController for authenticated pages - redirects to login if not authenticated
type AuthController struct {
	GorpController
	User *models.User // Populated in Before() - guaranteed to exist in action methods
}

func (c *AuthController) Before() revel.Result {
	if user, ok := c.connected(); ok {
		c.addUser(user)
		c.User = user // Populate field for direct access in controller actions
		return nil
	}
	return c.Redirect("/")
}
