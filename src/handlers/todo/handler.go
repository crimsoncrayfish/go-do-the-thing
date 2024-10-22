package todo

import (
	"database/sql"
	"errors"
	"go-do-the-thing/src/database"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	users_repo "go-do-the-thing/src/database/repos/users"
	templ_todo "go-do-the-thing/src/handlers/todo/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	fm "go-do-the-thing/src/models/forms"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type Handler struct {
	repo      tasks_repo.TasksRepo
	usersRepo users_repo.UsersRepo
	logger    slog.Logger
}

var activeScreens models.NavBarObject

var source = assert.Source{"TasksHandler"}

func SetupTodoHandler(
	tasksRepo tasks_repo.TasksRepo,
	usersRepo users_repo.UsersRepo,
	router *http.ServeMux,
	mw_stack middleware.Middleware,
) {
	logger := slog.NewLogger(source.Name)

	activeScreens = models.NavBarObject{ActiveScreens: models.ActiveScreens{IsTodoList: true}}
	todoHandler := &Handler{
		repo:      tasksRepo,
		usersRepo: usersRepo,
		logger:    logger,
	}

	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.getItemUI)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.listItemsUI)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItemStatusUI)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.createItemUI)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItemUI)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.deleteItemUI)))
	router.Handle("GET /error", mw_stack(http.HandlerFunc(todoHandler.testError)))
}

type idResponse struct {
	Id int64 `json:"id" json:"id"`
}

var defaultForm = fm.NewDefaultTaskForm()

func (h *Handler) createItemUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	form := fm.NewTaskForm()
	name, err := models.GetPropertyFromRequest(r, "name", "Task Name", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description, _ := models.GetPropertyFromRequest(r, "description", "Description", false)

	dateRaw, err := models.GetPropertyFromRequest(r, "due_date", "Due on", true)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(date),
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_todo.TaskFormContent("Create", form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
		}
		return
	}

	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      database.NewSqliteTime(date),
		AssignedTo:   currentUserId, // TODO: need to update this
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
	}

	// NOTE: Validate data
	form, isValid := formDataFromItem(task, currentUserEmail)
	if !isValid {
		h.logger.Info("invalid data")
		if err := templ_todo.TaskFormContent("Create", form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
			// TODO: better error handling. Honestly i might as well panic here
		}
		return
	}

	// NOTE: Take action
	id, err := h.repo.InsertItem(task)
	if err != nil {
		h.logger.Error(err, "failed to insert task")
		form.Errors["Task"] = "failed to create task"
		if err := templ_todo.TaskFormContent("Create", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	task, err = h.repo.GetItem(id)
	if err != nil {
		assert.NoError(err, source, "failed to get newly inserted task")
		// TODO: what should happen if the fetch fails after create
		return
	}

	// NOTE: Success zone
	assignedToUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	if ok := h.handleUserIdNotFound(err, task.AssignedTo); !ok {
		assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
		// TODO: what should happen if the fetch fails after create
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen if the fetch fails after create
			return
		}
	}
	taskListItem := models.TaskToViewModel(task, assignedToUser, createdBy)
	if err := templ_todo.TaskRowOOB(taskListItem).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render new task row with id %d", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		//if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		assert.NoError(err, source, "failed to render no data row")
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_todo.TaskFormContent("Create", defaultForm).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render the task form after creation")
		// TODO: what should happen if rendering fails
		return
	}
}

type NoItemRowData struct {
	HideNoData bool
}

