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
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                        ),
                                                0
						),
					2
				),
				0
			) AS completion_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'cancelled'
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                ),
                                                0
						),
					2
				),
				0
			) AS cancellation_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'no_show'
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                ),
                                                0
						),
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

func GetServicePerformance(clientID int) ([]model.ServicePerformance, error) {
	rows, err := database.DB.Query(`
		SELECT
			service,

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
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                ),
                                                0
						),
					2
				),
				0
			) AS completion_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'cancelled'
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                ),
                                                0
						),
					2
				),
				0
			) AS cancellation_rate,

			COALESCE(
				ROUND(
					100.0 * COUNT(*) FILTER (
						WHERE status = 'no_show'
					) / NULLIF(
                                                COUNT(*) FILTER (
                                                        WHERE status IN ('completed', 'cancelled', 'no_show')
                                                ),
                                                0
						),
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
		GROUP BY service
		ORDER BY total_appointments DESC, service ASC
	`, clientID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []model.ServicePerformance{}

	for rows.Next() {
		var service model.ServicePerformance

		if err := rows.Scan(
			&service.Service,
			&service.TotalAppointments,
			&service.Completed,
			&service.Cancelled,
			&service.NoShow,
			&service.CompletionRate,
			&service.CancellationRate,
			&service.NoShowRate,
			&service.CompletedRevenue,
		); err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}
