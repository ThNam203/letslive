package repository

import "testing"

func TestCreate(t *testing.T) {
	dbConn := GetDatabaseConnection()
	userRepo := NewUserRepository(*dbConn)
}
