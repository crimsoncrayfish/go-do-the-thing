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

func (item *Item) formDataFromCreateItem() (shared.FormData, bool) {
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

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	acceptHeaderSwitch(w, r, h.createItemAPI, h.createItemUI)
}

func (h *Handler) createItemAPI(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) createItemUI(w http.ResponseWriter, r *http.Request) {
	errorForm := shared.NewFormData()
	description, errorForm := getRequiredPropertyFromRequest(r, "description", errorForm)
	assignedTo, errorForm := getRequiredPropertyFromRequest(r, "assigned_to", errorForm)
	tag, errorForm := getRequiredPropertyFromRequest(r, "tag", errorForm)
	dateRaw, errorForm := getRequiredPropertyFromRequest(r, "due_date", errorForm)
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil || len(errorForm.Errors) > 0 {
		if err := h.templates.RenderWithCode(w, http.StatusUnprocessableEntity, "create-task-form-content", errorForm); err != nil {
			httpError("Failed to render template for formData", err, w)
		}
		return
	}
	item := Item{
		Description: description,
		AssignedTo:  assignedTo,
		DueDate:     &database.SqLiteTime{Time: date},
		CreatedBy:   "CurrentUser", //todo logins
		CreateDate:  &database.SqLiteTime{Time: time.Now()},
		IsDeleted:   false,
		Tag:         tag,
	}

	//Check if form is valid and respond with any error
	formData, isValid := item.formDataFromCreateItem()
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

func getRequiredPropertyFromRequest(r *http.Request, propName string, formData shared.FormData) (string, shared.FormData) {
	value := r.FormValue(propName)
	if len(value) == 0 {
		formData.Errors[propName] = propName + " is required"
		return value, formData
	}
	return value, formData
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	acceptHeaderSwitch(w, r, h.getItemAPI, h.getItemUI)
}

func (h *Handler) getItemAPI(w http.ResponseWriter, r *http.Request) {
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

type ItemPageModel struct {
	Task          Item
	ActiveScreens navigation.NavBarObject
}

func (h *Handler) getItemUI(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	acceptHeaderSwitch(w, r, h.updateItemStatusAPI, h.updateItemStatusUI)
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
	err = h.repo.UpdateItemStatus(id, newStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) updateItemStatusUI(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	acceptHeaderSwitch(w, r, h.listItemsAPI, h.listItemsUI)
}

func (h *Handler) listItemsAPI(w http.ResponseWriter, _ *http.Request) {
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
	shared.FormData
}

func (h *Handler) listItemsUI(w http.ResponseWriter, _ *http.Request) {
	tasks, err := h.repo.GetItems()
	if err != nil {
		fmt.Println("failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate.Time.Before(tasks[j].DueDate.Time)
	})

	formData := shared.NewFormData()
	formData.Values["due_date"] = time.Now().Add(time.Hour * 24).Format("2006-01-02")
	responseObject := ListModel{tasks, h.activeScreens, formData}
	err = h.templates.RenderOk(w, "task-list", responseObject)
	if err != nil {
		fmt.Println("Failed to execute tmpl for the item list page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	acceptHeaderSwitch(w, r, h.deleteItemAPI, h.deleteItemUI)
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

func acceptHeaderSwitch(w http.ResponseWriter, r *http.Request, jsonFunc func(w http.ResponseWriter, r *http.Request), uiFunc func(w http.ResponseWriter, r *http.Request)) {
	contentType := r.Header.Get("accept")
	if contentType == "application/json" {
		jsonFunc(w, r)
	} else if contentType == "text/html" {
		uiFunc(w, r)
	} else {
		httpError("No Content-type specified", nil, w)
	}
}
