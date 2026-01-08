package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "SuperS3cureP4s5wOrd"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("An error occured whilst hashing the password: %s", err)
	}
	if hash == "" || password == hash {
		t.Fatalf("The hashing process didn't go as expected. password: %v, hash: %v", password, hash)
	}
}

func TestCorrectCheckPasswordHash(t *testing.T) {
	password := "1L1K3G00DP455W0RD5"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("An error occured whilst hashing the password: %s", err)
	}

	b, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("An error occured whilst checking the hash: %s", err)
	}
	if !b {
		t.Fatalf("The password and hash didn't match... password: %v, hash: %v", password, hash)
	}
}

func TestWrongCheckPasswordHash(t *testing.T) {
	password := "1L1K3G00DP455W0RD5"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("An error occured whilst hashing the password: %s", err)
	}

	b, err := CheckPasswordHash("T0T4LLYTH3R!GHTP455WoRD", hash)
	if err != nil {
		t.Fatalf("An error occured whilst checking the hash: %s", err)
	}
	if b {
		t.Fatalf("That's weird, the two passwords matched (They should have been different) password: %v, hash: %v", password, hash)
	}
}
