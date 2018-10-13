package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/pkg/errors"
	"github.com/shashankp/cpjudge/models"
)

// HostRegisterGet displays a register form
func HostsRegisterGet(c buffalo.Context) error {
	// Make host available inside the html template
	c.Set("host", &models.Host{})
	return c.Render(200, r.HTML("hosts/register.html"))
}

func HostHomePage(c buffalo.Context) error {
	c.Set("host", &models.Host{})
	return c.Render(200, r.HTML("hosts/home.html"))
}

// HostsRegisterPost adds a host to the DB. This function is mapped to the
// path POST /accounts/register
func HostsRegisterPost(c buffalo.Context) error {
	// Allocate an empty Host
	host := &models.Host{}
	// Bind host to the html form elements
	if err := c.Bind(host); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := host.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		// Make host available inside the html template
		c.Set("host", host)
		// Make the errors available inside the html template
		c.Set("errors", verrs.Errors)
		// Render again the register.html template that the host can
		// correct the input.
		return c.Render(422, r.HTML("hosts/register.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Account created successfully.")
	// and redirect to the home page
	return c.Redirect(302, "/")
}

// HostsLoginGet displays a login form
func HostsLoginGet(c buffalo.Context) error {
	return c.Render(200, r.HTML("hosts/login"))
}

// HostsLoginPost logs in a host.
func HostsLoginPost(c buffalo.Context) error {
	host := &models.Host{}
	// Bind the host to the html form elements
	if err := c.Bind(host); err != nil {
		return errors.WithStack(err)
	}
	tx := c.Value("tx").(*pop.Connection)
	err := host.Authorize(tx)
	if err != nil {
		c.Set("host", host)
		verrs := validate.NewErrors()
		verrs.Add("Login", "Invalid email or password.")
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("hosts/login"))
	}
	c.Session().Set("current_host_id", host.ID)
	c.Flash().Add("success", "Welcome back!")
	return c.Redirect(302, "/contests/host_index")
}

// HostsDashboard displays host's dashboard
func HostsDashboard(c buffalo.Context) error {
	return c.Render(200, r.HTML("hosts/dashboard.html"))
}

// HostsLogout clears the session and logs out the host.
func HostsLogout(c buffalo.Context) error {
	c.Session().Clear()
	c.Flash().Add("success", "Goodbye!")
	return c.Redirect(302, "/")
}

// SetCurrentHost attempts to find a host based on the current_host_id
// in the session. If one is found it is set on the context.
func SetCurrentHost(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_host_id"); uid != nil {
			u := &models.Host{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_host", u)
		}
		return next(c)
	}
}

// HostRequired requires a host to be logged in before accessing a route.
func HostRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		_, ok := c.Value("current_host").(*models.Host)
		if ok {
			return next(c)
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}
