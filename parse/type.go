package parse

import (
	"errors"
	"fmt"
	"io"
	"log"
	"vector-ai/model"
	"vector-ai/parse/docx"
	"vector-ai/parse/epub"
	"vector-ai/parse/pdf"
	"vector-ai/parse/txt"
)

func ParseLocal(header model.CoreDocumentProps, data []byte) (string, error) {
	var parsedDoc string
	var err error

	mType := header.MimeType
	if mType == "text/plain" {
		parsedDoc = txt.ParseLocal(data)
	} else if mType == "application/epub+zip" {
		parsedDoc, err = epub.ParseLocal(data)
		check(err)
	} else if mType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		parsedDoc, err = docx.ParseLocal(data)
		check(err)
	} else if mType == "application/pdf" {
		parsedDoc, err = pdf.ParseLocal(data)
		check(err)
	} else if mType == "application/msword" {
		parsedDoc = ".DOC FILE"
		err = errors.New("convert to .docx to upload")
	} else {
		fmt.Println("contentType TYPE", mType)
		err = errors.New("not a supported file format")
	}

	return parsedDoc, err
}

func BodyText(body io.ReadCloser, fileSize int64, exportType string) (string, error) {
	var parsedDoc string
	var err error

	if exportType == "text/plain" {
		parsedDoc, err = txt.ParseBody(body, fileSize)
		check(err)
	} else if exportType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		parsedDoc, err = docx.ParseBody(body, fileSize)
		check(err)
	} else if exportType == "application/epub+zip" {
		parsedDoc, err = epub.ParseBody(body, fileSize)
		check(err)
	} else if exportType == "application/pdf" {
		parsedDoc, err = pdf.ParseBody(body, fileSize)
		check(err)
	} else {
		fmt.Println("Exported Type:", exportType)
		err = errors.New("not a supported file format")
	}

	return parsedDoc, err
}

func GoogleDriveExportType(mimeType string) string {
	// Google Doc application/vnd.google-apps.document
	if mimeType == "application/vnd.google-apps.document" {
		return "text/plain"
	} else {
		// .docx (DOCX) application/vnd.openxmlformats-officedocument.wordprocessingml.document
		// .zip (Web Page HTML)	application/zip
		// .epub (EPUB)	application/epub+zip
		// .odt (OpenDocument)	application/vnd.oasis.opendocument.text
		// .rtf (Rich Text)	application/rtf
		// .txt (Plain Text)	text/plain
		// .pdf (PDF)	application/pdf
		// fmt.Printf("Drive MimeType: %s \n", mimeType)
		return mimeType
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
