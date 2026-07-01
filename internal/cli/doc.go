// Package cli wires flags and runtime decisions into the hilighter executable.
//
// The package is intentionally thin. It should stay focused on translating user
// input into calls into the underlying packages rather than owning highlighting
// behavior itself.
package cli
