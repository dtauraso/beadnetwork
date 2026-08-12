// String-tokenizing helpers for the parsed InputLayoutFingerprint value (input_layout_parse.go
// finds and AST-extracts the raw string; these split it into the marker-delimited lists and
// derive TS names from them).
package main

import (
	"fmt"
	"strings"
)

// fpListToken mirrors input_fingerprint.go's parseFPList body-extraction, returning the raw
// comma-joined token text (without splitting into a slice).
func fpListToken(fp, marker string) string {
	i := strings.Index(fp, marker)
	if i < 0 {
		return ""
	}
	rest := fp[i+len(marker):]
	if sp := strings.IndexByte(rest, ' '); sp >= 0 {
		rest = rest[:sp]
	}
	return rest
}

// fpList mirrors input_fingerprint.go's parseFPList exactly (returns nil if marker absent).
func fpList(fp, marker string) []string {
	tok := fpListToken(fp, marker)
	if tok == "" {
		return nil
	}
	return strings.Split(tok, ",")
}

// unquoteGoString strips the surrounding double quotes from a Go string-literal AST value.
// input_fingerprint.go's fingerprint is a plain double-quoted literal (no escapes), so a simple
// trim suffices; a raw/backtick literal or one containing escapes is rejected loudly rather
// than mis-parsed.
func unquoteGoString(lit string) (string, error) {
	if len(lit) < 2 || lit[0] != '"' || lit[len(lit)-1] != '"' {
		return "", fmt.Errorf("InputLayoutFingerprint literal %q is not a plain double-quoted string", lit)
	}
	body := lit[1 : len(lit)-1]
	if strings.ContainsRune(body, '\\') {
		return "", fmt.Errorf("InputLayoutFingerprint literal contains an escape sequence, which this generator does not decode: %q", lit)
	}
	return body, nil
}

// kindConstName derives the TS export name for a kebab-case kind name from the fingerprint's
// kinds= token, e.g. "save" -> "IN_KIND_SAVE", "raw-input" -> "IN_KIND_RAW_INPUT".
func kindConstName(kind string) string {
	upper := strings.ToUpper(strings.ReplaceAll(kind, "-", "_"))
	return "IN_KIND_" + upper
}
