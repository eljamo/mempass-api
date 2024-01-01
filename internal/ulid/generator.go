package ulid

import (
	"crypto/rand"

	"github.com/oklog/ulid/v2"
)

func Generate() (string, error) {
	ulid, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return "", err
	}

	return ulid.String(), nil
}

func ValidateUlid(s string) error {
	_, err := ulid.Parse(s)
	if err != nil {
		return err
	}

	return nil
}
