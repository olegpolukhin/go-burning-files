go-burning-files
========================

Pet project

This Golang based project provides a service convert PDF's and Images to Text, 
using Tesseract OCR scanner.

----

### How to use:

#### Install dependency

Install tesseract-ocr: https://github.com/tesseract-ocr/tesseract/wiki

sudo apt install tesseract-ocr

sudo apt install libtesseract-dev

go get -t github.com/otiai10/gosseract

---

#### Init convert image to text

```
    handlers.NewBurning(uploadDir)
    
    item, err := h.BurningImage(pathToFileImage)
    if err != nil {
        t.Errorf("BurningImage Error: %v", err)
    }
```
