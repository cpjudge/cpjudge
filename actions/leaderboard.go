package actions

import (
	"sort"

	"github.com/cpjudge/cpjudge/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

type SubmissionCount struct {
	// User     uuid.UUID `json:"user"`
	Username string `json:"username"`
	Correct  int    `json:"correct" default:"0"`
	Wrong    int    `json:"wrong" default:"0"`
}

// LeaderboardDisplay default implementation.
func LeaderboardDisplay(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	contest := &models.Contest{}
	if err := tx.Find(contest, c.Param("cid")); err != nil {
		return c.Error(404, err)
	}

	submissions := models.Submissions{}
	if err := tx.BelongsTo(contest).All(&submissions); err != nil {
		return errors.WithStack(err)
	}

	userMap := make(map[string]*SubmissionCount)
	var Leaderboard []SubmissionCount

	for _, submission := range submissions {
		user := &models.User{}
		if err := tx.Find(user, submission.UserID); err != nil {
			return errors.WithStack(err)
		}

		if _, ok := userMap[user.Username]; !ok {
			userMap[user.Username] = new(SubmissionCount)
		}
		if submission.Status == "Correct Answer" {
			userMap[user.Username].Correct++
		} else if submission.Status == "Runtime Error" ||
			submission.Status == "Wrong answer" ||
			submission.Status == "Time Limit Exceeded" {
			userMap[user.Username].Wrong++
		}
	}

	for key, value := range userMap {
		Leaderboard = append(Leaderboard, SubmissionCount{key, value.Correct, value.Wrong})
	}
	sort.Slice(Leaderboard[:], func(i, j int) bool {
		return (Leaderboard[i].Correct > Leaderboard[j].Correct ||
			(Leaderboard[i].Correct == Leaderboard[j].Correct &&
				Leaderboard[i].Wrong < Leaderboard[j].Wrong))
	})
	// Make submissions available inside the html template
	c.Set("leaderboard", Leaderboard)
	c.Set("contest_name", contest.Title)
	// return c.Render(200, r.HTML("submissions/index.html"))
	return c.Render(200, r.HTML("leaderboard/display.html"))
}
