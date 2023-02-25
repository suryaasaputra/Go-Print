package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/print", routeSubmitPost)
	r.GET("/", routeIndexGet)

	if err := r.Run(":9000"); err != nil {
		fmt.Println(err)
	}
}

func routeSubmitPost(c *gin.Context) {
	pdfData := c.PostForm("pdf")
	if pdfData == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No PDF data found"})
		return
	}

	// Decode the base64-encoded PDF data
	pdfBytes, err := base64.StdEncoding.DecodeString(string(pdfData))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// err = sendPrintJobUsingIP("0.0.0.0", "tes", pdfBytes)
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
	// 		"message": err.Error(),
	// 	})
	// 	return
	// }

	tempFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println(err)
		return
	}
	// create unix timestamp and convert to string
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	fileLocation := fmt.Sprintf("files/%s.pdf", timestamp)
	err = os.WriteFile(fileLocation, pdfBytes, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer tempFile.Close()
	tempFile.Write(pdfBytes)

	err = printWithDefaulPrinter(tempFile.Name())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	fmt.Println("Print Sukses")
	fmt.Println("=====================================")

	c.JSON(http.StatusOK, gin.H{
		"message": "Sukses Cetak Kutipan",
	})
}

func routeIndexGet(c *gin.Context) {

	var tmpl = template.Must(template.ParseFiles("view.html"))
	var err = tmpl.Execute(c.Writer, nil)

	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}
