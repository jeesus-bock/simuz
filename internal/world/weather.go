// Package world defines the simulation world model, locations, travel rules, and weather systems.
package world

import "math/rand"

type WeatherType int

const (
	Clear WeatherType = iota
	Cloudy
	Overcast
	Fog
	LightRain
	HeavyRain
	Thunderstorm
	LightSnow
	HeavySnow
	Blizzard
	Windy
)

func (wt WeatherType) String() string {
	switch wt {
	case Clear:
		return "clear"
	case Cloudy:
		return "cloudy"
	case Overcast:
		return "overcast"
	case Fog:
		return "fog"
	case LightRain:
		return "light_rain"
	case HeavyRain:
		return "heavy_rain"
	case Thunderstorm:
		return "thunderstorm"
	case LightSnow:
		return "light_snow"
	case HeavySnow:
		return "heavy_snow"
	case Blizzard:
		return "blizzard"
	case Windy:
		return "windy"
	default:
		return "clear"
	}
}

func ParseWeatherType(s string) WeatherType {
	switch s {
	case "clear":
		return Clear
	case "cloudy":
		return Cloudy
	case "overcast":
		return Overcast
	case "fog":
		return Fog
	case "light_rain":
		return LightRain
	case "heavy_rain":
		return HeavyRain
	case "thunderstorm":
		return Thunderstorm
	case "light_snow":
		return LightSnow
	case "heavy_snow":
		return HeavySnow
	case "blizzard":
		return Blizzard
	case "windy":
		return Windy
	default:
		return Clear
	}
}

type Weather struct {
	Type        WeatherType `json:"type"`
	Temperature float64     `json:"temperature"`
	Visibility  float64     `json:"visibility"`
	WindSpeed   float64     `json:"wind_speed"`
	WindDir     string      `json:"wind_dir"`
	Humidity    float64     `json:"humidity"`
}

func NewWeather(t WeatherType, temp float64) *Weather {
	return &Weather{
		Type:        t,
		Temperature: temp,
		Visibility:  20,
		WindSpeed:   0,
		WindDir:     "N",
	}
}

func (w *Weather) VisibilityModifier() float64 {
	switch w.Type {
	case Clear:
		return 1.0
	case Cloudy:
		return 0.9
	case Overcast:
		return 0.7
	case Fog:
		return 0.2
	case LightRain:
		return 0.6
	case HeavyRain:
		return 0.4
	case Thunderstorm:
		return 0.3
	case LightSnow:
		return 0.5
	case HeavySnow:
		return 0.3
	case Blizzard:
		return 0.1
	case Windy:
		return 0.8
	default:
		return 1.0
	}
}

func (w *Weather) TravelSpeedModifier() float64 {
	switch w.Type {
	case Clear, Cloudy:
		return 1.0
	case Overcast:
		return 0.95
	case Fog, LightRain, LightSnow:
		return 0.8
	case HeavyRain, HeavySnow:
		return 0.6
	case Thunderstorm, Blizzard:
		return 0.3
	case Windy:
		return 0.85
	default:
		return 1.0
	}
}

// ClimateBias adjusts seasonal weather by region id.
type ClimateBias struct {
	TempOffset float64
	// PreferSnow, PreferRain, PreferClear, PreferFog shift rolls (0-1 additions to thresholds)
	Arid   bool // desert: less rain, more clear/windy
	Wet    bool // marches: more rain/fog
	Cold   bool // highlands: colder, more snow
	Forest bool // forest: more fog/overcast
}

func ClimateForRegion(regionID string) ClimateBias {
	switch regionID {
	case "northern_highlands":
		return ClimateBias{TempOffset: -8, Cold: true}
	case "sunken_marches":
		return ClimateBias{TempOffset: -2, Wet: true}
	case "golden_plains":
		return ClimateBias{TempOffset: 2}
	case "crystal_forest":
		return ClimateBias{TempOffset: -1, Forest: true}
	case "ash_desert":
		return ClimateBias{TempOffset: 10, Arid: true}
	default:
		return ClimateBias{}
	}
}

func GenerateWeather(season Season, rng *rand.Rand) *Weather {
	return GenerateWeatherFor(season, "", rng)
}

