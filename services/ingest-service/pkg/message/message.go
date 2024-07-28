package message

import "strings"

func ExtractPlusAddressTag(email string) string {
	// Split the email address into the local and domain parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	// Split the local part into the user and tag parts
	localParts := strings.Split(parts[0], "+")
	if len(localParts) < 2 {
		return ""
	}

	// Return the tag part
	return localParts[1]
}
