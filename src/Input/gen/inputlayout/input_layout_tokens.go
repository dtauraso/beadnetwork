package inputlayout

import (
	"fmt"
	"strings"
)

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

func fpList(fp, marker string) []string {
	tok := fpListToken(fp, marker)
	if tok == "" {
		return nil
	}
	return strings.Split(tok, ",")
}

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

func kindConstName(kind string) string {
	upper := strings.ToUpper(strings.ReplaceAll(kind, "-", "_"))
	return "IN_KIND_" + upper
}
