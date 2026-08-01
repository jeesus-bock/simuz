package gen

import (
	"simuz/internal/entity"
)

// LanguageDef defines a language in the simulation world.
type LanguageDef struct {
	ID             string   // unique key, e.g. "common", "orcish", "elvish_high"
	Name           string   // display name, e.g. "Common Tongue"
	NativeSpeakers []string // species that natively speak this language
	Complexity     int      // 1 (trivial) to 10 (near-impossible to learn as a second language)
	Dialects       []string // named variations
}

// LanguageDefs is the master registry of all languages in the world.
var LanguageDefs = []LanguageDef{
	// --- Lingua Franca ---
	{
		ID:             "common",
		Name:           "Common Tongue",
		NativeSpeakers: []string{"human", "hobbit"},
		Complexity:     1,
		Dialects:       []string{"northern", "southern", "trader"},
	},

	// --- Greenskin Languages ---
	{
		ID:             "orcish",
		Name:           "Orcish",
		NativeSpeakers: []string{"orc", "goblin", "ogre", "hobgoblin", "gnoll", "bugbear"},
		Complexity:     3,
		Dialects:       []string{"tribal", "war_cry", "diplomatic"},
	},
	{
		ID:             "goblin_tongue",
		Name:           "Goblin Tongue",
		NativeSpeakers: []string{"goblin"},
		Complexity:     4,
		Dialects:       []string{"cave", "market"},
	},

	// --- Elvish Languages ---
	{
		ID:             "elvish_high",
		Name:           "High Elvish",
		NativeSpeakers: []string{"elf"},
		Complexity:     5,
		Dialects:       []string{"court", "scholarly", "archaic"},
	},
	{
		ID:             "elvish_wood",
		Name:           "Wood Elvish",
		NativeSpeakers: []string{"elf"},
		Complexity:     4,
		Dialects:       []string{"forest", "river"},
	},

	// --- Dwarvish ---
	{
		ID:             "dwarvish",
		Name:           "Dwarvish",
		NativeSpeakers: []string{"dwarf"},
		Complexity:     4,
		Dialects:       []string{"deep", "mountain", "forge"},
	},

	// --- Ancient / Divine Languages ---
	{
		ID:             "ancient_divine",
		Name:           "Ancient Divine",
		NativeSpeakers: []string{"divine"},
		Complexity:     9,
		Dialects:       []string{"celestial", "infernal"},
	},
	{
		ID:             "draconic",
		Name:           "Draconic",
		NativeSpeakers: []string{"divine"},
		Complexity:     8,
		Dialects:       []string{"primal", "formal"},
	},

	// --- Beast Languages ---
	{
		ID:             "beast_speech",
		Name:           "Beast Speech",
		NativeSpeakers: []string{"bear", "wolf", "hawk", "spider", "rat", "bat", "snake", "boar", "deer", "lion", "tiger", "hyena"},
		Complexity:     7,
		Dialects:       []string{"predator", "prey", "avian"},
	},

	// --- Underdark / Aberration ---
	{
		ID:             "undercommon",
		Name:           "Undercommon",
		NativeSpeakers: []string{"mind_flayer", "beholder"},
		Complexity:     7,
		Dialects:       []string{"psionic", "aberrant"},
	},

	// --- Fey ---
	{
		ID:             "sylvan",
		Name:           "Sylvan",
		NativeSpeakers: []string{"elf"},
		Complexity:     5,
		Dialects:       []string{"whisper", "song"},
	},
}

// langIDIndex is a lookup from language ID → LanguageDef, built at init.
var langIDIndex = map[string]LanguageDef{}

// speciesNativeLangs maps species → list of native language IDs.
var speciesNativeLangs = map[string][]string{}

func init() {
	for _, ld := range LanguageDefs {
		langIDIndex[ld.ID] = ld
		for _, sp := range ld.NativeSpeakers {
			speciesNativeLangs[sp] = append(speciesNativeLangs[sp], ld.ID)
		}
	}
}

// GetLanguageDef returns the LanguageDef for a given language ID.
func GetLanguageDef(id string) (LanguageDef, bool) {
	ld, ok := langIDIndex[id]
	return ld, ok
}

// NativeLanguages returns the language IDs natively spoken by a species.
func NativeLanguages(speciesID string) []string {
	return speciesNativeLangs[speciesID]
}

// AssignLanguages sets native language skills on an entity based on its species.
// Civilized species get Common at proficiency 5+. Native tongues are set to 7-10.
// Educated professions (scholar, diplomat, merchant, bard, wizard, priest) may
// learn additional languages, scaled by the language's Complexity.
func AssignLanguages(e *entity.Entity, rng interface{ Intn(int) int }) {
	if e.LanguageSkills == nil {
		e.LanguageSkills = make(map[string]int)
	}

	sp := e.Species
	if sp == "divine" || sp == "deity" {
		for _, ld := range LanguageDefs {
			e.LanguageSkills[ld.ID] = 10
		}
		return
	}

	// Native languages at high proficiency.
	native := NativeLanguages(sp)
	for _, langID := range native {
		e.LanguageSkills[langID] = 7 + rng.Intn(4) // 7-10
	}

	// Civilized species know Common.
	civilized := map[string]bool{
		"human": true, "hobbit": true, "elf": true, "dwarf": true,
		"orc": true, "goblin": true, "ogre": true, "hobgoblin": true,
		"gnoll": true, "lizardfolk": true, "kobold": true,
		"half_orc": true, "half_elf": true, "half_dwarf": true,
		"half_goblin": true, "half_hobgoblin": true, "half_gnoll": true,
		"half_kobold": true, "half_fey": true,
	}
	if civilized[sp] {
		if _, has := e.LanguageSkills["common"]; !has {
			e.LanguageSkills["common"] = 3 + rng.Intn(5) // 3-7
		}
	}

	// Educated professions learn extra languages contextualized by complexity.
	educated := map[string]bool{
		"scholar": true, "diplomat": true, "merchant": true,
		"bard": true, "wizard": true, "priest": true,
		"scribe": true, "ambassador": true,
	}
	if educated[e.Profession] {
		var candidates []LanguageDef
		for _, ld := range LanguageDefs {
			if _, has := e.LanguageSkills[ld.ID]; !has && ld.ID != "beast_speech" && ld.ID != "undercommon" {
				candidates = append(candidates, ld)
			}
		}

		numExtra := 1 + rng.Intn(2) // 1-2
		for i := 0; i < numExtra && len(candidates) > 0; i++ {
			idx := rng.Intn(len(candidates))
			chosenLang := candidates[idx]

			// Higher complexity languages scale down the starting proficiency.
			baseProficiency := 5 - (chosenLang.Complexity / 2)
			if baseProficiency < 1 {
				baseProficiency = 1
			}
			e.LanguageSkills[chosenLang.ID] = baseProficiency + rng.Intn(3) // +0-2 variability

			candidates = append(candidates[:idx], candidates[idx+1:]...)
		}
	}
}
