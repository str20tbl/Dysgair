package tests

import (
	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/testing"
)

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	revel.AppLog.Info("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) After() {
	revel.AppLog.Info("Tear down")
}
