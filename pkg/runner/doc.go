// Package runner owns input acquisition for hilighter.
//
// It is responsible for deciding whether text comes from stdin or a spawned
// command, then streaming lines through the configured highlighter.
package runner
