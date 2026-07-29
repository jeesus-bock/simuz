package species

// Species defines the base data for a creature species in the simulation.
// It is the single source of truth for all species-related information.
type Species struct {
	ID                  string
	Name                string
	MaxAge              int
	AdultAge            int
	CanLevelUp          bool
	CanReproduce        bool
	IsCaveman           bool
	IsImmortal          bool
	GestationTicks      int
	DefaultScripts      []string
	DefaultSleepCycle   string // "diurnal", "nocturnal", "none"
	AutoFeed            bool
	StarvationThreshold int // ticks before starvation damage begins; 0 means immune
	MaleNames           []string
	FemaleNames         []string
	BaseAttrs           Attributes
}

// GetByID returns the Species definition for a given ID.
func GetByID(id string) (Species, bool) {
	s, exists := Registry[id]
	return s, exists
}

var StarvationDamageInterval = 10
var StarvationDamageMin = 10
var StarvationDamageMax = 20
