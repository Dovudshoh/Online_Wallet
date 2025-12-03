package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"online_bank/internal/user"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres password=Upup1748$$ dbname=online_bank sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo, "3b294c6ae8ae4dc1bebe1e3b50fbd216")


	templates := template.Must(template.ParseGlob("templates/*.html"))


	userHandler := user.NewUserHandler(userService, templates)


	http.HandleFunc("/register", userHandler.RegisterPage)
	http.HandleFunc("/login", userHandler.LoginPage)
	http.HandleFunc("/dashboard", userHandler.DashboardPage)
	http.HandleFunc("/deposit", userHandler.DepositPage)
	http.HandleFunc("/transfer", userHandler.TransferPage)
	http.HandleFunc("/convert", userHandler.ConvertPage)

	log.Println("Сервер запущен на http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
