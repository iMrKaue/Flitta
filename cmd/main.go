package main

import (
	"flitta/internal/database"
	"flitta/internal/handler"
	"flitta/internal/middleware"
	"flitta/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Aviso: .env não carregado")
	} else {
		log.Println(".env carregado com sucesso")
	}

	database.ConnectDB()

	reminderScheduler := service.NewReminderScheduler(1*time.Minute, 24)
	reminderScheduler.Start()

	http.HandleFunc("/webhook", handler.WebhookHandler)
	http.HandleFunc("/appointments", middleware.AuthMiddleware(handler.GetAppointmentsHandler))

	http.HandleFunc("/admin/service/create", middleware.AuthMiddleware(handler.CreateServiceHandler))
	http.HandleFunc("/admin/service/list", middleware.AuthMiddleware(handler.GetServicesHandler))
	http.HandleFunc("/admin/service/delete", middleware.AuthMiddleware(handler.DeleteServiceHandler))

	http.HandleFunc("/admin/hours/set", middleware.AuthMiddleware(handler.SetWorkingHoursHandler))
	http.HandleFunc("/admin/hours/get", middleware.AuthMiddleware(handler.GetWorkingHourHandler))

	http.HandleFunc("/admin/company/get", middleware.AuthMiddleware(handler.GetCompanySettingsHandler))
	http.HandleFunc("/admin/company/update", middleware.AuthMiddleware(handler.UpdateCompanySettingsHandler))

	http.HandleFunc("/admin/reminders/pending", middleware.AuthMiddleware(handler.GetPendingRemindersHandler))
	http.HandleFunc("/admin/reminders/mark-sent", middleware.AuthMiddleware(handler.MarkReminderSentHandler))
	http.HandleFunc("/admin/reminders/send", middleware.AuthMiddleware(handler.SendReminderHandler))

	http.HandleFunc("/auth/register", handler.RegisterHandler)
	http.HandleFunc("/auth/login", handler.LoginHandler)

	http.HandleFunc("/dashboard/today", middleware.AuthMiddleware(handler.DashboardToday))

	log.Println("Server running on :8080")
	h := middleware.EnableCORS(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":8080", h))
}
