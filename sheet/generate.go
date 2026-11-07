package sheet

import (
	"math/rand/v2"
	"sort"
	"strings"
)

// Attribute is a characteristic or a skill with its full, half and fifth values.
type Attribute struct {
	AttributeName string `json:"attributeName"`
	MainValue     int    `json:"mainValue"`
	HalfValue     int    `json:"halfValue"`
	FifthValue    int    `json:"fifthValue"`
}

func attr(name string, v int) Attribute {
	return Attribute{AttributeName: name, MainValue: v, HalfValue: v / 2, FifthValue: v / 5}
}

// Person is a detail of the character with a value: a language or the credit rating.
type Person struct {
	SkillName string `json:"skillName"`
	Detail    string `json:"detail"`
	Value     int    `json:"value"`
}

// Record is the whole sheet, as the API answers it.
type Record struct {
	HitPoints       int         `json:"hitPoints"`
	Age             int         `json:"age"`
	MovementRate    int         `json:"movementRate"`
	Build           int         `json:"build"`
	BonusDamage     string      `json:"bonusDamage"`
	MagicPoints     int         `json:"magicPoints"`
	NativeLanguage  *Person     `json:"nativeLanguage"`
	ForeignLanguage *Person     `json:"foreignLanguage"`
	CreditRating    *Person     `json:"creditRating"`
	CharFirst       []Attribute `json:"charFirst"`
	CharSecond      []Attribute `json:"charSecond"`
	SkillFirst      []Attribute `json:"skillFirst"`
	SkillSecond     []Attribute `json:"skillSecond"`
}

