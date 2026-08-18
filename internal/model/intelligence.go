package model

// IntelligenceSummary representa os principais indicadores do estabelecimento.
type IntelligenceSummary struct {
	TotalAppointments int64   `json:"total_appointments"`
	Completed         int64   `json:"completed"`
	Cancelled         int64   `json:"cancelled"`
	NoShow            int64   `json:"no_show"`
	CompletionRate    float64 `json:"completion_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
	NoShowRate        float64 `json:"no_show_rate"`
	CompletedRevenue  float64 `json:"completed_revenue"`
}
