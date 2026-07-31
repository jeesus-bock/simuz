package gen

import (
	"fmt"
	"log"
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
	log.Printf("[gen] generateWildSites: starting wilderness site generation")
	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	log.Printf("[gen] generateWildSites: found %d regions", len(regions))
	totalSites := 0

	for _, reg := range regions {
		biomeKey := biomeFromRegionID(reg.ID)

		templates := wildTemplates[biomeKey]

		numSitesToSpawn := 2 + g.RNG.Intn(3)
		log.Printf("[gen] region %s biome=%s: spawning %d wild sites", reg.ID, biomeKey, numSitesToSpawn)

		for i := 0; i < numSitesToSpawn; i++ {
			tmpl := templates[g.RNG.Intn(len(templates))]
			siteName := tmpl.nameVariants[g.RNG.Intn(len(tmpl.nameVariants))]
			siteID := fmt.Sprintf("%s_wild_%d", reg.ID, i)

			angle := float64(i)*(2*math.Pi/float64(numSitesToSpawn)) + (g.RNG.Float64() * 0.5)
			offsetRadius := 30.0 + (g.RNG.Float64() * 40.0)

			sx := math.Round(reg.Position.X + offsetRadius*math.Cos(angle))
			sy := math.Round(reg.Position.Y + offsetRadius*math.Sin(angle))

			loc := world.NewLocation(siteID, siteName, tmpl.locType, reg.ID, world.Position{X: sx, Y: sy})
			loc.IsOutside = tmpl.isOutside
			loc.Tags = append([]string{}, tmpl.tags...)

			g.World.AddLocation(loc)
			totalSites++
			log.Printf("[gen] wild_site: %q id=%s type=%s tags=%v pos=(%.0f,%.0f)", siteName, siteID, tmpl.locType, tmpl.tags, sx, sy)
		}
	}

	log.Printf("[gen] generateWildSites: done, total %d sites spawned", totalSites)
}
