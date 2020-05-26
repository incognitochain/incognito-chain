package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("main run ....")
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()
	r.HandleFunc("/{filename}", processLog)

	http.Handle("/", r)
	http.ListenAndServe(":9000", nil)
}

func processLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	filename = strings.Replace(filename, "file.data.", "", -1)

	if strings.Index(filename, "beacon") == -1 && strings.Index(filename, "shard") == -1 {
		return
	}

	bs, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(bs))
	fd, e := os.OpenFile("/data/"+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if e != nil {
		fmt.Println(e)
	}
	fd.Write(append(bs, []byte("\n")...))
	fd.Close()
}
