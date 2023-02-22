package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jadefox10200/goprint"
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

func printWithDefaulPrinter(file string) error {
	fmt.Println("=====================================")
	printerName, _ := goprint.GetDefaultPrinterName()
	fmt.Println("Default printer name: ", printerName)

	//open the printer
	printerHandle, err := goprint.GoOpenPrinter(printerName)
	if err != nil {
		log.Fatalln("Failed to open printer")
		return err
	}
	defer goprint.GoClosePrinter(printerHandle)

	filePath := file

	//Send to printer:
	err = goprint.GoPrint(printerHandle, filePath)
	if err != nil {
		log.Fatalln("during the func sendToPrinter, there was an error")
		return err
	}

	return nil
}

func sendPrintJobUsingIP(printerIP string, printerQueue string, data []byte) error {
	// Connect to the LPD server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:515", printerIP))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send the control file header
	fmt.Fprintf(conn, "\x02%s\x00%s\x00", printerQueue, "user")
	fmt.Fprintf(conn, "H%d\n", len(data))

	// Send the print job data
	writer := bufio.NewWriter(conn)
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	writer.Flush()

	// Send the end-of-file marker
	fmt.Fprint(conn, "\x00")

	// Read the response from the server
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		return err
	}

	// Check the response code for errors
	if response[0] != 0 {
		return fmt.Errorf("print job failed with error code %d", response[0])
	}

	return nil
}
