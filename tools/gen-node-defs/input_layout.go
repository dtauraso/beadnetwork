// input-layout-gen.ts emitter: mirrors the nodes/Wiring InputLayoutFingerprint const
// (input_fingerprint.go) into tools/topology-vscode/src/schema/input-layout-gen.ts.
//
// The const is located by SCANNING the package, never by a hardcoded filename
// (memory/feedback_guards_hardcoding_single_file_break_on_split.md — it used to live in
// input_codec.go and moved when that file was split by job).
//
// Go's InputLayoutFingerprint string is the single source of truth for the TS→Go input
// record layout (record kind bytes + enum orderings). Go derives its own kind consts and
// enum arrays from that same string via parseFPList (see input_fingerprint.go) — this generator
// derives the TS-side equivalents the same way, so the two languages can never carry a
// hand-copied fingerprint that silently drifts. Only the fingerprint STRING and its
// directly-derived constants/arrays are generated here; the codec functions
// (encode*/decode*, ByteWriter/ByteReader, etc.) stay hand-written in input-encode.ts,
// input-decode.ts, byte-writer.ts, byte-reader.ts, and input-attrs.ts.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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

// writeInputLayout emits tools/topology-vscode/src/schema/input-layout-gen.ts from the
// parsed fingerprint. Value-identical to what input-layout.ts previously hand-carried for
// these constants; only structure/provenance comments differ.
func writeInputLayout(outPath string, fp *inputLayoutFingerprint) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	fmt.Fprintln(w, `// GENERATED by tools/gen-node-defs — do not edit.`)
	fmt.Fprintln(w, `// Source: nodes/Wiring/input_fingerprint.go's InputLayoutFingerprint const.`)
	fmt.Fprintln(w, `// Regenerate with: npm run gen:node-defs`)
	fmt.Fprintln(w, `//`)
	fmt.Fprintln(w, `// Go is the single source of the TS<->Go input-record layout: Go's InputLayoutFingerprint`)
	fmt.Fprintln(w, `// string is parsed here (the same tokenization Go's own parseFPList uses) to emit the`)
	fmt.Fprintln(w, `// fingerprint string and its directly-derived constants/arrays for TS. There is no`)
	fmt.Fprintln(w, `// separate TS-side fingerprint to hand-keep in lockstep, so the two languages cannot`)
	fmt.Fprintln(w, `// drift apart on record kinds or enum orderings.`)
	fmt.Fprintln(w)

	fmt.Fprintln(w, `// INPUT_LAYOUT_FINGERPRINT: `+fp.raw)
	fmt.Fprintf(w, "export const INPUT_LAYOUT_FINGERPRINT =\n  %q;\n\n", fp.raw)

	fmt.Fprintln(w, `// Record kind bytes (first byte of every record). Must match input_fingerprint.go.`)
	for _, name := range fp.kindNames {
		fmt.Fprintf(w, "export const %s = %s;\n", kindConstName(name), fp.kindValues[name])
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, `// Enum orderings (u8 index -> string), shared with input_fingerprint.go.`)
	writeTSArray(w, "IN_EVENT_KINDS", fp.eventKinds)
	writeTSArray(w, "IN_HIT_KINDS", fp.hitKinds)
	// EDIT_UPDATE_KINDS_START / _END bound tools/check-edit-op-parity.sh's axis-2 extraction
	// (the 3rd update-kind parity source, alongside messages.ts EditMsg + stdin_reader.go
	// applyUpdate). Keep the sentinel lines immediately around IN_UPDATE_KINDS.
	fmt.Fprintln(w, `// EDIT_UPDATE_KINDS_START`)
	writeTSArray(w, "IN_UPDATE_KINDS", fp.updateKinds)
	fmt.Fprintln(w, `// EDIT_UPDATE_KINDS_END`)
	writeTSArray(w, "IN_UPDATE_ATTRS", fp.updateAttrs)

	w.Flush()
	out := strings.TrimRight(buf.String(), "\n") + "\n"
	return os.WriteFile(outPath, []byte(out), 0644)
}

func writeTSArray(w *bufio.Writer, name string, values []string) {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	fmt.Fprintf(w, "export const %s = [%s] as const;\n", name, strings.Join(quoted, ", "))
}
