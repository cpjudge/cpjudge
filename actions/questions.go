package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
	"github.com/shashankp/cpjudge/models"
)

// QuestionsIndex default implementation.
func QuestionsIndex(c buffalo.Context) error {
	return c.Render(200, r.HTML("questions/index.html"))
}

// QuestionsCreate default implementation.
func QuestionsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	host := c.Value("current_host").(*models.Host)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}
	// make sure the contest was made by the logged in host
	if host.ID != contest.HostID {
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/contests/detail/%s", contest.ID)
	}
	c.Set("contest", contest)
	c.Set("question", &models.Question{})
	return c.Render(200, r.HTML("questions/create"))
}

func QuestionsCreatePost(c buffalo.Context) error {
	fmt.Printf("\n\nQuestion is being created\n\n")
	question := &models.Question{}
	//host := c.Value("current_host").(*models.User)
	if err := c.Bind(question); err != nil {
		return errors.WithStack(err)
	}
	tx := c.Value("tx").(*pop.Connection)
	//question.AuthorID = host.ID
	contestID, err := uuid.FromString(c.Param("cid"))
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(question)

	question.ContestID = contestID

	//testCasesFile, err := c.File("someFile")
	if err != nil {
		return errors.WithStack(err)
	}

	// f, err := c.File("TestCasesZipFile")
	// if err != nil {
	// 	return errors.WithStack(err)
	// }

	// question.TestCasesZipFile = f

	verrs, err := tx.ValidateAndCreate(question)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Flash().Add("danger", "There was an error adding your question.")
		return c.Redirect(302, "/contests/detail/%s", c.Param("cid"))
	}
	c.Flash().Add("success", "Question added successfully.")

	tx = c.Value("tx").(*pop.Connection)
	query := tx.Where("testcases_path = ''")
	questions := []models.Question{}
	err = query.All(&questions)
	if err != nil {
		fmt.Print("\n\nAll filled!!\n\n")
		fmt.Printf("%v\n", err)
	} else {
		for i := 0; i < len(questions); i++ {
			fmt.Println("\n\nFound!!!\n\n")
			question := questions[i]
			question.TestCasesPath = "../testcases/testcase_" + question.ID.String()
			tx.ValidateAndSave(&question)
			fmt.Print("Success!\n")
			//fmt.Printf("%v\n", user)
		}
	}

	return c.Redirect(302, "/contests/detail/%s", c.Param("cid"))
}

// QuestionsEdit default implementation.
func QuestionsEdit(c buffalo.Context) error {
	return c.Render(200, r.HTML("questions/edit.html"))
}

// QuestionsDelete default implementation.
func QuestionsDelete(c buffalo.Context) error {
	return c.Render(200, r.HTML("questions/delete.html"))
}

// QuestionsDetail default implementation.
func QuestionsDetail(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	question := &models.Question{}
	if err := tx.Find(question, c.Param("qid")); err != nil {
		return c.Error(404, err)
	}
	contest := &models.Contest{}
	if err := tx.Find(contest, question.ContestID); err != nil {
		return c.Error(404, err)
	}
	c.Set("question", question)
	c.Set("contest", contest)
	return c.Render(200, r.HTML("questions/detail.html"))
}
