package todo

import (
	"encoding/json"
	"errors"
	"go-do-the-thing/app/handlers"
	"go-do-the-thing/app/models"
	"go-do-the-thing/app/repos"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type Handler struct {
	repo          repos.TasksRepo
	templates     helpers.Templates
	activeScreens models.NavBarObject
	logger        slog.Logger
}

func SetupTodoHandler(
	tasksRepo repos.TasksRepo,
	router *http.ServeMux,
	templates helpers.Templates,
	mw_stack middleware.Middleware,
) error {
	logger := slog.NewLogger("Tasks")

	todoHandler := &Handler{
		repo:          tasksRepo,
		templates:     templates,
		activeScreens: models.NavBarObject{ActiveScreens: models.ActiveScreens{IsTodoList: true}},
		logger:        *logger,
	}

	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.GetItem)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.ListItems)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.UpdateItemStatus)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.CreateItem)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.UpdateItem)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.DeleteItem)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(todoHandler.TestError)))
	//	router.HandleFunc("POST /todo/restore/{id}", todoHandler.RestoreItemUI)
	return nil
}

type idResponse struct {
	Id int64 `json:"id" json:"id"`
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.createItemAPI, h.createItemUI)
}

func (h *Handler) createItemAPI(w http.ResponseWriter, r *http.Request) {
	var item models.Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&item)
	if err != nil {
		h.logger.Error(err, "failed to decode todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.repo.InsertItem(item)
	if err != nil {
		h.logger.Error(err, "failed to insert item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	idResponse := idResponse{
		Id: id,
	}
	jsonBytes, err := json.Marshal(idResponse)
	if err != nil {
		h.logger.Error(err, "failed to marshal id response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		h.logger.Error(err, "failed to write response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createItemUI(w http.ResponseWriter, r *http.Request) {
	errorForm := models.NewFormData()
	name, errorForm := models.GetRequiredPropertyFromRequest(r, "name", errorForm, true)
	description, errorForm := models.GetOptionalPropertyFromRequest(r, "description", errorForm, true)
	assignedTo, errorForm := models.GetRequiredPropertyFromRequest(r, "assigned_to", errorForm, true)
	tag, errorForm := models.GetRequiredPropertyFromRequest(r, "tag", errorForm, true)
	dateRaw, errorForm := models.GetRequiredPropertyFromRequest(r, "due_date", errorForm, true)
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil || len(errorForm.Errors) > 0 {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "task-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
		}
		return
	}
	now := time.Now()
	item := models.Task{
		Name:        name,
		Description: description,
		AssignedTo:  assignedTo,
		DueDate:     &database.SqLiteTime{Time: &date},
		CreatedBy:   "CurrentUser", //todo logins
		CreateDate:  &database.SqLiteTime{Time: &now},
		IsDeleted:   false,
		Tag:         tag,
	}
	//Check if form is valid and respond with any error
	formData, isValid := item.FormDataFromItem()
	if !isValid {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "task-form-content", formData); err != nil {
			handlers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
		}
		return
	}

	//update data
	id, err := h.repo.InsertItem(item)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to insert todo item", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to get todo item", err, w)
		return
	}
	//Respond with templates
	if err = h.templates.RenderOk(w, "task-row-oob", task); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render item row", err, w)
		return
	}
	to := NoItemRowData{
		HideNoData: true,
	}
	if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render item row", err, w)
		return
	}
	err = h.templates.RenderOk(w, "task-form-content", models.NewFormData())
	if err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

type NoItemRowData struct {
	HideNoData bool
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.updateItemAPI, h.updateItemUI)
}

func (h *Handler) updateItemAPI(w http.ResponseWriter, r *http.Request) {
	var item models.Task
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&item); err != nil {
		handlers.HttpError("failed to decode item", err, w)
		return
	}
	if id != item.Id {
		handlers.HttpError("id mismatch", errors.New("The id in the path does not match the id in the request object"), w)
		return
	}

	if err = h.repo.UpdateItem(item); err != nil {
		handlers.HttpError("failed to update todo item", err, w)
		return
	}
}

