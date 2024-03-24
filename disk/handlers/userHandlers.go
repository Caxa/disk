package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	IDuser int
}

var Db *sql.DB
var Tmpl1 = template.Must(template.ParseFiles("templates/register.html"))
var Files []string

func GetUserNameByIDFromDB(userID int) (string, error) {
	var userName string
	query := "SELECT login FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя из базы данных:", err)
		return "", err
	}
	return userName, nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	files, err := GetFilesByUserIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении файлов из базы данных", http.StatusInternalServerError)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		Files    []string
		UserName string
	}{
		Files:    files,
		UserName: userName,
	}

	tmpl := template.Must(template.ParseFiles("templates/upload.html"))

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetFilesByUserIDFromDB(userID int) ([]string, error) {
	var files []string
	query := "SELECT stored_filename FROM files WHERE recipient_id = $1"
	rows, err := Db.Query(query, userID)
	if err != nil {
		log.Println("Ошибка при получении файлов из базы данных:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fileName string
		if err := rows.Scan(&fileName); err != nil {
			log.Println("Ошибка при сканировании строки:", err)
			return nil, err
		}
		files = append(files, fileName)
	}

	return files, nil
}

func Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case "GET":
		Tmpl1.Execute(w, nil)
	case "POST":
		login := r.FormValue("login")
		password := r.FormValue("password")
		if login != "" && password != "" {
			id := AuthenticateUser(ctx, login, password)
			if id.IDuser != 0 {
				http.Redirect(w, r, "/main?id="+strconv.Itoa(id.IDuser), http.StatusFound)
				return
			} else {
				RegisterHandler(w, r) // Регистрация нового пользователя
				return
			}
		}
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
func AuthenticateUser(ctx context.Context, login, password string) User {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	query := "SELECT id FROM users WHERE login = $1 AND password = $2"
	var currentUser User
	err := Db.QueryRowContext(ctx, query, login, password).Scan(&currentUser.IDuser)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return User{} // Возвращаем пустую структуру User в случае ошибки
	}
	return currentUser
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method == "POST" {
		// Обработка POST запроса для регистрации пользователя
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusInternalServerError)
			return
		}
		login := r.Form.Get("login")
		password := r.Form.Get("password")

		query := "INSERT INTO users (login, password) VALUES ($1, $2)"
		_, err = Db.Exec(query, login, password)
		if err != nil {
			log.Fatal("Failed to insert user:", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}
		id := AuthenticateUser(ctx, login, password)

		// Перенаправление на страницу загрузки файла
		http.Redirect(w, r, "/main?id="+strconv.Itoa(id.IDuser), http.StatusFound)
		return
	}
	Tmpl1.Execute(w, nil)
}

func OpenDatabase() {
	var err error
	Db, err = sql.Open("postgres", "user=postgres password=1234 dbname=airat sslmode=disable")
	if err != nil {
		log.Println("Failed to connect to the database:", err)
	} else {
		err = Db.Ping()
		if err != nil {
			log.Println("Failed to ping the database:", err)
		} else {
			log.Println("Successfully connected to the database")
		}
	}
}
