package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const endpoint string = "/customers/"

var ahmad = map[string]string{"id": "", "name": "Ahmad Abdallah", "country": "Cameroon", "phone": "(237) 25557800", "valid": "true"}
var boat = map[string]string{"id": "", "name": "Boaty McBoatFace", "country": "Cameroon", "phone": "(237) 95552372", "valid": "false"}

func TestGetSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	router := setupRouter(db)
	mapExpected := [2]map[string]string{ahmad, boat}
	expected, err := json.Marshal(mapExpected)
	if err != nil {
		fmt.Println(err)
	}
	rows := sqlmock.NewRows([]string{"name", "phone"}).
		AddRow(ahmad["name"], ahmad["phone"]).
		AddRow(boat["name"], boat["phone"])

	mock.ExpectQuery("SELECT name, phone FROM customer").WillReturnRows(rows)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", endpoint, nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(expected), w.Body.String())
}

func TestNonExistentCountry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	router := setupRouter(db)
	expected := "[]"
	rows := sqlmock.NewRows([]string{"name", "phone"}).
		AddRow(ahmad["name"], ahmad["phone"]).
		AddRow(boat["name"], boat["phone"])

	mock.ExpectQuery("SELECT name, phone FROM customer").WillReturnRows(rows)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", endpoint+"?country=Narnia", nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expected, w.Body.String())
}

func TestPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	router := setupRouter(db)
	mapExpected := []map[string]string{}
	for i := 0; i < 5; i++ {
		// would be a better idea to replace this with a fake data generator in the future for testing
		newEntry := map[string]string{
			"id": "", "name": "Boaty McBoatFace", "country": "Cameroon", "phone": "(237) 95552372", "valid": "false",
		}
		mapExpected = append(mapExpected, newEntry)
	}
	expected, err := json.Marshal(mapExpected)
	if err != nil {
		fmt.Println(err)
	}

	rows := sqlmock.NewRows([]string{"name", "phone"})
	for i := 0; i < 5; i++ {
		rows.AddRow(boat["name"], boat["phone"])
	}

	mock.ExpectQuery("SELECT name, phone FROM customer limit 5").WillReturnRows(rows)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", endpoint+"?page_size=5", nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(expected), w.Body.String())
}

func TestValid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	router := setupRouter(db)
	mapExpected := [1]map[string]string{
		ahmad,
	}
	expected, err := json.Marshal(mapExpected)
	if err != nil {
		fmt.Println(err)
	}
	rows := sqlmock.NewRows([]string{"name", "phone"}).
		AddRow(ahmad["name"], ahmad["phone"]).
		AddRow(boat["name"], boat["phone"])

	mock.ExpectQuery("SELECT name, phone FROM customer").WillReturnRows(rows)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", endpoint+"?valid=valid", nil)
	router.ServeHTTP(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, string(expected), w.Body.String())
}
