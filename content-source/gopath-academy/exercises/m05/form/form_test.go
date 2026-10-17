package form

import (
	"errors"
	"slices"
	"testing"
)

func TestValid(t *testing.T) {
	err := Validate(Signup{Username: "ana", Email: "ana@example.com", Password: "correct horse"})
	if err != nil {
		t.Errorf("valid form returned %v", err)
	}
	if Fields(nil) != nil {
		t.Error("Fields(nil) should be nil")
	}
}

func TestAllProblemsReported(t *testing.T) {
	err := Validate(Signup{Email: "not-an-email", Password: "short"})
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "username is required\nemail is invalid\npassword is too short"
	if err.Error() != want {
		t.Errorf("Error() =\n%s\nwant\n%s", err, want)
	}
	if got := Fields(err); !slices.Equal(got, []string{"username", "email", "password"}) {
		t.Errorf("Fields = %v", got)
	}
	for _, sentinel := range []error{ErrRequired, ErrInvalid, ErrTooShort} {
		if !errors.Is(err, sentinel) {
			t.Errorf("errors.Is(err, %v) = false", sentinel)
		}
	}
}

func TestEmptyFieldsAreRequired(t *testing.T) {
	err := Validate(Signup{})
	if got := Fields(err); !slices.Equal(got, []string{"username", "email", "password"}) {
		t.Errorf("Fields = %v", got)
	}
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrTooShort) {
		t.Error("empty fields should only report ErrRequired")
	}
}

func TestSingleProblem(t *testing.T) {
	err := Validate(Signup{Username: "bo", Email: "bo@example.com", Password: "1234567"})
	if got := Fields(err); !slices.Equal(got, []string{"password"}) {
		t.Errorf("Fields = %v", got)
	}
	var fe *FieldError
	if !errors.As(err, &fe) || fe.Field != "password" || !errors.Is(fe, ErrTooShort) {
		t.Errorf("errors.As found %+v", fe)
	}
}
