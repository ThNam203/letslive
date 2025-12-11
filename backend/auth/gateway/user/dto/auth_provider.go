package dto

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
	ProviderLocal  AuthProvider = "local"
)