// Title turns an enum name into a label: FAST_TALK -> Fast Talk.
func Title(s string) string {
	words := strings.Fields(strings.ToLower(strings.ReplaceAll(s, "_", " ")))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

func has(list []Attribute, name string) bool {
	for _, a := range list {
		if strings.EqualFold(a.AttributeName, name) {
			return true
		}
	}
	return false
}

func value(list []Attribute, name string) int {
	for _, a := range list {
		if strings.EqualFold(a.AttributeName, name) {
			return a.MainValue
		}
	}
	return 0
}

func shuffled[T any](r *rand.Rand, in []T) []T {
	out := append([]T(nil), in...)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// between draws an int in [min, max).
func between(r *rand.Rand, min, max int) int { return min + r.IntN(max-min) }

func findStereotype(name string) *Stereotype {
	for i := range Stereotypes {
		if strings.EqualFold(Stereotypes[i].Name, name) {
			return &Stereotypes[i]
		}
	}
	return nil
}

// pool hands out values in order, each once.
type pool struct {
	vals []int
	next int
}

func (p *pool) take() (int, bool) {
	if p.next >= len(p.vals) {
		return 0, false
	}
	p.next++
	return p.vals[p.next-1], true
}

// Generate builds a sheet. stereotype and the two regions are optional (empty: at random).
func Generate(r *rand.Rand, stereotype, nativeRegion, foreignRegion string) Record {
	st := findStereotype(stereotype)

	// characteristics: the stereotype's first, with the best values; the rest shuffled, with the rest of the values shuffled
	var chars []Attribute
	cp := &pool{vals: CharacteristicValues}
	if st != nil {
		for _, c := range st.Chars {
			if has(chars, c) {
				continue
			}
			if v, ok := cp.take(); ok {
				chars = append(chars, attr(Title(c), v))
			}
		}
	}
	rest := shuffled(r, cp.vals[cp.next:])
	rp := &pool{vals: rest}
	for _, c := range shuffled(r, MainCharacteristics) {
		if has(chars, c) {
			continue
		}
		if v, ok := rp.take(); ok {
			chars = append(chars, attr(Title(c), v))
		}
	}
	chars = append(chars, attr("Luck", (rollDice(r, 2, 6)+6)*5))
	chars = append(chars, attr("Sanity", value(chars, Power)))

	// skills: credit rating, then the stereotype's skills with the best points, then every other skill with what is left
	var skills []Attribute
	credit := Credits[r.IntN(len(Credits))]
	money := between(r, credit.Min, credit.Max)
	creditRating := &Person{SkillName: Title("CREDIT_RATING"), Detail: Title(credit.Name), Value: money}
	skills = append(skills, attr(Title("CREDIT_RATING"), money))
	sp := &pool{vals: SkillValues}
	add := func(name string) {
		if has(skills, Title(name)) {
			return
		}
		bonus, _ := sp.take()
		skills = append(skills, attr(Title(name), skillBase(name)+bonus))
	}
	if st != nil {
		for _, s := range shuffled(r, st.Skills) {
			add(s)
		}
	}
	for _, s := range shuffled(r, Skills) {
		if s.Base != 0 {
			add(s.Name)
		}
	}
	skills = append(skills, attr(Title("DODGE"), value(chars, Dexterity)/2))
	skills = append(skills, attr(Title("NATIVE_LANGUAGE"), value(chars, Education)))
	skills = append(skills, attr(Title("FOREIGN_LANGUAGE"), value(chars, Education)/2))

	sort.SliceStable(chars, func(i, j int) bool { return chars[i].MainValue > chars[j].MainValue })
	sort.SliceStable(skills, func(i, j int) bool { return skills[i].MainValue > skills[j].MainValue })
	c1, c2 := split(chars)
	s1, s2 := split(skills)
	rec := Record{CharFirst: c1, CharSecond: c2, SkillFirst: s1, SkillSecond: s2, CreditRating: creditRating}

	con, siz, str, dex, pow, edu := value(chars, Constitution), value(chars, Size), value(chars, Strength), value(chars, Dexterity), value(chars, Power), value(chars, Education)
	rec.HitPoints = (con + siz) / 5
	rec.MagicPoints = pow / 5
	for _, b := range Bonuses {
		if str+siz >= b.Min && str+siz <= b.Max {
			if b.Signed {
				rec.BonusDamage, rec.Build = itoa(b.SignedDamage), b.SignedBuild
			} else {
				rec.BonusDamage, rec.Build = itoa(b.DieNumber)+"d"+itoa(b.DieFaces), b.DieBuild
			}
			break
		}
	}
	switch {
	case dex < siz && str < siz:
		rec.MovementRate = 7
	case dex > siz && str > siz:
		rec.MovementRate = 9
	default:
		rec.MovementRate = 8
	}
	rec.Age = between(r, 15, 91)
	rec.NativeLanguage = &Person{SkillName: Title("NATIVE_LANGUAGE"), Detail: Title(pickLanguage(r, nativeRegion, true)), Value: edu}
	rec.ForeignLanguage = &Person{SkillName: Title("FOREIGN_LANGUAGE"), Detail: Title(pickLanguage(r, foreignRegion, false)), Value: edu / 2}
	return rec
}

// pickLanguage draws a language that can (or cannot) be a native one, from the region when it names one that has it.
func pickLanguage(r *rand.Rand, region string, native bool) string {
	var any, local []string
	for _, l := range Languages {
		if l.Native != native {
			continue
		}
		any = append(any, l.Name)
		if strings.EqualFold(l.Region, region) {
			local = append(local, l.Name)
		}
	}
	if len(local) > 0 {
		return local[r.IntN(len(local))]
	}
	return any[r.IntN(len(any))]
}

func rollDice(r *rand.Rand, n, faces int) int {
	t := 0
	for i := 0; i < n; i++ {
		t += 1 + r.IntN(faces)
	}
	return t
}

// split cuts a list in two, the first half one longer when the length is odd.
func split(list []Attribute) ([]Attribute, []Attribute) {
	mid := (len(list) + 1) / 2
	return append([]Attribute{}, list[:mid]...), append([]Attribute{}, list[mid:]...)
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

// Option is an entry of the lists the front shows: a stereotype or a region.
type Option struct {
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
}

// StereotypeOptions lists the stereotypes.
func StereotypeOptions() []Option {
	out := make([]Option, 0, len(Stereotypes))
	for _, s := range Stereotypes {
		out = append(out, Option{s.Name, Title(s.Name), "Based on " + strings.Join(s.Chars, ", ")})
	}
	return out
}

// RegionOptions lists the regions with their native (or foreign) languages as examples.
func RegionOptions(native bool) []Option {
	out := make([]Option, 0, len(Regions))
	for _, reg := range Regions {
		var ls []string
		for _, l := range Languages {
			if l.Region == reg && l.Native == native {
				ls = append(ls, l.Name)
			}
		}
		out = append(out, Option{reg, Title(reg), "Ex: " + strings.Join(ls, ", ")})
	}
	return out
}