func (h *Handler) updateItemUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	form := fm.NewTaskForm()
	name, err := models.GetPropertyFromRequest(r, "name", "Task Name", true)
	if err != nil {
		form.Errors["name"] = err.Error()
	}
	description, _ := models.GetPropertyFromRequest(r, "description", "Description", false)

	dateRaw, err := models.GetPropertyFromRequest(r, "due_date", "Due on", true)
	if err != nil {
		form.Errors["due_on"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(date),
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_todo.TaskFormContent("Update", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to render task form")
			return
		}
		return
	}

	// TODO: Get task from the db and only update relevant fields
	assignedToUser, err := h.usersRepo.GetUserById(currentUserId)
	if ok := h.handleUserIdNotFound(err, currentUserId); !ok {
		assert.NoError(err, source, "how did we get here when i cant read the current user from the db")
		return
	}
	item := models.Task{
		Id:           id,
		Name:         name,
		Description:  description,
		AssignedTo:   assignedToUser.Id,
		DueDate:      database.NewSqliteTime(date),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
	}

	// NOTE: Take action
	if err = h.repo.UpdateItem(item); err != nil {
		h.logger.Error(err, "failed to update task")
		form.Errors["Task"] = "failed to update task"
		if err := templ_todo.TaskFormContent("Update", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	// NOTE: Success zone
	task, err := h.repo.GetItem(id)
	if err != nil {
		assert.NoError(err, source, "failed to get updated task")
		// TODO: what should happen if the fetch fails after update
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}

	model := models.TaskToViewModel(task, assignedToUser, createdBy)
	if err := templ_todo.TaskItemContentOOB(model).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render new task row with id %d", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}

	formData := formDataFromItemNoValidation(item, currentUserEmail)
	if err := templ_todo.TaskFormContent("Update", formData).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render form content after update")
		// TODO: what should happen if the fetch fails after create
		return
	}
}

type ItemPageModel struct {
	Task   models.TaskView
	NavBar models.NavBarObject
}

func (h *Handler) getItemUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: Take action
	assignedUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	assert.NoError(err, source, "this user should exist since they are assigned to a task: %d", task.AssignedTo)
	formData := formDataFromItemNoValidation(task, assignedUser.Email)

	// NOTE: Success zone
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}

	taskView := models.TaskToViewModel(task, assignedUser, createdBy)
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_todo.TaskItem(taskView, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the item page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskItemWithBody(taskView, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the item page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func (h *Handler) updateItemStatusUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Validate data
	task, err := h.repo.GetItem(id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(errors.New("failed to get task"), "failed to get task %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assert.IsTrue(task.AssignedTo == currentUserId, source, "for now you can only update tasks assigned to you. %d tried to update task %d thats owned by %d", currentUserId, task.Id, task.AssignedTo)

	// NOTE: Take action
	task.ToggleStatus(currentUserId)
	if err = h.repo.UpdateItemStatus(id, task.CompleteDate, int64(task.Status), currentUserId); err != nil {
		// TODO: what to do here
		h.logger.Error(err, "Failed to update todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Success zone
	task, err = h.repo.GetItem(id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to get todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	assignedToUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	if ok := h.handleUserIdNotFound(err, task.AssignedTo); !ok {
		assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
		// TODO: what should happen to the row if an error occurs while updating?
		return
	}
	var createdBy models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen to the row if an error occurs while updating?
			return
		}
	}
	taskListItem := models.TaskToViewModel(task, assignedToUser, createdBy)
	if err = templ_todo.TaskRowContent(taskListItem).Render(r.Context(), w); err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to render task list item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ListModel struct {
	Tasks    []models.TaskView
	NavBar   models.NavBarObject
	FormData fm.TaskForm
}

func (h *Handler) listItemsUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Take action
	tasks, err := h.repo.GetItemsForUser(currentUserId)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Before(tasks[j].DueDate)
	})

	users := make(map[int64]models.User)
	var tasksList []models.TaskView
	for _, task := range tasks {

		if _, exists := users[task.AssignedTo]; !exists {
			user, err := h.usersRepo.GetUserById(task.AssignedTo)
			assert.NoError(err, source, "how does a task with an assigned user id of %d even exist?", task.AssignedTo)
			users[task.AssignedTo] = user
		}
		if _, exists := users[task.CreatedBy]; !exists {
			user, err := h.usersRepo.GetUserById(task.CreatedBy)
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.CreatedBy)
			users[task.CreatedBy] = user
		}
		tasksList = append(tasksList, models.TaskToViewModel(task, users[task.AssignedTo], users[task.CreatedBy]))
	}

	// NOTE: Success zone
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_todo.TaskList(activeScreens, defaultForm, tasksList).Render(r.Context(), w); err != nil {
			// TODO: Should this panic
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskListWithBody(activeScreens, defaultForm, tasksList).Render(r.Context(), w); err != nil {
			// TODO: Should this panic
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func (h *Handler) deleteItemUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Take action
	err = h.repo.DeleteItem(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to delete todo item")
		return
	}

	// NOTE: Success zone
	hasData, err := h.repo.GetItemsCount(currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to update ui")
		return
	}

	if err := templ_shared.NoDataRowOOB(hasData > 0).Render(r.Context(), w); err != nil {
		//if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		assert.NoError(err, source, "failed to render no data row")
		// TODO: what should happen if the fetch fails after create
		return
	}
}

func (h *Handler) testError(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")
}

func formDataFromItemNoValidation(task models.Task, assignedUser string) fm.TaskForm {
	formData := fm.NewTaskForm()
	formData.Task.Name = task.Name
	formData.Task.Description = task.Description
	formData.Task.AssignedTo = assignedUser
	formData.Task.DueDate = task.DueDate

	return formData
}

func formDataFromItem(task models.Task, assignedUser string) (fm.TaskForm, bool) {
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
		assert.NoError(err, source, "some error occurred. probably fialed to read from the db while checking user %s", userEmail)
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
		assert.NoError(err, source, "some error occurred. probably fialed to read from the db while checking user %d", userId)
	}
	return false
}
