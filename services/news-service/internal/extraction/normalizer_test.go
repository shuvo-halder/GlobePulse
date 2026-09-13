package extraction

import (
	"testing"
)

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard lowercase domain",
			input:   "example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "Uppercase domain",
			input:   "EXAMPLE.COM",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "Domain with port",
			input:   "example.com:8080",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "Domain with URL scheme and path",
			input:   "https://Sub.Domain.Co.UK/test/path?query=1",
			want:    "sub.domain.co.uk",
			wantErr: false,
		},
		{
			name:    "Domain with trailing dot",
			input:   "example.org.",
			want:    "example.org",
			wantErr: false,
		},
		{
			name:    "Whitespace trimming",
			input:   "   test.net   ",
			want:    "test.net",
			wantErr: false,
		},
		{
			name:    "IP address must be rejected as domain",
			input:   "192.168.1.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "IPv6 must be rejected as domain",
			input:   "::1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Localhost without TLD rejected",
			input:   "localhost",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Consecutive dots rejected",
			input:   "example..com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Leading hyphen label rejected",
			input:   "-bad.domain.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Trailing hyphen label rejected",
			input:   "bad-.domain.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeDomain(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NormalizeDomain(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("NormalizeDomain(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard HTTP URL",
			input:   "http://example.com/path",
			want:    "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "URL with port 80 stripped for HTTP",
			input:   "http://example.com:80/path",
			want:    "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "URL with port 443 stripped for HTTPS and case normalization",
			input:   "https://Example.COM:443/Path",
			want:    "https://example.com/Path",
			wantErr: false,
		},
		{
			name:    "Non-default ports preserved",
			input:   "https://example.com:8443/api",
			want:    "https://example.com:8443/api",
			wantErr: false,
		},
		{
			name:    "Trailing DNS dot on host stripped",
			input:   "https://example.com.:443/api",
			want:    "https://example.com/api",
			wantErr: false,
		},
		{
			name:    "Query parameter order preserved as semantically meaningful",
			input:   "https://example.com/api?z=3&a=1&m=2",
			want:    "https://example.com/api?z=3&a=1&m=2",
			wantErr: false,
		},
		{
			name:    "Repeated query parameters preserved in original sequence",
			input:   "https://example.com/search?category=intel&category=threats&page=1",
			want:    "https://example.com/search?category=intel&category=threats&page=1",
			wantErr: false,
		},
		{
			name:    "Fragment is discarded",
			input:   "https://example.com/page#section-1",
			want:    "https://example.com/page",
			wantErr: false,
		},
		{
			name:    "Missing path defaults to root slash",
			input:   "https://example.com",
			want:    "https://example.com/",
			wantErr: false,
		},
		{
			name:    "IPv6 host default port stripped",
			input:   "http://[2001:db8::1]:80/path",
			want:    "http://[2001:db8::1]/path",
			wantErr: false,
		},
		{
			name:    "Invalid scheme rejected",
			input:   "ftp://ftp.example.com/file",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty input rejected",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeURL(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NormalizeURL(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard IPv4",
			input:   "192.168.1.1",
			want:    "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "IPv4 with port stripped",
			input:   "203.0.113.10:8080",
			want:    "203.0.113.10",
			wantErr: false,
		},
		{
			name:    "Standard IPv6 loopback",
			input:   "::1",
			want:    "::1",
			wantErr: false,
		},
		{
			name:    "IPv6 canonical compression",
			input:   "2001:0db8:0000:0000:0000:0000:0000:0001",
			want:    "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "IPv6 bracketed with port",
			input:   "[2001:db8::1]:8080",
			want:    "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "Invalid IPv4 octet > 255",
			input:   "256.0.0.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Software version number rejected as IP",
			input:   "3.11.4",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty input rejected",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeIP(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NormalizeIP(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("NormalizeIP(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeCountry(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Country name Japan to ISO JP",
			input:   "Japan",
			want:    "JP",
			wantErr: false,
		},
		{
			name:    "Country name Philippines to ISO PH",
			input:   "Philippines",
			want:    "PH",
			wantErr: false,
		},
		{
			name:    "Country code US preserved uppercase",
			input:   "us",
			want:    "US",
			wantErr: false,
		},
		{
			name:    "Country name United States to US",
			input:   "United States",
			want:    "US",
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeCountry(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NormalizeCountry(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("NormalizeCountry(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeLocation(t *testing.T) {
	input := "  12km   NW of   Tokyo  "
	want := "12km nw of tokyo"
	got, err := NormalizeLocation(input)
	if err != nil {
		t.Fatalf("NormalizeLocation unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("NormalizeLocation(%q) = %q, want %q", input, got, want)
	}
}
