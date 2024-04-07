func getUserIDFromRequest(r *http.Request) int {
	userIDParam := r.URL.Query().Get("userID")
	if userIDParam == "" {
		log.Println("ID пользователя не передан в параметрах запроса")
		return 0
	}
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		log.Println("Ошибка при извлечении ID пользователя из запроса:", err)
		return 0
	}
	return userID
}
func isFileVisibleToUser(fileName string, userID int) bool {
	query := "SELECT COUNT(*) FROM filess WHERE stored_filename = $1 AND recipient = $2"
	var count int
	err := Db.QueryRow(query, fileName, userID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Файл не найден для пользователя:", userID)
			return false
		}
		log.Println("Ошибка при выполнении запроса:", err)
		return false
	}
	return count > 0
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

	// Генерация уникального имени файла
	fileName := handler.Filename
	fileExt := filepath.Ext(fileName)
	uniqueID := time.Now().UnixNano()
	newFileName := fmt.Sprintf("%d_%s%s", uniqueID, fileName, fileExt)

	recipientName := r.FormValue("recipient") // Получаем имя получателя из формы
	userID := r.FormValue("userID")           // Получаем ID пользователя из формы

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

	saveFileToDB(fileName, newFileName, recipientName, userID) // Сохраняем информацию о файле в базе данных

	fmt.Fprintf(w, "Файл успешно загружен и сохранен на сервере.")
	http.Redirect(w, r, "/main", http.StatusFound)
}
func saveFileToDB(originalFilename, storedFilename, recipientName, userID string) {
	query := "INSERT INTO filess (original_filename, stored_filename, recipient, user_id, upload_timestamp) VALUES ($1, $2, $3, $4, $5)"
	_, err := Db.Exec(query, originalFilename, storedFilename, recipientName, userID, time.Now())
	if err != nil {
		log.Println("Ошибка при сохранении файла в базе данных:", err)
	}
}
func AuthenticateUser1(ctx context.Context, login, password string) User {
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

func RegisterHandler1(w http.ResponseWriter, r *http.Request) {
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

		// Проверяем, существует ли пользователь с таким логином
		var existingUserID int
		queryExistingUser := "SELECT id FROM users WHERE login = $1"
		err = Db.QueryRow(queryExistingUser, login).Scan(&existingUserID)
		if err == nil {
			// Пользователь с таким логином уже существует, выполняем аутентификацию
			id := AuthenticateUser(ctx, login, password)
			if id.IDuser != 0 {
				http.Redirect(w, r, "/main?id="+strconv.Itoa(id.IDuser), http.StatusFound)
				return
			}
		}

		// Создаем нового пользователя, если пользователь с таким логином не найден
		queryInsertUser := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"
		var userID int
		err = Db.QueryRow(queryInsertUser, login, password).Scan(&userID)
		if err != nil {
			log.Fatal("Failed to insert user:", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}

		fmt.Println("ID пользователя:", userID) // Вывод ID пользователя в терминал

		id := User{IDuser: userID}

		// Перенаправление на страницу загрузки файла с передачей ID в качестве параметра
		http.Redirect(w, r, "/main?id="+strconv.Itoa(id.IDuser), http.StatusFound)
		return
	}

	Tmpl1.Execute(w, nil)
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

	fmt.Println("ID пользователя:", currentUser.IDuser) // Вывод ID пользователя в терминал

	return currentUser
}

func RegisterHandlermain(w http.ResponseWriter, r *http.Request) {
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

func AuthenticateUsermain(ctx context.Context, login, password string) User {
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