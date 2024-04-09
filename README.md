

# Быстрая Передача Файлов

Это веб-приложение предназначено для быстрой передачи файлов между пользователями. Проект включает в себя функционал регистрации пользователей и главную страницу для передачи файлов.

## Особенности

- Регистрация пользователей.
  - Хеширование пароля и проверка при аутификации.
  - Админ, который видит пользоваетелей и файлы,которые пользователи отправли.
  - Форма для ввода логина и пароля
- Главная страница для передачи файлов.
  - Отправка пользователю файла.
    - Сохранение файла в базе данных postgresSQL.
    - Сохранение файла в хранилище для загрузки. 
    - Оригинальный файл.
      - Название файла: время, отправитель, получатель, название файла (оригинальное).
  - Функционал поиска файлов. 
  - Демонстрация файлов.
  - Удаление файлов.
    - Удаление из базы данных postgres.
    - Удаление файла из Хранилища.
  - Выход из аккаунта.
  - Демонстрация имени пользователя.   
  - Темная тема.
  - Руссификация. 
  
- Фоновое изображение


## Стилизация

Используется шрифт Montserrat с помощью Google Fonts. Фоновое изображение добавляет эстетику к странице регистрации.

## Использование

- Зарегистрируйтесь, введя свой логин и пароль.
- Перейдите на главную страницу для передачи файлов.
- На главной странице вы сможете:
  - Загружать файлы, которые хотите отправить другим пользователям.
  - Скачивать файлы, которые были отправлены вам.
  - Просматривать список доступных файлов для загрузки.
  - Удалять ненужные файлы из системы.
  - Воспользоваться функционалом поиска для быстрого поиска конкретных файлов.
  - Выйти из аккаунта для безопасного завершения сеанса.


## Зависимости

- [github.com/lib/pq](https://github.com/lib/pq) - PostgreSQL driver for Go
- [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Package bcrypt implements Provos and Mazières's bcrypt adaptive hashing algorithm

## Настройка приложения

1. Убедитесь, что у вас установлен Go на вашей системе.
2. Клонируйте репозиторий на ваш компьютер.
3. Установите зависимости с помощью `go get`.
4. Настройте базу данных PostgreSQL и обновите детали подключения в коде.
5. Запустите приложение с помощью `go run main.go`.

## Использование

- Доступ к приложению по адресу `http://localhost:8080`.
- Зарегистрируйтесь как новый пользователь или войдите с существующими учетными данными.
- Загружайте файлы, скачивайте файлы и удаляйте файлы по мере необходимости.
- Используйте функционал поиска для поиска конкретных файлов.

## Структура кода

Приложение структурировано на различные пакеты:

- `main` содержит главную точку входа в приложение.
- `handlers` содержит обработчики HTTP-запросов для различных функциональностей.
- `templates` содержит Frontend оформление сервиса.
- `storage` репозиторий, в который сохраняются файлы. 
  

## Авторы

- [Шарушев Айрат](https://github.com/Caxa) - Разработчик Backend.
- [Маликов Арсений](https://github.com/ne0xis) - Разработчик Frontend.
 
