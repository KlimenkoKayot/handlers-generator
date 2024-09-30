package main

import (
	"strconv"
	"encoding/json"
	"io"
	"net/http"
)

// Result from wrappers
type resValue map[string]interface{}

// ...
// generated for type: MyApi
// ...

// main.methodOptions{URL:"/user/profile", Auth:false, Method:""}
// [Wrapper for MyApi] method: Profile
func (node *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	// validation ProfileParams
	r.ParseForm()
	paramLogin := r.Form.Get("login")
	// tplRequired
	if paramLogin == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "login must me not empty"})
		io.WriteString(w, string(data))
		return
	}

	params := ProfileParams{
		Login: paramLogin,
	}
	ctx := r.Context()
	response, err := node.Profile(ctx, params)
	if err != nil {
		switch err.(type) {
		case ApiError:
			data, _ := json.Marshal(resValue{"error": err.(ApiError).Err.Error()})
			w.WriteHeader(err.(ApiError).HTTPStatus)
			io.WriteString(w, string(data))
		default:
			data, _ := json.Marshal(resValue{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, string(data))
		}
		return
	}
	data, _ := json.Marshal(resValue{"error": "", "response": response})
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(data))	
}

// main.methodOptions{URL:"/user/create", Auth:true, Method:"POST"}
// [Wrapper for MyApi] method: Create
func (node *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	// Authorization checker	
	authGood := "100500"
	auth := r.Header.Get("X-Auth")
	if auth != authGood {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resValue{"error": "unauthorized"})
		io.WriteString(w, string(data))
		return
	} 	

	// Method checker
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		data, _ := json.Marshal(resValue{"error": "bad method"})
		io.WriteString(w, string(data))	
		return
	}

	// validation CreateParams
	r.ParseForm()
	paramLogin := r.Form.Get("login")
	// tplRequired
	if paramLogin == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "login must me not empty"})
		io.WriteString(w, string(data))
		return
	}

	// tplMin
	if len([]rune(paramLogin)) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "login len must be >= 10"})
		io.WriteString(w, string(data))
		return 
	}
	
	paramName := r.Form.Get("full_name")
	paramStatus := r.Form.Get("status")
	// tplDefault	
	if paramStatus == "" {
		paramStatus = "user"
	}

	// tplEnum
	enumFlag := false	
	if paramStatus == "user" {
		enumFlag = true
	}
	if paramStatus == "moderator" {
		enumFlag = true
	}
	if paramStatus == "admin" {
		enumFlag = true
	}
	if !enumFlag {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "status must be one of [user, moderator, admin]"})
		io.WriteString(w, string(data))
		return
	}

	paramAge := r.Form.Get("age")
	// tplMin
	paramAgeIntMin, err := strconv.Atoi(paramAge)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "age must be int"})
		io.WriteString(w, string(data))
		return
	}
	if paramAgeIntMin < 0 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "age must be >= 0"})
		io.WriteString(w, string(data))
		return 
	}
	
	// tplMax
	paramAgeIntMax, err := strconv.Atoi(paramAge)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "age must be int"})
		io.WriteString(w, string(data))
		return
	}
	if paramAgeIntMax > 128 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "age must be <= 128"})
		io.WriteString(w, string(data))
		return 
	}
	
	paramAgeInt, _ := strconv.Atoi(paramAge)
	params := CreateParams{
		Login: paramLogin,
		Name: paramName,
		Status: paramStatus,
		Age: paramAgeInt,
	}
	ctx := r.Context()
	response, err := node.Create(ctx, params)
	if err != nil {
		switch err.(type) {
		case ApiError:
			data, _ := json.Marshal(resValue{"error": err.(ApiError).Err.Error()})
			w.WriteHeader(err.(ApiError).HTTPStatus)
			io.WriteString(w, string(data))
		default:
			data, _ := json.Marshal(resValue{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, string(data))
		}
		return
	}
	data, _ := json.Marshal(resValue{"error": "", "response": response})
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(data))	
}

// ServeHTTP for MyApi
func (node *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		node.wrapperProfile(w, r)
	case "/user/create":
		node.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resValue{"error": "unknown method"})
		io.WriteString(w, string(data))
	}
}

// ...
// generated for type: OtherApi
// ...

// main.methodOptions{URL:"/user/create", Auth:true, Method:"POST"}
// [Wrapper for OtherApi] method: Create
func (node *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	// Authorization checker	
	authGood := "100500"
	auth := r.Header.Get("X-Auth")
	if auth != authGood {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resValue{"error": "unauthorized"})
		io.WriteString(w, string(data))
		return
	} 	

	// Method checker
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		data, _ := json.Marshal(resValue{"error": "bad method"})
		io.WriteString(w, string(data))	
		return
	}

	// validation OtherCreateParams
	r.ParseForm()
	paramUsername := r.Form.Get("username")
	// tplRequired
	if paramUsername == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "username must me not empty"})
		io.WriteString(w, string(data))
		return
	}

	// tplMin
	if len([]rune(paramUsername)) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "username len must be >= 3"})
		io.WriteString(w, string(data))
		return 
	}
	
	paramName := r.Form.Get("account_name")
	paramClass := r.Form.Get("class")
	// tplDefault	
	if paramClass == "" {
		paramClass = "warrior"
	}

	// tplEnum
	enumFlag := false	
	if paramClass == "warrior" {
		enumFlag = true
	}
	if paramClass == "sorcerer" {
		enumFlag = true
	}
	if paramClass == "rouge" {
		enumFlag = true
	}
	if !enumFlag {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "class must be one of [warrior, sorcerer, rouge]"})
		io.WriteString(w, string(data))
		return
	}

	paramLevel := r.Form.Get("level")
	// tplMin
	paramLevelIntMin, err := strconv.Atoi(paramLevel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "level must be int"})
		io.WriteString(w, string(data))
		return
	}
	if paramLevelIntMin < 1 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "level must be >= 1"})
		io.WriteString(w, string(data))
		return 
	}
	
	// tplMax
	paramLevelIntMax, err := strconv.Atoi(paramLevel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "level must be int"})
		io.WriteString(w, string(data))
		return
	}
	if paramLevelIntMax > 50 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "level must be <= 50"})
		io.WriteString(w, string(data))
		return 
	}
	
	paramLevelInt, _ := strconv.Atoi(paramLevel)
	params := OtherCreateParams{
		Username: paramUsername,
		Name: paramName,
		Class: paramClass,
		Level: paramLevelInt,
	}
	ctx := r.Context()
	response, err := node.Create(ctx, params)
	if err != nil {
		switch err.(type) {
		case ApiError:
			data, _ := json.Marshal(resValue{"error": err.(ApiError).Err.Error()})
			w.WriteHeader(err.(ApiError).HTTPStatus)
			io.WriteString(w, string(data))
		default:
			data, _ := json.Marshal(resValue{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, string(data))
		}
		return
	}
	data, _ := json.Marshal(resValue{"error": "", "response": response})
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(data))	
}

// ServeHTTP for OtherApi
func (node *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		node.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resValue{"error": "unknown method"})
		io.WriteString(w, string(data))
	}
}
