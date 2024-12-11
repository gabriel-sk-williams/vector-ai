package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
)

// Package epub provides basic support for reading EPUB archives.

// https://github.com/taylorskalyo/goreader/blob/master/epub/epub.go#L188
const containerPath = "META-INF/container.xml"

var (
	// ErrNoRootfile occurs when there are no rootfile entries found in
	// container.xml.
	ErrNoRootfile = errors.New("epub: no rootfile found in container")

	// ErrBadRootfile occurs when container.xml references a rootfile that does
	// not exist in the zip.
	ErrBadRootfile = errors.New("epub: container references non-existent rootfile")

	// ErrNoItemref occurrs when a content.opf contains a spine without any
	// itemref entries.
	ErrNoItemref = errors.New("epub: no itemrefs found in spine")

	// ErrBadItemref occurs when an itemref entry in content.opf references an
	// item that does not exist in the manifest.
	ErrBadItemref = errors.New("epub: itemref references non-existent item")

	// ErrBadManifest occurs when a manifest in content.opf references an item
	// that does not exist in the zip.
	ErrBadManifest = errors.New("epub: manifest references non-existent item")
)

func ParseLocal(data []byte) (string, error) {
	fileSize := int64(len(data))
	z, err := zip.NewReader(bytes.NewReader(data), fileSize)
	check(err)

	return ParseChapters(z, fileSize)
}

func ParseBody(body io.ReadCloser, fileSize int64) (string, error) {
	defer body.Close()
	buff, _ := io.ReadAll(body)

	z, err := zip.NewReader(bytes.NewReader(buff), fileSize)
	check(err)

	return ParseChapters(z, fileSize)
}

func ParseChapters(zipReader *zip.Reader, fileSize int64) (string, error) {
	var parsedDoc string
	var err error

	r, err := OpenEpub(zipReader)
	check(err)

	book := r.Rootfiles[0]

	fmt.Println(book.Title, fileSize)

	for _, item := range book.Spine.Itemrefs {
		rc, err := item.Open()
		check(err)

		cb, err := ParseText(rc, book.Manifest.Items)
		check(err)

		parsedDoc += cb.Text
	}

	return parsedDoc, err
}

// OpenReader will open the epub file specified by name and return a
// ReadCloser.
func OpenEpub(z *zip.Reader) (*Reader, error) {
	r := new(Reader)
	if err := r.init(z); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reader) init(z *zip.Reader) error {
	// Create a file lookup table
	r.files = make(map[string]*zip.File)
	for _, f := range z.File {
		r.files[f.Name] = f
	}

	err := r.setContainer()
	if err != nil {
		return err
	}
	err = r.setPackages()
	if err != nil {
		return err
	}
	err = r.setItems()
	if err != nil {
		return err
	}

	return nil
}

// setContainer unmarshals the epub's container.xml file.
func (r *Reader) setContainer() error {
	f, err := r.files[containerPath].Open()
	if err != nil {
		return err
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(b.Bytes(), &r.Container)
	if err != nil {
		return err
	}

	if len(r.Container.Rootfiles) < 1 {
		return ErrNoRootfile
	}

	return nil
}

// setPackages unmarshal's each of the epub's content.opf files.
func (r *Reader) setPackages() error {
	for _, rf := range r.Container.Rootfiles {
		if r.files[rf.FullPath] == nil {
			return ErrBadRootfile
		}

		f, err := r.files[rf.FullPath].Open()
		if err != nil {
			return err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, f)
		if err != nil {
			return err
		}

		err = xml.Unmarshal(b.Bytes(), &rf.Package)
		if err != nil {
			return err
		}
	}

	return nil
}

// setItems associates Itemrefs with their respective Item and Items with
// their zip.File.
func (r *Reader) setItems() error {
	itemrefCount := 0
	for _, rf := range r.Container.Rootfiles {
		itemMap := make(map[string]*Item)
		for i := range rf.Manifest.Items {
			item := &rf.Manifest.Items[i]
			itemMap[item.ID] = item

			abs := path.Join(path.Dir(rf.FullPath), item.HREF)
			item.f = r.files[abs]
		}

		for i := range rf.Spine.Itemrefs {
			itemref := &rf.Spine.Itemrefs[i]
			itemref.Item = itemMap[itemref.IDREF]
			if itemref.Item == nil {
				return ErrBadItemref
			}
		}
		itemrefCount += len(rf.Spine.Itemrefs)
	}

	if itemrefCount < 1 {
		return ErrNoItemref
	}

	return nil
}

// Open returns a ReadCloser that provides access to the Items's contents.
// Multiple items may be read concurrently.
func (item *Item) Open() (r io.ReadCloser, err error) {
	if item.f == nil {
		return nil, ErrBadManifest
	}

	return item.f.Open()
}

// NewReader returns a new Reader reading from ra, which is assumed to have the
// given size in bytes.
func NewReader(ra io.ReaderAt, size int64) (*Reader, error) {
	z, err := zip.NewReader(ra, size)
	if err != nil {
		return nil, err
	}

	r := new(Reader)
	if err = r.init(z); err != nil {
		return nil, err
	}

	return r, nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
