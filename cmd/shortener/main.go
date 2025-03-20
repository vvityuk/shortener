package main

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/exp/rand"
)

var mylist = make(map[string]string)

func mainPage(w http.ResponseWriter, r *http.Request) {

	// Получение ссылки по сокращенному коду
	if r.Method == http.MethodGet {

		val, ok := mylist[r.URL.Path]
		if ok {
			w.Header().Set("Location", val)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	// Получение сокращенной ссылки
	if r.Method == http.MethodPost {
		defer r.Body.Close()

		myurl, _ := io.ReadAll(r.Body)
		ok := true
		shortURL := ""
		for ok {
			shortURL = randStr(4)
			_, ok = mylist["/"+shortURL]
		}
		mylist["/"+shortURL] = string(myurl)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortURL))
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)

}

func main() {

	http.HandleFunc(`/`, mainPage)
	err := http.ListenAndServe(`:8080`, nil)

	if err != nil {
		panic(err)
	}

}

func randStr(n int) string {

	rnd := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}
