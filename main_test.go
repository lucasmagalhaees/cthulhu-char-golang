package main

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sort"
	"strconv"
	"testing"
)

// The same rules as CharacterApiTest in the Java service: the specification of this rewrite.
const runs = 300

var srv = httptest.NewServer(routes(func() *rand.Rand { return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())) }))

type attr struct {
	AttributeName string `json:"attributeName"`
	MainValue     int    `json:"mainValue"`
	HalfValue     int    `json:"halfValue"`
	FifthValue    int    `json:"fifthValue"`
}
type person struct {
	SkillName, Detail string
	Value             int
}
type record struct {
	HitPoints, Age, MovementRate, Build, MagicPoints int
	BonusDamage                                      string
	NativeLanguage, ForeignLanguage, CreditRating    person
	CharFirst, CharSecond, SkillFirst, SkillSecond   []attr
}

func getJSON(t *testing.T, path string, out any) *http.Response {
	t.Helper()
	res, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("%s: status %d", path, res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("%s: no CORS header", path)
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
	return res
}

func sheetOf(t *testing.T, q string) record {
	var r record
	getJSON(t, "/char-generator/random-char"+q, &r)
	return r
}

func byName(a, b []attr) map[string]int {
	m := map[string]int{}
	for _, x := range append(append([]attr{}, a...), b...) {
		m[x.AttributeName] = x.MainValue
	}
	return m
}

var mainChars = []string{"Strength", "Constitution", "Size", "Dexterity", "Appearance", "Intelligence", "Power", "Education"}

func TestCharacteristicsAreTheEightValuesPlusLuckAndSanity(t *testing.T) {
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "")
		c := byName(s.CharFirst, s.CharSecond)
		if len(c) != 10 {
			t.Fatalf("10 characteristics, got %v", c)
		}
		var vals []int
		for _, n := range mainChars {
			v, ok := c[n]
			if !ok {
				t.Fatalf("missing %s", n)
			}
			vals = append(vals, v)
		}
		sort.Ints(vals)
		if !slices.Equal(vals, []int{40, 50, 50, 60, 60, 70, 80, 90}) {
			t.Fatalf("values %v", vals)
		}
		if c["Sanity"] != c["Power"] {
			t.Fatalf("sanity %d power %d", c["Sanity"], c["Power"])
		}
		if l := c["Luck"]; l < 40 || l > 90 || l%5 != 0 {
			t.Fatalf("luck %d", l)
		}
		if len(s.CharFirst) != 5 || len(s.CharSecond) != 5 {
			t.Fatalf("halves %d %d", len(s.CharFirst), len(s.CharSecond))
		}
	}
}

func TestEveryAttributeCarriesItsHalfAndFifthAndTheListsAreSortedBestFirst(t *testing.T) {
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "")
		for _, pair := range [][2][]attr{{s.CharFirst, s.CharSecond}, {s.SkillFirst, s.SkillSecond}} {
			all := append(append([]attr{}, pair[0]...), pair[1]...)
			for i, a := range all {
				if a.HalfValue != a.MainValue/2 || a.FifthValue != a.MainValue/5 {
					t.Fatalf("%+v", a)
				}
				if i > 0 && all[i-1].MainValue < a.MainValue {
					t.Fatalf("not sorted at %d: %v", i, all)
				}
			}
			if d := len(pair[0]) - len(pair[1]); d < 0 || d > 1 {
				t.Fatalf("halves %d %d", len(pair[0]), len(pair[1]))
			}
		}
	}
}

func bonus(n int) (string, int) {
	switch {
	case n <= 64:
		return "-2", -2
	case n <= 84:
		return "-1", -1
	case n <= 124:
		return "0", 0
	case n <= 164:
		return "1d4", 1
	case n <= 204:
		return "1d6", 2
	case n <= 284:
		return "2d6", 3
	case n <= 364:
		return "3d6", 4
	case n <= 444:
		return "4d6", 5
	}
	return "5d6", 6
}

func TestDerivedValuesFollowTheRules(t *testing.T) {
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "")
		c := byName(s.CharFirst, s.CharSecond)
		con, siz, str, dex, pow, edu := c["Constitution"], c["Size"], c["Strength"], c["Dexterity"], c["Power"], c["Education"]
		if s.HitPoints != (con+siz)/5 || s.MagicPoints != pow/5 {
			t.Fatalf("hp %d mp %d", s.HitPoints, s.MagicPoints)
		}
		if s.Age < 15 || s.Age > 90 {
			t.Fatalf("age %d", s.Age)
		}
		if s.MovementRate < 7 || s.MovementRate > 9 || (s.MovementRate == 7) != (dex < siz && str < siz) {
			t.Fatalf("move %d dex %d str %d siz %d", s.MovementRate, dex, str, siz)
		}
		if bd, b := bonus(str + siz); s.BonusDamage != bd || s.Build != b {
			t.Fatalf("bonus %s/%d for %d", s.BonusDamage, s.Build, str+siz)
		}
		sk := byName(s.SkillFirst, s.SkillSecond)
		if sk["Dodge"] != dex/2 || sk["Native Language"] != edu || sk["Foreign Language"] != edu/2 {
			t.Fatalf("derived skills %v", sk)
		}
		if s.NativeLanguage.Value != edu || s.ForeignLanguage.Value != edu/2 {
			t.Fatalf("languages %+v %+v", s.NativeLanguage, s.ForeignLanguage)
		}
	}
}

