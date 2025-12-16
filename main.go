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
	// Загружаем конфиг
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к БД через конфиг
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Репозиторий и сервис
	userRepo := user.NewUserRepository(database)
	userService := user.NewUserService(userRepo, "3b294c6ae8ae4dc1bebe1e3b50fbd216")

	// Шаблоны
	templates := template.Must(template.ParseGlob("templates/*.html"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Handler
	userHandler := user.NewUserHandler(userService, templates)

	// Роуты
	http.HandleFunc("/register", userHandler.RegisterPage)
	http.HandleFunc("/login", userHandler.LoginPage)
	http.HandleFunc("/dashboard", userHandler.DashboardPage)
	http.HandleFunc("/deposit", userHandler.DepositPage)
	http.HandleFunc("/transfer", userHandler.TransferPage)
	http.HandleFunc("/convert", userHandler.ConvertPage)
	http.HandleFunc("/transactions", userHandler.TransactionsPage)
	http.HandleFunc("/logout", userHandler.LogoutPage)
	http.HandleFunc("/about", userHandler.AboutPage)

	log.Println("Сервер запущен на http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
