package cbcluster

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _data_couchbase_node_service_template = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xac\x93\x5f\xcb\xd3\x30\x14\xc6\xef\xfb\x29\x72\x21\x0c\x84\xbc\xf1\x42\x6f\x5e\xe9\x45\x9d\x7d\xa5\x37\xeb\x68\xeb\x10\xc6\x28\x59\x7a\xe6\xc2\xd2\x24\xe6\x4f\x37\x19\xfb\xee\x66\xeb\x58\x75\x45\x27\xe2\x5d\xf8\xf1\x3c\x4f\xce\x73\xe0\x2c\x3f\x4b\xee\x56\xd1\x47\xb0\xcc\x70\xed\xb8\x92\x31\x53\x9e\x6d\xd7\xd4\x42\x2d\x55\x03\x51\xb2\x71\x60\xe2\x46\xb1\x1d\x98\x27\x0b\xa6\xe3\x0c\xa2\x02\xbe\x79\x6e\xc0\xde\xf3\x5e\x0c\x8e\x35\x63\xe9\x2f\x34\x5a\x96\xfd\x6b\x15\x55\xbc\x05\xe5\x5d\xe9\xa8\x71\x25\xb0\xf8\xcd\x40\x94\xee\x41\x2a\x3b\x6e\x94\x6c\x41\xba\x17\x2e\x20\x26\x21\x8b\xc0\x00\xa3\xf4\x00\xec\x12\x30\x37\x10\x63\xe2\xad\x21\x6b\x2e\x49\x3f\x1d\xda\x71\x21\xd0\xad\xd6\x03\xb1\x69\x7f\x27\xbd\x57\x6a\x1f\x62\x9d\x80\xef\x0d\xc8\x77\x7c\x7f\x20\x37\x1f\x3e\xd7\x04\x83\x8f\x47\xf4\x34\xfd\x50\x2f\xd2\xa2\xcc\xf2\x19\x3a\x9d\x9e\x2f\x24\x9f\x55\x49\x36\x4b\x8b\xba\x4a\x3e\x05\xf8\xcf\xbf\x30\xe1\x6d\xd8\x37\xfe\xaa\x1e\xe4\xc6\x97\xc0\xe0\xd9\x22\xcc\xd0\x64\x54\xd9\x4b\x84\xb1\xa4\x2d\x0c\xd5\x11\xee\x10\x51\xda\x0d\xdf\x91\x8e\x9a\xe7\x31\x3a\x3b\xc1\xc5\x5b\x65\xdd\xff\x58\x06\xfa\xc9\x77\x1e\x7d\x72\x6d\xa1\xf4\xdf\x95\xf8\xf3\x28\x0f\x36\x86\xbc\x6e\xa8\x03\xbc\x37\x54\xeb\x90\x39\x32\x22\x03\xad\xea\x00\x53\xd9\x60\x03\x6b\x2a\xa8\x64\x61\x55\x58\x28\x46\x05\xe6\x1a\xbd\x9a\xe6\x45\x9a\x97\xf5\xbc\xc8\x16\x49\x95\xd6\xd9\x7c\xf1\xf6\x3d\xb2\xbe\x51\xe8\x3a\xa7\x0d\x55\x86\xe0\x49\xb8\x84\x2f\xf8\x45\x00\x84\x2b\x9c\x2a\xb9\x11\x9c\x39\x7b\x77\x83\xaf\x6f\x67\xf3\x23\x00\x00\xff\xff\x6f\x86\x56\xbc\xaf\x03\x00\x00")

func data_couchbase_node_service_template_bytes() ([]byte, error) {
	return bindata_read(
		_data_couchbase_node_service_template,
		"data/couchbase_node@.service.template",
	)
}

func data_couchbase_node_service_template() (*asset, error) {
	bytes, err := data_couchbase_node_service_template_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "data/couchbase_node@.service.template", size: 943, mode: os.FileMode(420), modTime: time.Unix(1425610643, 0)}
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
	"data/couchbase_node@.service.template": data_couchbase_node_service_template,
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
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func     func() (*asset, error)
	Children map[string]*_bintree_t
}

var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"data": &_bintree_t{nil, map[string]*_bintree_t{
		"couchbase_node@.service.template": &_bintree_t{data_couchbase_node_service_template, map[string]*_bintree_t{}},
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
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

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	if err != nil { // File
		return RestoreAsset(dir, name)
	} else { // Dir
		for _, child := range children {
			err = RestoreAssets(dir, path.Join(name, child))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
