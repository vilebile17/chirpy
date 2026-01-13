package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	type Input struct {
		userID      uuid.UUID
		buildSecret string
		testSecret  string
		expiresIn   time.Duration
	}
	type Output struct {
		userID            uuid.UUID
		errorInCreation   bool
		errorInValidation bool
	}

	inputs := []Input{
		{
			uuid.New(),
			"I'm in the thick of it",
			"I'm in the thick of it",
			time.Duration(time.Minute),
		},
		{
			uuid.New(),
			"This should expire",
			"This should expire",
			time.Duration(-time.Hour),
		},
		{
			uuid.New(),
			"This is top secret secret",
			"This isn't the right secret",
			time.Duration(time.Minute),
		},
	}
	outputs := []Output{
		{
			inputs[0].userID,
			false,
			false,
		},
		{
			uuid.Nil,
			false,
			true,
		},
		{
			uuid.Nil,
			false,
			true,
		},
	}

	for i := range inputs {
		signedString, err := MakeJWT(inputs[i].userID, inputs[i].buildSecret, inputs[i].expiresIn)
		if err != nil {
			if !outputs[i].errorInCreation {
				t.Fatalf("An error occured unexpectedly during creation of JWT: %s", err)
			}
			continue
		}
		if outputs[i].errorInCreation {
			t.Fatalf("An error was expected but not found: '%s', '%s', '%s'", inputs[i].userID, inputs[i].buildSecret, inputs[i].expiresIn)
		}

		id, err := ValidateJWT(signedString, inputs[i].testSecret)
		if err != nil {
			if !outputs[i].errorInValidation {
				t.Fatalf("An error occured unexpectedly during validation of JWT: %s", err)
			}
			continue
		}
		if outputs[i].errorInValidation {
			t.Fatalf("An error was expected but not found: '%s', '%s'", inputs[i].testSecret, signedString)
		}

		if id != outputs[i].userID {
			t.Fatalf("Subject ID wasn't what was expected: %v != %v", id, outputs[i].userID)
		}
	}
}

func TestBadSignedString(t *testing.T) {
	_, err := ValidateJWT("not.a.jwt", "secret")
	if err == nil {
		t.Fatalf("Expected an error to occur when trying to validate 'not.a.jwt' with 'secret'")
	}
}
