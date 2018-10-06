package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/shashankp/cpjudge/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
