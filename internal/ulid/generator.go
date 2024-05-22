package ulid

import (
	"crypto/rand"
	"fmt"

	"github.com/oklog/ulid/v2"
)

func Generate() (string, error) {
	ulid, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ULID: %w", err)
	}

	return ulid.String(), nil
}

func ValidateUlid(s string) error {
	_, err := ulid.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid ULID: %w", err)
	}

	return nil
}
