package router

import (
	"api/cfg"
	"api/service"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type response struct {
	Name    string
	Message string
	Status  int
}

// Get info about all services from file
func getServicesInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service.ServInfo)
}

// Get info about a specific service
func getServiceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	s := strings.ToLower(vars["service"])
	for i, v := range service.ServInfo {
		if v.Name == s {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(service.ServInfo[i])
			break
		}
	}
}

// Get service with max access time
func maxTime(w http.ResponseWriter, _ *http.Request) {
	var (
		maxT float64
		sId  int
	)
	for i, v := range service.ServInfo {
		if t := v.AccessTime; t > maxT {
			maxT = t
			sId = i
			continue
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service.ServInfo[sId])
}

// Get service with min access time
func minTime(w http.ResponseWriter, _ *http.Request) {
	var (
		minT float64
		sId  int
	)
	for i, v := range service.ServInfo {
		if t := v.AccessTime; t < minT {
			minT, sId = t, i
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service.ServInfo[sId])
}

// Get all inactive services
func notActive(w http.ResponseWriter, _ *http.Request) {
	var notActiveS []service.SInfo
	for _, v := range service.ServInfo {
		if s := v.Status; s != 200 {
			notActiveS = append(notActiveS, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notActiveS)
}

// Getting services by access time
func accessTime(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	time := vars["access-time"]
	accTime, err := strconv.ParseFloat(time, 64)
	var response []service.SInfo
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	for _, v := range service.ServInfo {
		if t := v.AccessTime; t == accTime {
			response = append(response, v)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Add new service to file
func addService(w http.ResponseWriter, r *http.Request) {
	config := cfg.GetConfig()
	vars := mux.Vars(r)
	site := vars["service"]
	if isSetService(site) {
		json.NewEncoder(w).Encode(response{Name: site, Message: "error: site is already there ¯\\_(ツ)_/¯", Status: 1})
		return
	}
	f, err := os.OpenFile(config.FileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println("File not open: ", err)
	}
	defer f.Close()

	if _, err = f.WriteString("\n" + site); err != nil {
		log.Fatal("write string: ", err)
	}
	json.NewEncoder(w).Encode(response{Name: site, Message: "all perfectly: site is added! ヽ(・∀・)ﾉ", Status: 0})
	return
}

// Remove service from file
func removeService(w http.ResponseWriter, r *http.Request) {
	config := cfg.GetConfig()
	vars := mux.Vars(r)
	site := vars["service"]
	if !isSetService(site) {
		json.NewEncoder(w).Encode(response{Name: site, Message: "error: there is no such service ¯\\_(ツ)_/¯", Status: 1})
		return
	}
	input, err := ioutil.ReadFile(config.FileName)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if line == site {
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(config.FileName, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(response{Name: site, Message: "all perfectly: site has been removed! ヽ(・∀・)ﾉ", Status: 0})
	return
}

func isSetService(site string) bool {
	config := cfg.GetConfig()
	f, err := os.Open(config.FileName)
	if err != nil {
		log.Println("open file err: ", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if string(line) == site && err == nil {
			return true
		}
		continue
	}
	return false
}
