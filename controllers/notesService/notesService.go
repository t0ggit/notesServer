package notesService

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"notesServer/gates/storage"
	"notesServer/models/dto"
	"notesServer/models/entity"
	"notesServer/pkg"
	"sort"
)

type NotesService struct {
	server  http.Server
	storage storage.Storage
}

func NewNotesService(addr string, storage storage.Storage) (service *NotesService) {
	service = new(NotesService)
	service.server = http.Server{}
	router := http.NewServeMux()
	router.HandleFunc("/create", service.handleCreateNote)
	router.HandleFunc("/get", service.handleGetNote)
	router.HandleFunc("/update", service.handleUpdateNote)
	router.HandleFunc("/delete", service.handleDeleteNoteByID)
	router.HandleFunc("/get-all", service.handleGetAllNotes)
	service.server.Handler = router
	service.server.Addr = addr
	service.storage = storage
	return service
}

// Start запускает сервис заметок
func (ns *NotesService) Start() {
	wErr := pkg.NewWrappedError("(ns *NotesService) Start()")

	err := ns.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		wErr.Specify(err, "ns.server.ListenAndServe()").LogError()
		return
	}
	wErr.LogMsg("Server closed")
	return
}

func (ns *NotesService) Close() error {
	return ns.server.Close()
}

// handleCreateNote обрабатывает запрос на создание записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
{"name": "Имя", "last_name": "Фамилия", "note": "Содержимое заметки"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
{"result": "OK", "data": {"id": 1}, "error": ""}

В случае ошибки:
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (ns *NotesService) handleCreateNote(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(ns *NotesService) handleCreateNote()")
	if err != nil {
		log.Println("(ns *NotesService) handleCreateNote: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		messageString := fmt.Sprintf("invalid request method: '%s' (need '%s')", req.Method, http.MethodPost)
		wErr.LogMsg(messageString)
		resp.Update("ERROR", nil, messageString)
		return
	}

	// Парсинг запроса
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		errorString := fmt.Sprintf("cannot read request bytes: %s", err)
		resp.Update("ERROR", nil, errorString)
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	creatableNote := dto.NewNote()
	err = json.Unmarshal(requestBytes, &creatableNote)
	if err != nil {
		messageString := fmt.Sprintf("cannot unmarshal request json: %s", err.Error())
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Проверка наличия необходимых данных в запросе
	if creatableNote.Name == "" || creatableNote.LastName == "" || creatableNote.Content == "" {
		err = errors.New("required data is missing")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s: {name: '%s', lastName: '%s', note: '%s'}",
			err.Error(), creatableNote.Name, creatableNote.LastName, creatableNote.Content))
		return
	}

	// Вставка записи в хранилище
	id, err := ns.storage.Add(entity.GetPureNote(creatableNote))
	creatableNote.ID = id
	if err != nil {
		resp.Update("ERROR", nil, errors.New("cannot add note: "+err.Error()).Error())
		wErr.Specify(err, "ns.storage.Add(creatableNote)").LogError()
		return
	}

	// Формирование содержимого для ответа
	idMap := map[string]int64{
		"id": id,
	}
	idJson, err := json.Marshal(idMap)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		wErr.Specify(err, "json.Marshal(idMap)").LogError()
		return
	}

	resp.Update("OK", idJson, "")
	wErr.LogMsg(fmt.Sprintf("OK - create: {id: %d}", creatableNote.ID))
}

// handleGetNote обрабатывает запрос на получение одной записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
{"id": 1}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
{"result": "OK", "data": {"id": 1, "name": "Имя", "last_name": "Фамилия", "note": "Содержимое заметки"}, "error": ""}

