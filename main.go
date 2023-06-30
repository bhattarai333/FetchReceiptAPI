package main

import (
	"github.com/gin-gonic/gin"       //go get github.com/gin-gonic/gin
	"github.com/lithammer/shortuuid" // go get github.com/lithammer/shortuuid
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var receiptsMap = make(map[string]receipt)

type receipt struct {
	ID     string `json:"id"`
	Points int64  `json:"points"`
}

type requestData struct {
	Retailer     string     `json:"retailer"`
	PurchaseDate string     `json:"purchaseDate"`
	PurchaseTime string     `json:"purchaseTime"`
	Total        string     `json:"total"`
	Items        []itemData `json:"items"`
}

type itemData struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

func generateID() string {
	return shortuuid.New()
}

func calculatePoints(c *gin.Context) int64 {
	var data requestData
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Println("Invalid JSON payload:", err)
		return 0
	}

	var totalPoints = 0

	//Calculate points for alphanumeric characters in retailer name
	for _, char := range data.Retailer {
		if unicode.IsDigit(char) || unicode.IsLetter(char) {
			totalPoints += 1
		}
	}

	//Calculate points for if total is round dollar number
	totalFloat, _ := strconv.ParseFloat(data.Total, 64)
	if totalFloat == math.Floor(totalFloat) {
		totalPoints += 50
	}

	//Calculate points for if total is multiple of 0.25
	if math.Mod(totalFloat, 0.25) == 0 {
		totalPoints += 25
	}

	//Calculate points for if item name's length is a multiple of 3
	for _, item := range data.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			priceFloat, _ := strconv.ParseFloat(item.Price, 64)
			totalPoints += int(math.Ceil(0.2 * priceFloat))
		}
	}

	//Calculate points for number of items bought
	totalPoints += (len(data.Items) / 2) * 5

	//Calculate points for if day in date of purchase is odd
	t, _ := time.Parse("2006-01-02", data.PurchaseDate)

	if t.Day()%2 != 0 {
		totalPoints += 6
	}

	// Calculate points for if time of purchase is between 2:00 PM and 4:00 PM
	t, _ = time.Parse("15:04", data.PurchaseTime)

	if t.Hour() >= 14 && t.Hour() < 16 {
		totalPoints += 10
	}
	return int64(totalPoints)
}

func postReceipt(c *gin.Context) {
	var newReceipt receipt

	newReceipt.ID = generateID()
	newReceipt.Points = calculatePoints(c)

	receiptsMap[newReceipt.ID] = newReceipt                //add new receipt to receipt dict
	c.JSON(http.StatusCreated, gin.H{"id": newReceipt.ID}) //return id
}

func getReceipt(c *gin.Context) {
	id := c.Param("id")
	var receipt = receiptsMap[id]                          //get receipt from receipt dict
	c.JSON(http.StatusOK, gin.H{"points": receipt.Points}) //return points
}

func main() {

	router := gin.Default()
	router.GET("/receipts/:id/points", getReceipt)
	router.POST("/receipts/process", postReceipt)

	router.Run("localhost:8080")
}
