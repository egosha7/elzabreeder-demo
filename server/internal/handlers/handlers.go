package handlers

import (
	"github.com/egosha7/site-go/internal/domain"
	"github.com/egosha7/site-go/internal/service"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"time"
)

// Handler представляет обработчик HTTP-запросов.
type Handler struct {
	Services        *service.Service
	logger          *zap.Logger
	ParseTemplate   func(filenames ...string) (*template.Template, error)
	ExecuteTemplate func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(services *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		Services:      services,
		logger:        logger,
		ParseTemplate: template.ParseFiles,
		ExecuteTemplate: func(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
			return t.ExecuteTemplate(w, name, data)
		},
	}
}

func (h *Handler) AddFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB max size
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")

	phone := r.FormValue("phone")

	title := r.FormValue("title")

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	h.logger.Info(
		"Puppy updated successfully",
		zap.String("Name", name),
		zap.String("Price", phone),
		zap.String("Title", title),
	)

	now := time.Now()

	// Create Puppy struct
	feedback := &domain.Feedback{
		PuppyID:  0,
		Name:     name,
		Number:   phone,
		Title:    title,
		Verified: false,
		Date:     now.Format("02.01.2006"),
		Urls:     []string{}, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.FeedbackAdd(feedback, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to update puppy", zap.Error(err))
		http.Error(w, "Failed to update puppy", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppy", http.StatusSeeOther)
}
