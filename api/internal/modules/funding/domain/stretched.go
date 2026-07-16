package domain

// CalculateStretchedFundedAllowanceMinutes returns the monthly funded allowance
// in minutes for a stretched funding model. Formula: hours_per_week * 52 * 60 / 12.
func CalculateStretchedFundedAllowanceMinutes(fundedHoursPerWeek float64) int {
	if fundedHoursPerWeek < 0 {
		return 0
	}
	return int(fundedHoursPerWeek * 52 * 60 / 12)
}
