package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	IDuser int
}

var Db *sql.DB
var Tmpl1 = template.Must(template.ParseFiles("register.html"))
var Files []string

func UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Устанавливаем максимальный размер файла 10MB

	storageDir := "Хранилище"
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755)
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusInternalServerError)
		log.Println("Ошибка при получении файла:", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "Файл загружен: %+v\n", handler.Filename)
	fmt.Fprintf(w, "Размер файла: %+v\n", handler.Size)

	dst, err := os.Create(filepath.Join(storageDir, handler.Filename))

	if err != nil {
		http.Error(w, "Ошибка при создании файла", http.StatusInternalServerError)
		log.Println("Ошибка при создании файла:", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Ошибка при сохранении файла на сервере", http.StatusInternalServerError)
		log.Println("Ошибка при сохранении файла на сервере:", err)
		return
	}

	fmt.Fprintf(w, "Файл успешно загружен и сохранен на сервере.")
	http.Redirect(w, r, "/main", http.StatusFound)
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
func Index(w http.ResponseWriter, r *http.Request) {
	// Получение списка файлов из папки "Хранилище" в текущей директории
	storageDir := "Хранилище"
	files, err := ioutil.ReadDir(storageDir)
	if err != nil {
		http.Error(w, "Ошибка при чтении файлов", http.StatusInternalServerError)
		log.Println("Ошибка при чтении файлов:", err)
		return
	}

	var Files []string
	for _, file := range files {
		Files = append(Files, file.Name())
	}

	data := struct {
		Files []string
	}{
		Files: Files,
	}

	tmpl := template.Must(template.ParseFiles("upload.html"))

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len("/download/"):]
	file, err := os.Open("/Users/airat/Downloads/Прога/Хранилище/" + fileName)
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, fileName, time.Now(), file)
}

func openDatabase() {
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

func main() {
	openDatabase()
	http.HandleFunc("/", Home)
	http.HandleFunc("/main", Index)
	http.HandleFunc("/upload", UploadFile)
	http.HandleFunc("/download/", DownloadFile)
	fmt.Println("Server is running on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
