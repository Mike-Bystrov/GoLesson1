package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var users map[int]UserInfo = make(map[int]UserInfo)
var uniq_id = 0

var mu sync.Mutex
var count int
var startTime = time.Now()

type LoginRequest struct {
	Name string `json:"name"`
}

type DepositRequest struct {
	Id     int `json:"id"`
	Amount int `json:"amount"`
}

type UserInfo struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}

type LoginResponse struct {
	Id      int `json:"id"`
	Balance int `json:"balance"`
}

func loginFunc(w http.ResponseWriter, r *http.Request) {
	request := LoginRequest{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("decode" + err.Error())
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if request.Name != "" {
		// здесь логинимся
		uniq_id++

		new_User := UserInfo{
			Id:      uniq_id,
			Name:    request.Name,
			Balance: 0,
		}
		users[uniq_id] = new_User
		loginResp := LoginResponse{
			Id:      uniq_id,
			Balance: new_User.Balance,
		}

		contentJSON, err := json.Marshal(loginResp)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(contentJSON)
	} else {
		fmt.Println("name is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func depositFunc(w http.ResponseWriter, r *http.Request) {
	request := DepositRequest{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("decode" + err.Error())
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	_, is := users[request.Id]
	if is {
		user := users[request.Id]
		user.Balance += request.Amount
		users[request.Id] = user

		contentJSON, err := json.Marshal(users[request.Id])

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(contentJSON)

	} else {
		fmt.Println("no such user")
		w.Write([]byte("no such user"))
		w.WriteHeader(http.StatusBadRequest)
	}

}

func withdrawFunc(w http.ResponseWriter, r *http.Request) {
	request := DepositRequest{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("decode" + err.Error())
		w.WriteHeader(http.StatusBadRequest)

		return
	}
	_, is := users[request.Id]
	if is {
		user := users[request.Id]
		user.Balance -= request.Amount
		users[request.Id] = user

		contentJSON, err := json.Marshal(users[request.Id])

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(contentJSON)

	} else {
		fmt.Println("no such user")
		w.Write([]byte("no such user"))
		w.WriteHeader(http.StatusBadRequest)
	}

}

type Response struct {
	UserObjects []UserInfo `json:"users_info"`
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Write([]byte("wrong id"))
		fmt.Printf("Error")
		return
	}
	_, is := users[id]

	if is {
		contentJSON, err := json.Marshal(users[id])

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(contentJSON)
	} else {
		fmt.Println("no such user")
		w.Write([]byte("no such user"))
		w.WriteHeader(http.StatusBadRequest)
	}

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()

		if time.Since(startTime) > time.Minute {
			startTime = time.Now()
			count = 0
		}

		if count >= 3 {
			mu.Unlock()
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		} else {
			count++
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /oplati/login", loginFunc)
	mux.HandleFunc("POST /oplati/deposit", depositFunc)
	mux.HandleFunc("POST /oplati/withdraw", withdrawFunc)
	mux.HandleFunc("GET /oplati/balance/{id}", getBalance)

	server := http.Server{
		Addr:    ":8081",
		Handler: loggingMiddleware(mux),
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
