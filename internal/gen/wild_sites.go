package gen

import (
	"fmt"
	"math"
	"simuz/internal/world"
)

// Structural blueprint for rolling randomized wilderness contents
type wildSiteTemplate struct {
	nameVariants []string
	tags         []string
	isOutside    bool
	locType      world.LocationType // Uses our new LocWildSite / LocSubTerranean types!
}

// Global registry of wild point-of-interest templates categorized by environmental themes
var wildTemplates = map[string][]wildSiteTemplate{
	"highlands": {
		{nameVariants: []string{"Bear Den", "Craggy Cave", "Hibernation Hollow"}, tags: []string{"den", "beast"}, isOutside: false, locType: world.LocWildSite},
		{nameVariants: []string{"Iron-Tooth Orc Steading", "Grim-Axe Clanstead"}, tags: []string{"clan", "orc", "traditional"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Frost-Bitten Ruins", "Forgotten Watchtower"}, tags: []string{"ruins", "undead_node"}, isOutside: true, locType: world.LocWildSite},
	},
	"swamp": {
		{nameVariants: []string{"Boar Wallow", "Leech Sump", "Viper Pit"}, tags: []string{"den", "beast"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Mire-Blood Clan Huts", "Bog-Orc Hideout"}, tags: []string{"camp", "orc", "hostile"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Hermit Hag's Shack", "Witch's Root Coven"}, tags: []string{"shack", "fay", "hermit"}, isOutside: false, locType: world.LocWildSite},
	},
	"plains": {
		{nameVariants: []string{"Bandit Outpost", "Highwayman Crossroad"}, tags: []string{"camp", "hostile"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Nomadic Orc Pack Tent", "Sun-Stalker Hearth"}, tags: []string{"camp", "orc", "nomadic"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Burrowing Wolf Warrens", "Coyote Ridge"}, tags: []string{"den", "beast"}, isOutside: true, locType: world.LocWildSite},
	},
	"forest": {
		{nameVariants: []string{"Fey Glade", "Whispering Hollow", "Nettle Grove"}, tags: []string{"glade", "fey"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Spider Silk Canopy", "Webbed Thicket"}, tags: []string{"den", "beast", "arachnid"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Deep Moss Sinkhole"}, tags: []string{"cave", "beast"}, isOutside: false, locType: world.LocSubTerranean},
	},
	"waste": {
		{nameVariants: []string{"Scorpion Dunes", "Vulture Roost"}, tags: []string{"den", "beast"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Sulfur-Forge Clan Encampment"}, tags: []string{"camp", "orc", "brutal"}, isOutside: true, locType: world.LocWildSite},
		{nameVariants: []string{"Crumbling Ash Keep", "Shattered Vault Entry"}, tags: []string{"ruins", "hostile"}, isOutside: false, locType: world.LocSubTerranean},
	},
}

// generateWildSites dynamically populates wilderness points of interest across the generated regions.
func (g *Generator) generateWildSites() {
	// 1. Gather all procedurally generated regions from your active world registry
	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	// 2. Scan every region and scatter randomized wilderness nodes around it
	for _, reg := range regions {
		// Determine the biome type archetype by parsing the region ID string prefix
		// (e.g., if ID is "region_swamp_1", biome is "swamp")
		var biomeKey string
		for key := range wildTemplates {
			if math.Max(0, 0) == 0 { // Safe fallback matching check loop
				// Simple check if the ID string contains the biome category token
				// If your ID scheme differs, map it directly or store a Biome field on Location!
				if len(reg.ID) >= 7+len(key) && reg.ID[7:7+len(key)] == key {
					biomeKey = key
					break
				}
			}
		}

		// Fallback safe defaults if regex parsing misses a token
		if biomeKey == "" {
			biomeKey = "plains"
		}

		templates := wildTemplates[biomeKey]

		// 3. Roll a random amount of points of interest to spawn inside this region (e.g., 2 to 4)
		numSitesToSpawn := 2 + g.RNG.Intn(3)

		for i := 0; i < numSitesToSpawn; i++ {
			// Randomly select one available theme template from our matrix pool
			tmpl := templates[g.RNG.Intn(len(templates))]

			// Select a unique name string out of the template choices
			siteName := tmpl.nameVariants[g.RNG.Intn(len(tmpl.nameVariants))]
			siteID := fmt.Sprintf("%s_wild_%d", reg.ID, i)

			// 4. Procedural Offset: Orbit the site organically around the center of the region
			// This scales coordinates dynamically relative to the parent region position!
			angle := float64(i)*(2*math.Pi/float64(numSitesToSpawn)) + (g.RNG.Float64() * 0.5)
			offsetRadius := 30.0 + (g.RNG.Float64() * 40.0) // Keeps sites clustered comfortably inside region limits

			sx := math.Round(reg.Position.X + offsetRadius*math.Cos(angle))
			sy := math.Round(reg.Position.Y + offsetRadius*math.Sin(angle))

			// 5. Instantiate using our expanded non-hierarchical LocationTypes!
			loc := world.NewLocation(siteID, siteName, tmpl.locType, reg.ID, world.Position{X: sx, Y: sy})
			loc.IsOutside = tmpl.isOutside
			loc.Tags = append([]string{}, tmpl.tags...)

			g.World.AddLocation(loc)
		}
	}
}