func GenerateWeatherFor(season Season, regionID string, rng *rand.Rand) *Weather {
	bias := ClimateForRegion(regionID)

	baseTemp := 15.0
	switch season {
	case Spring:
		baseTemp = 12
	case Summer:
		baseTemp = 25
	case Autumn:
		baseTemp = 10
	case Winter:
		baseTemp = -2
	}
	baseTemp += bias.TempOffset

	temp := baseTemp + (rng.Float64()-0.5)*10
	var wType WeatherType
	roll := rng.Float64()

	if bias.Arid {
		switch {
		case roll < 0.45:
			wType = Clear
		case roll < 0.7:
			wType = Cloudy
		case roll < 0.85:
			wType = Windy
		case roll < 0.93:
			wType = Overcast
		case roll < 0.97:
			wType = LightRain
		default:
			wType = Fog
		}
	} else if bias.Wet {
		switch {
		case roll < 0.15:
			wType = Clear
		case roll < 0.3:
			wType = Cloudy
		case roll < 0.45:
			wType = Overcast
		case roll < 0.65:
			wType = LightRain
		case roll < 0.8:
			wType = HeavyRain
		case roll < 0.9:
			wType = Fog
		case roll < 0.96:
			wType = Thunderstorm
		default:
			wType = Windy
		}
	} else if bias.Cold && season == Winter {
		switch {
		case roll < 0.2:
			wType = Clear
		case roll < 0.4:
			wType = Cloudy
		case roll < 0.55:
			wType = Overcast
		case roll < 0.7:
			wType = LightSnow
		case roll < 0.85:
			wType = HeavySnow
		case roll < 0.95:
			wType = Blizzard
		default:
			wType = Fog
		}
	} else if bias.Forest {
		switch {
		case roll < 0.25:
			wType = Clear
		case roll < 0.45:
			wType = Cloudy
		case roll < 0.65:
			wType = Overcast
		case roll < 0.78:
			wType = Fog
		case roll < 0.9:
			wType = LightRain
		default:
			wType = Windy
		}
	} else {
		switch season {
		case Summer:
			switch {
			case roll < 0.4:
				wType = Clear
			case roll < 0.7:
				wType = Cloudy
			case roll < 0.85:
				wType = LightRain
			case roll < 0.93:
				wType = HeavyRain
			case roll < 0.98:
				wType = Thunderstorm
			default:
				wType = Windy
			}
		case Winter:
			switch {
			case roll < 0.3:
				wType = Clear
			case roll < 0.5:
				wType = Cloudy
			case roll < 0.65:
				wType = Overcast
			case roll < 0.75:
				wType = LightSnow
			case roll < 0.85:
				wType = HeavySnow
			case roll < 0.92:
				wType = Blizzard
			case roll < 0.97:
				wType = Fog
			default:
				wType = Windy
			}
		default:
			switch {
			case roll < 0.35:
				wType = Clear
			case roll < 0.6:
				wType = Cloudy
			case roll < 0.75:
				wType = Overcast
			case roll < 0.85:
				wType = LightRain
			case roll < 0.92:
				wType = HeavyRain
			case roll < 0.96:
				wType = Fog
			default:
				wType = Windy
			}
		}
	}

	winds := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	windSpeed := rng.Float64() * 20
	if bias.Arid {
		windSpeed += 5
	}
	humidity := 0.5 + rng.Float64()*0.5
	if bias.Arid {
		humidity = 0.1 + rng.Float64()*0.3
	} else if bias.Wet {
		humidity = 0.7 + rng.Float64()*0.3
	}

	vis := 20.0
	if wType == Fog {
		vis = 0.5 + rng.Float64()*1.5
	}

	return &Weather{
		Type:        wType,
		Temperature: temp,
		Visibility:  vis,
		WindSpeed:   windSpeed,
		WindDir:     winds[rng.Intn(len(winds))],
		Humidity:    humidity,
	}
}

func (w *Weather) IsHarsh() bool {
	if w == nil {
		return false
	}
	switch w.Type {
	case Thunderstorm, Blizzard, HeavyRain, HeavySnow, Fog:
		return true
	default:
		return w.Temperature < -5 || w.Temperature > 35
	}
}

func (w *Weather) IsStormy() bool {
	if w == nil {
		return false
	}
	switch w.Type {
	case Thunderstorm, Blizzard, HeavyRain, HeavySnow:
		return true
	default:
		return false
	}
}
