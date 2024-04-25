package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"text/template"

	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"time"

	_ "github.com/lib/pq"
)

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
	recipientName := r.FormValue("recipient")
	senderName := r.FormValue("sender")

	// Создаем уникальное имя файла на основе времени отправления
	timeFormat := time.Now().Format("2006-01-02_15-04-05")
	newFileName := fmt.Sprintf("%s_%s->%s_%s", timeFormat, senderName, recipientName, fileName)

	userID, _ := getUserIDFromDB(recipientName)

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

	saveFileToDB1(fileName, newFileName, recipientName, userID)

	// Чтение содержимого HTML-шаблона
	tmpl, err := template.ParseFiles("templates/modal.html")
	if err != nil {
		http.Error(w, "Ошибка при чтении файла шаблона", http.StatusInternalServerError)
		log.Println("Ошибка при чтении файла шаблона:", err)
		return
	}

	data := struct {
		FileName      string
		FileSize      int64
		RecipientName string
	}{
		FileName:      fileName,
		FileSize:      handler.Size,
		RecipientName: recipientName,
	}

	tmpl.Execute(w, data)

}
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
	senderName := r.FormValue("sender")

	// Создаем уникальное имя файла на основе времени отправления
	timeFormat := time.Now().Format("2006-01-02_15-04-05")
	newFileName := fmt.Sprintf("%s_%s->%s_%s", timeFormat, senderName, recipientName, fileName)

	userID, _ := getUserIDFromDB(recipientName)
	senderID, _ := getUserIDFromDB(senderName)

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

	saveFileToDB(fileName, newFileName, recipientName, userID, senderName, senderID)

	// Чтение содержимого HTML-шаблона
	tmpl, err := template.ParseFiles("templates/modal.html")
	if err != nil {
		http.Error(w, "Ошибка при чтении файла шаблона", http.StatusInternalServerError)
		log.Println("Ошибка при чтении файла шаблона:", err)
		return
	}

	data := struct {
		FileName      string
		FileSize      int64
		RecipientName string
	}{
		FileName:      fileName,
		FileSize:      handler.Size,
		RecipientName: recipientName,
	}

	tmpl.Execute(w, data)

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

func saveFileToDB1(originalFilename, storedFilename, recipientName string, userID string) {
	query := "INSERT INTO files (original_filename, stored_filename, recipient_id, recipient_name, upload_timestamp) VALUES ($1, $2, $3, $4, $5)"
	_, err := Db.Exec(query, originalFilename, storedFilename, userID, recipientName, time.Now())
	if err != nil {
		log.Fatal("Failed to insert file into database:", err)
	}
}
func saveFileToDB(originalFilename, storedFilename, recipientName string, userID string, senderName string, senderID string) {
	query := "INSERT INTO filess (original_filename, stored_filename, recipient_id, recipient_name, sender_id,sender_name, upload_timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := Db.Exec(query, originalFilename, storedFilename, userID, recipientName, senderID, senderName, time.Now())
	if err != nil {
		log.Fatal("Failed to insert file into database:", err)
	}
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	fileToDelete := r.FormValue("fileToDelete")

	// Удаление файла из базы данных
	err := DeleteFileFromDB(fileToDelete)
	if err != nil {
		http.Error(w, "Ошибка при удалении файла из базы данных", http.StatusInternalServerError)
		return
	}

	// Удаление файла из хранилища
	err = DeleteFileFromStorage(fileToDelete)
	if err != nil {
		http.Error(w, "Ошибка при удалении файла из хранилища", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Файл успешно удален")
}

func DeleteFileFromDB(fileName string) error {
	query := "DELETE FROM filess WHERE stored_filename = $1"
	_, err := Db.Exec(query, fileName)
	if err != nil {
		log.Println("Ошибка при удалении файла из базы данных:", err)
		return err
	}
	return nil
}

func DeleteFileFromStorage(fileName string) error {
	storageDir := "storage"
	filePath := filepath.Join(storageDir, fileName)

	err := os.Remove(filePath)
	if err != nil {
		log.Println("Ошибка при удалении файла из хранилища:", err)
		return err
	}

	return nil
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
