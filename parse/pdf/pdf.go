package pdf

import (
	"bytes"
	"io"
	"log"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func ParseLocal(data []byte) (string, error) {

	pdfReader, err := model.NewPdfReader(bytes.NewReader(data))
	check(err)

	parsedDoc, err := parseWithPDFReader(pdfReader)

	return parsedDoc, err
}

func ParseBody(body io.ReadCloser, fileSize int64) (string, error) {

	defer body.Close()
	data, _ := io.ReadAll(body)
	pdfReader, err := model.NewPdfReader(bytes.NewReader(data))
	check(err)

	parsedDoc, err := parseWithPDFReader(pdfReader)

	return parsedDoc, err
}

func parseWithPDFReader(reader *model.PdfReader) (string, error) {
	var parsedDoc string

	numPages, err := reader.GetNumPages()
	check(err)

	// fmt.Printf("--------------------\n")
	// fmt.Printf("PDF to text extraction:\n")
	// fmt.Printf("--------------------\n")

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := reader.GetPage(pageNum)
		check(err)

		ex, err := extractor.New(page)
		check(err)

		text, err := ex.ExtractText()
		check(err)

		parsedDoc += text

		// fmt.Println("------------------------------")
		// fmt.Printf("Page %d:\n", pageNum)
		// fmt.Printf("\"%s\"\n", text)
		// fmt.Println("------------------------------")

	}

	return parsedDoc, nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
