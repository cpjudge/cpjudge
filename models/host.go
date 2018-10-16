package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Host struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Hostname        string    `json:"hostname" db:"hostname"`
	Email           string    `json:"email" db:"email"`
	Admin           bool      `json:"admin" db:"admin"`
	PasswordHash    string    `json:"-" db:"password_hash"`
	Password        string    `json:"-" db:"-"`
	PasswordConfirm string    `json:"-" db:"-"`
}

// String is not required by pop and may be deleted
func (h Host) String() string {
	jh, _ := json.Marshal(h)
	return string(jh)
}

// Hosts is not required by pop and may be deleted
type Hosts []Host

// String is not required by pop and may be deleted
func (h Hosts) String() string {
	jh, _ := json.Marshal(h)
	return string(jh)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (h *Host) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: h.Hostname, Name: "Hostname"},
		&validators.StringIsPresent{Field: h.Email, Name: "Email"},
		&validators.EmailIsPresent{Name: "Email", Field: h.Email},
		&validators.StringIsPresent{Field: h.Hostname, Name: "Hostname"},
		&validators.StringIsPresent{Field: h.Password, Name: "Password"},
		&validators.StringsMatch{Name: "Password", Field: h.Password, Field2: h.PasswordConfirm, Message: "Passwords do not match."},
		&HostnameNotTaken{Name: "Hostname", Field: h.Hostname, tx: tx},
		&HostEmailNotTaken{Name: "Email", Field: h.Email, tx: tx},
	), nil
}

type HostnameNotTaken struct {
	Name  string
	Field string
	tx    *pop.Connection
}

func (v *HostnameNotTaken) IsValid(errors *validate.Errors) {
	query := v.tx.Where("hostname = ?", v.Field)
	queryHost := Host{}
	err := query.First(&queryHost)
	if err == nil {
		// found a host with same hostname
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The hostname %s is not available.", v.Field))
	}
}

type HostEmailNotTaken struct {
	Name  string
	Field string
	tx    *pop.Connection
}

// IsValid performs the validation check for unique emails
func (v *HostEmailNotTaken) IsValid(errors *validate.Errors) {
	query := v.tx.Where("email = ?", v.Field)
	queryHost := Host{}
	err := query.First(&queryHost)
	if err == nil {
		// found a host with the same email
		errors.Add(validators.GenerateKey(v.Name), "An account with that email already exists.")
	}
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (h *Host) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (h *Host) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Create validates and creates a new Host.
func (h *Host) Create(tx *pop.Connection) (*validate.Errors, error) {
	h.Email = strings.ToLower(h.Email)
	h.Admin = false
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(h.Password), bcrypt.DefaultCost)
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	h.PasswordHash = string(pwdHash)
	return tx.ValidateAndCreate(h)
}

// Authorize checks host's password for logging in
func (u *Host) Authorize(tx *pop.Connection) error {
	err := tx.Where("email = ?", strings.ToLower(u.Email)).First(u)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			// couldn't find an host with that email address
			return errors.New("Host not found.")
		}
		return errors.WithStack(err)
	}
	// confirm that the given password matches the hashed password from the db
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(u.Password))
	if err != nil {
		return errors.New("Invalid password.")
	}
	return nil
}
