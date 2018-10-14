package models

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gobuffalo/buffalo/binding"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
)

type Question struct {
	ID               uuid.UUID    `json:"id" db:"id"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
	Title            string       `json:"title" db:"title"`
	Description      string       `json:"description" db:"description"`
	ContestID        uuid.UUID    `json:"contest_id" db:"contest_id"`
	Contest          Contest      `json:"-" db:"-"`
	TestCasesZipFile binding.File `json:"test_cases_zip_file" db:"-" form:"TestCasesZipFile"`
	TestCasesPath    string       `json:"testcases_path" db:"testcases_path"`
}

type Questions []Question

func (q *Question) AfterSave(tx *pop.Connection) error {

	if !q.TestCasesZipFile.Valid() {
		return nil
	}
	dir := filepath.Join(".", "testcases")

	fmt.Println(dir)

	fmt.Printf("\n\nFile path %q\n\n", dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.WithStack(err)
	}

	testCaseZipFileName := "testcase_" + q.ID.String() + ".zip"
	f, err := os.Create(filepath.Join(dir, testCaseZipFileName))
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	_, err = io.Copy(f, q.TestCasesZipFile)
	if err != nil {
		return errors.WithStack(err)
	}

	zipReader, err := zip.OpenReader(dir + "/" + testCaseZipFileName)
	if err != nil {
		fmt.Println("Zip file could not be opened")
		return errors.WithStack(err)
	}
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			fmt.Println("File could not be opened")
			return errors.WithStack(err)
		}
		defer zippedFile.Close()

		extractedFilePath := filepath.Join(dir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				fmt.Println("File could not be edited")
				return errors.WithStack(err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				fmt.Println("File could not be copied")
				return errors.WithStack(err)
			}
		}
	}

	//q.TestCasesPath = "./testcases/testcase_" + q.ID.String()

	err = os.Rename("./testcases/testcases", "./testcases/testcase_"+q.ID.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return err
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (q *Question) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: q.Title, Name: "Title"},
		&validators.StringIsPresent{Field: q.Description, Name: "Description"},
	), nil
}
