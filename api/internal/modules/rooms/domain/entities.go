package domain

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	BranchID    uuid.UUID
	Name        string
	Description *string
	AgeGroup    string
	Capacity    int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AgeGroup string

const (
	AgeGroupBaby     AgeGroup = "baby"
	AgeGroupToddler  AgeGroup = "toddler"
	AgeGroupPreschool AgeGroup = "preschool"
	AgeGroupMixed    AgeGroup = "mixed"
)

var ValidAgeGroups = map[AgeGroup]struct{}{
	AgeGroupBaby:     {},
	AgeGroupToddler:  {},
	AgeGroupPreschool: {},
	AgeGroupMixed:    {},
}

func IsValidAgeGroup(s string) bool {
	_, ok := ValidAgeGroups[AgeGroup(s)]
	return ok
}
