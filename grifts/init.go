package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/cpjudge/cpjudge/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
