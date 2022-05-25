package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/jasonlvhit/gocron"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

type Datos struct {
	ID      int    `gorm:"primaryKey;autoIncrement"`
	DataID  int    `json:"DataID"`
	Votes   string `json:"Votes"`
	Answers string `json:"Answers"`
	Views   string `json:"Views"`
	Title   string `json:"Title"`
	Author  string `json:"Author"`
	Content string `json:"Content"`
	Time    string `json:"Time"`
}

var erro = godotenv.Load()

var datas []Datos

//Connect DB
var DBUrl = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_ROOT_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE"))
var db, err = gorm.Open("mysql", DBUrl)

func main() {
	if erro != nil {
		log.Fatal("Error loading .env file")
	}

	//Scheduler crawl data
	s := gocron.NewScheduler()
	s.Every(10).Seconds().Do(crawl)
	s.Start()

	//router
	router := gin.Default()
	router.POST("/api/v1/data/create", createData)
	router.GET("/api/v1/data", fetchAllData)        //[GET] http://localhost:8080/api/v1/data
	router.GET("/api/v1/data/:id", fetchDataSingle) //[GET] http://localhost:8080/api/v1/data/61593377
	router.PUT("/api/v1/data/:id", updateData)      //[PUT] http://localhost:8080/api/v1/data/61593377
	router.DELETE("/api/v1/data/:id", deteleData)   //[POST] http://localhost:8080/api/v1/data/61593377
	router.Run()
}

func crawl() {
	if err != nil {
		panic("failed to connect database") // Check connect to database
	}
	fmt.Println("Connect Database!!!")

	if (!db.HasTable(&Datos{})) {
		db.CreateTable(&Datos{})
		return
	}

	//crawl Data
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnHTML("div.s-post-summary", func(e *colly.HTMLElement) {
		data := Datos{}
		idStr := e.Attr("data-post-id")
		voteStr := e.ChildText("div.js-post-summary-stats > div:nth-child(1) > span.s-post-summary--stats-item-number")
		answerStr := e.ChildText("div.js-post-summary-stats > div:nth-child(2) > span.s-post-summary--stats-item-number")
		viewStr := e.ChildText("div.js-post-summary-stats > div:nth-child(3) > span.s-post-summary--stats-item-number")
		titleStr := e.ChildText(".s-post-summary--content-title > .s-link")
		authorStr := e.ChildText(".s-user-card--info > .s-user-card--link > .flex--item")
		contentStr := e.ChildText("div.s-post-summary--content > .s-post-summary--content-excerpt")
		timeStr := e.ChildText("div.s-user-card__minimal > .s-user-card--time > span.relativetime")
		data.DataID, _ = strconv.Atoi(idStr)
		data.Votes = voteStr
		data.Answers = answerStr
		data.Views = viewStr
		data.Title = titleStr
		data.Author = authorStr
		data.Time = timeStr
		data.Content = contentStr

		//Check data in database if data exist => update data, else => insert new data
		info := Datos{}
		db.Where("data_id = ?", data.DataID).First(&info)
		fmt.Println(data.DataID)
		fmt.Println(info.DataID)
		if info.DataID == 0 {
			fmt.Println("Create: ", data)
			db.Create(&data)
		} else {
			if data == info {
				fmt.Println("Don't update ", data)
				return
			} else {
				fmt.Println("Update ", info)
				info.DataID = data.DataID
				info.Votes = data.Votes
				info.Answers = data.Answers
				info.Views = data.Views
				info.Title = data.Title
				info.Author = data.Author
				info.Content = data.Content
				info.Time = data.Time
				db.Save(&info)
			}

		}
		datas = append(datas, data) // insert to datas

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished ", r.Request.URL)
	})

	//param("pagesize")
	pagesize, _ := strconv.Atoi("30")
	if pagesize != 15 && pagesize != 30 && pagesize != 50 {
		pagesize = 30
	}

	for i := 1; i < 4; i++ {
		fullURL := fmt.Sprintf("https://stackoverflow.com/questions/tagged/ibm-blockchain?tab=newest&page=%d&pagesize=%d", i, pagesize)
		c.Visit(fullURL)
	}

	//Print to json
	file, _ := json.MarshalIndent(datas, "", "")
	_ = ioutil.WriteFile("data.json", file, 0644)

}

func createData(c *gin.Context) {
	dataid, _ := strconv.Atoi(c.PostForm("DataID"))
	data := Datos{
		DataID:  dataid,
		Votes:   c.PostForm("votes"),
		Answers: c.PostForm("answers"),
		Views:   c.PostForm("views"),
		Title:   c.PostForm("title"),
		Author:  c.PostForm("author"),
		Content: c.PostForm("content"),
		Time:    c.PostForm("time"),
	}
	db.Save(&data)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": data})
}

func fetchAllData(c *gin.Context) {
	var data []Datos
	db.Find(&data)
	if len(data) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Not data found!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": data})
}

func fetchDataSingle(c *gin.Context) {
	var data Datos
	dataId := c.Param("id")
	db.Where("data_id = ?", dataId).First(&data)
	db.First(&data, dataId)
	if data.DataID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Not data found!"})
	}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": data})

}

func updateData(c *gin.Context) {
	var data Datos
	dataId := c.Param("id")
	db.Where("data_id = ?", dataId).First(&data)
	if data.DataID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Not data found!"})
		return
	}
	db.Model(&data).Update("votes", c.PostForm("votes"))
	db.Model(&data).Update("answers", c.PostForm("answers"))
	db.Model(&data).Update("views", c.PostForm("views"))
	db.Model(&data).Update("title", c.PostForm("title"))
	db.Model(&data).Update("author", c.PostForm("author"))
	db.Model(&data).Update("content", c.PostForm("content"))
	db.Model(&data).Update("time", c.PostForm("time"))
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": data})
}

func deteleData(c *gin.Context) {
	var data Datos
	dataId := c.Param("id")
	db.Where("data_id = ?", dataId).First(&data)
	if data.DataID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Not data found!"})
		return
	}
	db.Delete(&data)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": data})
}
