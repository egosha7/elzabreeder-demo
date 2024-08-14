package handlers

import (
	"fmt"
	"html"
	"strconv"
	"unicode"
)

func sanitize(input string) string {
	return html.EscapeString(input)
}

// ValidatePage функция для валидации параметра страницы
func ValidatePage(pageStr string) (int, error) {
	if pageStr == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1, fmt.Errorf("invalid page number")
	}
	return page, nil
}

// ValidateGender функция для валидации параметра gender
func ValidateGender(genders []string) ([]string, error) {
	if len(genders) < 1 {
		validatedGenders := []string{}
		return validatedGenders, nil
	}
	validGenders := map[string]bool{"Кобель": true, "Сука": true}
	validatedGenders := []string{}
	for _, gender := range genders {
		if validGenders[gender] {
			validatedGenders = append(validatedGenders, gender)
		} else {
			return nil, fmt.Errorf("invalid gender: %s", gender)
		}
	}
	return validatedGenders, nil
}

// ValidateReadyToMove функция для валидации параметра readyToMove
func ValidateReadyToMove(readyToMove string) (string, error) {
	if readyToMove == "" {
		return "", nil
	}
	if readyToMove == "true" || readyToMove == "false" {
		return readyToMove, nil
	}
	return "", fmt.Errorf("invalid readyToMove value")
}

// ValidateChocolates функция для валидации параметра chocolates
func ValidateChocolates(chocolates []string) ([]string, error) {
	validChocolates := map[string]bool{
		"Классический":     true,
		"Шоколадный":       true,
		"Черный":           true,
		"Биро":             true,
		"Бивер":            true,
		"Голддаст":         true,
		"Черный мерле":     true,
		"Шоколадный мерле": true,
	}
	validatedChocolates := []string{}
	for _, chocolate := range chocolates {
		if validChocolates[chocolate] {
			validatedChocolates = append(validatedChocolates, chocolate)
		} else {
			return nil, fmt.Errorf("invalid chocolate color: %s", chocolate)
		}
	}
	return validatedChocolates, nil
}

// isValidID проверяет, состоит ли ID только из цифр и не превышает ли он заданное значение.
func isValidID(id string, maxID int) bool {
	if id == "" {
		return false
	}

	for _, ch := range id {
		if !unicode.IsDigit(ch) {
			return false
		}
	}

	num, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	return num <= maxID
}
