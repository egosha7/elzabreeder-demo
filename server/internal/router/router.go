package routes

import (
	"github.com/egosha7/site-go/internal/authMiddleware"
	"github.com/egosha7/site-go/internal/compress"
	"github.com/egosha7/site-go/internal/handlers"
	"github.com/go-chi/chi/middleware"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// SetupRoutes настраивает и возвращает обработчик HTTP-маршрутов.
func SetupRoutes(h *handlers.Handler, logger *zap.Logger) *chi.Mux {
	// Middleware для сжатия ответа
	gzipMiddleware := compress.GzipMiddleware{}

	// Middleware для определения гео
	GeoMiddleware := authMiddleware.NewMiddlewareRobotTaskDenied(logger)

	// Создание middleware для аутентификации
	JWTMiddleware := authMiddleware.NewMiddlewareJWT(logger, authMiddleware.SigningKey)

	// Создание роутера
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Добавляем обработчик статических файлов
	staticFileServer := http.FileServer(http.Dir("cmd/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", staticFileServer))

	// Группа роутов
	r.Group(
		func(route chi.Router) {
			route.Use(gzipMiddleware.Apply)
			route.Use(GeoMiddleware.RequestToIP)
			// Авторизованные маршруты
			route.Route(
				"/admin", func(r chi.Router) {

					// middleware авторизации
					r.Use(JWTMiddleware.JWTAuth)

					r.Get(
						"/puppies", func(w http.ResponseWriter, r *http.Request) {
							archived := false
							h.AdminPuppiesHandler(w, r, archived)
						},
					)

					r.Get(
						"/archive", func(w http.ResponseWriter, r *http.Request) {
							archived := true
							h.AdminPuppiesHandler(w, r, archived)
						},
					)
					r.Get(
						"/dogs", func(w http.ResponseWriter, r *http.Request) {
							archived := false
							h.AdminDogsHandler(w, r, archived)
						},
					)
					r.Get(
						"/archive/dogs", func(w http.ResponseWriter, r *http.Request) {
							archived := true
							h.AdminDogsHandler(w, r, archived)
						},
					)
					r.Post(
						"/puppies/update", func(w http.ResponseWriter, r *http.Request) {
							h.UpdatePuppy(w, r)
						},
					)
					r.Post(
						"/puppies/delete", func(w http.ResponseWriter, r *http.Request) {
							h.DeletePuppy(w, r)
						},
					)
					r.Post(
						"/puppies/add", func(w http.ResponseWriter, r *http.Request) {
							h.AddPuppy(w, r)
						},
					)
					r.Post(
						"/puppies/archived", func(w http.ResponseWriter, r *http.Request) {
							h.ChangeArchivedPuppy(w, r)
						},
					)
					r.Post(
						"/dogs/update", func(w http.ResponseWriter, r *http.Request) {
							h.UpdateDog(w, r)
						},
					)
					r.Post(
						"/dogs/add", func(w http.ResponseWriter, r *http.Request) {
							h.AddDog(w, r)
						},
					)
					r.Post(
						"/dogs/archived", func(w http.ResponseWriter, r *http.Request) {
							h.ChangeArchivedDog(w, r)
						},
					)
					r.Get(
						"/reviews", func(w http.ResponseWriter, r *http.Request) {
							checked := true
							h.AdminReviewsView(w, r, checked)
						},
					)
					r.Get(
						"/reviews/archive", func(w http.ResponseWriter, r *http.Request) {
							checked := false
							h.AdminReviewsView(w, r, checked)
						},
					)
					r.Post(
						"/reviews/update", func(w http.ResponseWriter, r *http.Request) {
							h.UpdateFeedback(w, r)
						},
					)
					r.Post(
						"/reviews/delete", func(w http.ResponseWriter, r *http.Request) {
							h.DeleteFeedback(w, r)
						},
					)
					r.Post(
						"/reviews/checked", func(w http.ResponseWriter, r *http.Request) {
							h.ChangeCheckedFeedback(w, r)
						},
					)
					r.Post(
						"/user/add", func(w http.ResponseWriter, r *http.Request) {
							h.AdminNewUser(w, r)
						},
					)
				},
			)

			// Регистрация обработчиков для различных маршрутов
			route.Get(
				"/", func(w http.ResponseWriter, r *http.Request) {
					h.MainView(w, r)
				},
			)
			route.Get(
				"/auth", func(w http.ResponseWriter, r *http.Request) {
					h.AuthView(w, r)
				},
			)
			// Регистрация обработчиков для различных маршрутов
			route.Post(
				"/auth", func(w http.ResponseWriter, r *http.Request) {
					h.AuthUser(w, r)
				},
			)
			route.Get(
				"/puppies", func(w http.ResponseWriter, r *http.Request) {
					var archived = false
					h.PuppiesView(w, r, archived)
				},
			)
			route.Get(
				"/archive", func(w http.ResponseWriter, r *http.Request) {
					var archived = true
					h.PuppiesView(w, r, archived)
				},
			)
			route.Get(
				"/puppies/{id}", func(w http.ResponseWriter, r *http.Request) {
					h.PuppyView(w, r)
				},
			)
			route.Post(
				"/users/email/add", func(w http.ResponseWriter, r *http.Request) {
					h.AddEmail(w, r)
				},
			)
			route.Get(
				"/contacts", func(w http.ResponseWriter, r *http.Request) {
					h.ContactsView(w, r)
				},
			)
			route.Get(
				"/reviews", func(w http.ResponseWriter, r *http.Request) {
					h.ReviewsView(w, r)
				},
			)
			route.Get(
				"/reviews/new", func(w http.ResponseWriter, r *http.Request) {
					h.NewFeedbackView(w, r)
				},
			)
			route.Post(
				"/reviews/add", func(w http.ResponseWriter, r *http.Request) {
					h.AddFeedback(w, r)
				},
			)
			route.Get(
				"/article/{id}", func(w http.ResponseWriter, r *http.Request) {
					h.ArticleView(w, r)
				},
			)
		},
	)

	// Обработчик для 404
	r.NotFound(
		func(w http.ResponseWriter, r *http.Request) {
			h.NotFoundView(w, r)
		},
	)

	return r
}
