package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/egosha7/site-go/internal/domain"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
)

// PostgresRepository представляет интерфейс для работы с данными пользователей.
type PostgresRepository interface {
	PuppiesGet(chocolates, genders []string, idPuppy, readyToMove string, archived bool) ([]domain.Puppy, error)
	PuppyGet(idPuppy string) (*domain.Puppy, error)
	PuppyUpdate(puppy *domain.Puppy) (map[string]struct{}, map[string]struct{}, error)
	PuppyAdd(puppy *domain.Puppy) error
	PuppyDelete(puppyID string) ([]string, error)
	PuppyChangeArchived(puppyID, archived, city, phone string) error
	PuppiesWithReviewsGet() (map[int]int, error)
	DogsGet(chocolates, genders []string, id string, archived bool) ([]domain.Dog, error)
	DogGet(idDog string) (*domain.Dog, error)
	DogChangeArchived(puppyID string, archived string) error
	DogAdd(puppy *domain.Dog) error
	DogUpdate(dog *domain.Dog) (map[string]struct{}, map[string]struct{}, error)
	ReviewsPuppyNameGet() (map[int]string, error)
	ReviewsGet(idReview string, checked bool) ([]domain.Feedback, error)
	FeedbackGet(idPuppy, verified string) (*domain.Feedback, error)
	FeedbackUpdate(feedback *domain.Feedback) (map[string]struct{}, map[string]struct{}, error)
	FeedbackAdd(feedback *domain.Feedback) error
	FeedbackDelete(feedbackID string) ([]string, error)
	FeedbackChangeChecked(feedbackID string, checked string) error
	GetReviews(ctx context.Context, puppies []domain.Puppy) (map[int]string, error)
	GetByUsername(username string) (*domain.User, error)
	CheckUniqUser(login string) (bool, error)
	CheckValidUser(login string) (int, string, error)
	SaveUser(login string, hash string) error
}

// PostgresRepo представляет репозиторий для работы с PostgresSQL.
type PostgresRepo struct {
	logger *zap.Logger
	pool   *pgxpool.Pool
}

// FeedbackChangeChecked меняет состояние у отзыва о щенке в базе данных
func (r *PostgresRepo) FeedbackChangeChecked(feedbackID string, checked string) error {
	// Подготовка SQL-запроса
	query := `UPDATE reviews SET verified = $1 WHERE id = $2`

	// Выполнение запроса
	_, err := r.pool.Exec(context.Background(), query, checked, feedbackID)
	if err != nil {
		return err
	}

	return nil
}

// FeedbackDelete удаляет отзыв о щенке в базе данных
func (r *PostgresRepo) FeedbackDelete(feedbackID string) ([]string, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Получаем ссылки на изображения для удаления из S3
	var imgUrls []string
	query := `SELECT i.url
		FROM img_urls i
		JOIN reviews_img ri ON ri.img_url_id = i.id
		WHERE ri.reviews_id = $1`

	rows, err := tx.Query(context.Background(), query, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch image URLs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var imgUrl string
		err = rows.Scan(&imgUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to scan image URL: %w", err)
		}
		imgUrls = append(imgUrls, imgUrl)
	}

	// Удаляем записи из таблицы reviews_img
	query = `DELETE FROM reviews_img
		WHERE reviews_id = $1`
	_, err = tx.Exec(context.Background(), query, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("unable to delete from reviews_img: %w", err)
	}

	// Удаляем записи из таблицы img_urls
	for _, imgUrl := range imgUrls {
		query = "DELETE FROM img_urls WHERE url = $1"
		_, err = tx.Exec(context.Background(), query, imgUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to delete from img_urls: %w", err)
		}
	}

	query = "DELETE FROM reviews WHERE id = $1"
	_, err = tx.Exec(context.Background(), query, feedbackID)

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return imgUrls, nil
}

