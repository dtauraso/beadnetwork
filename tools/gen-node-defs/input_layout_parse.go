// Parses the nodes/Wiring InputLayoutFingerprint const out of Go source. Emitting the
// TS mirror from the parsed result is input_layout.go's job.
package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

// inputLayoutFingerprint holds the parsed pieces of Go's InputLayoutFingerprint string.
type inputLayoutFingerprint struct {
	raw         string            // the full fingerprint string, verbatim
	kindNames   []string          // ordered kind names, e.g. ["save","raw-input","edit-update"]
	kindValues  map[string]string // kind name -> numeric value string, e.g. "save"->"4"
	eventKinds  []string
	hitKinds    []string
	updateKinds []string
	updateAttrs []string
}

// errFingerprintNotFound says "this particular file does not declare the const" — as
// opposed to "the const is there and is malformed". parseInputLayoutFingerprintDir needs
// the distinction so it can keep scanning past the package's other files while still
// surfacing a real parse error from the file that DOES declare it.
var errFingerprintNotFound = fmt.Errorf("InputLayoutFingerprint const not found")

// parseInputLayoutFingerprintDir finds the one file in the nodes/Wiring package that
// declares InputLayoutFingerprint and parses it. Scanning rather than naming a file is what
// keeps this generator working across a file split
// (memory/feedback_guards_hardcoding_single_file_break_on_split.md); finding no declaration
// at all is an ERROR, never a silent skip that would emit an empty layout.
func parseInputLayoutFingerprintDir(dir string) (*inputLayoutFingerprint, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, p := range paths {
		fp, err := parseInputLayoutFingerprint(p)
		if errors.Is(err, errFingerprintNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return fp, nil
	}
	return nil, fmt.Errorf("no file in %s declares InputLayoutFingerprint", dir)
}

// parseInputLayoutFingerprint reads one Go file via go/ast and extracts the
// InputLayoutFingerprint string-literal constant, then tokenizes it the same way Go's own
// parseFPList does (space-delimited tokens, comma-separated lists after "marker=").
func parseInputLayoutFingerprint(goPath string) (*inputLayoutFingerprint, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, goPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var raw string
	found := false
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if name.Name != "InputLayoutFingerprint" {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return nil, fmt.Errorf("InputLayoutFingerprint is not a string literal")
				}
				val, err := unquoteGoString(lit.Value)
				if err != nil {
					return nil, err
				}
				raw = val
				found = true
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("%w in %s", errFingerprintNotFound, goPath)
	}

	fp := &inputLayoutFingerprint{raw: raw, kindValues: map[string]string{}}

	kindsTok := fpListToken(raw, "kinds=")
	if kindsTok == "" {
		return nil, fmt.Errorf("InputLayoutFingerprint is missing the kinds= token")
	}
	for _, pair := range strings.Split(kindsTok, ",") {
		nv := strings.SplitN(pair, ":", 2)
		if len(nv) != 2 {
			return nil, fmt.Errorf("InputLayoutFingerprint kinds= entry %q is not name:value", pair)
		}
		fp.kindNames = append(fp.kindNames, nv[0])
		fp.kindValues[nv[0]] = nv[1]
	}

	fp.eventKinds = fpList(raw, "eventKinds=")
	fp.hitKinds = fpList(raw, "hitKinds=")
	fp.updateKinds = fpList(raw, "updateKinds=")
	fp.updateAttrs = fpList(raw, "updateAttrs=")
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"eventKinds=", fp.eventKinds},
		{"hitKinds=", fp.hitKinds},
		{"updateKinds=", fp.updateKinds},
		{"updateAttrs=", fp.updateAttrs},
	} {
		if len(e.list) == 0 {
			return nil, fmt.Errorf("InputLayoutFingerprint is missing the %s token", e.marker)
		}
	}
	return fp, nil
}

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
