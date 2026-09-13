package extraction

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidDomain   = errors.New("invalid domain")
	ErrInvalidURL      = errors.New("invalid url")
	ErrInvalidIP       = errors.New("invalid ip address")
	ErrInvalidCountry  = errors.New("invalid country")
	ErrInvalidLocation = errors.New("invalid location")

	// Pre-compiled regexes for strict syntactic validation
	domainLabelRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	tldRegex         = regexp.MustCompile(`^[a-z]{2,24}$`)
	multipleSpaces   = regexp.MustCompile(`\s+`)
)

// Standard ISO-3166-1 alpha-2 mapping for common countries
var countryToISO = map[string]string{
	"japan":          "JP",
	"jp":             "JP",
	"philippines":    "PH",
	"ph":             "PH",
	"indonesia":      "ID",
	"id":             "ID",
	"united states":  "US",
	"usa":            "US",
	"us":             "US",
	"turkey":         "TR",
	"türkiye":        "TR",
	"tr":             "TR",
	"taiwan":         "TW",
	"tw":             "TW",
	"mexico":         "MX",
	"mx":             "MX",
	"chile":          "CL",
	"cl":             "CL",
	"peru":           "PE",
	"pe":             "PE",
	"greece":         "GR",
	"gr":             "GR",
	"italy":          "IT",
	"it":             "IT",
	"new zealand":    "NZ",
	"nz":             "NZ",
	"iceland":        "IS",
	"is":             "IS",
	"afghanistan":    "AF",
	"af":             "AF",
	"ukraine":        "UA",
	"ua":             "UA",
	"united kingdom": "GB",
	"uk":             "GB",
	"gb":             "GB",
	"canada":         "CA",
	"ca":             "CA",
	"australia":      "AU",
	"au":             "AU",
	"china":          "CN",
	"cn":             "CN",
	"germany":        "DE",
	"de":             "DE",
	"france":         "FR",
	"fr":             "FR",
	"india":          "IN",
	"in":             "IN",
	"russia":         "RU",
	"ru":             "RU",
	"brazil":         "BR",
	"br":             "BR",
	"south africa":   "ZA",
	"za":             "ZA",
}

// NormalizeDomain performs strict RFC-compliant domain normalization and validation.
// It strips schemes, ports, and trailing dots, and verifies that the value is not an IP address.
func NormalizeDomain(raw string) (string, error) {
	d := strings.TrimSpace(raw)
	if d == "" {
		return "", ErrInvalidDomain
	}

	// Strip URL scheme if passed
	if strings.Contains(d, "://") {
		u, err := url.Parse(d)
		if err == nil && u.Host != "" {
			d = u.Host
		} else {
			parts := strings.SplitN(d, "://", 2)
			d = parts[1]
		}
	}

	// Remove trailing slash and path if attached
	if idx := strings.Index(d, "/"); idx != -1 {
		d = d[:idx]
	}

	// Strip port if present
	if strings.Contains(d, ":") {
		host, _, err := net.SplitHostPort(d)
		if err == nil {
			d = host
		} else {
			// Might be missing brackets on IPv6 or simple trailing colon
			idx := strings.LastIndex(d, ":")
			d = d[:idx]
		}
	}

	// Strip trailing dot
	d = strings.TrimSuffix(d, ".")
	d = strings.ToLower(d)

	if d == "" || len(d) > 253 {
		return "", ErrInvalidDomain
	}

	// Must NOT be an IP address (IPs belong to EntityTypeIP)
	if net.ParseIP(d) != nil {
		return "", ErrInvalidDomain
	}

	labels := strings.Split(d, ".")
	if len(labels) < 2 {
		return "", ErrInvalidDomain
	}

	// Validate each DNS label
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return "", ErrInvalidDomain
		}
		if !domainLabelRegex.MatchString(label) {
			return "", ErrInvalidDomain
		}
	}

	// TLD check (last label must be alphabetic and at least 2 characters)
	tld := labels[len(labels)-1]
	if !tldRegex.MatchString(tld) {
		return "", ErrInvalidDomain
	}

	return d, nil
}

// NormalizeURL canonicalizes a URL safely and conservatively:
// - Lowercases scheme and host
// - Strips default ports (:80 for http, :443 for https)
// - Strips trailing DNS dot from host
// - Strips client fragment
// - Preserves query parameter order and repeated query parameters
func NormalizeURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}

	host := strings.ToLower(u.Host)
	if host == "" {
		return "", ErrInvalidURL
	}

	// Strip default ports
	if scheme == "http" && strings.HasSuffix(host, ":80") {
		host = strings.TrimSuffix(host, ":80")
	} else if scheme == "https" && strings.HasSuffix(host, ":443") {
		host = strings.TrimSuffix(host, ":443")
	}

	// Strip trailing DNS dot from host if present
	host = strings.TrimSuffix(host, ".")

	u.Scheme = scheme
	u.Host = host
	u.Fragment = "" // Discard client fragment

	if u.Path == "" {
		u.Path = "/"
	}

	// Query parameters are preserved as-is: order and repeated parameters are preserved
	// to ensure semantically safe URL identity without altering query semantics.

	return u.String(), nil
}

// NormalizeIP validates and canonicalizes IPv4 and IPv6 addresses.
// IPv4 is output in standard dotted-decimal; IPv6 is output in canonical RFC 5952 format.
func NormalizeIP(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrInvalidIP
	}

	// If enclosed in brackets (e.g. [::1]:80) or has port
	if strings.HasPrefix(s, "[") && strings.Contains(s, "]") {
		host, _, err := net.SplitHostPort(s)
		if err == nil {
			s = host
		} else {
			s = strings.Trim(s, "[]")
		}
	} else if strings.Count(s, ":") == 1 {
		// Possible IPv4:port
		host, _, err := net.SplitHostPort(s)
		if err == nil {
			s = host
		}
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return "", ErrInvalidIP
	}

	// Check if valid IPv4
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String(), nil
	}

	// Canonical IPv6 representation
	return ip.String(), nil
}

// NormalizeCountry maps country names and codes to standardized uppercase ISO alpha-2 codes,
// or standardized clean uppercase/TitleCase if unmapped.
func NormalizeCountry(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrInvalidCountry
	}

	lower := strings.ToLower(s)
	if iso, exists := countryToISO[lower]; exists {
		return iso, nil
	}

	// If 2-letter alpha code, return uppercase
	if len(s) == 2 && isASCIIAlpha(s) {
		return strings.ToUpper(s), nil
	}

	// Return clean title-cased representation
	return toTitleCase(lower), nil
}

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			r := []rune(w)
			r[0] = unicode.ToUpper(r[0])
			words[i] = string(r)
		}
	}
	return strings.Join(words, " ")
}

// NormalizeLocation cleans and collapses location descriptions into a canonical lowercase representation.
func NormalizeLocation(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrInvalidLocation
	}

	cleaned := multipleSpaces.ReplaceAllString(s, " ")
	return strings.ToLower(cleaned), nil
}

func isASCIIAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) || r > unicode.MaxASCII {
			return false
		}
	}
	return true
}
