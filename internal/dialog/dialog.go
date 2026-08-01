package dialog

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"simuz/internal/entity"
	"simuz/internal/gen"
)

// DialogueType indicates the social tone of the conversation pass
type DialogueType string

const (
	DialogueGreeting    DialogueType = "greeting"
	DialogueTrade       DialogueType = "trade"
	DialogueGossip      DialogueType = "gossip"
	DialogueHostile     DialogueType = "hostile"
	DialogueFarewell    DialogueType = "farewell"
	DialogueNegotiation DialogueType = "negotiation"
	DialogueDemand      DialogueType = "demand"
	DialogueTreaty      DialogueType = "treaty"
	DialogueDeclaration DialogueType = "declaration"
	DialogueInsult      DialogueType = "insult"
	DialogueTribute     DialogueType = "tribute"
	DialogueAudience    DialogueType = "audience"
)

// Dialogue handles a linguistic interaction session between two entities.
type Dialogue struct {
	Speaker    *entity.Entity
	Listener   *entity.Entity
	LanguageID string       // The specific language being spoken during this interaction
	Dialect    string       // The specific dialect variant used by the speaker
	Type       DialogueType // The psychological/social framework of the conversation
	RNG        *rand.Rand   // Caller-scoped RNG; NOT safe for concurrent use
}

// NewDialogue instantiates a conversation session, automatically choosing the optimal shared language.
func NewDialogue(speaker, listener *entity.Entity, rng *rand.Rand, dType DialogueType) (*Dialogue, error) {
	chosenLang := "none"
	bestScore := -1

	// 1. Contextual Language Arbitration: Find the language that maximizes shared understanding.
	// Iterate over sorted IDs for deterministic tie-breaking.
	langIDs := make([]string, 0, len(speaker.LanguageSkills))
	for langID := range speaker.LanguageSkills {
		langIDs = append(langIDs, langID)
	}
	sort.Strings(langIDs)

	for _, langID := range langIDs {
		speakerProf := speaker.LanguageSkills[langID]
		listenerProf, has := listener.LanguageSkills[langID]
		if has {
			combinedScore := speakerProf + listenerProf
			if combinedScore > bestScore || (combinedScore == bestScore && langID < chosenLang) {
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
	case DialogueNegotiation:
		options = d.getPoliticalTemplates()
	case DialogueDemand:
		options = []string{
			"You will submit to our terms, or face the consequences.",
			"This is not a request. Comply immediately.",
			"We have the strength to take what we want. Cooperate and keep your holdings.",
		}
	case DialogueTreaty:
		options = []string{
			"Let us formalize this agreement. Our peoples have much to gain from peace.",
			"A treaty between our factions would stabilize the region. Shall we draft terms?",
			"I propose a non-aggression pact. Your borders will be respected in exchange for trade access.",
		}
	case DialogueDeclaration:
		options = []string{
			"Hear this: our faction declares its sovereign right to these lands.",
			"Let it be known that we stand united. Any aggression against one is aggression against all.",
			"We claim this territory by right of strength and legacy.",
		}
	case DialogueInsult:
		options = []string{
			"Your lineage is as weak as your sword arm.",
			"I have seen goblins with more honor than your entire faction.",
			"You are unfit to lead a pack of rats, let alone a people.",
		}
	case DialogueTribute:
		options = []string{
			"Your faction will deliver tribute to our coffers. This is non-negotiable.",
			"We demand a tithe of your goods. Consider it protection insurance.",
			"Hand over your valuables and we may allow you to continue operating.",
		}
	case DialogueAudience:
		options = []string{
			"Speak. I will hear what you have to say.",
			"You have my attention. Make your case.",
			"I grant you this audience. Choose your words carefully.",
		}
	default:
		options = []string{"Hello there."}
	}

	return options[d.RNG.Intn(len(options))]
}

// getPoliticalTemplates returns faction-aware negotiation templates.
// Speaker and listener species/faction/profession influence the wording.
func (d *Dialogue) getPoliticalTemplates() []string {
	speakerSpecies := d.Speaker.Species
	listenerSpecies := d.Listener.Species

	// Species-specific negotiation flavor
	if speakerSpecies == "orc" {
		return []string{
			"Your warriors are strong. An alliance would make us both stronger.",
			"Submit to our clan and share in our conquests. Resist and be crushed.",
			"We do not negotiate with the weak. Show us your strength first.",
		}
	}
	if speakerSpecies == "dwarf" {
		return []string{
			"Let us discuss terms. Every clause must be ironclad, mind you.",
			"We propose a trade agreement. Fair exchange, no tricks.",
			"Clan honor demands we settle this properly. Sit down and talk.",
		}
	}
	if speakerSpecies == "elf" {
		return []string{
			"The winds of change blow. Perhaps our paths were meant to converge.",
			"We see far into the future. An alliance now would bear fruit for centuries.",
			"Your people have much to learn, but we are willing to teach—for a price.",
		}
	}
	if speakerSpecies == "human" {
		return []string{
			"I believe we can find common ground. Let us discuss terms.",
			"Our factions have complementary interests. A partnership would be profitable.",
			"The council has authorized me to negotiate. What do you offer in return?",
		}
	}

	// Listener species flavor
	if listenerSpecies == "orc" {
		return []string{
			"We respect your warrior tradition. Let us forge a pact of mutual strength.",
			"Your raids have cost us dearly. Perhaps we can redirect that energy elsewhere.",
			"An orcish alliance is worth ten human treaties. Name your price.",
		}
	}

	// Default negotiation
	return []string{
		"I propose we negotiate. Both sides have much to gain.",
		"Let us discuss terms that benefit our peoples equally.",
		"This conflict serves no one. Shall we find a resolution?",
	}
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
	case "court":
		return text + " *adjusts formal attire*"
	case "tribal":
		return text + " *bangs chest*"
	case "ancient":
		return text + " *speaks with measured, timeless cadence*"
	case "diplomatic":
		return fmt.Sprintf("*speaking formally* %s *pauses for effect*", text)
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
		// Strip leading and trailing punctuation to isolate the core word,
		// then re-attach it after replacement so sentence structure is preserved.
		cleanWord := strings.Trim(word, ".,?!\"()*")
		if len(cleanWord) <= 2 {
			continue
		}

		if d.RNG.Intn(100) > comprehensionChance {
			// Find leading/trailing punctuation by trimming one side at a time
			leading := strings.TrimRight(word, ".,?!\"()*")
			leading = leading[:len(leading)-len(cleanWord)]
			trailing := strings.TrimLeft(word, ".,?!\"()*")
			trailing = trailing[len(cleanWord):]

			var replacement string
			if d.RNG.Intn(2) == 0 {
				replacement = "..."
			} else {
				replacement = d.scrambleString(cleanWord)
			}
			words[i] = leading + replacement + trailing
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
