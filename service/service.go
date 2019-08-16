package service

import (
	"api/cfg"
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SInfo struct {
	Name       string  `json:"name"`
	Status     int     `json:"status"`
	AccessTime float64 `json:"access_time"`
}

// Service data slice
var ServInfo []SInfo

//
func Check() time.Duration {
	file, ok := os.Open("sites.txt")
	if ok != nil {
		log.Fatalf("File was not open: %s", ok)
	}
	defer file.Close()

	// ServiceCh - in channel, resultCh - out channel
	serviceCh, resultCh := make(chan string, 10), make(chan SInfo, 10)

	f := bufio.NewReader(file)
	wg := &sync.WaitGroup{}

	// For Waiting Handler Goroutine
	wgDone := &sync.WaitGroup{}

	cl := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 7 * time.Second,
			}).DialContext,
		},
		Timeout: time.Second * 30,
	}
	now := time.Now()

	// create pull workers
	for i := 0; i <= 20; i++ {
		wg.Add(1)
		go checkService(serviceCh, resultCh, wg, &cl)
	}

	index := 0

	var jData []SInfo // json data of services
	// collects data from workers
	go func(ch <-chan SInfo, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		for result := range ch {
			jData = append(jData, result)
			fmt.Println(result)
			index++
		}
		ServInfo = nil
		ServInfo = append(ServInfo, jData...)
		sendToDB() // send to db ServInfo for statistic, query time: 368ms for 43 sites
		return
	}(resultCh, wgDone)

	for {
		data, _, err := f.ReadLine()
		if err == io.EOF {
			close(serviceCh)
			break
		}
		serviceCh <- string(data)
	}
	wg.Wait()
	close(resultCh)
	wgDone.Wait()
	runtime := now.Sub(time.Now())
	fmt.Printf("runtime: %v, cnt sites: %v", runtime, index)
	return runtime
}

// Worker. Make requests to service and return info about them.
// Return its runtime
func checkService(in <-chan string, out chan<- SInfo, wg *sync.WaitGroup, client *http.Client) {
	defer wg.Done()
	for service := range in {
		r, _ := http.NewRequest("GET", "http://"+service, nil)
		r.Header.Set("User-Agent", "Mozilla/5.0")

		startTime := time.Now()
		resp, err := client.Do(r) // Execute request
		resTime := time.Since(startTime)
		if err != nil {
			out <- SInfo{
				Name:       service,
				Status:     000,
				AccessTime: 0,
			}
			continue
		}
		// Access Time
		t, err := checkTime(service)
		// if site ignore ping command
		if err != nil {
			out <- SInfo{
				Name:       service,
				Status:     resp.StatusCode,
				AccessTime: resTime.Seconds() * 1000,
			}
			resp.Body.Close()
			continue
		}
		out <- SInfo{
			Name:       service,
			Status:     resp.StatusCode,
			AccessTime: t,
		}
		resp.Body.Close()
	}
}

// Check Access Time of services
func checkTime(url string) (float64, error) {
	cmd, err := exec.Command("ping", url, "-c 1").Output()
	if err != nil {
		return 0, err
	}
	output := string(cmd)
	startStr := strings.Index(output, "time=") + 5
	endStr := strings.Index(output[startStr:], " ms")

	return strconv.ParseFloat(output[startStr:startStr+endStr], 64)
}

// Send to db info for statistics
func sendToDB() {
	config := cfg.GetConfig()
	db, err := sql.Open("mysql", config.DbUser+":"+config.DbPassword+"@/"+config.DbName)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	data := ServInfo
	for _, v := range data {
		_, err = db.Exec("insert into statistics.services (name, status, access_time, timestamp) values (?, ?, ?, now())",
			v.Name, v.Status, v.AccessTime)
		if err != nil {
			log.Println("db exec error: ", err)
		}
	}
	return
}
