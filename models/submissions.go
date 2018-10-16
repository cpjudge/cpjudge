package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gobuffalo/buffalo/binding"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
)

type Submission struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	UserID         uuid.UUID    `json:"user_id" db:"user_id"`
	QuestionID     uuid.UUID    `json:"question_id" db:"question_id"`
	SubmissionFile binding.File `json:"submission_file" db:"-" form:"SubmissionFile"`
	SubmissionPath string       `json:"submission_path" db:"submission_path"`
}

type Submissions []Submission

func (s *Submission) AfterSave(tx *pop.Connection) error {

	if !s.SubmissionFile.Valid() {
		fmt.Printf("\n\nFile is not valid\n\n")
		return nil
	}
	dir := filepath.Join("..", "submissions")

	fmt.Println(dir)

	fmt.Printf("\n\nFile path %q\n\n", dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.WithStack(err)
	}

	submissionFileName := "submission_" + s.ID.String() + ".c"
	f, err := os.Create(filepath.Join(dir, submissionFileName))
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	_, err = io.Copy(f, s.SubmissionFile)
	if err != nil {
		return errors.WithStack(err)
	}

	return err
}
