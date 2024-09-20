package todo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/handlers"
	templ_todo "go-do-the-thing/src/handlers/todo/templ"

	//templ_todo "go-do-the-thing/src/handlers/todo/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type Handler struct {
	repo          repos.TasksRepo
	usersRepo     repos.UsersRepo
	templates     helpers.Templates
	activeScreens models.NavBarObject
	logger        slog.Logger
}

func SetupTodoHandler(
	tasksRepo repos.TasksRepo,
	usersRepo repos.UsersRepo,
	router *http.ServeMux,
	templates helpers.Templates,
	mw_stack middleware.Middleware,
) error {
	logger := slog.NewLogger("Tasks")

	todoHandler := &Handler{
		repo:          tasksRepo,
		usersRepo:     usersRepo,
		templates:     templates,
		activeScreens: models.NavBarObject{ActiveScreens: models.ActiveScreens{IsTodoList: true}},
		logger:        logger,
	}

	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.getItem)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.listItems)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItemStatus)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.createItem)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItem)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.deleteItem)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(todoHandler.testError)))
	//	router.HandleFunc("POST /todo/restore/{id}", todoHandler.RestoreItemUI)
	return nil
}

type idResponse struct {
	Id int64 `json:"id" json:"id"`
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.createItemAPI, h.createItemUI)
}

