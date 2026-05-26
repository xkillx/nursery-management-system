package domain

import "testing"

func TestCalculateFundingDeduction(t *testing.T) {
	t.Run("core-only deduction", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 600}
		result, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result.CoreAttendedMinutes != 600 {
			t.Errorf("CoreAttendedMinutes: got %d want 600", result.CoreAttendedMinutes)
		}
		if result.FundedAllowanceMinutes != 120 {
			t.Errorf("FundedAllowanceMinutes: got %d want 120", result.FundedAllowanceMinutes)
		}
		if result.FundedDeductionMinutes != 120 {
			t.Errorf("FundedDeductionMinutes: got %d want 120", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 480 {
			t.Errorf("CoreBillableMinutes: got %d want 480", result.CoreBillableMinutes)
		}
	})

	t.Run("zero-floor behavior: allowance exceeds attendance", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 60}
		result, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result.FundedDeductionMinutes != 60 {
			t.Errorf("FundedDeductionMinutes: got %d want 60", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 0 {
			t.Errorf("CoreBillableMinutes: got %d want 0", result.CoreBillableMinutes)
		}
	})

	t.Run("zero allowance", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 600}
		result, err := CalculateFundingDeduction(att, 0)
		if err != nil {
			t.Fatal(err)
		}
		if result.FundedDeductionMinutes != 0 {
			t.Errorf("FundedDeductionMinutes: got %d want 0", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 600 {
			t.Errorf("CoreBillableMinutes: got %d want 600", result.CoreBillableMinutes)
		}
	})

	t.Run("zero attendance", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 0}
		result, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result.FundedDeductionMinutes != 0 {
			t.Errorf("FundedDeductionMinutes: got %d want 0", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 0 {
			t.Errorf("CoreBillableMinutes: got %d want 0", result.CoreBillableMinutes)
		}
	})

	t.Run("exact allowance match", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 120}
		result, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result.FundedDeductionMinutes != 120 {
			t.Errorf("FundedDeductionMinutes: got %d want 120", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 0 {
			t.Errorf("CoreBillableMinutes: got %d want 0", result.CoreBillableMinutes)
		}
	})

	t.Run("deterministic output", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 600}
		result1, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		result2, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result1 != result2 {
			t.Errorf("repeated calls with same inputs produced different results: %+v vs %+v", result1, result2)
		}
	})

	t.Run("result exposes only core funding components", func(t *testing.T) {
		att := AttendanceMinuteCalculation{
			RoundedBillableMinutes: 600,
			RawElapsedMinutes:      590,
			IncludedSessionCount:   3,
		}
		result, err := CalculateFundingDeduction(att, 120)
		if err != nil {
			t.Fatal(err)
		}
		if result.CoreAttendedMinutes != 600 {
			t.Errorf("CoreAttendedMinutes: got %d want 600", result.CoreAttendedMinutes)
		}
		if result.FundedAllowanceMinutes != 120 {
			t.Errorf("FundedAllowanceMinutes: got %d want 120", result.FundedAllowanceMinutes)
		}
		if result.FundedDeductionMinutes != 120 {
			t.Errorf("FundedDeductionMinutes: got %d want 120", result.FundedDeductionMinutes)
		}
		if result.CoreBillableMinutes != 480 {
			t.Errorf("CoreBillableMinutes: got %d want 480", result.CoreBillableMinutes)
		}
	})

	t.Run("negative funded allowance returns error", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: 600}
		_, err := CalculateFundingDeduction(att, -1)
		if err == nil {
			t.Fatal("expected error for negative funded allowance")
		}
	})

	t.Run("negative core attended minutes returns error", func(t *testing.T) {
		att := AttendanceMinuteCalculation{RoundedBillableMinutes: -1}
		_, err := CalculateFundingDeduction(att, 120)
		if err == nil {
			t.Fatal("expected error for negative core attended minutes")
		}
	})
}
