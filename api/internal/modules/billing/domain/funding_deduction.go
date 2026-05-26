package domain

import "fmt"

type FundingDeductionCalculation struct {
	CoreAttendedMinutes    int
	FundedAllowanceMinutes int
	FundedDeductionMinutes int
	CoreBillableMinutes    int
}

func CalculateFundingDeduction(attendance AttendanceMinuteCalculation, fundedAllowanceMinutes int) (FundingDeductionCalculation, error) {
	coreAttendedMinutes := attendance.RoundedBillableMinutes

	if coreAttendedMinutes < 0 {
		return FundingDeductionCalculation{}, fmt.Errorf("core attended minutes must not be negative: %d", coreAttendedMinutes)
	}
	if fundedAllowanceMinutes < 0 {
		return FundingDeductionCalculation{}, fmt.Errorf("funded allowance minutes must not be negative: %d", fundedAllowanceMinutes)
	}

	fundedDeductionMinutes := min(coreAttendedMinutes, fundedAllowanceMinutes)
	coreBillableMinutes := max(0, coreAttendedMinutes-fundedAllowanceMinutes)

	return FundingDeductionCalculation{
		CoreAttendedMinutes:    coreAttendedMinutes,
		FundedAllowanceMinutes: fundedAllowanceMinutes,
		FundedDeductionMinutes: fundedDeductionMinutes,
		CoreBillableMinutes:    coreBillableMinutes,
	}, nil
}
