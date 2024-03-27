package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Устанавливаем максимальный размер файла 10MB

	storageDir := "storage"
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

	fileName := handler.Filename
	recipientName := r.FormValue("recipient")
	// Создаем уникальное имя файла на основе времени отправления
	timeFormat := time.Now().Format("2006-01-02_15-04-05")
	//newFileName := fmt.Sprintf("%s_%s", timeFormat, fileName)

	newFileName := fmt.Sprintf("%s_%s_%s", timeFormat, recipientName, fileName)

	userID, _ := getUserIDFromDB(recipientName)

	fmt.Fprintf(w, "Файл загружен: %+v\n", handler.Filename)
	fmt.Fprintf(w, "Размер файла: %+v\n", handler.Size)
	fmt.Fprintf(w, "Отправлено пользователю: %s\n", recipientName)

	dst, err := os.Create(filepath.Join(storageDir, newFileName))
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

	saveFileToDB(fileName, newFileName, recipientName, userID)
}

func UploadFile1(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Устанавливаем максимальный размер файла 10MB

	storageDir := "storage"
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

	fileName := handler.Filename
	fileExt := filepath.Ext(fileName)

	// Получаем имя пользователя из запроса
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Создаем уникальное имя файла на основе времени отправления, имени пользователя и имени файла
	timeFormat := time.Now().Format("2006-01-02_15")
	newFileName := fmt.Sprintf("%s_%s_%s%s", timeFormat, userName, fileName, fileExt)

	fmt.Fprintf(w, "Файл загружен: %+v\n", handler.Filename)
	fmt.Fprintf(w, "Размер файла: %+v\n", handler.Size)
	fmt.Fprintf(w, "Отправлено пользователю: %s\n", userName)

	dst, err := os.Create(filepath.Join(storageDir, newFileName))
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

	// Преобразуем userID в строку
	userIDStr = strconv.Itoa(userID)

	// Сохраняем информацию о файле в базе данных
	saveFileToDB(fileName, newFileName, userName, userIDStr)
}

func getUserIDFromDB(username string) (string, error) {
	var userID string
	query := "SELECT id FROM users WHERE login = $1"
	err := Db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Пользователь с таким именем не найден в базе данных")
			return "", err
		}
		log.Println("Ошибка при получении ID пользователя из базы данных:", err)
		return "", err
	}
	return userID, nil
}

func saveFileToDB(originalFilename, storedFilename, recipientName string, userID string) {
	query := "INSERT INTO files (original_filename, stored_filename, recipient_id, recipient_name, upload_timestamp) VALUES ($1, $2, $3, $4, $5)"
	_, err := Db.Exec(query, originalFilename, storedFilename, userID, recipientName, time.Now())
	if err != nil {
		log.Fatal("Failed to insert file into database:", err)
	}
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len("/download/"):]
	filePath := filepath.Join("storage", fileName)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Ошибка при получении информации о файле", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Ошибка при отправке файла", http.StatusInternalServerError)
		return
	}
}
