package router

import "github.com/gorilla/mux"

func New() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", getServicesInfo).Methods("GET")

	r.HandleFunc("/max-time", maxTime).Methods("GET")
	r.HandleFunc("/min-time", minTime).Methods("GET")
	r.HandleFunc("/not-active", notActive).Methods("GET")
	r.HandleFunc("/{access-time:[0-9]*[.]?[0-9]+}", accessTime).Methods("GET")
	r.HandleFunc("/{service}", getServiceInfo).Methods("GET")

	r.HandleFunc("/{service}", addService).Methods("PUT")
	r.HandleFunc("/{service}", removeService).Methods("DELETE")

	return r
}
