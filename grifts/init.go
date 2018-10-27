package grifts

import (
	"github.com/cpjudge/cpjudge/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
