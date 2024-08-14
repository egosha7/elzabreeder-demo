package domain

// User структура для представления пользователя
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
