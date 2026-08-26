package core

import "uuid"

func NewID() string {
	return uuid.NewV7().String()
}
