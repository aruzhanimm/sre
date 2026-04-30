package service

import "github.com/google/uuid"

func parseUUIDFromString(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