func (h *Handler) updateItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	errorForm := models.NewFormData()
	name, errorForm := models.GetRequiredPropertyFromRequest(r, "name", errorForm, true)
	description, errorForm := models.GetOptionalPropertyFromRequest(r, "description", errorForm, true)
	assignedTo, errorForm := models.GetRequiredPropertyFromRequest(r, "assigned_to", errorForm, true)
	tag, errorForm := models.GetRequiredPropertyFromRequest(r, "tag", errorForm, true)
	dateRaw, errorForm := models.GetRequiredPropertyFromRequest(r, "due_date", errorForm, true)
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil || len(errorForm.Errors) > 0 {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "task-form-content", errorForm); err != nil {
			handlers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
		}
		return
	}
	now := time.Now()
	item := models.Task{
		Id:          id,
		Name:        name,
		Description: description,
		AssignedTo:  assignedTo,
		DueDate:     &database.SqLiteTime{Time: &date},
		CreatedBy:   "CurrentUser",
		CreateDate:  &database.SqLiteTime{Time: &now},
		IsDeleted:   false,
		Tag:         tag,
	}

	//Check if form is valid and respond with any error
	formData := item.FormDataFromItemNoValidation()
	formData.Submit = "Update"

	//update data
	if err = h.repo.UpdateItem(item); err != nil {
		handlers.HttpErrorUI(h.templates, "failed to insert update item", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to get todo item", err, w)
		return
	}
	//Respond with templates
	model := ItemPageModel{task, h.activeScreens, formData}
	if err = h.templates.RenderOk(w, "task-item-content-oob", model); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render item row", err, w)
		return
	}
	err = h.templates.RenderOk(w, "task-form-content", formData)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render form", err, w)
		return
	}
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.getItemAPI, h.getItemUI)
}

func (h *Handler) getItemAPI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := h.repo.GetItem(id)
	if err != nil {
		h.logger.Error(err, "failed to get todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	items := make([]models.Task, 1)
	items[0] = item
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		h.logger.Error(err, "failed to marshal todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		h.logger.Error(err, "failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type ItemPageModel struct {
	Task     models.Task
	NavBar   models.NavBarObject
	FormData models.FormData
}

func (h *Handler) getItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to parse id from path", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	formData := task.FormDataFromItemNoValidation()
	formData.Submit = "Update"
	model := ItemPageModel{task, h.activeScreens, formData}
	err = h.templates.RenderOk(w, "task-item", model)
	if err != nil {
		h.logger.Error(err, "Failed to execute template for the item page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.updateItemStatusAPI, h.updateItemStatusUI)
}

func (h *Handler) updateItemStatusAPI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newStatus, err := strconv.ParseInt(r.FormValue("status"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	completeDate := database.SqLiteTime{}
	if newStatus == int64(models.Completed) {
		now := time.Now()
		completeDate = database.SqLiteTime{Time: &now}
	}
	err = h.repo.UpdateItemStatus(id, completeDate, newStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) updateItemStatusUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to parse id from path", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to get todo item", err, w)
		return
	}

	task.ToggleStatus()

	if err = h.repo.UpdateItemStatus(id, *task.CompleteDate, int64(task.Status)); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to update todo item", err, w)
		return
	}
	task, err = h.repo.GetItem(id)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to get todo item", err, w)
		return
	}

	if err = h.templates.RenderOk(w, "task-row", task); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to execute template for the home page", err, w)
		return
	}
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.listItemsAPI, h.listItemsUI)
}

func (h *Handler) listItemsAPI(w http.ResponseWriter, _ *http.Request) {
	items, err := h.repo.GetItems()
	if err != nil {
		h.logger.Error(err, "failed to get todo items")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		h.logger.Error(err, "failed to marshal todo items")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		h.logger.Error(err, "failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type ListModel struct {
	Tasks    []models.Task
	NavBar   models.NavBarObject
	FormData models.FormData
}

func (h *Handler) listItemsUI(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repo.GetItems()
	if err != nil {
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Time.Before(*tasks[j].DueDate.Time)
	})

	formData := models.NewFormData()
	formData.Values["due_date"] = time.Now().Add(time.Hour * 24).Format("2006-01-02")

	data := ListModel{tasks, h.activeScreens, formData}
	email, name, err := helpers.GetUserFromContext(r)
	if err != nil {
		h.logger.Error(err, "could not get user details from http context")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.NavBar = data.NavBar.SetUser(name, email)

	err = h.templates.RenderOk(w, "task-list", data)
	if err != nil {
		h.logger.Error(err, "Failed to execute template for the item list page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.deleteItemAPI, h.deleteItemUI)
}

func (h *Handler) deleteItemAPI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.repo.DeleteItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) deleteItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to parse id from path", err, w)
		return
	}
	err = h.repo.DeleteItem(id)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to delete todo item", err, w)
		return
	}
	//get count of items
	hasData, err := h.repo.GetItemsCount()
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to update ui", err, w)
		return
	}
	to := NoItemRowData{
		HideNoData: hasData > 0,
	}
	if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		handlers.HttpErrorUI(h.templates, "Failed to render item row", err, w)
		return
	}
}

func (h *Handler) TestError(w http.ResponseWriter, _ *http.Request) {
	handlers.HttpErrorUI(h.templates, "Testing the error page", errors.New("Testing the error page"), w)
}
