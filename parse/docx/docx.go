package docx

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

var (
	ErrTemplatePlaceholdersNotFound = errors.New("placeholders not found in template")
	ErrDelimitersNotPassed          = errors.New("delimiters did not passed")
	ErrBadTemplate                  = errors.New("can't open template file, it's broken")
	ErrPlaceholderWithWhitespaces   = errors.New("some placeholders template has leading or tailing whitespace")

	documentXmlPathInZip = "word/document.xml"
	// xmlTextTag           = "t"
)

func ParseLocal(data []byte) (string, error) {
	fileSize := int64(len(data))
	z, err := zip.NewReader(bytes.NewReader(data), fileSize)
	check(err)

	reader, err := z.Open(documentXmlPathInZip)
	check(err)

	texto := getAllXmlText(reader)

	return texto, nil
}

func ParseBody(body io.ReadCloser, fileSize int64) (string, error) {
	defer body.Close()
	buff, _ := io.ReadAll(body)

	z, err := zip.NewReader(bytes.NewReader(buff), fileSize)
	check(err)

	reader, err := z.Open(documentXmlPathInZip)
	check(err)

	texto := getAllXmlText(reader)

	return texto, nil
}

// Gets `word/document.xml` as string from given `docx` file (it's basically `zip`).
// Returns error if file is not a valid `docx`
func getAllXmlText(reader io.Reader) string {

	var output string
	tokenizer := html.NewTokenizer(reader)
	prevToken := tokenizer.Token()
loop:
	for {
		tok := tokenizer.Next()
		switch {
		case tok == html.ErrorToken:
			break loop // End of the document,  done
		case tok == html.StartTagToken: // <a>
			prevToken = tokenizer.Token()
			switch {
			case prevToken.Data == "w:br": // single newline
				output += "\n"
			case prevToken.Data == "w:p": // double newline
				output += "\n\n"
			}
			// fmt.Println(prevToken.String())
		case tok == html.TextToken:
			if prevToken.Data == "script" {
				continue
			}
			TxtContent := html.UnescapeString(string(tokenizer.Text()))
			if len(TxtContent) > 0 {
				output += TxtContent
			}
		}
	}
	return output
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		//log.Fatal(err)
	}
}
