package actions

import "github.com/gobuffalo/buffalo"

// Handle user related requests
func UserHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("user.html"))
}
