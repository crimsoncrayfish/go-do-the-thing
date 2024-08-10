package todo

import (
	"encoding/json"
	"fmt"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
	"go-do-the-thing/navigation"
	"go-do-the-thing/shared"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type Handler struct {
	repo          Repo
	templates     helpers.Templates
	activeScreens navigation.NavBarObject
}

func New(r Repo, templates helpers.Templates) *Handler {
	return &Handler{repo: r, templates: templates, activeScreens: navigation.NavBarObject{IsTodoList: true}}
}

type idResponse struct {
	Id int64 `json:"id" json:"id"`
}

func (item *Item) isValid() (bool, map[string]string) {
	errors := make(map[string]string)
	isValid := true

	now := time.Now()
	if item.DueDate.Before(now) {
		isValid = false
		errors["due_date"] = "Due date is before now"
	}
	return isValid, errors
}

func (item *Item) FormDataFromCreateItem() (shared.FormData, bool) {
	formData := shared.NewFormData()
	formData.Values["description"] = item.Description
	formData.Values["assigned_to"] = item.AssignedTo
	formData.Values["due_date"] = item.DueDate.StringF(database.DateFormat)
	isValid, errors := item.isValid()
	if !isValid {
		formData.Errors = errors
	}
	return formData, isValid
}

func (h *Handler) CreateItemAPI(w http.ResponseWriter, r *http.Request) {
	var item Item
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&item)
	if err != nil {
		println("failed to decode todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.repo.InsertItem(item)
	if err != nil {
		println("failed to insert item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	idResponse := idResponse{
		Id: id,
	}
	jsonBytes, err := json.Marshal(idResponse)
	if err != nil {
		println("failed to marshal id response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		println("failed to write response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetItemAPI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := h.repo.GetItem(id)
	if err != nil {
		println("failed to get todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	items := make([]Item, 1)
	items[0] = item
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		println("failed to marshal todo item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		println("failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) DeleteItemAPI(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateItemStatusAPI(w http.ResponseWriter, r *http.Request) {
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
	err = h.repo.UpdateItemStatus(id, newStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) ListItemsAPI(w http.ResponseWriter, _ *http.Request) {
	items, err := h.repo.GetItems()
	if err != nil {
		println("failed to get todo items")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonBytes, err := json.Marshal(items)
	if err != nil {
		println("failed to marshal todo items")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		println("failed to write response")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

type ListModel struct {
	Tasks         []Item
	ActiveScreens navigation.NavBarObject
}

func (h *Handler) ListItemsUI(w http.ResponseWriter, _ *http.Request) {
	tasks, err := h.repo.GetItems()
	if err != nil {
		fmt.Println("failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Time.Before(tasks[j].DueDate.Time)
	})
	responseObject := ListModel{tasks, h.activeScreens}
	err = h.templates.RenderOk(w, "task-list", responseObject)
	if err != nil {
		fmt.Println("Failed to execute tmpl for the item list page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ItemPageModel struct {
	Task          Item
	ActiveScreens navigation.NavBarObject
}

func (h *Handler) GetItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpError("failed to parse id from path", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		fmt.Println("failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	model := ItemPageModel{task, h.activeScreens}
	err = h.templates.RenderOk(w, "task-item", model)
	if err != nil {
		fmt.Println("Failed to execute tmpl for the item page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ToggleItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpError("failed to parse id from path", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		httpError("failed to get todo item", err, w)
		return
	}
	var newStatus ItemStatus
	if task.Status == Scheduled {
		newStatus = Completed
	} else {
		newStatus = Scheduled
	}

	if err = h.repo.UpdateItemStatus(id, int64(newStatus)); err != nil {
		httpError("Failed to update todo item", err, w)
		return
	}
	task, err = h.repo.GetItem(id)
	if err != nil {
		httpError("failed to get todo item", err, w)
		return
	}

	if err = h.templates.RenderOk(w, "task-row", task); err != nil {
		httpError("Failed to execute tmpl for the home page", err, w)
		return
	}
}

func (h *Handler) CreateItemUI(w http.ResponseWriter, r *http.Request) {
	//Read data from submitted form
	dateRaw := r.FormValue("due_date")
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		httpError("failed to parse due date", err, w)
		return
	}
	item := Item{
		Description: r.FormValue("description"),
		AssignedTo:  r.FormValue("assigned_to"),
		DueDate:     &database.SqLiteTime{Time: date},
		CreatedBy:   "CurrentUser", //todo logins
		CreateDate:  &database.SqLiteTime{Time: time.Now()},
		IsDeleted:   false,
		Tag:         r.FormValue("tag"),
	}

	//Check if form is valid and respond with any error
	formData, isValid := item.FormDataFromCreateItem()
	if !isValid {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "create-task-form-content", formData); err != nil {
			httpError("Failed to render template for formData", err, w)
		}
		return
	}

	//update data
	id, err := h.repo.InsertItem(item)
	if err != nil {
		httpError("failed to insert todo item", err, w)
		return
	}
	task, err := h.repo.GetItem(id)
	if err != nil {
		httpError("failed to get todo item", err, w)
		return
	}
	//Respond with templates
	if err = h.templates.RenderOk(w, "task-row-oob", task); err != nil {
		httpError("Failed to render item row", err, w)
		return
	}
	err = h.templates.RenderOk(w, "create-task-form-content", shared.NewFormData())
	if err != nil {
		httpError("Failed to render form", err, w)
		return
	}

}

func (h *Handler) DeleteItemUI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpError("failed to parse id from path", err, w)
		return
	}
	err = h.repo.DeleteItem(id)
	if err != nil {
		httpError("failed to delete todo item", err, w)
		return
	}
}

func httpError(message string, err error, w http.ResponseWriter) {
	fmt.Println(message)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