В случае ошибки:
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (ns *NotesService) handleGetNote(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(ns *NotesService) handleGetNote()")
	if err != nil {
		log.Println("(ns *NotesService) handleCreateNote: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		messageString := fmt.Sprintf("invalid request method: '%s' (need '%s')", req.Method, http.MethodPost)
		wErr.LogMsg(messageString)
		resp.Update("ERROR", nil, messageString)
		return
	}

	// Парсинг запроса
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		errorString := fmt.Sprintf("cannot read request bytes: %s", err)
		resp.Update("ERROR", nil, errorString)
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	gettableNote := dto.NewNote()
	err = json.Unmarshal(requestBytes, &gettableNote)
	if err != nil {
		messageString := fmt.Sprintf("cannot unmarshal request json: %s", err.Error())
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Проверка валидности полученного ID
	if gettableNote.ID < 1 {
		err = errors.New("invalid note id")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s %d", err.Error(), gettableNote.ID))
		return
	}

	// Получение нужной записки по ID
	foundPureNoteAny, status := ns.storage.GetByID(gettableNote.ID)
	if !status {
		err = errors.New(fmt.Sprintf("cannot find note with id %d", gettableNote.ID))
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(err.Error())
		return
	}
	foundPureNote, ok := foundPureNoteAny.(entity.PureNote)
	if !ok {
		err = errors.New("cannot convert interface{} to PureNote")
		resp.Update("ERROR", nil, "internal server error")
		wErr.Specify(err, "foundPureNote, ok := foundPureNoteAny.(PureNote)").LogError()
		return
	}
	foundNote := foundPureNote.ToNoteWithID(gettableNote.ID)

	// Формирование содержимого для ответа
	noteJson, err := json.Marshal(foundNote)
	if err != nil {
		resp.Update("ERROR", nil, err.Error())
		return
	}
	resp.Update("OK", noteJson, "")
	wErr.LogMsg(fmt.Sprintf("OK - get: {id: %d}", gettableNote.ID))
}

// handleUpdateNote обрабатывает запрос на обновлениe записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
{"id": 1, "name": "Имя", "last_name": "Фамилия", "note": "Содержимое заметки"}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
{"result": "OK", "data": null, "error": ""}

В случае ошибки: (например, запись с таким ID не существует)
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (ns *NotesService) handleUpdateNote(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(ns *NotesService) handleUpdateNote()")
	if err != nil {
		log.Println("(ns *NotesService) handleUpdateNote: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		messageString := fmt.Sprintf("invalid request method: '%s' (need '%s')", req.Method, http.MethodPost)
		wErr.LogMsg(messageString)
		resp.Update("ERROR", nil, messageString)
		return
	}

	// Парсинг запроса
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		errorString := fmt.Sprintf("cannot read request bytes: %s", err)
		resp.Update("ERROR", nil, errorString)
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	updatableNote := dto.NewNote()
	err = json.Unmarshal(requestBytes, &updatableNote)
	if err != nil {
		messageString := fmt.Sprintf("cannot unmarshal request json: %s", err.Error())
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Проверка наличия необходимых данных в запросе
	if updatableNote.Name == "" || updatableNote.LastName == "" || updatableNote.Content == "" || updatableNote.ID < 1 {
		err = errors.New("required data is missing")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s: {name: '%s', lastName: '%s', note: '%s'}",
			err.Error(), updatableNote.Name, updatableNote.LastName, updatableNote.Content))
		return
	}

	// Обновление записи
	ok, err := ns.storage.UpdateByID(updatableNote.ID, entity.GetPureNote(updatableNote))
	if err != nil {
		resp.Update("ERROR", nil, "internal server error")
		wErr.Specify(err, "ns.storage.UpdateByID(updatableNote.ID, updatableNote)").LogError()
		return
	}
	if !ok {
		messageString := fmt.Sprintf("cannot update non-existing note: %s", err.Error())
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	resp.Update("OK", nil, "")
	wErr.LogMsg(fmt.Sprintf("OK - update: {id: %d}", updatableNote.ID))
}

// handleDeleteNoteByID обрабатывает запрос на удаление записи
/*
Запрос должен быть с методом POST и с содержимым в формате JSON следующего вида:
{"id": 1}

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
{"result": "OK", "data": null, "error": ""}

В случае ошибки: (например, удаление несуществующей записи)
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (ns *NotesService) handleDeleteNoteByID(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(ns *NotesService) handleDeleteNoteByID()")
	if err != nil {
		log.Println("(ns *NotesService) handleDeleteNoteByID: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodPost {
		messageString := fmt.Sprintf("invalid request method: '%s' (need '%s')", req.Method, http.MethodPost)
		wErr.LogMsg(messageString)
		resp.Update("ERROR", nil, messageString)
		return
	}

	// Парсинг запроса
	requestBytes, err := io.ReadAll(req.Body)
	if err != nil {
		errorString := fmt.Sprintf("cannot read request bytes: %s", err)
		resp.Update("ERROR", nil, errorString)
		wErr.Specify(err, "io.ReadAll(req.Body)").LogError()
		return
	}
	deletableNote := dto.NewNote()
	err = json.Unmarshal(requestBytes, &deletableNote)
	if err != nil {
		messageString := fmt.Sprintf("cannot unmarshal request json: %s", err.Error())
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Проверка наличия необходимых данных в запросе
	if deletableNote.ID < 1 {
		err = errors.New("invalid note id")
		resp.Update("ERROR", nil, err.Error())
		wErr.LogMsg(fmt.Sprintf("%s %d", err.Error(), deletableNote.ID))
		return
	}

	// Проверка наличия записи с таким ID
	_, status := ns.storage.GetByID(deletableNote.ID)
	if !status {
		messageString := fmt.Sprintf("note with this ID doesn't exist: %d", deletableNote.ID)
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Удаление записи
	ns.storage.RemoveByID(deletableNote.ID)

	resp.Update("OK", nil, "")
	wErr.LogMsg(fmt.Sprintf("OK - delete: {id: %d}", deletableNote.ID))
}

// handleGetAllNotes обрабатывает запрос на получение всех записей
/*
Запрос должен быть с методом GET. Тело запроса игнорируется.

Возвращает клиенту ответ с содержимым в формате JSON следующего вида:
{"result": "OK", "data": [
{"id": 1, "name": "Иванов", "last_name": "Иван", "note": "Привет, друг!"},
{"id": 2, "name": "Петров", "last_name": "Петр", "note": "Привет, друг!"}
], "error": ""}

В случае ошибки:
{"result": "ERROR", "data": null, "error": "error description"}
*/
func (ns *NotesService) handleGetAllNotes(w http.ResponseWriter, req *http.Request) {
	setHttpHeaders(w)

	wErr, err := pkg.NewWrappedErrorWithFile("(ns *NotesService) handleGetAllNotes()")
	if err != nil {
		log.Println("(ns *NotesService) handleGetAllNotes: NewWrappedErrorWithFile()", err)
	}

	// Создание ответа
	resp := &dto.Response{}
	defer writeResponseContent(w, resp, wErr)

	// Проверка метода
	if req.Method != http.MethodGet {
		messageString := fmt.Sprintf("invalid request method: '%s' (need '%s')", req.Method, http.MethodGet)
		wErr.LogMsg(messageString)
		resp.Update("ERROR", nil, messageString)
		return
	}

	// Тело запроса игнорируется, поэтому его парсинг не производится

	// Получение всех записей
	allNotesMap, status := ns.storage.GetAll()
	if !status {
		messageString := "no records found"
		resp.Update("ERROR", nil, messageString)
		wErr.LogMsg(messageString)
		return
	}

	// Преобразование из map[int64]interface{} в []*dto.Note
	allNotes := make([]*dto.Note, 0, len(allNotesMap))
	for id, pureNoteAny := range allNotesMap {
		pureNote, ok := pureNoteAny.(entity.PureNote)
		if !ok {
			err = errors.New("cannot convert interface{} to PureNote")
			resp.Update("ERROR", nil, "internal server error")
			wErr.Specify(err, "pureNote, ok := pureNoteAny.(PureNote)").LogError()
			return
		}
		allNotes = append(allNotes, pureNote.ToNoteWithID(id))
	}

	// Сортировка []*dto.Note по возрастанию ID
	sort.Slice(allNotes, func(i, j int) bool {
		return allNotes[i].ID < allNotes[j].ID
	})

	// Формирование содержимого ответа в формате JSON
	allNotesJson, err := json.Marshal(allNotes)
	if err != nil {
		errorString := fmt.Sprintf("cannot marshal all notes: %s", err)
		resp.Update("ERROR", nil, errorString)
		wErr.Specify(err, "json.Marshal(notesSlice)").LogError()
		return
	}

	resp.Update("OK", allNotesJson, "")
	wErr.LogMsg(fmt.Sprintf("OK - get-all: {count: %d}", len(allNotes)))
}

func writeResponseContent(w http.ResponseWriter, resp *dto.Response, wErr *pkg.WrappedError) {
	defer wErr.Close()

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		wErr.Specify(err, "json.NewEncoder(w).Encode(resp)").LogError()
		resp.Update("ERROR", nil, "internal server error")
		return
	}
}

func setHttpHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}
