package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
	"github.com/shashankp/cpjudge/models"
)

// SubmissionsIndex default implementation.
func SubmissionsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := c.Value("current_user").(*models.User)

	submissions := &models.Submissions{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	// Retrieve all Submissions from the DB
	if err := q.BelongsTo(user).All(submissions); err != nil {
		return errors.WithStack(err)
	}
	// Make submissions available inside the html template
	c.Set("submissions", submissions)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.HTML("submissions/index.html"))
}

func SubmissionsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	submission := &models.Submission{}
	question := &models.Question{}
	if err := tx.Find(question, c.Param("qid")); err != nil {
		return c.Error(404, err)
	}

	contest := &models.Contest{}
	if err := tx.Find(contest, question.ContestID); err != nil {
		return c.Error(404, err)
	}

	c.Set("question", question)
	c.Set("submission", submission)
	c.Set("contest", contest)
	return c.Render(200, r.HTML("submissions/create"))
}

func SubmissionsCreatePost(c buffalo.Context) error {
	// Allocate an empty Submission
	fmt.Println("\n\nInside submission create\n\n")
	submission := &models.Submission{}
	user := c.Value("current_user").(*models.User)
	// Bind submission to the html form elements
	if err := c.Bind(submission); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	submission.UserID = user.ID
	questionID, err := uuid.FromString(c.Param("qid"))
	if err != nil {
		return errors.WithStack(err)
	}

	submission.QuestionID = questionID

	verrs, err := tx.ValidateAndCreate(submission)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("submission", submission)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("submissions/create"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Your code has been submitted. It is being evaluated now. Please wait.")

	query := tx.Where("submission_path = ''")
	submissions := []models.Submission{}
	err = query.All(&submissions)
	if err != nil {
		fmt.Print("\n\nAll filled!!\n\n")
		fmt.Printf("%v\n", err)
	} else {
		for i := 0; i < len(submissions); i++ {
			fmt.Println("\n\nFound!!!\n\n")
			submission := submissions[i]
			submission.SubmissionPath = "../submission/submission_" + submission.ID.String()
			tx.ValidateAndSave(&submission)
			fmt.Print("Success!\n")
			//fmt.Printf("%v\n", user)
		}
	}

	// and redirect to the index page
	return c.Redirect(302, "/submissions/index")
}

// SubmissionsDetail default implementation.
func SubmissionsDetail(c buffalo.Context) error {
	return c.Render(200, r.HTML("submissions/detail.html"))
}
