package models

type MilestoneDetails struct {
	ID          *int64  `json:"id,omitempty"`
	Number      *int    `json:"number,omitempty"`
	Description *string `json:"description,omitempty"`
}
