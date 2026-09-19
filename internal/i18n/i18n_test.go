package i18n

import (
	"testing"
)

func TestT(t *testing.T) {
	if got := T("en", "projects"); got != "Projects" {
		t.Errorf("en projects = %q, want Projects", got)
	}
	if got := T("sv", "projects"); got != "Projekt" {
		t.Errorf("sv projects = %q, want Projekt", got)
	}
	if got := T("sv", "all_writing"); got != "Allt skrivande →" {
		t.Errorf("sv all_writing = %q, want Allt skrivande →", got)
	}
	// Fallback to English for unknown language
	if got := T("de", "recently_shipped"); got != "Recently shipped" {
		t.Errorf("de fallback = %q, want Recently shipped", got)
	}
	// Fallback to key if unknown key
	if got := T("sv", "unknown_key_xyz"); got != "unknown_key_xyz" {
		t.Errorf("unknown key = %q, want unknown_key_xyz", got)
	}
	// Formatted translation
	if got := T("sv", "default_bio", "Alex"); got != "Jag är Alex, mjukvaruingenjör och designer. Fokuserad på enkla verktyg, idiomatisk Go och lugna gränssnitt." {
		t.Errorf("sv formatted bio = %q", got)
	}
}

func TestIsSectionMatch(t *testing.T) {
	cases := []struct {
		sec, label string
		want       bool
	}{
		{"projects", "Projects", true},
		{"projects", "projekt", true},
		{"projects", "PROJEKT", true},
		{"projects", "Writing", false},
		{"blog", "Blog", true},
		{"blog", "Blogg", true},
		{"blog", "Writing", true},
		{"blog", "Skrivande", true},
		{"about", "About", true},
		{"about", "About Me", true},
		{"about", "Om", true},
		{"about", "Om mig", true},
		{"about", "Projekt", false},
	}
	for _, tc := range cases {
		if got := IsSectionMatch(tc.sec, tc.label); got != tc.want {
			t.Errorf("IsSectionMatch(%q, %q) = %v, want %v", tc.sec, tc.label, got, tc.want)
		}
	}
}
