package sheet

import (
	"math/rand/v2"
	"testing"
)

func TestTitle(t *testing.T) {
	for in, want := range map[string]string{"FAST_TALK": "Fast Talk", "SUPER_RICH": "Super Rich", "luck": "Luck", "EASTERN_EUROPE": "Eastern Europe"} {
		if got := Title(in); got != want {
			t.Errorf("Title(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitPutsTheExtraInTheFirstHalf(t *testing.T) {
	for n, want := range map[int][2]int{0: {0, 0}, 1: {1, 0}, 9: {5, 4}, 10: {5, 5}} {
		a, b := split(make([]Attribute, n))
		if len(a) != want[0] || len(b) != want[1] {
			t.Errorf("split(%d) = %d/%d, want %v", n, len(a), len(b), want)
		}
	}
}

func TestTheSameSeedGivesTheSameSheet(t *testing.T) {
	a := Generate(rand.New(rand.NewPCG(1, 2)), "", "", "")
	b := Generate(rand.New(rand.NewPCG(1, 2)), "", "", "")
	if a.Age != b.Age || a.CharFirst[0] != b.CharFirst[0] || a.NativeLanguage.Detail != b.NativeLanguage.Detail {
		t.Fatal("not reproducible")
	}
}

func TestBothAboveSizeMovesAtNine(t *testing.T) {
	for k := uint64(0); k < 500; k++ {
		s := Generate(rand.New(rand.NewPCG(k, 7)), "", "", "")
		all := append(append([]Attribute{}, s.CharFirst...), s.CharSecond...)
		dex, str, siz := value(all, Dexterity), value(all, Strength), value(all, Size)
		if (dex > siz && str > siz) != (s.MovementRate == 9) {
			t.Fatalf("dex %d str %d siz %d move %d", dex, str, siz, s.MovementRate)
		}
	}
}
