package dto

type ChangePlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}
