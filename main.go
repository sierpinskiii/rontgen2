package main

import (
	"os"
	"os/exec"

	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

type Lecture struct {
	gorm.Model
	Title        string
	UUID         string
	DriveLink    string
	Dir          string
	Filename     string
	Length       int
	CurrentSlide int
	CreatedAt    time.Time
}

func intToString(n int) string {
	return strconv.Itoa(n)
}

var siteTitle = "RONTGEN2::REALTIME PRESENTATION SYSTEM"

var uploadPrefix = "./lectures/"

var UUID = "default-lecture-dir"

// var PORT = 8080

var users = map[string]string{
	"admin": "password123",
}

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	r := gin.Default()

	// Setup session middleware
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("tmpl/*")

	r.Static("/static", "./static")
	r.Static("/lectures", "./lectures")
	// r.StaticFS("/more_static", http.Dir("my_file_system"))
	// r.StaticFile("/favicon.ico", "./resources/favicon.ico")

	db, err := gorm.Open(sqlite.Open("lectures.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Lecture{})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"siteTitle": siteTitle,
		})
	})

	r.POST("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		username := c.PostForm("username")
		password := c.PostForm("password")

		expectedPassword, ok := users[username]
		if !ok || expectedPassword != password {
			c.JSON(401, gin.H{"status": "unauthorized"})
			return
		}

		session.Set("user", username)
		session.Save()

		// c.JSON(200, gin.H{"status": "you are logged in"})
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"siteTitle": siteTitle,
			"port":      PORT,
			"addr":      "localhost",
			"message":   "Logged in successfully",
		})
	})

	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			// c.JSON(401, gin.H{"status": "unauthorized"}) // Not logged in
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"siteTitle": siteTitle,
				"port":      PORT,
				"addr":      "localhost",
			})
			return
		}

		// Delete the user session (logout)
		session.Delete("user")
		session.Save()

		// c.JSON(200, gin.H{"status": "logged out"})
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"siteTitle": siteTitle,
			"port":      PORT,
			"addr":      "localhost",
			"alert":     "Logged out successfully",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"siteTitle": siteTitle,
			"port":      PORT,
			"addr":      "localhost",
		})
	})

	r.GET("/screen", func(c *gin.Context) {
		var lecture Lecture
		db.First(&lecture)
		slideFileName := "/lectures/" + UUID + "/slide-" + intToString(lecture.CurrentSlide) + ".png"
		c.HTML(http.StatusOK, "screen.tmpl", gin.H{
			"siteTitle":     siteTitle,
			"slideFileName": slideFileName,
		})
	})

	/*
		r.GET("/next", func(c *gin.Context) {
			var lecture Lecture
			db.First(&lecture)
			lecture.CurrentSlide++ // % lecture.Length
			db.Save(&lecture)

			// db.Save(&lecture)
			// Save the updated lecture with error handling
			if err := db.Save(&lecture).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// c.Redirect(http.StatusSeeOther, "/present/lecturer")

		})

		r.GET("/prev", func(c *gin.Context) {
			var lecture Lecture
			db.First(&lecture)
			if lecture.CurrentSlide == 0 {
				lecture.CurrentSlide = lecture.Length - 1
			} else {
				lecture.CurrentSlide--
			}
			db.Save(&lecture)

			// c.Redirect(http.StatusMovedPermanently, "/present/lecturer")
		})
	*/

	r.GET("/student", func(c *gin.Context) {
		var lectures []Lecture
		db.Find(&lectures)
		c.HTML(http.StatusOK, "student.tmpl", gin.H{
			"lectures":  lectures,
			"siteTitle": siteTitle,
		})
	})

	r.GET("/present/lecturer", authRequired(), func(c *gin.Context) {
		var lecture Lecture
		db.First(&lecture)

		nextStr := c.DefaultQuery("next", "false")
		prevStr := c.DefaultQuery("prev", "false")
		lectureUUID := c.DefaultQuery("lecture", "")

		// Convert string values to boolean
		next, err0 := strconv.ParseBool(nextStr)
		prev, err1 := strconv.ParseBool(prevStr)

		if err0 != nil || err1 != nil {
			c.JSON(400, gin.H{
				"error": "Invalid query parameters",
			})
			return
		}

		// Handle based on the boolean values of 'next' and 'prev'
		if next {
			lecture.CurrentSlide = (lecture.CurrentSlide + 1) % (lecture.Length + 1)
		} else if prev {
			if lecture.CurrentSlide == 0 {
				lecture.CurrentSlide = lecture.Length - 1
			} else {
				lecture.CurrentSlide--
			}
		} else {
			//do nothing
		}

		db.Save(&lecture)

		if lectureUUID != "" {
			UUID = lectureUUID
		}

		var lectures []Lecture
		db.Find(&lectures)
		c.HTML(http.StatusOK, "lecturer.tmpl", gin.H{
			"lectures":  lectures,
			"siteTitle": siteTitle,
		})
	})

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/upload", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File is required",
			})
			return
		}

		title, _ := c.GetPostForm("title")
		driveLink, _ := c.GetPostForm("link")
		log.Println(title)
		log.Println(file.Filename)

		lecUUID := uuid.New().String()

		lecDirPath := uploadPrefix + lecUUID + "/"
		if _, err := os.Stat(lecDirPath); os.IsNotExist(err) {
			err := os.Mkdir(lecDirPath, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}

		// Upload the file to specific dst.
		c.SaveUploadedFile(file, lecDirPath+"original.pdf")

		cmd := exec.Command(
			"nice", "-20",
			"convert", "-density", "300",
			lecDirPath+"original.pdf",
			"-resize", "25%",
			lecDirPath+"slide.png")
		err = cmd.Run()
		if err != nil {
			log.Println(err)
		}

		length, err := api.PageCountFile(lecDirPath + "original.pdf")
		if err != nil {
			log.Println(err)
		}

		db.Create(&Lecture{
			Title:        title,
			UUID:         lecUUID,
			Dir:          lecDirPath,
			Filename:     lecDirPath + file.Filename,
			DriveLink:    driveLink,
			Length:       length,
			CurrentSlide: 0})

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})

	// Form page for uploading article
	r.GET("/upload", authRequired(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.tmpl", gin.H{
			"title": "Upload a New Lecture",
		})
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":" + PORT)
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.JSON(401, gin.H{"status": "unauthorized"})
			c.Abort()
			return
		}
		// If user is found, pass to the next middleware
		c.Next()
	}
}
