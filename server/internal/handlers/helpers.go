package handlers

import (
	"github.com/egosha7/site-go/internal/domain"
	"log"
)

func contains(chocolates []string, genders []string, readyToMove string) string {
	var getParams string
	if len(chocolates) > 0 {
		for _, v := range chocolates {
			getParams += "&chocolate=" + v
		}
	}

	if len(genders) > 0 {
		for _, v := range genders {
			getParams += "&gender=" + v
		}
	}

	if readyToMove != "" {
		getParams += "&readyToMove=" + readyToMove
	}

	return getParams
}

func getPagedReviews(reviews []domain.Feedback, currentPage, perPage int) ([]domain.Feedback, int, error) {
	// Общее количество щенков
	reviewsCount := len(reviews)

	// Рассчитываем общее количество страниц
	totalPages := reviewsCount / perPage
	if reviewsCount%perPage != 0 {
		totalPages++
	}

	// Проверка текущей страницы на выход за пределы
	if currentPage > totalPages {
		currentPage = totalPages
	}
	if currentPage < 1 {
		currentPage = 1
	}

	// Вычисляем начальный и конечный индекс для текущей страницы
	startIndex := (currentPage - 1) * perPage
	endIndex := startIndex + perPage

	// Проверка на выход за границы
	if startIndex > reviewsCount {
		startIndex = reviewsCount
	}
	if endIndex > reviewsCount {
		endIndex = reviewsCount
	}

	log.Println(startIndex, "->", endIndex)

	// Получаем щенков для текущей страницы
	pagedReviews := reviews[startIndex:endIndex]

	return pagedReviews, totalPages, nil
}

func getPagedPuppies(puppies []domain.Puppy, currentPage, perPage int) ([]domain.Puppy, int, error) {
	// Общее количество щенков
	puppiesCount := len(puppies)

	// Рассчитываем общее количество страниц
	totalPages := puppiesCount / perPage
	if puppiesCount%perPage != 0 {
		totalPages++
	}

	// Проверка текущей страницы на выход за пределы
	if currentPage > totalPages {
		currentPage = totalPages
	}
	if currentPage < 1 {
		currentPage = 1
	}

	// Вычисляем начальный и конечный индекс для текущей страницы
	startIndex := (currentPage - 1) * perPage
	endIndex := startIndex + perPage

	// Проверка на выход за границы
	if startIndex > puppiesCount {
		startIndex = puppiesCount
	}
	if endIndex > puppiesCount {
		endIndex = puppiesCount
	}

	log.Println(startIndex, "->", endIndex)

	// Получаем щенков для текущей страницы
	pagedPuppies := puppies[startIndex:endIndex]

	return pagedPuppies, totalPages, nil
}

func getPagedDogs(dogs []domain.Dog, currentPage, perPage int) ([]domain.Dog, int, error) {
	// Общее количество щенков
	dogsCount := len(dogs)

	// Рассчитываем общее количество страниц
	totalPages := dogsCount / perPage
	if dogsCount%perPage != 0 {
		totalPages++
	}

	// Проверка текущей страницы на выход за пределы
	if currentPage > totalPages {
		currentPage = totalPages
	}
	if currentPage < 1 {
		currentPage = 1
	}

	// Вычисляем начальный и конечный индекс для текущей страницы
	startIndex := (currentPage - 1) * perPage
	endIndex := startIndex + perPage

	// Проверка на выход за границы
	if startIndex > dogsCount {
		startIndex = dogsCount
	}
	if endIndex > dogsCount {
		endIndex = dogsCount
	}

	log.Println(startIndex, "->", endIndex)

	// Получаем щенков для текущей страницы
	pagedDogs := dogs[startIndex:endIndex]

	return pagedDogs, totalPages, nil
}
