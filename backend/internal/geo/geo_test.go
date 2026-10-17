package geo

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		iso2, numeric, continent, en, ru string
	}{
		{"NL", "528", Europe, "Netherlands", "Нидерланды"},
		{"BR", "076", SouthAmerica, "Brazil", "Бразилия"},
		{"MX", "484", NorthAmerica, "Mexico", "Мексика"},
		{"JP", "392", Asia, "Japan", "Япония"},
		{"AU", "036", Oceania, "Australia", "Австралия"},
		{"EG", "818", Africa, "Egypt", "Египет"},
		{"AQ", "010", Antarctica, "Antarctica", "Антарктида"},
	}
	for _, tt := range tests {
		t.Run(tt.iso2, func(t *testing.T) {
			c, ok := Lookup(tt.iso2)
			if !ok {
				t.Fatalf("Lookup(%q) not found", tt.iso2)
			}
			if c.ISONumeric != tt.numeric || c.Continent != tt.continent || c.Name.EN != tt.en || c.Name.RU != tt.ru {
				t.Errorf("got %+v", c)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("XX"); ok {
		t.Error("Lookup(XX) should not be found")
	}
}

func TestAllComplete(t *testing.T) {
	if n := len(All()); n < 240 {
		t.Fatalf("got %d countries, want at least 240", n)
	}
	valid := map[string]bool{Africa: true, Antarctica: true, Asia: true, Europe: true,
		NorthAmerica: true, Oceania: true, SouthAmerica: true}
	for _, c := range All() {
		if len(c.ISO2) != 2 || c.Name.EN == "" || c.Name.RU == "" || !valid[c.Continent] {
			t.Errorf("invalid entry: %+v", c)
		}
		if c.ISONumeric != "" && len(c.ISONumeric) != 3 {
			t.Errorf("%s: isoNumeric %q is not 3 digits", c.ISO2, c.ISONumeric)
		}
	}
}
