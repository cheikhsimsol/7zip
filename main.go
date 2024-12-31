package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.Use(
		CompressResponse(
			Compress7zip,
		),
	)
	// Route to access the PDF
	r.GET("/pdf", PDFHandler("./sample.pdf"))
	// ... let's assume there are 100s more like
	// the endpoint above.

	// Start the server
	r.Run(":8080")
}

type CompressFunc func(data []byte) ([]byte, error)

// Compress function that takes the data, saves it to a temp file, runs 7z, and returns the compressed file.
func Compress7zip(data []byte) ([]byte, error) {
	// Create a temporary file to store the input data.
	tmpFile, err := os.CreateTemp("", "input_*.tmp")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name()) // Ensure the temp file is deleted after use.

	// Write the data to the temporary file.
	_, err = tmpFile.Write(data)
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	// Create a temporary directory for the compressed file output.
	tmpDir := filepath.Join(os.TempDir(), "compress_output")
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir) // Cleanup the output directory.

	// Output file path.
	outputFile := filepath.Join(tmpDir, "output.7z")

	// Run the 7z command to compress the file.
	cmd := exec.Command("7z", "a", "-mx=9", outputFile, tmpFile.Name())

	err = cmd.Run()
	if err != nil {
		return nil, errors.New("failed to run 7z command: " + err.Error())
	}

	// Read the compressed file into memory.
	compressedData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, err
	}

	return compressedData, nil
}

func CompressResponse(cf CompressFunc) gin.HandlerFunc {
	return func(c *gin.Context) {

		buffer := bytes.Buffer{}
		compressWriter := &CompressResponseWriter{
			headers: http.Header{},
			buffer:  &buffer,
			status:  http.StatusOK,
		}

		originalWriter := c.Writer
		c.Writer = compressWriter

		// will return after all
		// subsequent middleware/handlers have executed.
		c.Next()

		result := buffer.Bytes()

		fmt.Println("length of data before compression:", len(result))

		compressed, err := cf(result)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fmt.Println("length of compressed:", len(compressed))

		c.Writer = originalWriter

		if compressWriter.status == http.StatusOK {
			c.Writer.Header().Set("Content-Type", "application/x-7z-compressed")
		}

		c.Writer.WriteHeader(compressWriter.status)
		c.Writer.Write(compressed)
	}
}

// PDFHandler serves a given PDF file for requests to its route
func PDFHandler(filePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open the PDF file
		file, err := os.Open(filePath)
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not open PDF file: %v", err)
			return
		}
		defer file.Close()

		// Set headers for PDF response
		c.Header("Content-Type", "application/pdf")

		// the file name should never be hard-coded.
		c.Header("Content-Disposition", "inline; filename=\"download.pdf\"")

		// Stream the file to the response
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not serve PDF file: %v", err)
			return
		}

		c.Status(http.StatusOK)
	}
}
