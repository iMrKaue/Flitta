package main

import (
	"flitta/internal/database"
	"flitta/internal/handler"
	"flitta/internal/middleware"
	"log"
	"net/http"

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

	http.HandleFunc("/webhook", handler.WebhookHandler)
	http.HandleFunc("/appointments", middleware.AuthMiddleware(handler.GetAppointmentsHandler))

	http.HandleFunc("/admin/service/create", middleware.AuthMiddleware(handler.CreateServiceHandler))
	http.HandleFunc("/admin/service/list", middleware.AuthMiddleware(handler.GetServicesHandler))
	http.HandleFunc("/admin/service/delete", middleware.AuthMiddleware(handler.DeleteServiceHandler))
	http.HandleFunc("/admin/hours/set", middleware.AuthMiddleware(handler.SetWorkingHoursHandler))
	http.HandleFunc("/admin/hours/get", middleware.AuthMiddleware(handler.GetWorkingHourHandler))

	http.HandleFunc("/auth/register", handler.RegisterHandler)
	http.HandleFunc("/auth/login", handler.LoginHandler)

	http.HandleFunc("/dashboard/today", middleware.AuthMiddleware(handler.DashboardToday))

	log.Println("Server running on :8080")
	h := middleware.EnableCORS(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":8080", h))
}
