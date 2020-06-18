package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/olegpolukhin/go-burning-files/schema"
	"github.com/olegpolukhin/go-burning-files/wrappers"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const (
	NumberParallelRoutines = 4
)

type App struct {
	throttle  chan int
	uploadDir string
}

func NewBurning(uploadDir string) *App {
	return &App{
		throttle:  make(chan int, NumberParallelRoutines),
		uploadDir: uploadDir,
	}
}

func (h *App) BurningImage(pathFile string) (*schema.SubmissionDetails, error) {
	var (
		err                     error
		submission              schema.SubmissionDetails
		tempPath, txtOutputPath string
		outfile                 *os.File
	)

	submission.FileName = pathFile

	var infile multipart.File

	if infile, err = os.Open(pathFile); err != nil {
		return nil, fmt.Errorf("error open image file %w", err)
	}

	// Save the file into the docker container disk,
	generatedUUID := uuid.New()

	submission.UUID = generatedUUID.String()

	tempPath = path.Join(h.uploadDir, generatedUUID.String())

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write temporary folder %w", err)
	}

	outfile, err = os.Create(filepath.Join(tempPath, schema.DocumentImageName))
	if err != nil {
		return nil, fmt.Errorf("creating temporary handlers file error %w", err)
	}

	defer outfile.Close()

	// 32K buffer copy
	if _, err = io.Copy(outfile, infile); err != nil {
		return nil, fmt.Errorf("error while copying file %w", err)
	}

	txtOutputPath = path.Join(tempPath, schema.TextFolderName)
	if err := os.MkdirAll(txtOutputPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write text output folder error %w", err)
	}

	var wg sync.WaitGroup

	h.processParallelOCR(tempPath, "jpg", txtOutputPath, &wg)

	submission.Path = txtOutputPath
	submission.Pages, err = h.generateDetails(txtOutputPath)
	if err != nil {
		return nil, fmt.Errorf("generateDetails error %w", err)
	}

	return &submission, nil
}

func (h *App) BurningPDF(pathFile string) (*schema.SubmissionDetails, error) {
	var (
		err                     error
		submission              schema.SubmissionDetails
		tempPath, txtOutputPath string
		infile                  multipart.File
	)

	submission.FileName = pathFile

	if infile, err = os.Open(pathFile); err != nil {
		return nil, fmt.Errorf("error open image file %w", err)
	}

	// open destination
	var outfile *os.File

	generatedUUID := uuid.New()
	submission.UUID = generatedUUID.String()

	tempPath = path.Join(h.uploadDir, generatedUUID.String())

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating temporary directory error %w", err)
	}

	if outfile, err = os.Create(filepath.Join(tempPath, schema.DocumentFileName)); nil != err {
		return nil, fmt.Errorf("creating temporary handlers file error %w", err)
	}

	defer outfile.Close()

	// 32K buffer copy
	if _, err = io.Copy(outfile, infile); nil != err {
		return nil, fmt.Errorf("while copying file error %w", err)
	}

	// Generates Images from the PDF
	imagesOutputPath := path.Join(tempPath, schema.ImagesFolderName)
	if err := os.MkdirAll(imagesOutputPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating images output directory error %w", err)
	}

	pdfFilePath := path.Join(tempPath, schema.DocumentFileName)

	_, err = wrappers.ExtractPdfToImagesFromPDF(pdfFilePath, imagesOutputPath)
	if err != nil {
		return nil, fmt.Errorf("unable to extract images from PDF error %w", err)
	}

	var wg sync.WaitGroup

	txtOutputPath = path.Join(tempPath, schema.TextFolderName)

	if err := os.MkdirAll(txtOutputPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating texts output directory error %w", err)
	}

	h.processParallelOCR(imagesOutputPath, "jpg", txtOutputPath, &wg)

	submission.Path = txtOutputPath
	submission.Pages, err = h.generateDetails(txtOutputPath)
	if err != nil {
		return nil, fmt.Errorf("generateDetails error %w", err)
	}

	return &submission, err
}

func (h *App) processParallelOCR(imagesDirectoryPath string, imageExtension string, textOutPutDirectory string, wg *sync.WaitGroup) error {
	var err error

	imageFilesList, _ := ioutil.ReadDir(imagesDirectoryPath)

	for i := range imageFilesList {
		if !strings.HasSuffix(imageFilesList[i].Name(), imageExtension) || imageFilesList[i].IsDir() {
			continue
		}

		imagePath := path.Join(imagesDirectoryPath, imageFilesList[i].Name())

		h.throttle <- 1

		wg.Add(1)

		go func(name string) {
			err = wrappers.ExtractPlainTextFromImage(imagePath, textOutPutDirectory, name, []string{"rus+eng"}, wg, h.throttle)
		}(imageFilesList[i].Name())
	}

	wg.Wait()

	return err
}

func (h *App) generateDetails(textsDirectory string) ([]schema.Details, error) {
	txtFilesList, _ := ioutil.ReadDir(textsDirectory)

	pages := make([]schema.Details, len(txtFilesList))

	pageNumber := 0

	for i := range txtFilesList {
		txtPath := path.Join(textsDirectory, txtFilesList[i].Name())

		data, err := ioutil.ReadFile(txtPath)
		if err != nil {
			return nil, fmt.Errorf("error cannot read txt file %w", err)
		}

		pages[pageNumber] = schema.Details{
			Text: string(data),
		}
		pageNumber++
	}

	return pages, nil
}
