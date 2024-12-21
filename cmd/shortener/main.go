package main

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/exp/rand"
)

var mylist = make(map[string]string)

func mainPage(w http.ResponseWriter, r *http.Request) {

	// Получение ссылки по сокращенному коду
	if r.Method == http.MethodGet {

		val, ok := mylist[r.URL.Path]
		if ok {

			// Проверим URL
			parsedUrl, err := url.Parse(val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if parsedUrl.Scheme == "" {
				parsedUrl.Scheme = "http"
			}
			w.Header().Set("Location", parsedUrl.String())
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
			//http.Redirect(w, r, parsedUrl.String(), http.StatusTemporaryRedirect)
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
		shortUrl := ""
		for ok {
			shortUrl = randStr(4)
			_, ok = mylist["/"+shortUrl]
		}
		mylist["/"+shortUrl] = string(myurl)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortUrl))
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	return

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
