package epub

import (
	"archive/zip"
)

// Reader represents a readable epub file.
type Reader struct {
	Container
	files map[string]*zip.File
}

// ReadCloser represents a readable epub file that can be closed.
type ReadCloser struct {
	Reader
	// f *os.File
}

// Rootfile contains the location of a content.opf package file.
type Rootfile struct {
	FullPath string `xml:"full-path,attr"`
	Package
}

// Container serves as a directory of Rootfiles.
type Container struct {
	Rootfiles []*Rootfile `xml:"rootfiles>rootfile"`
}

// Package represents an epub content.opf file.
type Package struct {
	Metadata
	Manifest
	Spine
}

// Metadata contains publishing information about the epub.
type Metadata struct {
	Title       string `xml:"metadata>title"`
	Language    string `xml:"metadata>language"`
	Identifier  string `xml:"metadata>idenifier"`
	Creator     string `xml:"metadata>creator"`
	Contributor string `xml:"metadata>contributor"`
	Publisher   string `xml:"metadata>publisher"`
	Subject     string `xml:"metadata>subject"`
	Description string `xml:"metadata>description"`
	Event       []struct {
		Name string `xml:"event,attr"`
		Date string `xml:",innerxml"`
	} `xml:"metadata>date"`
	Type     string `xml:"metadata>type"`
	Format   string `xml:"metadata>format"`
	Source   string `xml:"metadata>source"`
	Relation string `xml:"metadata>relation"`
	Coverage string `xml:"metadata>coverage"`
	Rights   string `xml:"metadata>rights"`
}

// Manifest lists every file that is part of the epub.
type Manifest struct {
	Items []Item `xml:"manifest>item"`
}

// Item represents a file stored in the epub.
type Item struct {
	ID        string `xml:"id,attr"`
	HREF      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
	f         *zip.File
}

// Spine defines the reading order of the epub documents.
type Spine struct {
	Itemrefs []Itemref `xml:"spine>itemref"`
}

// Itemref points to an Item.
type Itemref struct {
	IDREF string `xml:"idref,attr"`
	*Item
}
