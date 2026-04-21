package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
)

func NewRouter(taskHandler *httphandlers.TaskHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	return newRouterWithoutRecurrence(taskHandler, docsHandler)
}

func NewRouterWithRecurrence(taskHandler *httphandlers.TaskHandler, recurrenceHandler *httphandlers.RecurrenceHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	// Task endpoints
	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	// Recurrence Rule endpoints
	api.HandleFunc("/recurrence-rules", recurrenceHandler.CreateRecurrenceRule).Methods(http.MethodPost)
	api.HandleFunc("/recurrence-rules", recurrenceHandler.ListRecurrenceRules).Methods(http.MethodGet)
	api.HandleFunc("/recurrence-rules/{id:[0-9]+}", recurrenceHandler.GetRecurrenceRule).Methods(http.MethodGet)
	api.HandleFunc("/recurrence-rules/{id:[0-9]+}", recurrenceHandler.UpdateRecurrenceRule).Methods(http.MethodPut)
	api.HandleFunc("/recurrence-rules/{id:[0-9]+}", recurrenceHandler.DeleteRecurrenceRule).Methods(http.MethodDelete)

	// Task with Recurrence endpoint
	api.HandleFunc("/tasks/with-recurrence", recurrenceHandler.CreateTaskWithRecurrence).Methods(http.MethodPost)

	return router
}

func newRouterWithoutRecurrence(taskHandler *httphandlers.TaskHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	return router
}
