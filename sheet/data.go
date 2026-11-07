// Package sheet generates Call of Cthulhu (7th edition) investigator sheets. A port of the original Java service:
// same rules, same JSON.
package sheet

// The eight main characteristics.
const (
	Strength     = "STRENGTH"
	Constitution = "CONSTITUTION"
	Size         = "SIZE"
	Dexterity    = "DEXTERITY"
	Appearance   = "APPEARANCE"
	Intelligence = "INTELLIGENCE"
	Power        = "POWER"
	Education    = "EDUCATION"
)

// MainCharacteristics in their canonical order.
var MainCharacteristics = []string{Strength, Constitution, Size, Dexterity, Appearance, Intelligence, Power, Education}

// CharacteristicValues are the values dealt to the characteristics, best first.
var CharacteristicValues = []int{90, 80, 70, 60, 60, 50, 50, 40}

// SkillValues are the points dealt to skills, best first.
var SkillValues = []int{80, 70, 70, 60, 60, 50, 50, 50, 40}

// Skill is a skill and its base value.
type Skill struct {
	Name string
	Base int
}

// Skills of the game with their base values; those at 0 are not dealt points (credit rating, mythos, dodge and the
// languages are worked out apart).
var Skills = []Skill{
	{"ACCOUNTING", 5}, {"ANTHROPOLOGY", 1}, {"APPRAISE", 5}, {"ARCHAEOLOGY", 1}, {"ART_CRAFT", 5}, {"CHARM", 15},
	{"CLIMB", 20}, {"CREDIT_RATING", 0}, {"CTHULHU_MYTHOS", 0}, {"DISGUISE", 5}, {"DODGE", 0}, {"DRIVE_AUTO", 20},
	{"ELECTRIC_REPAIR", 10}, {"FAST_TALK", 5}, {"FIGHTING_BRAWL", 25}, {"FIREARMS_HANDGUN", 20},
	{"FIREARMS_RIFLE_SHOTGUN", 25}, {"FIRST_AID", 30}, {"HISTORY", 5}, {"INTIMIDATE", 15}, {"JUMP", 20},
	{"FOREIGN_LANGUAGE", 0}, {"NATIVE_LANGUAGE", 0}, {"LAW", 5}, {"LIBRARY_USE", 20}, {"LISTEN", 20},
	{"LOCKSMITH", 1}, {"MECHANICAL_REPAIR", 10}, {"MEDICINE", 1}, {"NATURAL_WORLD", 10}, {"NAVIGATE", 10},
	{"OCCULT", 5}, {"PERSUADE", 10}, {"PILOT", 1}, {"PSYCHOANALYSIS", 1}, {"PSYCHOLOGY", 10}, {"RIDE", 5},
	{"SCIENCE", 1}, {"SLEIGHT_OF_HAND", 10}, {"SPOT_HIDDEN", 25}, {"STEALTH", 20}, {"SURVIVAL", 10}, {"SWIM", 20},
	{"THROW", 10}, {"TRACK", 10},
}

func skillBase(name string) int {
	for _, s := range Skills {
		if s.Name == name {
			return s.Base
		}
	}
	return 0
}

// Credit is a wealth bracket: a value is drawn in [Min, Max).
type Credit struct {
	Name     string
	Min, Max int
}

// Credits are the wealth brackets.
var Credits = []Credit{{"POOR", 1, 9}, {"AVERAGE", 10, 49}, {"WEALTHY", 50, 89}, {"RICH", 90, 98}, {"SUPER_RICH", 99, 100}}

// Bonus is the damage bonus and build for STR+SIZ in [Min, Max].
type Bonus struct {
	Min, Max                      int
	SignedDamage, SignedBuild     int
	DieNumber, DieFaces, DieBuild int
	Signed                        bool
}

// Bonuses by STR+SIZ.
var Bonuses = []Bonus{
	{2, 64, -2, -2, 0, 0, 0, true}, {65, 84, -1, -1, 0, 0, 0, true}, {85, 124, 0, 0, 0, 0, 0, true},
	{125, 164, 0, 0, 1, 4, 1, false}, {165, 204, 0, 0, 1, 6, 2, false}, {205, 284, 0, 0, 2, 6, 3, false},
	{285, 364, 0, 0, 3, 6, 4, false}, {365, 444, 0, 0, 4, 6, 5, false}, {445, 524, 0, 0, 5, 6, 6, false},
}

// Regions where languages come from.
var Regions = []string{"AMERICA", "EASTERN_EUROPE", "WESTERN_EUROPE", "MIDDLE_EAST", "ASIA", "AFRICA"}

// Language is a language, whether a character can have it as its native one, and its region.
type Language struct {
	Name   string
	Native bool
	Region string
}

