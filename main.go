package main

import (
	"html/template"
	"log"
	"net/http"

	"online_bank/config"
	"online_bank/db"
	"online_bank/internal/user"
)

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	userRepo := user.NewUserRepository(database)
	userService := user.NewUserService(userRepo, cfg.APIKey)

	templates := template.Must(template.ParseGlob("templates/*.html"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	userHandler := user.NewUserHandler(userService, templates)

	// Публичные маршруты — без авторизации
	http.HandleFunc("/register", userHandler.RegisterPage)
	http.HandleFunc("/login", userHandler.LoginPage)

	// Защищённые маршруты — через middleware
	http.Handle("/dashboard", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.DashboardPage)))
	http.Handle("/deposit", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.DepositPage)))
	http.Handle("/transfer", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.TransferPage)))
	http.Handle("/convert", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.ConvertPage)))
	http.Handle("/transactions", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.TransactionsPage)))
	http.Handle("/about", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.AboutPage)))
	http.Handle("/logout", userHandler.AuthMiddleware(http.HandlerFunc(userHandler.LogoutPage)))

	log.Println("Сервер запущен на http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}