package wrappers

import (
	"bufio"
	"fmt"
	"github.com/olegpolukhin/go-burning-files/schema"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/otiai10/gosseract"
)

// ExtractPdfToImagesFromPDF extracts Images from the PDF file and output an image per page.
func ExtractPdfToImagesFromPDF(pdfFullPath, outputDirectory string) (*string, error) {
	if err := os.Chdir(outputDirectory); err != nil {
		return nil, err
	}

	cmdArgs := []string{"-dNOPAUSE", "-dBATCH", "-sDEVICE=jpeg", "-r300", "-sOutputFile=p%03d.jpg", pdfFullPath}

	cmd := exec.Command("gs", cmdArgs...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating StdoutPipe for Cmd error %w", err)
	}

	var textDoc string

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			textDoc = scanner.Text()
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting Cmd error %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("waiting for Cmd error %w", err)
	}

	return &textDoc, nil
}

//ExtractPlainTextFromImage given a images file, Tesseract OCR generates a plain text file with the detected text.
func ExtractPlainTextFromImage(imageFullPath, outputDirectory, textFilePrefix string, languages []string, wg *sync.WaitGroup, throttle chan int) error {
	defer wg.Done()

	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetTessdataPrefix("/usr/local/share/tessdata/"); err != nil {
		return fmt.Errorf("error SetTessdataPrefix %w", err)
	}

	if err := client.SetLanguage(languages[0]); err != nil {
		return fmt.Errorf("error SetLanguage %w", err)
	}

	if err := client.SetImage(imageFullPath); err != nil {
		return fmt.Errorf("error SetImage %w", err)
	}

	client.Trim = true

	text, _ := client.Text()

	textFilePath := filepath.Join(outputDirectory, fmt.Sprintf("%s_%s", textFilePrefix, schema.TextFileName))

	outfile, err := os.Create(textFilePath)
	if err != nil {
		return fmt.Errorf("error creating text file %w", err)
	}

	defer outfile.Close()

	sanitizedTxt := strings.Replace(text, "\n", " ", -1)

	if _, err := outfile.WriteString(sanitizedTxt); err != nil {
		return fmt.Errorf("error WriteString %w", err)
	}

	<-throttle

	return nil
}
