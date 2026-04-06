package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

// Create godoc
// @Summary      Создать задачу
// @Description  Создает одну задачу или серию периодических задач на основе правила
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      taskMutationDTO  true  "Данные задачи"
// @Success      201   {object}  taskDTO
// @Failure      400   {object}  map[string]string
// @Router       /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	recurrence := mapDTOToDomainRecurrence(req.Recurrence)

	createdTasks, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		ScheduledDate:  req.ScheduledDate,
		RecurrenceRule: recurrence,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	if len(createdTasks) == 1 {
		writeJSON(w, http.StatusCreated, newTaskDTO(&createdTasks[0]))
	} else {
		response := make([]taskDTO, 0, len(createdTasks))
		for i := range createdTasks {
			response = append(response, newTaskDTO(&createdTasks[i]))
		}
		writeJSON(w, http.StatusCreated, response)
	}
}

// GetByID godoc
// @Summary      Получить задачу
// @Description  Получить подробную информацию о конкретной задаче по её ID
// @Tags         tasks
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  taskDTO
// @Failure      404  {object}  map[string]string "Задача не найдена"
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

// Update godoc
// @Summary      Обновить задачу
// @Description  Изменить заголовок, описание, статус или дату конкретной задачи
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id       path      int              true  "Task ID"
// @Param        request  body      taskMutationDTO  true  "Новые данные"
// @Success      200      {object}  taskDTO
// @Failure      404      {object}  map[string]string "Задача не найдена"
// @Router       /tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

// Delete godoc
// @Summary      Удалить задачу
// @Description  Удалить конкретный экземпляр задачи из системы
// @Tags         tasks
// @Param        id   path      int  true  "Task ID"
// @Success      204  "Без контента"
// @Failure      404  {object}  map[string]string "Задача не найдена"
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      Список задач
// @Description  Получить список задач с пагинацией и информацией о количестве
// @Tags         tasks
// @Produce      json
// @Param        start   query     string  false  "Дата начала"
// @Param        end     query     string  false  "Дата конца"
// @Param        limit   query     int     false  "Лимит"
// @Param        offset  query     int     false  "Смещение"
// @Success      200     {object}  taskListResponseDTO
// @Router       /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	start, _ := time.Parse("2006-01-02", query.Get("start"))
	end, _ := time.Parse("2006-01-02", query.Get("end"))

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	tasks, total, err := h.usecase.List(r.Context(), start, end, limit, offset)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	taskDTOs := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		taskDTOs = append(taskDTOs, newTaskDTO(&tasks[i]))
	}

	// Отдаем объект вместо массива
	writeJSON(w, http.StatusOK, taskListResponseDTO{
		Items:      taskDTOs,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	})
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func mapDTOToDomainRecurrence(dto *recurrenceDTO) *taskdomain.RecurrenceRule {
	if dto == nil {
		return nil
	}

	var parity *taskdomain.ParityType
	if dto.Parity != nil {
		p := taskdomain.ParityType(*dto.Parity)
		parity = &p
	}

	return &taskdomain.RecurrenceRule{
		RuleType:      taskdomain.RuleType(dto.Type),
		IntervalDays:  dto.IntervalDays,
		MonthlyDay:    dto.MonthlyDay,
		Parity:        parity,
		ValidUntil:    dto.ValidUntil,
		SpecificDates: dto.SpecificDates,
	}
}
