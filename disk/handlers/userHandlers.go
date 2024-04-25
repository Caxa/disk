package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"

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
		Id       int
	}{
		Files:    files,
		UserName: userName,
		Id:       userID,
	}

	tmpl := template.Must(template.ParseFiles("templates/upload1.html"))

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetMYFilesByUserIDFromDB(userID int) ([]string, error) {
	var files []string
	query := "SELECT stored_filename FROM filess WHERE sender_id = $1"
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
func GetFilesByUserIDFromDB(userID int) ([]string, error) {
	var files []string
	query := "SELECT stored_filename FROM filess WHERE recipient_id = $1"
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
func Sent(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	files, errFiles := GetMYFilesByUserIDFromDB(userID)

	if errFiles != nil {
		http.Error(w, "Ошибка при получении файлов из базы данных", http.StatusInternalServerError)
		return
	}

	userName, errUserName := GetUserNameByIDFromDB(userID)
	if errUserName != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		Files    []string
		UserName string
		Id       int
	}{
		Files:    files,
		UserName: userName,
		Id:       userID,
	}

	tmpl := template.Must(template.ParseFiles("templates/sent_files.html"))

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		if login == "admin" && password == "admin" {
			// Если логин и пароль равны "admin", выводим таблицы users и files
			showTables(w)
			return
		} else if login != "" && password != "" {
			id := AuthenticateUser(ctx, login, password)
			if id == -1 {
				http.ServeFile(w, r, "templates/errorModal.html") // Обработка неверного пароля
				return
			} else if id == 0 {
				RegisterHandler(w, r) // Регистрация нового пользователя
				return
			} else {
				http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
				return
			}
		}
		http.ServeFile(w, r, "templates/errorModal.html")
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
func Home1(w http.ResponseWriter, r *http.Request) {
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
		if login == "admin" && password == "admin" {
			// Если логин и пароль равны "admin", выводим таблицы users и files
			showTables(w)
			return
		} else if login != "" && password != "" {
			id := AuthenticateUser(ctx, login, password)
			if id == -1 { //
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Ваш пароль неверный"))
				return
			} else if id == 0 {
				RegisterHandler(w, r) // Регистрация нового пользователя
				return
			} else {
				http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
				return
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Ваш пароль неверный"))
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
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

		hashedPassword, err := HashPassword(password) // Хешируем пароль
		if err != nil {
			log.Fatal("Failed to hash password:", err)
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		query := "INSERT INTO users (login, password) VALUES ($1, $2)"
		_, err = Db.Exec(query, login, hashedPassword) // Сохраняем хешированный пароль в базу данных
		if err != nil {
			log.Fatal("Failed to insert user:", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}
		id := AuthenticateUser(ctx, login, password)

		// Перенаправление на страницу загрузки файла
		http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
		return
	}
	Tmpl1.Execute(w, nil)
}

func AuthenticateUser(ctx context.Context, login, password string) int {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	query := "SELECT id, password FROM users WHERE login = $1"
	var currentUser User
	var hashedPassword string
	err := Db.QueryRowContext(ctx, query, login).Scan(&currentUser.IDuser, &hashedPassword)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return 0 // Возвращаем 0, если пользователя не найден
	}

	if CheckPasswordHash(password, hashedPassword) {
		return currentUser.IDuser // Возвращаем ID пользователя, если пароль верный
	} else {
		log.Println("Неверный пароль для пользователя:", login)
		return -1 // Возвращаем -1, если пароль неверный
	}
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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

func SearchFilesByQuery(searchQuery string) []string {
	var searchResults []string

	query := "SELECT stored_filename FROM files WHERE original_filename ILIKE $1 OR recipient_name ILIKE $1"
	rows, err := Db.Query(query, "%"+searchQuery+"%")
	if err != nil {
		log.Println("Ошибка при выполнении запроса поиска файлов:", err)
		return searchResults
	}
	defer rows.Close()

	for rows.Next() {
		var fileName string
		if err := rows.Scan(&fileName); err != nil {
			log.Println("Ошибка при сканировании строки:", err)
			continue
		}
		searchResults = append(searchResults, fileName)
	}

	return searchResults
}

func SearchFiles(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.FormValue("searchQuery")

	searchResults := SearchFilesByQuery(searchQuery)

	fmt.Fprintf(w, "Search Results:\n")
	for _, file := range searchResults {
		fmt.Fprintf(w, "%s\n", file)
	}
}

func showTables(w http.ResponseWriter) {
	Db, err := sql.Open("postgres", "user=postgres password=1234 dbname=airat sslmode=disable")
	if err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		log.Println("Failed to connect to the database:", err)
		return
	}
	defer Db.Close()

	// Запрос к таблице users для получения количества пользователей
	rowsUsers, err := Db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		http.Error(w, "Failed to query users table", http.StatusInternalServerError)
		log.Println("Failed to query users table:", err)
		return
	}
	defer rowsUsers.Close()

	var userCount int
	for rowsUsers.Next() {
		if err := rowsUsers.Scan(&userCount); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}
	}

	// Запрос к таблице files для получения количества файлов
	rowsFiles, err := Db.Query("SELECT COUNT(*) FROM files")
	if err != nil {
		http.Error(w, "Failed to query files table", http.StatusInternalServerError)
		log.Println("Failed to query files table:", err)
		return
	}
	defer rowsFiles.Close()

	var fileCount int
	for rowsFiles.Next() {
		if err := rowsFiles.Scan(&fileCount); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}
	}

	// Вывод общей информации о пользователях и файлах
	fmt.Fprintf(w, "Общее количество пользователей: %d\n", userCount)
	fmt.Fprintf(w, "Общее количество файлов: %d\n", fileCount)

	// Вывод данных таблицы users
	rowsUsersData, err := Db.Query("SELECT id, login FROM users")
	if err != nil {
		http.Error(w, "Failed to query users table", http.StatusInternalServerError)
		log.Println("Failed to query users table:", err)
		return
	}
	defer rowsUsersData.Close()

	fmt.Fprintln(w, "\nТаблица пользователей (users):")
	for rowsUsersData.Next() {
		var id int
		var login string
		if err := rowsUsersData.Scan(&id, &login); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}
		fmt.Fprintf(w, "ID: %d, Login: %s\n", id, login)
	}

	rowsFilesData, err := Db.Query("SELECT id, stored_filename, original_filename, recipient_id, recipient_name, upload_timestamp FROM files")
	if err != nil {
		http.Error(w, "Failed to query files table", http.StatusInternalServerError)
		log.Println("Failed to query files table:", err)
		return
	}
	defer rowsFilesData.Close()

	fmt.Fprintln(w, "\nТаблица файлов (files):")
	for rowsFilesData.Next() {
		var id int
		var storedFilename, originalFilename, recipientName string
		var recipientID int
		var uploadTime time.Time

		if err := rowsFilesData.Scan(&id, &storedFilename, &originalFilename, &recipientID, &recipientName, &uploadTime); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}

		uploadTimeString := uploadTime.Format("2006-01-02 15:04:05") // Форматирование времени в строку

		fmt.Fprintf(w, "ID: %d, Stored Filename: %s, Original Filename: %s, Recipient ID: %d, Recipient Name: %s, Upload Time: %s\n", id, storedFilename, originalFilename, recipientID, recipientName, uploadTimeString)
	}
}
