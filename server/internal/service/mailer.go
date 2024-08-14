package service

// AddEmail добавляет электронную почту в базу данных.
func (s *ServiceImpl) AddEmail(email string) error {
	// Здесь может быть ваша бизнес-логика
	return s.Repository.AddEmail(email)
}
