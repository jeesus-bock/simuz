package dialog

import (
	"fmt"
	"math/rand"
	"strings"

	"simuz/internal/entity"
	"simuz/internal/gen"
)

// DialogueType indicates the social tone of the conversation pass
type DialogueType string

const (
	DialogueGreeting DialogueType = "greeting"
	DialogueTrade    DialogueType = "trade"
	DialogueGossip   DialogueType = "gossip"
	DialogueHostile  DialogueType = "hostile"
	DialogueFarewell DialogueType = "farewell"
)

// Dialogue handles a linguistic interaction session between two entities.
type Dialogue struct {
	Speaker    *entity.Entity
	Listener   *entity.Entity
	LanguageID string       // The specific language being spoken during this interaction
	Dialect    string       // The specific dialect variant used by the speaker
	Type       DialogueType // The psychological/social framework of the conversation
	RNG        *rand.Rand   // Thread-safe scoped random instance
}

// NewDialogue instantiates a conversation session, automatically choosing the optimal shared language.
func NewDialogue(speaker, listener *entity.Entity, rng *rand.Rand, dType DialogueType) (*Dialogue, error) {
	chosenLang := "none"
	bestScore := -1

	// 1. Contextual Language Arbitration: Find the language that maximizes shared understanding
	for langID, speakerProf := range speaker.LanguageSkills {
		listenerProf, has := listener.LanguageSkills[langID]
		if has {
			// Scoring favors the language where BOTH characters have high capability
			combinedScore := speakerProf + listenerProf
			if combinedScore > bestScore {
				bestScore = combinedScore
				chosenLang = langID
			}
		}
	}

	// Fallback to "common" if no native overlapping tongue is found and both are civilized
	if chosenLang == "none" {
		if speaker.LanguageSkills["common"] > 0 && listener.LanguageSkills["common"] > 0 {
			chosenLang = "common"
		} else {
			return nil, fmt.Errorf("dialogue failed: %s and %s share no common language framework", speaker.Name, listener.Name)
		}
	}

	// 2. Extract random dialect variation from the master registry
	var chosenDialect string
	if def, ok := gen.GetLanguageDef(chosenLang); ok && len(def.Dialects) > 0 {
		chosenDialect = def.Dialects[rng.Intn(len(def.Dialects))]
	}

	return &Dialogue{
		Speaker:    speaker,
		Listener:   listener,
		LanguageID: chosenLang,
		Dialect:    chosenDialect,
		Type:       dType,
		RNG:        rng,
	}, nil
}

// GenerateText produces a procedurally altered string based on the speaker's tone and dialect,
// and then structurally masks it based on the listener's specific proficiency level.
func (d *Dialogue) GenerateText() string {
	rawSpeech := d.getRawTemplate()

	// 1. Apply Speaker Dialect Layer Modification
	processedSpeech := d.applyDialectModifications(rawSpeech)

	// 2. Apply Listener Comprehension Filter Matrix
	listenerProficiency := d.Listener.LanguageSkills[d.LanguageID]

	return d.obfuscateForComprehension(processedSpeech, listenerProficiency)
}

// getRawTemplate uses your structured pool to fetch varied narrative strings
func (d *Dialogue) getRawTemplate() string {
	var options []string

	switch d.Type {
	case DialogueGreeting:
		options = []string{
			"Greetings, traveler. Safe journeys to you on this fine day.",
			"Hail! May your paths be clear and your blades sharp.",
			"Ah, a welcome face. What brings you to these parts?",
		}
	case DialogueTrade:
		options = []string{
			"Take a look at my wares. I offer only the finest craftsmanship.",
			"Let us talk coin. I am willing to barter if you have rare goods.",
			"Times are tough, but for a friend, I can offer a reasonable price.",
		}
	case DialogueGossip:
		options = []string{
			"They say strange beasts have been sighted roaming the outer borders.",
			"Keep your ears open; the local guards have been acting anxious lately.",
			"A hidden treasure is rumored to lie buried deep within the barren wastes.",
		}
	case DialogueHostile:
		options = []string{
			"You walk on dangerous ground, stranger. Turn back or face the consequences.",
			"Your presence here is an insult. State your business before I lose patience.",
			"I do not trust your kind. Keep your hands where my guards can see them.",
		}
	default:
		options = []string{"Hello there."}
	}

	return options[d.RNG.Intn(len(options))]
}

// applyDialectModifications introduces phonetic quirks depending on the chosen dialect tag
func (d *Dialogue) applyDialectModifications(text string) string {
	switch d.Dialect {
	case "northern":
		return text + " Aye, dynamic winds blow cold up here."
	case "trader":
		return text + " (The speaker checks their ledger attentively)."
	case "war_cry":
		return strings.ToUpper(text) + " Lok-tar!"
	case "whisper":
		return fmt.Sprintf("*softly rustling* %s", strings.ToLower(text))
	case "forge":
		return text + " Clang! The anvil sings."
	default:
		return text
	}
}

// obfuscateForComprehension scrambles or masks words if the listener's proficiency is poor
func (d *Dialogue) obfuscateForComprehension(text string, proficiency int) string {
	// If the listener is highly fluent (proficiency 8-10), they understand 100% clearly
	if proficiency >= 8 {
		return text
	}

	words := strings.Fields(text)
	// Map proficiency directly to the mathematical percentage of words understood clearly
	// Prof 1 = ~10% comprehension, Prof 4 = ~40% comprehension, Prof 7 = ~70% comprehension
	comprehensionChance := proficiency * 10
	if comprehensionChance < 10 {
		comprehensionChance = 10
	}

	for i, word := range words {
		// Clean up punctuation attached to words for isolated randomization checks
		cleanWord := strings.Trim(word, ".,?!\"()*")
		if len(cleanWord) <= 2 {
			continue // Skip processing tiny structural particle words
		}

		if d.RNG.Intn(100) > comprehensionChance {
			// Replace the word with randomized linguistic static gibberish placeholders
			if d.RNG.Intn(2) == 0 {
				words[i] = "..."
			} else {
				words[i] = d.scrambleString(cleanWord)
			}
		}
	}

	return strings.Join(words, " ")
}

// scrambleString creates chaotic phonetic clusters for words a listener fails to recognize
func (d *Dialogue) scrambleString(s string) string {
	runes := []rune(s)
	// Simple chaotic swap algorithm to mock garbled syllables
	for i := range runes {
		if d.RNG.Intn(100) < 40 {
			runes[i] = rune('a' + d.RNG.Intn(26))
		}
	}
	return "[" + string(runes) + "]"
}
