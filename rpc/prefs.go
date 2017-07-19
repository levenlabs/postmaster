package rpc

import (
	"net/http"

	"github.com/levenlabs/postmaster/db"
)

type UpdatePrefsArgs struct {
	Email string `json:"email" validate:"email,nonzero"`
	Flags int64  `json:"flags" validate:"nonzero"`
}

// UpdatePrefs updates an email addresses email preferences
func (Postmaster) UpdatePrefs(r *http.Request, args *UpdatePrefsArgs, reply *SuccessResult) error {
	if err := db.StoreEmailFlags(args.Email, args.Flags); err != nil {
		return err
	}
	reply.Success = true
	return nil
}

// MovePrefsArgs defines the arguments of MovePrefs
type MovePrefsArgs struct {
	OldEmail string `json:"oldEmail" validate:"email,nonzero"`
	NewEmail string `json:"newEmail" validate:"email,nonzero"`
}

// MovePrefs moves a set of email preferences to a new email address
func (Postmaster) MovePrefs(r *http.Request, args *MovePrefsArgs, reply *SuccessResult) error {
	if err := db.MoveEmailPrefs(args.OldEmail, args.NewEmail); err != nil {
		return err
	}
	reply.Success = true
	return nil
}

type EmailArgs struct {
	Email string `json:"email" validate:"email,nonzero"`
}

type PrefsRes struct {
	Flags int64 `json:"flags"`
}

// GetPrefs returns an email address's email preferences
func (Postmaster) GetPrefs(r *http.Request, args *EmailArgs, reply *PrefsRes) error {
	prefs, err := db.GetEmailFlags(args.Email)
	if err != nil {
		return err
	}
	reply.Flags = prefs
	return nil
}