// FeedbackUpdate обновляет отзыв о щенке в базе данных
func (r *PostgresRepo) FeedbackUpdate(feedback *domain.Feedback) (map[string]struct{}, map[string]struct{}, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(context.Background())

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `UPDATE reviews
	SET puppy_id=$1, name=$2, number=$3, title=$4, verified=$5, date=$6
	WHERE id=$7 RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, feedback.PuppyID, feedback.Name, feedback.Number, feedback.Title,
		feedback.Verified, feedback.Date, feedback.ID,
	).Scan(&feedback.ID)
	if err != nil {
		return nil, nil, err
	}

	// Получение текущих URL-адресов изображений для щенка
	currentUrls := []string{}
	query = `SELECT i.url FROM img_urls i
             JOIN reviews_img ri ON i.id = ri.img_url_id
             WHERE ri.reviews_id = $1`
	rows, err := tx.Query(context.Background(), query, feedback.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current image URLs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, nil, fmt.Errorf("failed to scan image URL: %w", err)
		}
		currentUrls = append(currentUrls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate over image URLs: %w", err)
	}

	// Создание набора текущих и новых URL-адресов для удобства сравнения
	currentUrlSet := make(map[string]struct{}, len(currentUrls))
	for _, url := range currentUrls {
		currentUrlSet[url] = struct{}{}
	}
	newUrlSet := make(map[string]struct{}, len(feedback.Urls))
	for _, url := range feedback.Urls {
		newUrlSet[url] = struct{}{}
	}

	// Удаление старых URL-адресов, которые больше не существуют
	for url := range currentUrlSet {
		if _, exists := newUrlSet[url]; !exists {
			// Удаление связи между щенком и изображением
			query = `DELETE FROM reviews_img WHERE reviews_id = $1 AND img_url_id = (SELECT id FROM img_urls WHERE url = $2)`
			_, err = tx.Exec(context.Background(), query, feedback.ID, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image relationship: %w", err)
			}
			// Удаление записи URL из таблицы img_urls
			query = `DELETE FROM img_urls WHERE url = $1`
			_, err = tx.Exec(context.Background(), query, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image URL: %w", err)
			}
		}
	}

	// Добавление новых URL-адресов
	for url := range newUrlSet {
		if _, exists := currentUrlSet[url]; !exists {
			// Добавление URL в таблицу img_urls
			var imgID int
			query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
			err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert new image URL: %w", err)
			}
			// Связывание изображения с щенком в таблице puppies_img
			query = `INSERT INTO reviews_img (reviews_id, img_url_id) VALUES ($1, $2)`
			_, err = tx.Exec(context.Background(), query, feedback.ID, imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert image relationship: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return currentUrlSet, newUrlSet, nil
}

// FeedbackAdd добавляет отзыв о щенке в базу данных
func (r *PostgresRepo) FeedbackAdd(feedback *domain.Feedback) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(
		context.Background(), `SELECT puppy_id FROM puppies_member WHERE "number" = $1`, feedback.Number,
	).Scan(&feedback.PuppyID)
	if err != nil {
		return err
	}

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `INSERT INTO reviews(
	puppy_id, name, "number", title, verified, date)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, feedback.PuppyID, feedback.Name, feedback.Number, feedback.Title,
		feedback.Verified, feedback.Date,
	).Scan(&feedback.ID)
	if err != nil {
		return err
	}

	// Добавляем URL изображений в таблицу img_urls и связываем их с щенком
	for _, url := range feedback.Urls {
		// Добавляем URL в таблицу img_urls
		var imgID int
		query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
		err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
		if err != nil {
			return err
		}

		// Связываем изображение с щенком в таблице puppies_img
		query = `INSERT INTO reviews_img (reviews_id, img_url_id) VALUES ($1, $2)`
		_, err = tx.Exec(context.Background(), query, feedback.ID, imgID)
		if err != nil {
			return err
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// FeedbackGet получает отзыв из базы данных
func (r *PostgresRepo) FeedbackGet(idPuppy, verified string) (*domain.Feedback, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT r.*, array_agg(i.url) as urls FROM reviews r"
	query += " LEFT JOIN reviews_img ri ON r.id = ri.reviews_id"
	query += " LEFT JOIN img_urls i ON ri.img_url_id = i.id"
	query += " WHERE 1=1"

	// Добавление условий в запрос
	switch verified {
	case "true":
		query += " AND r.verified = 'true'"
	case "false":
		query += " AND r.verified = 'false'"
	default:
		break
	}

	if idPuppy != "" {
		query += " AND r.puppy_id = '" + idPuppy + "'"
	}

	query += " GROUP BY r.id" // Группируем результаты по ID щенка

	log.Println(query)

	feedback := &domain.Feedback{}

	row := r.pool.QueryRow(context.Background(), query)

	err := row.Scan(
		&feedback.ID, &feedback.PuppyID, &feedback.Name, &feedback.Number, &feedback.Title, &feedback.Verified,
		&feedback.Date,
		pq.Array(&feedback.Urls),
	)

	log.Println(feedback.ID, feedback.Name)

	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// ReviewsPuppyNameGet получает список щенков у которых есть отзыв в базе данных
func (r *PostgresRepo) ReviewsPuppyNameGet() (map[int]string, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT r.id, p.name FROM reviews r"
	query += " LEFT JOIN puppies p ON r.puppy_id = p.id"
	query += " WHERE 1=1"
	query += " GROUP BY r.id, p.name" // Группируем результаты по ID отзыва

	log.Println(query)

	// Выполнение SQL-запроса
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puppyNames = make(map[int]string) // Словарь для хранения

	// Итерация по рядам
	for rows.Next() {
		// Создание новой структуры Puppy для хранения каждого ряда
		var feedback int

		var namePuppy sql.NullString

		// Сканирование значений из текущего ряда в поля структуры
		err := rows.Scan(
			&feedback, &namePuppy,
		)
		if err != nil {
			return nil, err

		}

		if namePuppy.Valid {
			puppyNames[feedback] = namePuppy.String
		}
	}

	// Проверка на ошибки во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return puppyNames, nil
}

// ReviewsGet получает список отзывов из базы данных
func (r *PostgresRepo) ReviewsGet(idReview string, checked bool) ([]domain.Feedback, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT r.*, array_agg(i.url) as urls FROM reviews r"
	query += " LEFT JOIN reviews_img ri ON r.id = ri.reviews_id"
	query += " LEFT JOIN img_urls i ON ri.img_url_id = i.id"
	query += " WHERE 1=1"

	// Добавление условий в запрос
	switch checked {
	case true:
		query += " AND r.verified = 'true'"
	case false:
		query += " AND r.verified = 'false'"
	default:
		break
	}

	if idReview != "" {
		query += " AND r.id = '" + idReview + "'"
	}

	query += " GROUP BY r.id" // Группируем результаты по ID щенка

	// Добавление сортировки по ID
	query += " ORDER BY r.id DESC"

	log.Println(query)

	// Выполнение SQL-запроса
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Создание слайса для хранения результатов
	reviews := make([]domain.Feedback, 0)

	// Итерация по рядам
	for rows.Next() {
		// Создание новой структуры Puppy для хранения каждого ряда
		var feedback domain.Feedback
		var puppyID sql.NullInt32

		// Сканирование значений из текущего ряда в поля структуры
		err := rows.Scan(
			&feedback.ID, &puppyID, &feedback.Name, &feedback.Number, &feedback.Title, &feedback.Verified,
			&feedback.Date,
			pq.Array(&feedback.Urls),
		)
		if err != nil {
			return nil, err
		}

		if puppyID.Valid {
			feedback.PuppyID = int(puppyID.Int32)
		} else {
			feedback.PuppyID = 0
		}

		// Добавление структуры к слайсу
		reviews = append(reviews, feedback)
	}

	// Проверка на ошибки во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// DogUpdate обновляет информацию о собаке в базе данных
func (r *PostgresRepo) DogUpdate(dog *domain.Dog) (map[string]struct{}, map[string]struct{}, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(context.Background())

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `UPDATE adult_dogs
	SET name=$1, title=$2, gender=$3, color=$4, archived=$5
	WHERE id=$6 RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, dog.Name, dog.Title, dog.Gender, dog.Color, dog.Archived, dog.ID,
	).Scan(&dog.ID)
	if err != nil {
		return nil, nil, err
	}

	// Получение текущих URL-адресов изображений для щенка
	currentUrls := []string{}
	query = `SELECT i.url FROM img_urls i
             JOIN adult_dogs_img di ON i.id = di.img_url_id
             WHERE di.adult_dogs_id = $1`
	rows, err := tx.Query(context.Background(), query, dog.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current image URLs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, nil, fmt.Errorf("failed to scan image URL: %w", err)
		}
		currentUrls = append(currentUrls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate over image URLs: %w", err)
	}

	// Создание набора текущих и новых URL-адресов для удобства сравнения
	currentUrlSet := make(map[string]struct{}, len(currentUrls))
	for _, url := range currentUrls {
		currentUrlSet[url] = struct{}{}
	}
	newUrlSet := make(map[string]struct{}, len(dog.Urls))
	for _, url := range dog.Urls {
		newUrlSet[url] = struct{}{}
	}

	// Удаление старых URL-адресов, которые больше не существуют
	for url := range currentUrlSet {
		if _, exists := newUrlSet[url]; !exists {
			// Удаление связи между щенком и изображением
			query = `DELETE FROM adult_dogs_img WHERE adult_dogs_id = $1 AND img_url_id = (SELECT id FROM img_urls WHERE url = $2)`
			_, err = tx.Exec(context.Background(), query, dog.ID, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image relationship: %w", err)
			}
			// Удаление записи URL из таблицы img_urls
			query = `DELETE FROM img_urls WHERE url = $1`
			_, err = tx.Exec(context.Background(), query, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image URL: %w", err)
			}
		}
	}

	// Добавление новых URL-адресов
	for url := range newUrlSet {
		if _, exists := currentUrlSet[url]; !exists {
			// Добавление URL в таблицу img_urls
			var imgID int
			query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
			err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert new image URL: %w", err)
			}
			// Связывание изображения с щенком в таблице puppies_img
			query = `INSERT INTO adult_dogs_img (adult_dogs_id, img_url_id) VALUES ($1, $2)`
			_, err = tx.Exec(context.Background(), query, dog.ID, imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert image relationship: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return currentUrlSet, newUrlSet, nil
}

// DogChangeArchived меняет состояние архива у собаки в базе данных
func (r *PostgresRepo) DogChangeArchived(dogID string, archived string) error {
	// Подготовка SQL-запроса
	query := `UPDATE adult_dogs SET archived = $1 WHERE id = $2`

	// Выполнение запроса
	_, err := r.pool.Exec(context.Background(), query, archived, dogID)
	if err != nil {
		return err
	}

	return nil
}

// DogAdd добавляет информацию о собаке в базу данных
func (r *PostgresRepo) DogAdd(dog *domain.Dog) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `INSERT INTO adult_dogs (name, title, gender, color, archived)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, dog.Name, dog.Title, dog.Gender, dog.Color, dog.Archived,
	).Scan(&dog.ID)
	if err != nil {
		return err
	}

	// Добавляем URL изображений в таблицу img_urls и связываем их с щенком
	for _, url := range dog.Urls {
		// Добавляем URL в таблицу img_urls
		var imgID int
		query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
		err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
		if err != nil {
			return err
		}

		// Связываем изображение с щенком в таблице puppies_img
		query = `INSERT INTO adult_dogs_img (adult_dogs_id, img_url_id) VALUES ($1, $2)`
		_, err = tx.Exec(context.Background(), query, dog.ID, imgID)
		if err != nil {
			return err
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// PuppyChangeArchived меняет состояние архива у щенка в базе данных
func (r *PostgresRepo) PuppyChangeArchived(puppyID, archived, city, phone string) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Обновляем поле archived и city в таблице puppies
	query := `UPDATE puppies SET archived = $1, city = $2 WHERE id = $3`
	_, err = tx.Exec(context.Background(), query, archived, city, puppyID)
	if err != nil {
		return err
	}

	// Проверяем, существует ли запись в таблице puppies_member с данным puppyID
	var count int
	query = `SELECT COUNT(*) FROM puppies_member WHERE puppy_id = $1`
	err = tx.QueryRow(context.Background(), query, puppyID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Если записи нет, вставляем новую запись
		query = `INSERT INTO puppies_member (puppy_id, number) VALUES ($1, $2)`
		_, err = tx.Exec(context.Background(), query, puppyID, phone)
		if err != nil {
			return err
		}
	} else {
		// Если запись есть, обновляем существующую запись
		query = `UPDATE puppies_member SET number = $1 WHERE puppy_id = $2`
		_, err = tx.Exec(context.Background(), query, phone, puppyID)
		if err != nil {
			return err
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// PuppyDelete удаляет щенка в базе данных
func (r *PostgresRepo) PuppyDelete(puppyID string) ([]string, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Получаем ссылки на изображения для удаления из S3
	var imgUrls []string
	query := `SELECT i.url
		FROM img_urls i
		JOIN puppies_img pi ON pi.img_url_id = i.id
		WHERE pi.puppy_id = $1
		UNION
		SELECT i.url
		FROM img_urls i
		JOIN reviews_img ri ON ri.img_url_id = i.id
		JOIN reviews r ON r.id = ri.reviews_id
		WHERE r.puppy_id = $1`

	rows, err := tx.Query(context.Background(), query, puppyID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch image URLs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var imgUrl string
		err = rows.Scan(&imgUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to scan image URL: %w", err)
		}
		imgUrls = append(imgUrls, imgUrl)
	}

	// Удаляем записи из таблицы puppies_img
	query = "DELETE FROM puppies_img WHERE puppy_id = $1"
	_, err = tx.Exec(context.Background(), query, puppyID)
	if err != nil {
		return nil, fmt.Errorf("unable to delete from puppies_img: %w", err)
	}

	// Удаляем записи из таблицы reviews_img
	query = `DELETE FROM reviews_img
		WHERE reviews_id IN (
			SELECT id FROM reviews WHERE puppy_id = $1
		)`
	_, err = tx.Exec(context.Background(), query, puppyID)
	if err != nil {
		return nil, fmt.Errorf("unable to delete from reviews_img: %w", err)
	}

	// Удаляем записи из таблицы img_urls
	for _, imgUrl := range imgUrls {
		query = "DELETE FROM img_urls WHERE url = $1"
		_, err = tx.Exec(context.Background(), query, imgUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to delete from img_urls: %w", err)
		}
	}

	// Удаляем щенка из других таблиц
	tables := []string{"puppies_member", "reviews"}
	for _, table := range tables {
		query = fmt.Sprintf("DELETE FROM %s WHERE puppy_id = $1", table)
		_, err = tx.Exec(context.Background(), query, puppyID)
		if err != nil {
			return nil, fmt.Errorf("unable to delete from table %s: %w", table, err)
		}
	}

	// Удаляем щенка из таблицы puppies
	query = "DELETE FROM puppies WHERE id = $1"
	_, err = tx.Exec(context.Background(), query, puppyID)
	if err != nil {
		return nil, fmt.Errorf("unable to delete from puppies: %w", err)
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return imgUrls, nil
}

// PuppyAdd добавляет информацию о щенке в базу данных
func (r *PostgresRepo) PuppyAdd(puppy *domain.Puppy) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `INSERT INTO puppies (name, title, gender, price, ready_out, archived, city, mother_id, father_id, date_birth, color) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, puppy.Name, puppy.Title, puppy.Sex, puppy.Price, puppy.ReadyOut, puppy.Archived,
		puppy.City, puppy.MotherID, puppy.FatherID, puppy.DateBirth, puppy.Color,
	).Scan(&puppy.ID)
	if err != nil {
		return err
	}

	// Добавляем URL изображений в таблицу img_urls и связываем их с щенком
	for _, url := range puppy.Urls {
		// Добавляем URL в таблицу img_urls
		var imgID int
		query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
		err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
		if err != nil {
			return err
		}

		// Связываем изображение с щенком в таблице puppies_img
		query = `INSERT INTO puppies_img (puppy_id, img_url_id) VALUES ($1, $2)`
		_, err = tx.Exec(context.Background(), query, puppy.ID, imgID)
		if err != nil {
			return err
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// PuppyUpdate обновляет информацию о щенке в базе данных
func (r *PostgresRepo) PuppyUpdate(puppy *domain.Puppy) (map[string]struct{}, map[string]struct{}, error) {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(context.Background())

	// Добавляем информацию о щенке в таблицу puppies и получаем созданный ID
	query := `UPDATE puppies
	SET name=$1, title=$2, gender=$3, price=$4, ready_out=$5, archived=$6, city=$7, mother_id=$8, father_id=$9, date_birth=$10, color=$11
	WHERE id=$12 RETURNING id`
	err = tx.QueryRow(
		context.Background(), query, puppy.Name, puppy.Title, puppy.Sex, puppy.Price, puppy.ReadyOut, puppy.Archived,
		puppy.City, puppy.MotherID, puppy.FatherID, puppy.DateBirth, puppy.Color, puppy.ID,
	).Scan(&puppy.ID)
	if err != nil {
		return nil, nil, err
	}

	// Получение текущих URL-адресов изображений для щенка
	currentUrls := []string{}
	query = `SELECT i.url FROM img_urls i
             JOIN puppies_img pi ON i.id = pi.img_url_id
             WHERE pi.puppy_id = $1`
	rows, err := tx.Query(context.Background(), query, puppy.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current image URLs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, nil, fmt.Errorf("failed to scan image URL: %w", err)
		}
		currentUrls = append(currentUrls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate over image URLs: %w", err)
	}

	// Создание набора текущих и новых URL-адресов для удобства сравнения
	currentUrlSet := make(map[string]struct{}, len(currentUrls))
	for _, url := range currentUrls {
		currentUrlSet[url] = struct{}{}
	}
	newUrlSet := make(map[string]struct{}, len(puppy.Urls))
	for _, url := range puppy.Urls {
		newUrlSet[url] = struct{}{}
	}

	// Удаление старых URL-адресов, которые больше не существуют
	for url := range currentUrlSet {
		if _, exists := newUrlSet[url]; !exists {
			// Удаление связи между щенком и изображением
			query = `DELETE FROM puppies_img WHERE puppy_id = $1 AND img_url_id = (SELECT id FROM img_urls WHERE url = $2)`
			_, err = tx.Exec(context.Background(), query, puppy.ID, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image relationship: %w", err)
			}
			// Удаление записи URL из таблицы img_urls
			query = `DELETE FROM img_urls WHERE url = $1`
			_, err = tx.Exec(context.Background(), query, url)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to delete image URL: %w", err)
			}
		}
	}

	// Добавление новых URL-адресов
	for url := range newUrlSet {
		if _, exists := currentUrlSet[url]; !exists {
			// Добавление URL в таблицу img_urls
			var imgID int
			query = `INSERT INTO img_urls (url) VALUES ($1) RETURNING id`
			err = tx.QueryRow(context.Background(), query, url).Scan(&imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert new image URL: %w", err)
			}
			// Связывание изображения с щенком в таблице puppies_img
			query = `INSERT INTO puppies_img (puppy_id, img_url_id) VALUES ($1, $2)`
			_, err = tx.Exec(context.Background(), query, puppy.ID, imgID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert image relationship: %w", err)
			}
		}
	}

	// Коммитим транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return currentUrlSet, newUrlSet, nil
}

// PuppiesWithReviewsGet получает список щенков у которых есть отзыв в базе данных
func (r *PostgresRepo) PuppiesWithReviewsGet() (map[int]int, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT r.id, r.puppy_id FROM reviews r"
	query += " WHERE r.puppy_id > 0"

	log.Println(query)

	// Выполнение SQL-запроса
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puppyReviews = make(map[int]int) // Словарь для хранения ID отзывов

	// Итерация по рядам
	for rows.Next() {
		// Создание новой структуры Puppy для хранения каждого ряда
		var puppyID int

		var reviewID int

		// Сканирование значений из текущего ряда в поля структуры
		err := rows.Scan(
			&puppyID, &reviewID,
		)
		if err != nil {
			return nil, err
		}

		puppyReviews[puppyID] = reviewID
	}

	// Проверка на ошибки во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return puppyReviews, nil
}

// PuppiesGet получает список щенков в базе данных
func (r *PostgresRepo) PuppiesGet(chocolates, genders []string, idPuppy, readyToMove string, archived bool) ([]domain.Puppy, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT p.*, array_agg(i.url) as urls FROM puppies p"
	query += " LEFT JOIN puppies_img pi ON p.id = pi.puppy_id"
	query += " LEFT JOIN img_urls i ON pi.img_url_id = i.id"
	query += " WHERE 1=1"

	// Добавление условий в запрос
	if len(chocolates) > 0 {
		query += " AND p.color IN ("
		for _, c := range chocolates {
			query += "'" + c + "',"
		}
		query = strings.TrimSuffix(query, ",") + ")"
	}

	if len(genders) > 0 {
		query += " AND p.gender IN ("
		for _, g := range genders {
			query += "'" + g + "',"
		}
		query = strings.TrimSuffix(query, ",") + ")"
	}

	if readyToMove != "" {
		query += " AND p.ready_out = '" + readyToMove + "'"
	}

	if archived {
		query += " AND p.archived = 'true'"
	} else {
		query += " AND p.archived = 'false'"
	}

	if idPuppy != "" {
		query += " AND p.id = '" + idPuppy + "'"
	}

	query += " GROUP BY p.id" // Группируем результаты по ID щенка

	// Добавление сортировки по ID
	query += " ORDER BY p.id DESC"

	log.Println(query)

	// Выполнение SQL-запроса
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Создание слайса для хранения результатов
	puppies := make([]domain.Puppy, 0)

	// Итерация по рядам
	for rows.Next() {
		// Создание новой структуры Puppy для хранения каждого ряда
		var puppy domain.Puppy

		// Сканирование значений из текущего ряда в поля структуры
		err := rows.Scan(
			&puppy.ID, &puppy.Name, &puppy.Title, &puppy.Sex, &puppy.Price, &puppy.ReadyOut,
			&puppy.Archived, &puppy.City, &puppy.MotherID, &puppy.FatherID, &puppy.DateBirth, &puppy.Color,
			pq.Array(&puppy.Urls),
		)
		if err != nil {
			return nil, err
		}

		// Добавление структуры к слайсу
		puppies = append(puppies, puppy)
	}

	// Проверка на ошибки во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Возвращение слайса структур
	return puppies, nil
}

// GetReviews получает отзывы и id щенка в базе данных
func (r *PostgresRepo) GetReviews(ctx context.Context, puppies []domain.Puppy) (map[int]string, error) {
	ids := make([]string, len(puppies))
	for i, puppy := range puppies {
		ids[i] = strconv.Itoa(puppy.ID)
	}
	query := "SELECT id, puppy_id FROM reviews WHERE puppy_id IN (" + strings.Join(ids, ",") + ")"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make(map[int]string)
	for rows.Next() {
		var reviewID int
		var puppyID int
		err := rows.Scan(&reviewID, &puppyID)
		if err != nil {
			return nil, err
		}

		reviews[puppyID] = strconv.Itoa(reviewID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// PuppyGet получает информацию о щенке в базе данных
func (r *PostgresRepo) PuppyGet(idPuppy string) (*domain.Puppy, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT p.*, array_agg(i.url) as urls FROM puppies p"
	query += " LEFT JOIN puppies_img pi ON p.id = pi.puppy_id"
	query += " LEFT JOIN img_urls i ON pi.img_url_id = i.id"
	query += " WHERE 1=1"

	if idPuppy != "" {
		query += " AND p.id = '" + idPuppy + "'"
	}

	query += " GROUP BY p.id" // Группируем результаты по ID щенка

	log.Println(query)

	puppy := &domain.Puppy{}

	row := r.pool.QueryRow(context.Background(), query)

	err := row.Scan(
		&puppy.ID, &puppy.Name, &puppy.Title, &puppy.Sex, &puppy.Price, &puppy.ReadyOut,
		&puppy.Archived, &puppy.City, &puppy.MotherID, &puppy.FatherID, &puppy.DateBirth, &puppy.Color,
		pq.Array(&puppy.Urls),
	)
	if err != nil {
		return nil, err
	}

	log.Println(puppy.ID, puppy.Name, puppy.MotherID, puppy.FatherID)

	// Возвращение слайса структур и количество страниц
	return puppy, nil
}

// DogGet получает информацию о собаке в базе данных
func (r *PostgresRepo) DogGet(idDog string) (*domain.Dog, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT d.*, array_agg(i.url) as urls FROM adult_dogs d"
	query += " LEFT JOIN adult_dogs_img di ON d.id = di.adult_dogs_id"
	query += " LEFT JOIN img_urls i ON di.img_url_id = i.id"
	query += " WHERE 1=1"

	if idDog != "" {
		query += " AND d.id = '" + idDog + "'"
	}

	query += " GROUP BY d.id" // Группируем результаты по ID щенка

	log.Println(query)

	dog := &domain.Dog{}

	row := r.pool.QueryRow(context.Background(), query)

	err := row.Scan(
		&dog.ID, &dog.Name, &dog.Title, &dog.Gender, &dog.Color, &dog.Archived,
		pq.Array(&dog.Urls),
	)

	log.Println(dog.ID, dog.Name)

	if err != nil {
		return nil, err
	}

	return dog, nil
}

// DogsGet получает список собак в базе данных
func (r *PostgresRepo) DogsGet(chocolates, genders []string, idDog string, archived bool) ([]domain.Dog, error) {
	// Подготовка SQL-запроса с условиями
	query := "SELECT d.*, array_agg(i.url) as urls FROM adult_dogs d"
	query += " LEFT JOIN adult_dogs_img di ON d.id = di.adult_dogs_id"
	query += " LEFT JOIN img_urls i ON di.img_url_id = i.id"
	query += " WHERE 1=1"

	// Добавление условий в запрос
	if len(chocolates) > 0 {
		query += " AND d.color IN ("
		for _, c := range chocolates {
			query += "'" + c + "',"
		}
		query = strings.TrimSuffix(query, ",") + ")"
	}

	if len(genders) > 0 {
		query += " AND d.gender IN ("
		for _, g := range genders {
			query += "'" + g + "',"
		}
		query = strings.TrimSuffix(query, ",") + ")"
	}

	if archived {
		query += " AND d.archived = 'true'"
	} else {
		query += " AND d.archived = 'false'"
	}

	if idDog != "" {
		query += " AND d.id = '" + idDog + "'"
	}

	query += " GROUP BY d.id" // Группируем результаты по ID щенка

	log.Println(query)

	// Выполнение SQL-запроса
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Создание слайса для хранения результатов
	dogs := make([]domain.Dog, 0)

	// Итерация по рядам
	for rows.Next() {
		// Создание новой структуры Puppy для хранения каждого ряда
		var dog domain.Dog

		// Сканирование значений из текущего ряда в поля структуры
		err := rows.Scan(
			&dog.ID, &dog.Name, &dog.Title, &dog.Gender, &dog.Color, &dog.Archived,
			pq.Array(&dog.Urls),
		)
		if err != nil {
			return nil, err
		}

		// Добавление структуры к слайсу
		dogs = append(dogs, dog)
	}

	// Проверка на ошибки во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dogs, nil
}

// GetByUsername возвращает пользователя из базы данных по его логину.
func (r *PostgresRepo) GetByUsername(username string) (*domain.User, error) {
	row := r.pool.QueryRow(context.Background(), "SELECT id, password FROM users WHERE username = $1", username)
	user := &domain.User{}
	err := row.Scan(&user.Login, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CheckValidUser проверяет валидность пользователя.
func (r *PostgresRepo) CheckValidUser(login string) (int, string, error) {
	var password string
	var id int
	query := "SELECT id, password FROM users WHERE login = $1"
	err := r.pool.QueryRow(context.Background(), query, login).Scan(&id, &password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, "", err
		}
		r.logger.Error("Failed to check user validity", zap.Error(err))
		return 0, "", err
	}
	return id, password, nil
}

// CheckUniqUser проверяет уникальность логина пользователя.
func (r *PostgresRepo) CheckUniqUser(login string) (bool, error) {
	var existingUser string
	query := "SELECT login FROM users WHERE login = $1"
	err := r.pool.QueryRow(context.Background(), query, login).Scan(&existingUser)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error("Failed to check unique user", zap.Error(err))
		return false, err
	}
	return true, nil
}

// SaveUser сохраняет пользователя в базе данных
func (r *PostgresRepo) SaveUser(login string, hash string) error {
	// Подготовка SQL-запроса
	query := `INSERT INTO users (login, password) VALUES ($1, $2)`
	println(login)
	println(hash)

	// Выполнение запроса
	_, err := r.pool.Exec(context.Background(), query, login, hash)
	if err != nil {
		return err
	}

	return nil
}
