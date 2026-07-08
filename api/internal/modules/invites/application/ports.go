package application

// TokenValidator validates invite tokens by hashing them for comparison.
// Implemented by the infrastructure-layer tokens.Manager.
type TokenValidator interface {
	Hash(token string) string
}
