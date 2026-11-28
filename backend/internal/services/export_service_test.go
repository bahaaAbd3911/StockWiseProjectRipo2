package services

import "testing"

func TestCSVCellNeutralisesFormulas(t *testing.T) {
	cases := map[string]string{
		"Sara":          "Sara",
		"":              "",
		"=1+1":          "'=1+1",
		"+962790000000": "'+962790000000",
		"-5":            "'-5",
		"@SUM(A1)":      "'@SUM(A1)",
		"\tcmd":         "'\tcmd",
		"سارة":          "سارة",
		"a=b":           "a=b",
	}
	for in, want := range cases {
		if got := cell(in); got != want {
			t.Errorf("cell(%q) = %q, want %q", in, got, want)
		}
	}
}
