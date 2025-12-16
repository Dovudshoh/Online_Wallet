package user

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

)

type UserHandler struct {
	service   *UserService
	templates *template.Template
}

func NewUserHandler(service *UserService, templates *template.Template) *UserHandler {
	return &UserHandler{service: service, templates: templates}
}

func (h *UserHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := h.service.Register(name, email, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
func (h *UserHandler) AboutPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	profile, err := h.service.GetProfile(userID)
	if err != nil {
		log.Fatal(err)
	}
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "about.html", profile)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20)
		name := r.FormValue("full_name")
		bio := r.FormValue("bio")

		var avatar_path string
		file, handler, err := r.FormFile("avatar")
		if err == nil{
			defer file.Close()

			os.MkdirAll("uploads", os.ModePerm)
			avatar_path = fmt.Sprintf("uploads/%d_%s", userID, handler.Filename)

			dst, _ := os.Create(avatar_path)
			defer dst.Close()
			io.Copy(dst, file)
		}
		oldAvatar, err := h.service.GetAvatar(userID)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}


		err = h.service.UpProfile(name, bio, avatar_path, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if oldAvatar != "" {
			err := os.Remove(oldAvatar)
		if err != nil {
			log.Println("Не удалось удалить старый аватар:", err)
	}
}


		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}


func (h *UserHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		token, err := h.service.Login(email, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}


func (h *UserHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.service.GetBalance(userID)
	if err != nil{
		http.Error(w, err.Error(), http.StatusBadRequest)
			return
	}
	user.Avatar_path, err = h.service.GetAvatar(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.templates.ExecuteTemplate(w, "dashboard.html", user)
}

func (h *UserHandler) DepositPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "deposit.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		amountStr := r.FormValue("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}

		err = h.service.Deposit(userID, amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}


func (h *UserHandler) TransferPage(w http.ResponseWriter, r *http.Request) {
	fromID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}


	users, err := h.service.GetAllUsersExcept(fromID)
	if err != nil {
		http.Error(w, "failed to get users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "transfer.html", users)
		return
	}

	if r.Method == http.MethodPost {
		toIDStr := r.FormValue("to_id")
		amountStr := r.FormValue("amount")

		toID, err := strconv.Atoi(toIDStr)
		if err != nil {
			http.Error(w, "invalid recipient ID", http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}

		err = h.service.Transfer(fromID, toID, amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}



func (h *UserHandler) ConvertPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "convert.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		from := r.FormValue("from")
		to := r.FormValue("to")
		amountStr := r.FormValue("amount")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}

		rate, err := h.service.GetCurrencyRate(from, to)
		if err != nil {
			http.Error(w, "failed to get currency rate: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.service.ConvertCurrency(userID, from, to, amount, rate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
func (h *UserHandler) TransactionsPage(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.service.GetTransactions(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	

	h.templates.ExecuteTemplate(w, "transactions.html", user)
}
func (h *UserHandler) LogoutPage(w http.ResponseWriter, r *http.Request) {
	h.Logout(w, r)
}


func (h *UserHandler) getUserIDFromCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return 0, err
	}

	return h.service.GetUserIDByToken(cookie.Value)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    "",
        Path:     "/",
        MaxAge:   -1, 
        HttpOnly: true,
    })
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

