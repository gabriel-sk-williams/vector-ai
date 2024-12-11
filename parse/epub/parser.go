package epub

import (
	"io"

	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/net/html"
)

// "github.com/taylorskalyo/goreader/epub"

type parser struct {
	// tagStack  []atom.Atom
	tokenizer *html.Tokenizer
	doc       Cellbuf
	items     []Item // []epub.Item
}

type Cellbuf struct {
	Text string
}

// takes in html content via an io.Reader and
// returns a buffer containing only plain text.
func ParseText(r io.Reader, items []Item) (Cellbuf, error) {
	tokenizer := html.NewTokenizer(r)
	doc := Cellbuf{Text: ""}
	p := parser{tokenizer: tokenizer, doc: doc, items: items}
	err := p.parse(r)
	if err != nil {
		return p.doc, err
	}
	return p.doc, nil
}

// parse walks an html document and renders elements to a cell buffer document.
func (p *parser) parse(io.Reader) (err error) {
	for {
		tokenType := p.tokenizer.Next()
		token := p.tokenizer.Token()

		switch tokenType {
		case html.ErrorToken:
			err = p.tokenizer.Err()
		case html.TextToken:
			p.doc.appendText(string(token.Data))
		}

		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// appendText appends text to the cell buffer document.
func (c *Cellbuf) appendText(str string) {
	if len(str) <= 0 {
		return
	}
	c.Text += str
}
