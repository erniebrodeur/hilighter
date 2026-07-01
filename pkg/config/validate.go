package config

// Validate checks a config for structural and semantic issues.
//
// Validation will eventually enforce required fields, accepted scopes, and
// coherence between rule declarations and theme references.
func Validate(_ Config) error {
	return nil
}
