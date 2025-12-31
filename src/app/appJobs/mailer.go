package appJobs

import (
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"

	"app/app/models"

	"github.com/str20tbl/revel"
)

type Mailer struct {
	User models.User
}

func (m Mailer) Run() {
	revel.AppLog.Info(m.User.FirstName + "<<<< >>>>" + m.User.LastName)
	serverEmail := revel.Config.StringDefault("email.user", "")
	fromEmail := revel.Config.StringDefault("email.from", "")
	host := revel.Config.StringDefault("email.host", "")
	port := revel.Config.IntDefault("email.port", 25)
	password := revel.Config.StringDefault("email.pass", "")
	auth := smtp.PlainAuth("", serverEmail, password, host)
	subject := "Subject: Cofrestru | dysgair.nehpets.co.uk | Register\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	ex, err := os.Executable()
	if err != nil {
		revel.AppLog.Errorf("could not send email %+v", err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
	mailTemplate, err := os.ReadFile("/usr/src/app/appJobs/registerEmail.html")
	if err != nil {
		revel.AppLog.Error(err.Error())
	} else {
		body := fmt.Sprintf(string(mailTemplate), m.User.FirstName, m.User.RegistrationHash)
		msg := []byte(subject + mime + body)
		err = smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, fromEmail, []string{m.User.Email}, msg)
		if err != nil {
			revel.AppLog.Errorf("Email send error :: %+v", err)
		}
	}
}

type ResetMailer struct {
	User models.User
}

func (m ResetMailer) Run() {
	serverEmail := revel.Config.StringDefault("email.user", "")
	fromEmail := revel.Config.StringDefault("email.from", "")
	host := revel.Config.StringDefault("email.host", "")
	port := revel.Config.IntDefault("email.port", 25)
	password := revel.Config.StringDefault("email.pass", "")
	auth := smtp.PlainAuth("", serverEmail, password, host)
	subject := "Subject: Cofrestru | dysgair.nehpets.co.uk | Register\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	ex, err := os.Executable()
	if err != nil {
		revel.AppLog.Errorf("could not send email %+v", err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
	mailTemplate, err := os.ReadFile("/usr/src/app/appJobs/resetEmail.html")
	if err != nil {
		revel.AppLog.Error(err.Error())
	} else {
		body := fmt.Sprintf(string(mailTemplate), m.User.FirstName, m.User.RegistrationHash)
		msg := []byte(subject + mime + body)
		err = smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, fromEmail, []string{m.User.Email}, msg)
		if err != nil {
			revel.AppLog.Errorf("Email send error :: %+v", err)
		}
	}
}
