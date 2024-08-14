package handlers

import (
	"github.com/egosha7/site-go/internal/domain"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *Handler) AdminNewUser(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	println(login)
	println(password)
	err = h.Services.AuthorizationServices.CreateUser(login, password)
	if err != nil {
		h.logger.Error("Invalid request add user", zap.Error(err))
		http.Error(w, "Invalid request add user", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateDog(w http.ResponseWriter, r *http.Request) {
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

	sex := r.FormValue("gender")

	color := r.FormValue("color")

	title := r.FormValue("title")

	dogID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		h.logger.Error("Invalid dog ID", zap.Error(err))
		http.Error(w, "Invalid dog ID", http.StatusBadRequest)
		return
	}

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	existingPhotos := r.MultipartForm.Value["existingPhotos"]

	h.logger.Info(
		"Puppy update",
		zap.Int("dogID", dogID),
		zap.String("Name", name),
		zap.String("Gender", sex),
		zap.String("Color", color),
		zap.String("Title", title),
	)

	// Create Puppy struct
	dog := &domain.Dog{
		ID:       dogID,
		Name:     name,
		Title:    title,
		Gender:   sex,
		Color:    color,
		Archived: false,          // Assuming this is not being set from form
		Urls:     existingPhotos, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.DogUpdate(dog, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to update dog", zap.Error(err))
		http.Error(w, "Failed to update dog", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dogs", http.StatusSeeOther)
}

func (h *Handler) ChangeArchivedDog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	dogID := r.FormValue("id")

	archived := r.FormValue("archived")

	h.logger.Info(
		"Dog change archive condition",
		zap.String("dogID", dogID),
		zap.String("Archived", archived),
	)

	err = h.Services.DogChangeArchived(dogID, archived)
	if err != nil {
		h.logger.Error("Failed to change archive condition dog", zap.Error(err))
		http.Error(w, "Failed to change archive condition dog", http.StatusInternalServerError)
		return
	}

	if archived == "true" {
		http.Redirect(w, r, "/admin/dogs", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/archive/dogs", http.StatusSeeOther)
	}

}

func (h *Handler) AddDog(w http.ResponseWriter, r *http.Request) {
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

	sex := r.FormValue("gender")

	color := r.FormValue("color")

	title := r.FormValue("title")

	dogID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		h.logger.Error("Invalid dog ID", zap.Error(err))
		http.Error(w, "Invalid dog ID", http.StatusBadRequest)
		return
	}

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	h.logger.Info(
		"Dog add",
		zap.Int("dogID", dogID),
		zap.String("Name", name),
		zap.String("Gender", sex),
		zap.String("Color", color),
		zap.String("Title", title),
	)

	// Create Puppy struct
	dog := &domain.Dog{
		ID:       dogID,
		Name:     name,
		Title:    title,
		Gender:   sex,
		Color:    color,
		Archived: false,      // Assuming this is not being set from form
		Urls:     []string{}, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.DogAdd(dog, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to add dog", zap.Error(err))
		http.Error(w, "Failed to add dog", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/dogs", http.StatusSeeOther)
}

func (h *Handler) ChangeArchivedPuppy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	puppyID := r.FormValue("id")

	archived := r.FormValue("archived")

	city := r.FormValue("city")

	phone := r.FormValue("phone")

	h.logger.Info(
		"Puppy change archive condition",
		zap.String("puppyID", puppyID),
		zap.String("Archived", archived),
		zap.String("City", city),
		zap.String("Phone", phone),
	)

	err = h.Services.PuppyChangeArchived(puppyID, archived, city, phone)
	if err != nil {
		h.logger.Error("Failed to change archive condition for puppy", zap.Error(err))
		http.Error(w, "Failed to change archive condition for puppy", http.StatusInternalServerError)
		return
	}

	if archived == "true" {
		http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/archive", http.StatusSeeOther)
	}
}

func (h *Handler) DeletePuppy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	puppyID := r.FormValue("id")

	h.logger.Info(
		"Puppy delete",
		zap.String("puppyID", puppyID),
	)
	err = h.Services.PuppyDelete(puppyID)
	if err != nil {
		h.logger.Error("Failed to delete puppy", zap.Error(err))
		http.Error(w, "Failed to delete puppy", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)

}

func (h *Handler) UpdatePuppy(w http.ResponseWriter, r *http.Request) {
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

	price := r.FormValue("price")

	sex := r.FormValue("gender")

	fatherID, err := strconv.Atoi(r.FormValue("father"))
	if err != nil {
		h.logger.Error("Invalid father ID", zap.Error(err))
		http.Error(w, "Invalid father ID", http.StatusBadRequest)
		return
	}

	motherID, err := strconv.Atoi(r.FormValue("mother"))
	if err != nil {
		h.logger.Error("Invalid mother ID", zap.Error(err))
		http.Error(w, "Invalid mother ID", http.StatusBadRequest)
		return
	}

	color := r.FormValue("color")

	dateBirth := r.FormValue("date")

	readyOut := r.FormValue("readyToMoveAdd") == "true"

	title := r.FormValue("title")

	puppyID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		h.logger.Error("Invalid puppy ID", zap.Error(err))
		http.Error(w, "Invalid puppy ID", http.StatusBadRequest)
		return
	}

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	existingPhotos := r.MultipartForm.Value["existingPhotos"]

	h.logger.Info(
		"Puppy update",
		zap.Int("puppyID", puppyID),
		zap.String("Name", name),
		zap.String("Price", price),
		zap.String("Gender", sex),
		zap.Int("FatherID", fatherID),
		zap.Int("MotherID", motherID),
		zap.String("Color", color),
		zap.String("DateBirth", dateBirth),
		zap.Bool("ReadyOut", readyOut),
		zap.String("Title", title),
	)

	// Create Puppy struct
	puppy := &domain.Puppy{
		ID:        puppyID,
		Name:      name,
		Title:     title,
		Sex:       sex,
		Price:     price,
		ReadyOut:  readyOut,
		Archived:  false, // Assuming this is not being set from form
		City:      "Уфа", // Assuming this is not being set from form
		MotherID:  motherID,
		FatherID:  fatherID,
		DateBirth: dateBirth,
		Color:     color,
		Urls:      existingPhotos, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.PuppyUpdate(puppy, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to update puppy", zap.Error(err))
		http.Error(w, "Failed to update puppy", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
}

func (h *Handler) AddPuppy(w http.ResponseWriter, r *http.Request) {
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

	price := r.FormValue("price")

	sex := r.FormValue("gender")

	fatherID, err := strconv.Atoi(r.FormValue("father"))
	if err != nil {
		h.logger.Error("Invalid father ID", zap.Error(err))
		http.Error(w, "Invalid father ID", http.StatusBadRequest)
		return
	}

	motherID, err := strconv.Atoi(r.FormValue("mother"))
	if err != nil {
		h.logger.Error("Invalid mother ID", zap.Error(err))
		http.Error(w, "Invalid mother ID", http.StatusBadRequest)
		return
	}

	color := r.FormValue("color")

	dateBirth := r.FormValue("date")

	readyOut := r.FormValue("readyToMoveAdd") == "true"

	title := r.FormValue("title")

	puppyID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		h.logger.Error("Invalid puppy ID", zap.Error(err))
		http.Error(w, "Invalid puppy ID", http.StatusBadRequest)
		return
	}

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	h.logger.Info(
		"Puppy add",
		zap.Int("puppyID", puppyID),
		zap.String("Name", name),
		zap.String("Price", price),
		zap.String("Gender", sex),
		zap.Int("FatherID", fatherID),
		zap.Int("MotherID", motherID),
		zap.String("Color", color),
		zap.String("DateBirth", dateBirth),
		zap.Bool("ReadyOut", readyOut),
		zap.String("Title", title),
	)

	// Create Puppy struct
	puppy := &domain.Puppy{
		ID:        puppyID,
		Name:      name,
		Title:     title,
		Sex:       sex,
		Price:     price,
		ReadyOut:  readyOut,
		Archived:  false, // Assuming this is not being set from form
		City:      "Уфа", // Assuming this is not being set from form
		MotherID:  motherID,
		FatherID:  fatherID,
		DateBirth: dateBirth,
		Color:     color,
		Urls:      []string{}, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.PuppyAdd(puppy, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to add puppy", zap.Error(err))
		http.Error(w, "Failed to add puppy", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
}

func (h *Handler) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
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

	title := r.FormValue("title")

	phone := r.FormValue("phone")

	date := r.FormValue("date")

	puppyID, err := strconv.Atoi(r.FormValue("puppyID"))
	if err != nil {
		h.logger.Error("Invalid puppyID", zap.Error(err))
		http.Error(w, "Invalid puppyID", http.StatusBadRequest)
		return
	}

	feedbackID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		h.logger.Error("Invalid feedbackID", zap.Error(err))
		http.Error(w, "Invalid feedbackID", http.StatusBadRequest)
		return
	}

	// Upload files to S3
	fileHeaders := r.MultipartForm.File["files"]

	existingPhotos := r.MultipartForm.Value["existingPhotos"]

	h.logger.Info(
		"Feedback update",
		zap.Int("feedbackID", feedbackID),
		zap.Int("PuppyID", puppyID),
		zap.String("Name", name),
		zap.String("Number", phone),
		zap.String("Title", title),
		zap.String("Date", date),
	)

	// Create Puppy struct
	feedback := &domain.Feedback{
		ID:       feedbackID,
		PuppyID:  puppyID,
		Name:     name,
		Number:   phone,
		Title:    title,
		Verified: false, // Assuming this is not being set from form
		Date:     date,
		Urls:     existingPhotos, // Placeholder URLs
	}

	//	Call the service layer to update the puppy
	err = h.Services.FeedbackUpdate(feedback, fileHeaders)
	if err != nil {
		h.logger.Error("Failed to update feedback", zap.Error(err))
		http.Error(w, "Failed to update feedback", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/reviews", http.StatusSeeOther)
}

func (h *Handler) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	feedbackID := r.FormValue("id")

	h.logger.Info(
		"Delete feedback",
		zap.String("id", feedbackID),
	)
	err = h.Services.FeedbackDelete(feedbackID)
	if err != nil {
		h.logger.Error("Failed to switch puppy", zap.Error(err))
		http.Error(w, "Failed to switch puppy", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/reviews", http.StatusSeeOther)
}

func (h *Handler) ChangeCheckedFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	feedbackID := r.FormValue("id")

	checked := r.FormValue("checked")

	h.logger.Info(
		"Change checked condition for feedback",
		zap.String("dogID", feedbackID),
		zap.String("checked", checked),
	)

	err = h.Services.FeedbackChangeChecked(feedbackID, checked)
	if err != nil {
		h.logger.Error("Failed to switch dog", zap.Error(err))
		http.Error(w, "Failed to switch dog", http.StatusInternalServerError)
		return
	}

	if checked == "true" {
		http.Redirect(w, r, "/admin/reviews", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/reviews/archive", http.StatusSeeOther)
	}

}
