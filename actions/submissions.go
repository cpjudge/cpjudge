package actions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpjudge/cpjudge/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
)

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

// SubmissionsIndex default implementation.
func SubmissionsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := c.Value("current_user").(*models.User)

	submissions := &models.Submissions{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.Order("created_at desc").PaginateFromParams(c.Params())
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
	submission.Status = "Pending"

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
			submission.SubmissionPath = "../submissions/submission_" + submission.ID.String() + ".c"
			tx.ValidateAndSave(&submission)
			fmt.Print("Success!\n")
			//fmt.Printf("%v\n", user)
		}
	}

	query = tx.Where("status = 'Pending'")
	submissions = []models.Submission{}
	err = query.All(&submissions)
	if err != nil {
		fmt.Print("\n\nAll filled!!\n\n")
		fmt.Printf("%v\n", err)
	} else {
		for i := 0; i < len(submissions); i++ {
			fmt.Println("\n\nFound!!!\n\n")
			submission := submissions[i]
			submission.Status = EvaluateSubmission(c, submission.SubmissionPath, questionID)
			tx.ValidateAndSave(&submission)
			fmt.Print("Success!\n")
			//fmt.Printf("%v\n", user)
		}
	}

	// and redirect to the index page
	return c.Redirect(302, "/submissions/detail/%s", submission.ID)
}

// SubmissionsDetail default implementation.
func SubmissionsDetail(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	submission := &models.Submission{}
	if err := tx.Find(submission, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("submission", submission)
	return c.Render(200, r.HTML("submissions/detail.html"))
}

func EvaluateSubmission(c buffalo.Context, submissionPath string, questionId uuid.UUID) string {
	//query := tx.Where("question_id = (?)", questionId.String())

	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	question := &models.Question{}
	err := tx.Find(question, questionId)
	if err != nil {
		fmt.Print("\n\nNo question with this id found!!\n\n")
		return "Some error occured"
	} else {
		testCasesPath := question.TestCasesPath

		// Compile code
		compile := exec.Command("gcc", submissionPath)
		err := compile.Run()
		if err != nil {
			return "Compilation error"
		}

		// Test cases inputs
		inputTestCaseFiles, err := ioutil.ReadDir(testCasesPath + "/inputs/")
		if err != nil {
			fmt.Println("\n\nInput test cases could not be read\n\n")
			return "Some error occurred"
		}

		// Test cases answers
		answerTestCaseFiles, err := ioutil.ReadDir(testCasesPath + "/answers/")
		if err != nil {
			fmt.Println("\n\nAnswers of test cases could not be read\n\n")
			return "Some error occurred"
		}

		// Number of input test cases
		numInputTestCaseFiles := len(inputTestCaseFiles)

		// Number of test cases answers
		numAnswersTestCaseFiles := len(answerTestCaseFiles)

		if numInputTestCaseFiles != numAnswersTestCaseFiles {
			fmt.Printf("Number of input files and answer files in test case folder are not equal.")
			return "Some error occurred"
		}

		for i := 0; i < numInputTestCaseFiles; i++ {

			input, err := os.Open(testCasesPath + "/inputs/" + inputTestCaseFiles[i].Name())
			if err != nil {
				fmt.Println("\n\nCould not read input test case file\n\n")
				return "Some error occurred"
			}

			cmd := exec.Command("./a.out")
			cmd.Stdin = input
			var out bytes.Buffer
			cmd.Stdout = &out

			if err := cmd.Start(); err != nil {
				log.Println(err)
				return "Runtime Error"
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case <-time.After(3 * time.Second): //TODO : Take time limit from question
				if err := cmd.Process.Kill(); err != nil {
					log.Println("failed to kill process: ", err)
					return "System Error"
				}
				log.Println("process killed as timeout reached")
				return "Time Limit Exceeded"
			case err := <-done:
				if err != nil {
					log.Println("process finished with error = %v", err)
				}
				log.Print("process finished successfully")

				outputString := out.String()
				outputString = strings.Trim(outputString, "\n")
				// fmt.Printf("Output: %q\n", outputString)

				dat, err := ioutil.ReadFile(testCasesPath + "/answers/" + answerTestCaseFiles[i].Name())
				if err != nil {
					fmt.Println("\n\nCould not read answer test case file")
					return "Some error occurred"
				}

				answerString := string(dat)
				answerString = strings.Trim(answerString, "\n")
				//fmt.Printf("%q\n", answerString)

				if strings.Compare(outputString, answerString) != 0 {
					return "Wrong answer"
				}
			}
		}
	}
	return "Correct answer"
}
