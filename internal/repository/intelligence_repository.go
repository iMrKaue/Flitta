package repository

import (
	"flitta/internal/database"
	"flitta/internal/model"
)

func GetIntelligenceSummary(clientID int) (model.IntelligenceSummary, error) {
	var summary model.IntelligenceSummary

	err := database.DB.QueryRow(`
		SELECT
			COUNT(*) AS total_appointments,

			COUNT(*) FILTER (
				WHERE status = 'completed'
			) AS completed,

			COUNT(*) FILTER (
				WHERE status = 'cancelled'
			) AS cancelled,

			COUNT(*) FILTER (
				WHERE status = 'no_show'
			) AS no_show,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'completed'
					) / NULLIF(COUNT(*), 0),
					2
				),
				0
			) AS completion_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'cancelled'
					) / NULLIF(COUNT(*), 0),
					2
				),
				0
			) AS cancellation_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'no_show'
					) / NULLIF(COUNT(*), 0),
					2
				),
				0
			) AS no_show_rate,

			COALESCE(
				ROUND(
					SUM(price_snapshot) FILTER (
						WHERE status = 'completed'
					),
					2
				),
				0
			) AS completed_revenue

		FROM appointments
		WHERE client_id = $1
	`, clientID).Scan(
		&summary.TotalAppointments,
		&summary.Completed,
		&summary.Cancelled,
		&summary.NoShow,
		&summary.CompletionRate,
		&summary.CancellationRate,
		&summary.NoShowRate,
		&summary.CompletedRevenue,
	)

	if err != nil {
		return model.IntelligenceSummary{}, err
	}

	return summary, nil
}
