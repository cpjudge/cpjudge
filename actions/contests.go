package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
	"github.com/shashankp/cpjudge/models"
)

// ContestsIndex default implementation.
func ContestsUserIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	contests := &models.Contests{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	// Retrieve all Contests from the DB
	if err := q.All(contests); err != nil {
		return errors.WithStack(err)
	}
	// Make contests available inside the html template
	c.Set("contests", contests)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.HTML("contests/index.html"))
}

// ContestsIndex default implementation.
func ContestsHostIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	host := c.Value("current_host").(*models.Host)

	contests := &models.Contests{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	// Retrieve all Contests from the DB
	if err := q.BelongsTo(host).All(contests); err != nil {
		return errors.WithStack(err)
	}
	// Make contests available inside the html template
	c.Set("contests", contests)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.HTML("contests/index.html"))
}

func ContestsCreateGet(c buffalo.Context) error {
	c.Set("contest", &models.Contest{})
	return c.Render(200, r.HTML("contests/create"))
}

func ContestsCreatePost(c buffalo.Context) error {
	// Allocate an empty Contest
	contest := &models.Contest{}
	host := c.Value("current_host").(*models.Host)
	// Bind contest to the html form elements
	if err := c.Bind(contest); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	contest.HostID = host.ID
	verrs, err := tx.ValidateAndCreate(contest)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("contest", contest)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("contests/create"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "New contest added successfully.")
	// and redirect to the index page
	return c.Redirect(302, "host_index")
}

// ContestsDetail displays a single contest.
func ContestsDetail(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}
	host := &models.Host{}
	if err := tx.Find(host, contest.HostID); err != nil {
		return c.Error(404, err)
	}
	c.Set("contest", contest)
	c.Set("host", host)

	question := &models.Question{}
	c.Set("question", question)
	questions := models.Questions{}
	qPage := tx.PaginateFromParams(c.Params())
	if err := qPage.BelongsTo(contest).All(&questions); err != nil {
		return errors.WithStack(err)
	}
	for i := 0; i < len(questions); i++ {
		contest := models.Contest{}
		if err := tx.Find(&contest, questions[i].ContestID); err != nil {
			return c.Error(404, err)
		}
		questions[i].Contest = contest
	}
	c.Set("questions", questions)
	c.Set("qPagination", qPage.Paginator)
	return c.Render(200, r.HTML("contests/detail"))
}

// ContestsEditGet displays a form to edit the contest.
func ContestsEditGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("contest", contest)
	return c.Render(200, r.HTML("contests/edit.html"))
}

// ContestsEditContest updates a contest.
func ContestsEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}
	if err := c.Bind(contest); err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(contest)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("contest", contest)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("contests/edit.html"))
	}
	c.Flash().Add("success", "Contest was updated successfully.")
	return c.Redirect(302, "/contests/detail/%s", contest.ID)
}

func ContestsDelete(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}
	if err := tx.Destroy(contest); err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "Contest was successfully deleted.")
	return c.Redirect(302, "/contests/host_index")
}
