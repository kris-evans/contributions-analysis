# Harris County Contributions Analyzer

## Getting Started

### Prequisites

This is a list of tools that need to be installed before you can run this CLI client. Tesseract OCR and PDFCPU all have dependenices that
have to be installed before the program will function.

* [Golang](https://go.dev/doc/install)
* [Tesseract OCR](https://github.com/tesseract-ocr/tessdoc)
* [PDFCPU](https://github.com/pdfcpu/pdfcpu)

## Commands

### Extract Contribution Images

Takes a Harris County Contributions document and extracts the embedded images from the PDF file. 

```sh
go run ./contributions-analyzer.go extract-contribution-images --pdf-file-path ./example-contributions.pdf --image-output-path ./tmp
```

### Parse Contribution Image

```
go run ./contributions-analyzer.go parse-image --image-file-path ./tmp/example-contributions_006_I0.png
```