func (h *Handler) createItemAPI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	assert.IsTrue(false, h.logger, "this has diverged a lot from the UI implementation")
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	var item models.Task
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&item)
	if err != nil {
		h.logger.Error(err, "failed to decode todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item.CreatedBy = currentUserId
	item.CreatedDate = database.SqLiteNow()
	item.ModifiedBy = currentUserId
	item.ModifiedDate = database.SqLiteNow()

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

var tagOptions = []string{"Project 1", "Project 2", "Personal"}

func (h *Handler) createItemUI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	//Collect data about new task
	form := models.NewTaskForm()
	name, err := models.GetPropertyFromRequest(r, "name", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description, _ := models.GetPropertyFromRequest(r, "description", false)
	tag, err := models.GetPropertyFromRequest(r, "tag", true)
	if err != nil {
		form.Errors["Tag"] = err.Error()
	}
	dateRaw, err := models.GetPropertyFromRequest(r, "due_date", true)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		Tag:         tag,
		DueDate:     &database.SqLiteTime{Time: &date},
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_todo.TaskFormContent("Create", form, tagOptions).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, h.logger, "Failed to render template for formData")
			// TODO: better error handling. Honestly i might as well panic here
		}
		return
	}

	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      &database.SqLiteTime{Time: &date},
		AssignedTo:   currentUserId, // TODO: need to update this
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
		Tag:          tag,
	}

	//Check if form is valid and respond with any error
	form, isValid := formDataFromItem(task, currentUserEmail)
	if !isValid {
		h.logger.Info("invalid data")
		if err := templ_todo.TaskFormContent("Create", form, tagOptions).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, h.logger, "Failed to render template for formData")
			// TODO: better error handling. Honestly i might as well panic here
		}
		return
	}

	//update data
	id, err := h.repo.InsertItem(task)
	if err != nil {
		h.logger.Error(err, "failed to insert task")
		form.Errors["Task"] = "failed to create task"
		if err := templ_todo.TaskFormContent("Create", form, tagOptions).Render(r.Context(), w); err != nil {
			assert.NoError(err, h.logger, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	task, err = h.repo.GetItem(id)
	if err != nil {
		assert.NoError(err, h.logger, "failed to get newly inserted task")
		// TODO: what should happen if the fetch fails after create
		return
	}
	//Respond with templates
	assignedToUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	if ok := h.handleUserIdNotFound(err, task.AssignedTo); !ok {
		assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
		// TODO: what should happen if the fetch fails after create
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen if the fetch fails after create
			return
		}
	}

	taskListItem := models.TaskToViewModel(task, assignedToUser, createdBy)
	if err := templ_todo.TaskRowOOB(taskListItem).Render(r.Context(), w); err != nil {
		assert.NoError(err, h.logger, "failed to render new task row with id %d", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}

	if err := templ_todo.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		//if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		assert.NoError(err, h.logger, "failed to render no data row", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}
	err = h.templates.RenderOk(w, "task-form-content", models.NewFormData())
	if err := templ_todo.TaskFormContent("Create", form, tagOptions).Render(r.Context(), w); err != nil {
		assert.NoError(err, h.logger, "failed to render task form content after create")
		// TODO: what should happen if the fetch fails after create
		return
	}
}

type NoItemRowData struct {
	HideNoData bool
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.updateItemAPI, h.updateItemUI)
}

func (h *Handler) updateItemAPI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	assert.IsTrue(false, h.logger, "this probably doesnt work anymore")
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	var item models.Task
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&item); err != nil {
		handlers.HttpError("failed to decode item", err, w)
		return
	}
	item.ModifiedBy = currentUserId
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
	// Get currentUser details
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	form := models.NewTaskForm()
	name, err := models.GetPropertyFromRequest(r, "name", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description, _ := models.GetPropertyFromRequest(r, "description", false)
	assignedTo, err := models.GetPropertyFromRequest(r, "assigned_to", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	tag, err := models.GetPropertyFromRequest(r, "tag", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	dateRaw, err := models.GetPropertyFromRequest(r, "due_date", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		Tag:         tag,
		DueDate:     &database.SqLiteTime{Time: &date},
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_todo.TaskFormContent("Update", form, tagOptions).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error(err, "failed to render task form after failed validation")
			handlers.HttpErrorUI(h.templates, "Failed to render template for formData", err, w)
			// TODO: better error handling. Honestly i might as well panic here
		}
		return
	}

	// TODO: Get task from the db and only update relevant fields
	assignedToUser, err := h.usersRepo.GetUserByEmail(assignedTo)
	if ok := h.handleUserNotFound(err, assignedTo); !ok {
		assert.NoError(err, h.logger, "how does a task with an assigned to user of %s even exist?", assignedTo)
		// TODO: what should happen if the fetch fails while updating
		return
	}
	item := models.Task{
		Id:           id,
		Name:         name,
		Description:  description,
		AssignedTo:   assignedToUser.Id,
		DueDate:      &database.SqLiteTime{Time: &date},
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
		Tag:          tag,
	}

	//update data
	if err = h.repo.UpdateItem(item); err != nil {
		h.logger.Error(err, "failed to update task")
		form.Errors["Task"] = "failed to update task"
		if err := templ_todo.TaskFormContent("Update", form, tagOptions).Render(r.Context(), w); err != nil {
			assert.NoError(err, h.logger, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		assert.NoError(err, h.logger, "failed to get updated task")
		// TODO: what should happen if the fetch fails after update
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}
	//Respond with templates

	model := models.TaskToViewModel(task, assignedToUser, createdBy)
	if err := templ_todo.TaskItemContentOOB(model).Render(r.Context(), w); err != nil {
		assert.NoError(err, h.logger, "failed to render new task row with id %d", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}

	formData := formDataFromItemNoValidation(item, assignedTo)
	if err := templ_todo.TaskFormContent("Create", formData, tagOptions).Render(r.Context(), w); err != nil {
		assert.NoError(err, h.logger, "failed to render form content after update", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.getItemAPI, h.getItemUI)
}

func (h *Handler) getItemAPI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := h.repo.GetItem(id)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	items := make([]models.Task, 1)
	items[0] = item
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to marshal todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type ItemPageModel struct {
	Task     models.TaskView
	NavBar   models.NavBarObject
	FormData models.FormData
}

func (h *Handler) getItemUI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		handlers.HttpErrorUI(h.templates, "failed to parse id from path", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	assignedUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	assert.NoError(err, h.logger, "this user should exist since they are assigned to a task: %d", task.AssignedTo)

	formData := formDataFromItemNoValidation(task, assignedUser.Email)

	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}
	taskView := models.TaskToViewModel(task, assignedUser, createdBy)
	if err = templ_todo.TaskItem(taskView, h.activeScreens, formData, tagOptions).Render(r.Context(), w); err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "Failed to execute template for the item page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateItemStatus(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.updateItemStatusAPI, h.updateItemStatusUI)
}

func (h *Handler) updateItemStatusAPI(w http.ResponseWriter, r *http.Request) {
	assert.IsTrue(false, h.logger, "this probably doesnt work anymore")
	// Get currentUser details
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newStatus, err := strconv.ParseInt(r.FormValue("status"), 10, 64)
	if err != nil {
		// TODO: what to do here
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	completeDate := database.SqLiteTime{}
	if newStatus == int64(models.Completed) {
		completeDate = *database.SqLiteNow()
	}
	err = h.repo.UpdateItemStatus(id, completeDate, newStatus, currentUserId)
	if err != nil {
		// TODO: what to do here
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) updateItemStatusUI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(errors.New("failed to read part of path"), "failed to parse id from path", err, w)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(errors.New("failed to get task"), "failed to get task %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assert.IsTrue(task.AssignedTo == currentUserId, h.logger, "for now you can only update tasks assigned to you. %d tried to update task %d thats owned by %d", currentUserId, task.Id, task.AssignedTo)

	task.ToggleStatus(currentUserId)

	if err = h.repo.UpdateItemStatus(id, *task.CompleteDate, int64(task.Status), currentUserId); err != nil {
		// TODO: what to do here
		h.logger.Error(err, "Failed to update todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err = h.repo.GetItem(id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to get todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assignedToUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	if ok := h.handleUserIdNotFound(err, task.AssignedTo); !ok {
		assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
		// TODO: what should happen to the row if an error occurs while updating?
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}
	taskListItem := models.TaskToViewModel(task, assignedToUser, createdBy)

	if err = templ_todo.TaskRow(taskListItem).Render(r.Context(), w); err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to render task list item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.listItemsAPI, h.listItemsUI)
}

func (h *Handler) listItemsAPI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

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
	Tasks    []models.TaskView
	NavBar   models.NavBarObject
	FormData models.FormData
}

func (h *Handler) listItemsUI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	currentUserId, currentUserEmail, currentUserName, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	tasks, err := h.repo.GetItemsForUser(currentUserId)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Time.Before(*tasks[j].DueDate.Time)
	})

	formData := models.NewTaskForm()
	formData.Values["due_date"] = time.Now().Add(time.Hour * 24).Format("2006-01-02")
	users := make(map[int64]models.User)

	var tasksList []models.TaskView
	for _, task := range tasks {

		if _, exists := users[task.AssignedTo]; !exists {
			user, err := h.usersRepo.GetUserById(task.AssignedTo)
			assert.NoError(err, h.logger, "how does a task with an assigned user id of %d even exist?", task.AssignedTo)
			users[task.AssignedTo] = user
		}
		if _, exists := users[task.CreatedBy]; !exists {
			user, err := h.usersRepo.GetUserById(task.CreatedBy)
			assert.NoError(err, h.logger, "how does a task with an created by user id of %d even exist?", task.CreatedBy)
			users[task.CreatedBy] = user
		}
		tasksList = append(tasksList, models.TaskToViewModel(task, users[task.AssignedTo], users[task.CreatedBy]))
	}

	data := ListModel{tasksList, h.activeScreens, formData}
	data.NavBar = data.NavBar.SetUser(currentUserName, currentUserEmail)

	if err = templ_todo.TaskList(data.NavBar, formdata, tasksList, tagOptions).Render(r.Context(), w); err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "Failed to execute template for the item list page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	handlers.AcceptHeaderSwitch(w, r, h.deleteItemAPI, h.deleteItemUI)
}

func (h *Handler) deleteItemAPI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.repo.DeleteItem(id, currentUserId, *database.SqLiteNow())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) deleteItemUI(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		handlers.HttpErrorUI(h.templates, "failed to parse id from path", err, w)
		return
	}
	err = h.repo.DeleteItem(id, currentUserId, *database.SqLiteNow())
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

func (h *Handler) testError(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	handlers.HttpErrorUI(h.templates, "Testing the error page", errors.New("Testing the error page"), w)
}

func formDataFromItemNoValidation(task models.Task, assignedUser string) models.TaskForm {
	formData := models.NewTaskForm()
	formData.Task.Name = task.Name
	formData.Task.Description = task.Description
	formData.Task.AssignedTo = assignedUser
	formData.Task.DueDate = task.DueDate
	formData.Task.Tag = task.Tag

	return formData
}

func formDataFromItem(task models.Task, assignedUser string) (models.TaskForm, bool) {
	formData := formDataFromItemNoValidation(task, assignedUser)
	isValid, errs := task.IsValid()
	if !isValid {
		formData.Errors = errs
	}
	return formData, isValid
}

func (h *Handler) handleUserNotFound(err error, userEmail string) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, sql.ErrNoRows) {
		h.logger.Error(err, "the entered email address does not corrispond to an existing user: %s", userEmail)
	} else {
		h.logger.Error(err, "some error occured reading from the db while querying for %s", userEmail)
		assert.NoError(err, h.logger, "some error occurred. probably fialed to read from the db while checking user %s", userEmail)
	}
	return false
}
func (h *Handler) handleUserIdNotFound(err error, userId int64) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, sql.ErrNoRows) {
		h.logger.Error(err, "the entered email address does not corrispond to an existing user: %d", userId)
	} else {
		h.logger.Error(err, "some error occured reading from the db while querying for %d", userId)
		assert.NoError(err, h.logger, "some error occurred. probably fialed to read from the db while checking user %d", userId)
	}
	return false
}
