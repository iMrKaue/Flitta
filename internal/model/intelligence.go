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

// ServicePerformance representa os indicadores de desempenho de um serviço.
type ServicePerformance struct {
	Service           string  `json:"service"`
	TotalAppointments int64   `json:"total_appointments"`
	Completed         int64   `json:"completed"`
	Cancelled         int64   `json:"cancelled"`
	NoShow            int64   `json:"no_show"`
	CompletionRate    float64 `json:"completion_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
	NoShowRate        float64 `json:"no_show_rate"`
	CompletedRevenue  float64 `json:"completed_revenue"`
}

// WeekdayPerformance representa os indicadores agrupados por dia da semana.
type WeekdayPerformance struct {
	WeekdayNumber     int     `json:"weekday_number"`
	Weekday           string  `json:"weekday"`
	TotalAppointments int64   `json:"total_appointments"`
	Completed         int64   `json:"completed"`
	Cancelled         int64   `json:"cancelled"`
	NoShow            int64   `json:"no_show"`
	CompletionRate    float64 `json:"completion_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
	NoShowRate        float64 `json:"no_show_rate"`
	CompletedRevenue  float64 `json:"completed_revenue"`
}

// HourPerformance representa os indicadores agrupados por horário.
type HourPerformance struct {
	Hour              string  `json:"hour"`
	TotalAppointments int64   `json:"total_appointments"`
	Completed         int64   `json:"completed"`
	Cancelled         int64   `json:"cancelled"`
	NoShow            int64   `json:"no_show"`
	CompletionRate    float64 `json:"completion_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
	NoShowRate        float64 `json:"no_show_rate"`
	CompletedRevenue  float64 `json:"completed_revenue"`
}
