package main

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func getCountry(code int) (country string, matchRegex string, errr error) {
	countryCode := map[int][2]string{
		237: {"Cameroon", "\\(237\\)\\ ?[2368]\\d{7,8}$"},
		251: {"Ethiopia", "\\(251\\)\\ ?[1-59]\\d{8}$"},
		212: {"Morocco", " \\(212\\)\\ ?[5-9]\\d{8}$"},
		258: {"Mozambique", "\\(258\\)\\ ?[28]\\d{7,8}$"},
		256: {"Uganda", "\\(256\\)\\ ?\\d{9}$"},
	}
	arr := countryCode[code]
	if len(arr) == 0 {
		return "", "", errors.New("Non-existent country")
	}
	country = countryCode[code][0]
	matchRegex = countryCode[code][1]
	return country, matchRegex, nil
}

type customer struct {
	// order alphabetically
	Country string `json:"country"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Valid   string `json:"valid"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func getData(db *sql.DB, filterCountry, filterValid, page_id, page_size string) []customer {
	query := "SELECT name, phone FROM customer"
	if page_size != "" {
		query = query + " limit "
		if page_id != "" {
			query = query + page_id + ", " + page_size
		} else {
			query = query + page_size
		}
	}
	rows, err := db.Query(query)
	checkErr(err)
	defer rows.Close()

	checkErr(rows.Err())

	customers := make([]customer, 0)

	for rows.Next() {
		newCustomer := customer{}
		err = rows.Scan(&newCustomer.Name, &newCustomer.Phone)
		checkErr(err)
		number := newCustomer.Phone
		code, err := strconv.Atoi(number[1:4])
		checkErr(err)
		country, matchRegex, err := getCountry(code)
		if err != nil {
			continue
		}
		if filterCountry != "" && country != filterCountry {
			continue
		}
		valid, err := regexp.MatchString(matchRegex, number)
		if filterValid != "" && filterValid != "any" {
			if filterValid == "valid" {
				if !valid {
					continue
				}
			} else if valid {
				continue
			}
		}
		checkErr(err)
		newCustomer.Country = country
		newCustomer.Valid = strconv.FormatBool(valid)
		customers = append(customers, newCustomer)
	}

	err = rows.Err()
	checkErr(err)

	return customers
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./sample.db")
	checkErr(err)
	return db
}
func setupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()
	router.Use(CORSMiddleware())
	router.GET("/customers/", func(c *gin.Context) {
		country := c.Query("country")
		valid := c.Query("valid")
		page_id := c.Query("page_id")
		page_size := c.Query("page_size")
		jsonList := getData(db, country, valid, page_id, page_size)
		c.JSON(200, jsonList)
	})

	return router
}

func main() {
	db := setupDB()
	router := setupRouter(db)
	defer db.Close()
	router.Run()
}
