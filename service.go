package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/jadefox10200/goprint"
)

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