// Languages known to the generator.
var Languages = []Language{
	{"ARABIC", true, "MIDDLE_EAST"}, {"TURKISH", true, "MIDDLE_EAST"}, {"PERSIAN", false, "MIDDLE_EAST"}, {"HEBREW", false, "MIDDLE_EAST"},
	{"ENGLISH", true, "WESTERN_EUROPE"}, {"FRENCH", true, "WESTERN_EUROPE"}, {"GERMAN", true, "WESTERN_EUROPE"}, {"DUTCH", true, "WESTERN_EUROPE"},
	{"ITALIAN", true, "WESTERN_EUROPE"}, {"GREEK", false, "WESTERN_EUROPE"}, {"LATIN", false, "WESTERN_EUROPE"}, {"CELTIC", false, "WESTERN_EUROPE"},
	{"AFRICANER", true, "AFRICA"}, {"ZULU", true, "AFRICA"}, {"UBUNDU", true, "AFRICA"}, {"CRIOLO", false, "AFRICA"},
	{"YORUBA", false, "AFRICA"}, {"EGYPTIAN", false, "AFRICA"},
	{"JAPANESE", true, "ASIA"}, {"THAI", true, "ASIA"}, {"KOREAN", true, "ASIA"}, {"MANDARIN", true, "ASIA"},
	{"HINDI", false, "ASIA"}, {"MONGOLIAN", false, "ASIA"},
	{"SPANISH", true, "AMERICA"}, {"PORTUGUESE", true, "AMERICA"}, {"ASTECA", false, "AMERICA"}, {"INCA", false, "AMERICA"},
	{"MAYA", false, "AMERICA"}, {"TUPI", false, "AMERICA"}, {"NAVAJO", false, "AMERICA"},
	{"RUSSIAN", true, "EASTERN_EUROPE"}, {"ROMANIAN", true, "EASTERN_EUROPE"}, {"SERBIAN", true, "EASTERN_EUROPE"},
	{"CZECH", true, "EASTERN_EUROPE"}, {"SLAV", false, "EASTERN_EUROPE"}, {"GYPSY", false, "EASTERN_EUROPE"}, {"NORSE", false, "EASTERN_EUROPE"},
}

// Stereotype favours some characteristics (dealt the best values first) and six skills (dealt the best points first).
type Stereotype struct {
	Name   string
	Chars  []string
	Skills []string
}

// Stereotypes of investigators.
var Stereotypes = []Stereotype{
	{"ADVENTURER", []string{Dexterity, Appearance}, []string{"CLIMB", "RIDE", "SWIM", "DRIVE_AUTO", "FIGHTING_BRAWL", "FIRST_AID"}},
	{"BEEFCAKE", []string{Strength}, []string{"FIGHTING_BRAWL", "INTIMIDATE", "THROW", "LISTEN", "PSYCHOLOGY", "CLIMB"}},
	{"BON_VIVANT", []string{Size}, []string{"APPRAISE", "ART_CRAFT", "CHARM", "FAST_TALK", "SPOT_HIDDEN", "PSYCHOLOGY"}},
	{"COLD_BLOODED", []string{Intelligence}, []string{"DISGUISE", "STEALTH", "FIGHTING_BRAWL", "FIREARMS_HANDGUN", "FIREARMS_RIFLE_SHOTGUN", "TRACK"}},
	{"EGGHEAD", []string{Intelligence, Education}, []string{"ELECTRIC_REPAIR", "MECHANICAL_REPAIR", "SCIENCE", "LIBRARY_USE", "ANTHROPOLOGY", "APPRAISE"}},
	{"EXPLORER", []string{Dexterity, Power}, []string{"SURVIVAL", "TRACK", "RIDE", "PILOT", "NATURAL_WORLD", "CLIMB"}},
	{"FEMME_FATALE", []string{Appearance, Intelligence}, []string{"CHARM", "FAST_TALK", "SLEIGHT_OF_HAND", "FIREARMS_HANDGUN", "PSYCHOLOGY", "LISTEN"}},
	{"GREASE_MONKEY", []string{Intelligence}, []string{"MECHANICAL_REPAIR", "DRIVE_AUTO", "LOCKSMITH", "ELECTRIC_REPAIR", "THROW", "FIGHTING_BRAWL"}},
	{"HARD_BOILED", []string{Constitution}, []string{"FIREARMS_RIFLE_SHOTGUN", "FIREARMS_HANDGUN", "INTIMIDATE", "FIGHTING_BRAWL", "LOCKSMITH", "DRIVE_AUTO"}},
	{"HARLEQUIN", []string{Appearance}, []string{"PERSUADE", "DISGUISE", "CHARM", "FAST_TALK", "SLEIGHT_OF_HAND", "STEALTH"}},
	{"HUNTER", []string{Intelligence, Constitution}, []string{"NATURAL_WORLD", "SURVIVAL", "SWIM", "TRACK", "FIRST_AID", "NAVIGATE"}},
	{"MYSTIC", []string{Power}, []string{"PSYCHOLOGY", "OCCULT", "HISTORY", "SCIENCE", "ART_CRAFT", "NATURAL_WORLD"}},
	{"OUTSIDER", []string{Intelligence, Constitution}, []string{"NATURAL_WORLD", "SURVIVAL", "SWIM", "TRACK", "FIRST_AID", "NAVIGATE"}},
	{"ROGUE", []string{Dexterity, Appearance}, []string{"STEALTH", "SLEIGHT_OF_HAND", "DISGUISE", "SPOT_HIDDEN", "APPRAISE", "ART_CRAFT"}},
	{"SCHOLAR", []string{Education}, []string{"SCIENCE", "LIBRARY_USE", "MEDICINE", "HISTORY", "NATURAL_WORLD", "ANTHROPOLOGY"}},
	{"SEEKER", []string{Intelligence}, []string{"SPOT_HIDDEN", "LISTEN", "OCCULT", "LIBRARY_USE", "DISGUISE", "APPRAISE"}},
	{"SIDEKICK", []string{Dexterity, Constitution}, []string{"JUMP", "NAVIGATE", "STEALTH", "TRACK", "CLIMB", "FIRST_AID"}},
	{"STEADFAST", []string{Constitution}, []string{"LAW", "PERSUADE", "ACCOUNTING", "HISTORY", "PSYCHOLOGY", "FIRST_AID"}},
	{"THRILL_SEEKER", []string{Dexterity, Power}, []string{"SURVIVAL", "CLIMB", "SWIM", "NAVIGATE", "PILOT", "DRIVE_AUTO"}},
	{"TWO_FISTED", []string{Strength, Size}, []string{"FIGHTING_BRAWL", "DRIVE_AUTO", "SWIM", "THROW", "FIREARMS_HANDGUN", "INTIMIDATE"}},
}
