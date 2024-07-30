package message

import (
	"net/mail"
	"strings"
)

func ExtractPlusAddress(email string) (string, error) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	parts := strings.Split(address.Address, "@")

	localParts := strings.Split(parts[0], "+")
	if len(localParts) < 2 {
		return "", nil
	}

	return localParts[1], nil
}
