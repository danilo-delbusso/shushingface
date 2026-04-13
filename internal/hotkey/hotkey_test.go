package hotkey

import "testing"

func TestParseSpec(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"Ctrl+Shift+B", "Ctrl+Shift+B", false},
		{"ctrl+b", "Ctrl+B", false},
		{"Super+Ctrl+B", "Ctrl+Super+B", false},
		{"Alt+F5", "Alt+F5", false},
		{"B", "", true},     // no modifier
		{"Ctrl+", "", true}, // missing key
		{"", "", true},      // empty
		{"Foo+B", "", true}, // unknown modifier
	}
	for _, tc := range cases {
		spec, err := ParseSpec(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseSpec(%q) err=%v wantErr=%v", tc.in, err, tc.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if got := FormatSpec(spec); got != tc.want {
			t.Errorf("ParseSpec(%q) round-trip = %q want %q", tc.in, got, tc.want)
		}
	}
}
