package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"sync"
)

type Mailer struct {
	SMTPHost     string // Адрес SMTP-сервера
	SMTPPort     int    // Порт SMTP-сервера
	SMTPUsername string // Имя пользователя (обычно адрес электронной почты)
	SMTPPassword string // Пароль для аутентификации на SMTP-сервере
	FromEmail    string // Адрес электронной почты отправителя
}

type EmailTask struct {
	To      string
	Subject string
	Body    string
}

func NewMailer(smtpHost string, smtpPort int, smtpUsername, smtpPassword, fromEmail string) *Mailer {
	return &Mailer{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		FromEmail:    fromEmail,
	}
}

func (m *Mailer) SendMail(to, subject, htmlContent string) error {
	auth := smtp.PlainAuth("", m.SMTPUsername, m.SMTPPassword, m.SMTPHost)

	msg := []byte("From: " + m.FromEmail + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlContent + "\r\n")

	err := smtp.SendMail(fmt.Sprintf("%s:%d", m.SMTPHost, m.SMTPPort), auth, m.FromEmail, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *Mailer) SendTemplatedMail(to, subject, templatePath string, data interface{}) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return err
	}

	return m.SendMail(to, subject, body.String())
}

func (m *Mailer) SendMailWorker(tasks <-chan EmailTask, wg *sync.WaitGroup, results chan<- error) {
	defer wg.Done()
	for task := range tasks {
		err := m.SendMail(task.To, task.Subject, task.Body)
		results <- err
	}
}

func (m *Mailer) SendEmailsParallel(tasks []EmailTask, workerCount int) []error {
	taskChan := make(chan EmailTask, len(tasks))
	resultChan := make(chan error, len(tasks))
	var wg sync.WaitGroup

	// Запускаем рабочих горутин
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go m.SendMailWorker(taskChan, &wg, resultChan)
	}

	// Отправляем задачи в канал
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// Ждем завершения всех рабочих горутин
	wg.Wait()
	close(resultChan)

	// Собираем результаты
	var errs []error
	for err := range resultChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (m *Mailer) RenderTemplate(templatePath string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return "", err
	}

	return body.String(), nil
}

// CheckSMTPConnection проверяет подключение к SMTP-серверу, отправляя тестовое письмо.
func (m *Mailer) CheckSMTPConnection() error {
	testEmail := "eg8989646@gmail.com" // Вы можете использовать другой адрес, если хотите
	subject := "Test Email"
	// HTML-контент email
	htmlContent := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Elza Breeder</title>
    <style>
        body {
            background-color: #000;
            color: #fff;
            font-family: 'Rubik', sans-serif;
            margin: 0;
            padding: 0;
        }
        .container {
background-color: #000;
            width: 100%;
            padding: 20px;
            max-width: 600px;
            margin: 0 auto;
        }
        .navbar {
            background-color: #000;
            padding: 10px;
            text-align: center;
        }
        .navbar img {
            height: 35px;
        }
        .carousel {
            margin-bottom: 20px;
        }
        .carousel img {
            width: 100%;
            border-radius: 8px;
        }
        .footer {
            background-color: #0a0a0a;
            color: #575757;
            padding: 20px;
            text-align: center;
        }
        .footer img {
            height: 40px;
            margin-bottom: 10px;
        }
        .badge {
            background-color: #c2c2c2;
            color: #000;
            padding: 2px 6px;
            border-radius: 4px;
        }
        h2 {
            color: #fff;
        }
        p {
            color: #aaa;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="navbar">
            <a href="/">
                <img src="https://elzabreeder-space.s3.cloud.ru/logo.png" alt="Логотип">
            </a>
        </div>
        <h2>Новый щенок в продаже!</h2>
        <div class="carousel">
            <img src="https://elzabreeder-space.s3.cloud.ru/Cesar_1719304103_scale_1200.png" alt="Puppy">
        </div>
        <h2>Cesar <span class="badge">Мальчик</span></h2>
        <p>Умная!1</p>
        <p class="h5">Цена: 35000 руб.</p>
        <div class="footer">
            <img src="https://elzabreeder-space.s3.cloud.ru/logo.png" alt="Логотип">
            <p>&copy; 2024 Elza Breeder</p>
        </div>
    </div>
</body>
</html>`

	return m.SendMail(testEmail, subject, htmlContent)
}
