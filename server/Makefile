.PHONY: test cover open-cover clean

# Устанавливаем переменные
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

# Цель для запуска тестов и генерации профиля покрытия
test:
	@echo "Running tests..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...

# Цель для генерации HTML-отчета покрытия
cover: test
	@echo "Generating coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Цель для открытия HTML-отчета покрытия в браузере
open-cover: cover
	@echo "Opening coverage report..."
	@start "" $(COVERAGE_HTML)

# Цель для удаления файлов покрытия
clean: open-cover
	@echo "Cleaning up..."
	@timeout /t 3 /nobreak
	@del $(COVERAGE_FILE) $(COVERAGE_HTML)
