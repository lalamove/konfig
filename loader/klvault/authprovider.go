package klvault

import "time"

// AuthProvider is the interface for a Vault authentication provider
type AuthProvider interface {
	Token() (string, time.Duration, error)
}
