package tests

import (
	"github.com/olegpolukhin/go-burning-files/handlers"
	"log"
	"testing"
)

const (
	uploadDir       = "/path/uploaded-test"
	pathToFileImage = "/path/1.png"
	pathToFilePDF   = "/path/1.pdf"
)

func TestUploadImage(t *testing.T) {
	h := handlers.NewBurning(uploadDir)

	item, err := h.BurningImage(pathToFileImage)
	if err != nil {
		t.Errorf("BurningImage Error: %v", err)
	}

	if len(item.Pages) == 0 {
		t.Errorf("BurningImage Error: Items empty.")
	}

	log.Println(item.Pages)
}

func TestUploadPDF(t *testing.T) {
	h := handlers.NewBurning(uploadDir)

	item, err := h.BurningPDF(pathToFilePDF)
	if err != nil {
		t.Errorf("BurningPDF Error: %v", err)
	}

	if len(item.Pages) == 0 {
		t.Errorf("BurningPDF Error: Items empty.")
	}

	log.Println(item.Pages)
}
