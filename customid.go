// Custom-ID slug parser. Pattern-matching logic ported from the pat HTTP mux
// (https://github.com/bmizerany/pat), MIT-licensed.
package disgoplus

func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

func match(
	s string,
	f func(byte) bool,
	i int,
) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}

	if j < len(s) {
		next = s[j]
	}

	return s[i:j], next, j
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

// trySlug attempts to match customID against the slug pattern.
// Returns the extracted params and true if it matches.
// Slug params use the :name syntax: "LEADERBOARD/:page".
func trySlug(pattern, customID string) (map[string]string, bool) {
	p := make(map[string]string)

	var i, j int
	for i < len(customID) {
		switch {
		case j >= len(pattern):
			if pattern != "/" && len(pattern) > 0 &&
				pattern[len(pattern)-1] == '/' {
				return nil, true
			}

			return nil, false
		case pattern[j] == ':':
			var (
				name, val string
				nextc     byte
			)

			name, nextc, j = match(pattern, isAlnum, j+1)
			val, _, i = match(customID, matchPart(nextc), i)
			p[name] = val
		case customID[i] == pattern[j]:
			i++
			j++
		default:
			return nil, false
		}
	}

	if j != len(pattern) {
		return nil, false
	}

	return p, true
}
