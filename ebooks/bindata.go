// Code generated by go-bindata.
// sources:
// content.opf
// cover.html
// part.html
// story.css
// toc.ncx
// DO NOT EDIT!

package ebooks

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data, name string) ([]byte, error) {
	gz, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _contentOpf = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x84\x92\x51\x8b\xdc\x20\x14\x85\xdf\xf7\x57\x04\xdf\x37\xce\xee\x4b\x21\x24\x81\x29\x4b\x61\x1f\x0a\x85\xe9\x3c\x17\x47\x6f\x26\x97\x31\x6a\xf5\xda\x29\x84\xfc\xf7\xa2\x71\x26\x1d\x68\xd9\x37\x3d\xe7\xdc\x4f\x3d\xd8\x3a\x21\x2f\xe2\x0c\xd5\xef\x49\x9b\xd0\xb1\x91\xc8\x35\x9c\x5f\xaf\xd7\x1a\x95\x1b\x6a\xeb\xcf\xfc\x75\xb7\xfb\xc4\xad\x1b\x58\x15\x0d\xfe\x8c\xf0\x8c\x0a\x0c\xe1\x80\xe0\x3b\xf6\xd9\xda\xcb\xbb\x62\xd5\x2f\xf0\x01\xad\xe9\xd8\x6b\xbd\x63\xfd\x53\x3b\x01\x09\x25\x48\xac\xe0\x46\xc9\x3b\xdb\x45\xaf\x33\x57\x49\x0e\x1a\x26\x30\x14\xf8\x4b\xfd\xc2\x59\xc9\x5a\x37\x7c\x70\x91\xfe\xa9\x55\xb2\x91\x1e\x04\x59\x5f\x59\x37\x34\xde\x6a\xe8\x98\x88\xc4\xf2\x76\x40\x0d\xcf\x22\x74\x6c\x9e\xeb\x7d\xa4\xd1\xfa\x2f\xa8\x61\x1f\x96\x85\xf5\x77\x69\x59\x5a\xbe\x61\x56\xe6\xf6\xb4\x0a\xd5\xf6\xba\xc4\x0c\x72\x84\x09\x3a\x76\x3c\xbe\xbf\xb1\x3e\x7a\xd3\xc4\x88\xaa\x99\xe7\x3a\x29\x05\xb6\xcd\xaf\x3c\x25\x08\xd2\x89\x6f\x82\xa0\x44\xb2\x94\x4d\x42\xd2\xd9\xfd\x9e\x16\xc5\x5e\xc5\xec\xbb\x78\xd2\x18\x46\xf0\x29\xf3\xed\xb6\x29\xb9\xcd\x5c\xcb\xae\x8c\x48\xb7\x93\x42\xe3\xc9\xc3\x8a\xf9\x11\xac\x27\x56\x49\x6b\x08\x0c\xe5\x36\xf2\x51\xf7\x32\xf8\xbf\x87\x03\x78\x84\xf0\x38\x78\xc8\x5a\x99\x79\x6c\xea\xef\x76\x0a\x22\xd7\xfc\xbf\x5e\xf8\xed\x73\xa4\xd0\x57\x61\x70\x80\x40\x7b\xa3\x0e\x0e\x4d\xaa\xe1\x1c\x51\x01\xef\x5b\x5e\x3e\x67\xff\x27\x00\x00\xff\xff\x57\x39\xc6\xde\xa6\x02\x00\x00"

func contentOpfBytes() ([]byte, error) {
	return bindataRead(
		_contentOpf,
		"content.opf",
	)
}

func contentOpf() (*asset, error) {
	bytes, err := contentOpfBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "content.opf", size: 678, mode: os.FileMode(420), modTime: time.Unix(1476255196, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _coverHtml = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x6c\x8d\x4d\x4f\xc3\x30\x0c\x86\xef\xf9\x15\x26\xf7\xc6\x8a\x76\xa1\xe0\xe6\x40\x3b\x09\xa4\x01\x13\x0a\x02\x8e\x65\x35\x74\x52\x3f\xc6\xe6\xd1\x4e\x53\xff\x3b\x4a\x0a\x37\x7c\xb1\x5f\xe9\x7d\x1e\xd3\x45\xf1\x98\xfb\xb7\xf5\x12\x6a\x69\x1b\x58\x3f\xdf\xac\xee\x72\xd0\x09\xe2\xcb\x22\x47\x2c\x7c\x01\xaf\xb7\xfe\x7e\x05\xd6\x58\xc4\xe5\x83\x56\xf0\x3b\xba\x16\xd9\x5d\x21\x0e\xc3\x60\x86\x85\xe9\xf7\x9f\xe8\x9f\x70\x0c\x1a\x6b\x03\xf8\x77\x9b\x4a\x2a\xed\x14\xc5\x07\x63\xdb\x74\x87\xec\x1f\xd6\xa6\x69\x3a\x13\xb1\xcb\x65\xe5\x14\xb5\x2c\x25\x84\x6e\xc2\x5f\xc7\xed\x77\xa6\xf3\xbe\x13\xee\x24\xf1\xa7\x1d\x6b\xd8\xcc\x29\xd3\xc2\xa3\x60\x60\xaf\x61\x53\x97\xfb\x03\x4b\x76\x94\x8f\xe4\x52\xa3\x53\x24\x5b\x69\xd8\x11\xce\x5b\x11\x46\x39\xbd\xf7\xd5\xc9\xa9\xf3\xd9\x4c\x93\x22\x8c\x89\xa2\xc3\xa9\x9f\x00\x00\x00\xff\xff\xde\x87\x7a\x3d\x15\x01\x00\x00"

func coverHtmlBytes() ([]byte, error) {
	return bindataRead(
		_coverHtml,
		"cover.html",
	)
}

func coverHtml() (*asset, error) {
	bytes, err := coverHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cover.html", size: 277, mode: os.FileMode(420), modTime: time.Unix(1476255196, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _partHtml = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x6c\x8e\xcd\x4e\xeb\x30\x10\x85\xf7\x7e\x0a\xdf\xd9\xc7\xa3\xa8\x9b\x5b\x70\xb2\x68\x5a\x04\x52\x81\x8a\x1a\x01\xcb\xd0\x4c\x49\x85\xf3\x43\x3c\x25\x8d\xa2\xbc\x3b\xb2\x5b\x76\x78\x35\x1e\x7d\xdf\x9c\xa3\xff\x2d\x1f\x33\xf3\xb6\x59\xc9\x92\x2b\x2b\x37\xcf\x8b\xf5\x5d\x26\x21\x42\x7c\x99\x65\x88\x4b\xb3\x94\xaf\xb7\xe6\x7e\x2d\x63\x15\x23\xae\x1e\x40\xc8\xcb\x83\x92\xb9\xbd\x42\xec\xfb\x5e\xf5\x33\xd5\x74\x1f\x68\x9e\xf0\xe4\xcf\xc4\xb1\x17\x7f\x67\x55\x70\x01\xa9\xd0\x21\xe0\x54\xd9\xda\x25\x7f\xb8\xf1\x7c\x3e\x3f\x1b\x81\xa5\xbc\x48\x85\xae\x88\x73\xe9\xd9\x88\xbe\x8e\x87\xef\x04\xb2\xa6\x66\xaa\x39\x32\x43\x4b\x20\x77\xe7\x5f\x02\x4c\x27\x46\xef\x5e\xcb\x5d\x99\x77\x8e\x38\x39\xf2\x3e\xfa\x0f\x98\x0a\xcd\x07\xb6\x94\x8e\xa3\x32\x7e\x98\x26\x8d\xe7\x8d\xd0\xf6\x50\x7f\xca\xb2\xa3\x7d\x02\x4a\xe1\x96\x07\x4b\x0e\x1d\x37\xdd\xa0\x76\xce\x81\xec\xc8\x26\xe0\xc2\xba\x24\x62\x90\x3c\xb4\x74\x49\x0b\x80\x3f\x8f\xa1\xab\x7e\x6f\x8a\x21\x15\xe3\xa8\xb6\xde\x5f\x34\xc5\x70\xd3\x74\x86\xaa\xd6\xe6\x4c\xd3\x24\x34\x06\x42\x87\x9a\xe9\x4f\x00\x00\x00\xff\xff\x4f\x73\xcb\xe8\x77\x01\x00\x00"

func partHtmlBytes() ([]byte, error) {
	return bindataRead(
		_partHtml,
		"part.html",
	)
}

func partHtml() (*asset, error) {
	bytes, err := partHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "part.html", size: 375, mode: os.FileMode(420), modTime: time.Unix(1476255196, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _storyCss = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x9c\x54\xcd\x6e\xdb\x3c\x10\xbc\xeb\x29\xf6\x4b\x10\xe0\x6b\x10\xd3\x4a\xdc\x38\xad\x7a\x4c\xe1\x22\x28\x92\xf6\x50\x20\xe8\x91\x22\x57\xf2\x36\x14\x49\x50\x74\x6c\xb7\xf0\xbb\x17\xd4\x8f\x25\xd9\x6a\x13\xd4\x17\x43\xb3\xd4\xec\xcc\xee\x88\xa9\x91\x5b\xf8\x15\x01\x00\x14\x7c\x33\x59\x93\xf4\xcb\xe4\x3a\xc6\xe2\x43\x83\xb9\x9c\xf4\x44\x61\xe6\x13\xbe\xf2\x66\x80\x3a\xca\x97\x2d\xbc\x8b\x6c\x43\xe3\x71\xe3\x27\xa4\x25\x6a\x9f\x00\xbb\x26\x1d\x8a\x54\xe4\x4d\xb9\xee\x00\x97\x71\x7c\x06\xff\x51\x61\x8d\xf3\x5c\xfb\x9a\x77\x89\x15\x23\x04\xca\xa3\x62\xdd\x34\xbc\x69\x37\x10\x07\x52\xa6\xb6\x8e\x44\x79\x01\xd3\x73\xf8\xcc\xb7\x92\xc3\xf9\x14\x98\x74\xc8\x8b\xbd\xa5\x4e\x3e\xc4\xb5\x98\x23\x03\xfd\xc2\x40\xfc\x21\xce\x15\xe5\x3a\x81\xc0\x16\xda\x47\xbd\xb6\xd1\xa0\x6d\x66\xb4\x9f\x94\x7e\xab\x30\x01\xf2\x5c\x91\x68\xcf\xa3\x96\xe0\xf0\x99\x70\x8d\x12\xaa\x13\x65\x78\x3b\x62\x8f\x98\x72\xe7\x49\x28\x6c\x38\x84\x51\xc6\x25\x70\xba\xa8\x7e\xb5\x8a\x94\x8b\xa7\xdc\x99\x95\x96\x93\xb6\x1c\xc7\xf3\x79\x1c\x0f\x5c\x79\x63\xeb\x11\x8e\x79\x7d\x77\x76\xbc\xd8\x3d\x98\x1a\x27\xd1\x25\xe0\x97\xa4\xa1\x34\x8a\x24\x9c\xbe\xaf\x7e\x95\x7e\x96\xd2\x4f\xc1\x9d\xfc\x8b\xcb\x3f\xa8\xec\x9b\xe8\x8f\x52\xa0\xf6\xe8\x6a\xfc\x19\x83\x7f\xae\xda\x5a\xca\x4b\x54\xa4\x71\xb8\xfd\x6b\xbb\xa9\x01\xcb\xa5\x24\x9d\xf7\x90\x36\x3d\x57\xed\xd2\x9a\xa8\xcd\xda\xe7\xd6\xde\xa5\xdd\xb4\xee\x16\x8b\x38\x6e\xc7\xd7\xd3\x6d\x4d\x49\x9e\xcc\x81\x42\xa1\x90\xbb\x04\x52\xe3\x97\x35\x90\x29\xc3\x7d\x02\xda\x04\x95\xbb\x88\x7d\xbd\x2d\x85\x43\xd4\xfd\x01\x65\xbc\x20\xb5\x4d\xe0\xe4\xd6\xac\x1c\xa1\x83\x07\x5c\x9f\x5c\x40\xf3\x74\x01\x85\xd1\xa6\xb4\x5c\x34\x3e\x5f\xbd\xf6\x4e\xf7\x60\x95\xec\xed\x78\xc6\x3b\x7c\x3f\x38\x76\x79\x38\x98\xab\x6e\x30\x3c\xe3\xd9\x8d\xe8\x6d\x4c\xa2\x30\x8e\xd7\x43\xd9\x1b\x7e\x40\xdf\x78\x1d\xd1\x38\x9b\xc5\xf1\xde\x42\xaf\x83\x34\xab\x54\x61\xf0\xd0\x59\x3c\xb4\xbd\x8b\xd8\x82\x1c\xae\x8d\x7b\x2a\x8f\x3b\x50\xc1\x73\x4c\x60\xe5\xd4\xff\x8c\x4d\xef\xc2\x53\x39\xcd\x9a\xf3\x57\x37\x2c\xa7\xec\xcd\x8b\x3b\xdd\xaf\x36\x24\x7b\x89\x5c\x4a\x74\xfd\x2b\xec\x20\xa2\x95\xdb\x75\x29\x14\x59\x4b\x3a\x1f\x5e\x30\xcd\x90\x67\x63\x1f\xd7\xec\x15\x1f\x57\x60\x7f\xe4\x4e\x93\xce\xbf\xa3\x52\x66\x3d\x72\x09\x8c\xc4\xb4\x2b\xd7\x69\xd8\x45\xec\xfe\xee\xe3\xf0\xa2\x9d\xc5\x56\x8c\xa4\xb5\x97\xe7\x0e\x68\x25\x36\xea\xe6\xf3\x79\xe0\xbc\xef\x38\xff\x25\xd1\x23\x7a\xe7\xf3\x5e\xb8\x0f\x7b\xee\x03\xf0\xed\xcb\xc3\xa7\x57\xee\xde\x72\x8b\x8e\xfd\xb0\xf9\xf1\xda\x1d\x5a\x0c\xb6\xeb\xff\x8a\xd8\x72\x27\x96\x05\xea\x91\xec\xbe\xc8\xbe\x8b\x98\xb3\x24\xfc\xca\xb5\x17\x75\x33\xd6\x2a\x02\xa1\xfe\x3b\x00\x00\xff\xff\x52\x46\x52\x7c\x52\x07\x00\x00"

func storyCssBytes() ([]byte, error) {
	return bindataRead(
		_storyCss,
		"story.css",
	)
}

func storyCss() (*asset, error) {
	bytes, err := storyCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "story.css", size: 1874, mode: os.FileMode(420), modTime: time.Unix(1476255196, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _tocNcx = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x84\xd0\xc1\x4a\xc4\x30\x10\x80\xe1\x7b\x9f\xa2\xe4\x6e\xa6\x2a\x8a\x96\x34\x17\xbd\x78\x70\xf1\xe0\x3e\x40\xb6\x19\xb6\x81\x66\x52\xda\xc9\x36\x1a\xf2\xee\xd2\x45\x44\xd8\xc3\x1e\xe7\xe7\x1b\x18\x46\x51\x9f\xea\xe4\x47\x5a\x3a\x31\x30\x4f\x2d\xc0\xba\xae\xd2\x1a\xb7\x7c\xc9\x30\x1f\xe1\xfb\xfe\xf9\xe9\x11\xee\x9a\xe6\x01\xa8\x4f\x20\xea\x13\xce\x8b\x0b\xd4\x89\xad\xdd\xdc\x0a\x5d\xa9\x01\x8d\xd5\x95\xf2\xc8\xa6\xee\x03\x31\x12\x77\x22\xce\xd4\xc6\xe8\x6c\x9b\xb3\xdc\xef\xdf\x5e\x4b\x11\x35\x19\x8f\x9d\xb0\x7c\x68\xa3\xb3\x02\x2e\x76\x9a\xff\xc4\xe2\xc4\xc3\x35\xc4\x81\xcd\xf8\x61\x8e\xf8\x12\x22\xf1\x35\xed\x4d\xda\xec\x2e\xfa\x03\xce\x02\xb4\x82\xdf\xd3\x6d\xe8\x3f\x1d\x8f\xa8\x15\x63\x62\x9d\xb3\x3c\x8f\xa5\x28\x38\x07\x05\x7f\xa2\xca\x59\xee\xcc\xe9\xdd\x4c\xa5\x54\x6a\x7b\x8a\xfe\x09\x00\x00\xff\xff\xf6\x95\x75\xd6\x45\x01\x00\x00"

func tocNcxBytes() ([]byte, error) {
	return bindataRead(
		_tocNcx,
		"toc.ncx",
	)
}

func tocNcx() (*asset, error) {
	bytes, err := tocNcxBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "toc.ncx", size: 325, mode: os.FileMode(420), modTime: time.Unix(1476255196, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"content.opf": contentOpf,
	"cover.html": coverHtml,
	"part.html": partHtml,
	"story.css": storyCss,
	"toc.ncx": tocNcx,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"content.opf": &bintree{contentOpf, map[string]*bintree{}},
	"cover.html": &bintree{coverHtml, map[string]*bintree{}},
	"part.html": &bintree{partHtml, map[string]*bintree{}},
	"story.css": &bintree{storyCss, map[string]*bintree{}},
	"toc.ncx": &bintree{tocNcx, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

