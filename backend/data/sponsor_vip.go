package data

func IsLocalFeatureEnabled(feature string) bool {
	switch feature {
	case "kline_analysis", "ai_assistant_web":
		return true
	default:
		return true
	}
}

// EffectiveSponsorVipLevel preserves the legacy feature-gating contract.
// This distribution has no paid local tier, so all local features are active.
func EffectiveSponsorVipLevel() (level int, active bool) {
	return 2, true
}
