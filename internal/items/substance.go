package items

type SubstanceEffect struct {
	Name          string `json:"name"`
	Duration      int    `json:"duration"`
	CrashDuration int    `json:"crash_duration"`
	HealHP        int    `json:"heal_hp"`
	HealFP        int    `json:"heal_fp"`
	HealPerTick   int    `json:"heal_per_tick"`
	FPPerTick     int    `json:"fp_per_tick"`
	STR           int    `json:"str"`
	DEX           int    `json:"dex"`
	CON           int    `json:"con"`
	INT           int    `json:"int"`
	WIS           int    `json:"wis"`
	CHA           int    `json:"cha"`
	CrashSTR      int    `json:"crash_str"`
	CrashDEX      int    `json:"crash_dex"`
	CrashCON      int    `json:"crash_con"`
	CrashINT      int    `json:"crash_int"`
	CrashWIS      int    `json:"crash_wis"`
	CrashCHA      int    `json:"crash_cha"`
}