func TestSkillsStartAtTheirBaseAndCreditRatingMatchesItsBracket(t *testing.T) {
	base := map[string]int{"Persuade": 10, "Fast Talk": 5, "Charm": 15, "Psychology": 10, "First Aid": 30, "Spot Hidden": 25}
	brackets := map[string][2]int{"Poor": {1, 8}, "Average": {10, 48}, "Wealthy": {50, 88}, "Rich": {90, 97}, "Super Rich": {99, 99}}
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "")
		sk := byName(s.SkillFirst, s.SkillSecond)
		for n, b := range base {
			if v, ok := sk[n]; !ok || v < b {
				t.Fatalf("%s = %d (base %d)", n, v, b)
			}
		}
		if _, ok := sk["Cthulhu Mythos"]; ok {
			t.Fatal("no mythos at creation")
		}
		cr := s.CreditRating
		r, ok := brackets[cr.Detail]
		if cr.SkillName != "Credit Rating" || !ok || cr.Value < r[0] || cr.Value > r[1] || sk["Credit Rating"] != cr.Value {
			t.Fatalf("credit %+v skill %d", cr, sk["Credit Rating"])
		}
	}
}

func TestAStereotypeGetsTheBestValueInItsCharacteristic(t *testing.T) {
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "?stereotype=scholar")
		if c := byName(s.CharFirst, s.CharSecond); c["Education"] != 90 {
			t.Fatalf("scholar education %d", c["Education"])
		}
		s = sheetOf(t, "?stereotype=TWO_FISTED")
		c := byName(s.CharFirst, s.CharSecond)
		if got := []int{c["Strength"], c["Size"]}; !(slices.Contains(got, 90) && slices.Contains(got, 80)) {
			t.Fatalf("two-fisted %v", got)
		}
	}
}

func TestLanguagesComeFromTheAskedRegions(t *testing.T) {
	for k := 0; k < runs; k++ {
		s := sheetOf(t, "?nativeLanguageRegion=EASTERN_EUROPE&foreignLanguageRegion=africa")
		if s.NativeLanguage.SkillName != "Native Language" || !slices.Contains([]string{"Russian", "Romanian", "Serbian", "Czech"}, s.NativeLanguage.Detail) {
			t.Fatalf("native %+v", s.NativeLanguage)
		}
		if s.ForeignLanguage.SkillName != "Foreign Language" || !slices.Contains([]string{"Criolo", "Yoruba", "Egyptian"}, s.ForeignLanguage.Detail) {
			t.Fatalf("foreign %+v", s.ForeignLanguage)
		}
	}
}

func TestHealthAnswersWithoutBuildingASheet(t *testing.T) {
	var h map[string]string
	getJSON(t, "/char-generator/health", &h)
	if h["status"] != "UP" {
		t.Fatalf("%v", h)
	}
}

type option struct{ Name, Alias, Description string }

func TestTheListsTheFrontShows(t *testing.T) {
	var st, nat, fo []option
	getJSON(t, "/char-generator/random-char/stereotype", &st)
	getJSON(t, "/char-generator/random-char/region/native", &nat)
	getJSON(t, "/char-generator/random-char/region/foreign", &fo)
	if len(st) != 20 || st[0] != (option{"ADVENTURER", "Adventurer", "Based on DEXTERITY, APPEARANCE"}) {
		t.Fatalf("stereotypes %v", st[:1])
	}
	if len(nat) != 6 || nat[1] != (option{"EASTERN_EUROPE", "Eastern Europe", "Ex: RUSSIAN, ROMANIAN, SERBIAN, CZECH"}) {
		t.Fatalf("native %v", nat)
	}
	if fo[5] != (option{"AFRICA", "Africa", "Ex: CRIOLO, YORUBA, EGYPTIAN"}) {
		t.Fatalf("foreign %v", fo[5])
	}
}

// The response has every field the game reads, with the same names and types as the recorded Java answer.
func TestContractMatchesTheRecordedJavaAnswer(t *testing.T) {
	raw, err := os.ReadFile("testdata/random-char_eastern-europe_2026-11-03.json")
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	res, err := http.Get(srv.URL + "/char-generator/random-char")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	for k, v := range want {
		g, ok := got[k]
		if !ok {
			t.Fatalf("missing field %q", k)
		}
		if kind(g) != kind(v) {
			t.Fatalf("field %q: %s, want %s", k, kind(g), kind(v))
		}
	}
}

func kind(v any) string {
	switch x := v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case float64:
		return "number"
	case string:
		_ = x
		return "string"
	}
	return strconv.Quote("other")
}

func TestUnknownRoutesAre404AndTheRootGoesToHealth(t *testing.T) {
	c := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, _ := c.Get(srv.URL + "/nope")
	if res.StatusCode != 404 {
		t.Fatalf("status %d", res.StatusCode)
	}
	res, _ = c.Get(srv.URL + "/")
	if res.StatusCode != 302 || res.Header.Get("Location") != "/char-generator/health" {
		t.Fatalf("root %d %s", res.StatusCode, res.Header.Get("Location"))
	}
}
