// Code generated by go-bindata.
// sources:
// input/createLesson.gql
// input/createStudy.gql
// input/createUser.gql
// input/deleteLesson.gql
// input/deleteStudy.gql
// input/deleteUser.gql
// input/updateLesson.gql
// input/updateStudy.gql
// input/updateUser.gql
// interfaces/deletable.gql
// interfaces/node.gql
// interfaces/study_node.gql
// interfaces/uniform_resource_locatable.gql
// interfaces/updateable.gql
// scalars/html.gql
// scalars/time.gql
// scalars/uri.gql
// schema.gql
// type/lesson.gql
// type/page_info.gql
// type/study.gql
// type/user.gql
// DO NOT EDIT!

package schema

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

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
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

var _inputCreatelessonGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x2e\x4a\x4d\x2c\x49\xf5\x49\x2d\x2e\xce\xcf\xd3\xe3\xca\x04\x4b\x22\x8b\x41\x94\x57\x73\x29\x28\x24\xe5\xa7\x54\x5a\x29\x80\x40\x70\x49\x51\x66\x5e\x3a\x97\x82\x42\x71\x49\x69\x4a\xa5\x67\x8a\x95\x82\x82\xa7\x8b\x22\x97\x82\x42\x49\x66\x49\x4e\xaa\x15\x42\x89\x22\x57\x2d\x17\x20\x00\x00\xff\xff\xae\x5e\x3b\xfc\x72\x00\x00\x00")

func inputCreatelessonGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputCreatelessonGql,
		"input/createLesson.gql",
	)
}

func inputCreatelessonGql() (*asset, error) {
	bytes, err := inputCreatelessonGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/createLesson.gql", size: 114, mode: os.FileMode(420), modTime: time.Unix(1525788725, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputCreatestudyGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x2e\x4a\x4d\x2c\x49\x0d\x2e\x29\x4d\xa9\xd4\xe3\xca\x04\xcb\x21\x09\x41\x14\x57\x73\x29\x28\xa4\xa4\x16\x27\x17\x65\x16\x94\x64\xe6\xe7\x59\x29\x28\x04\x97\x14\x65\xe6\xa5\x73\x29\x28\xe4\x25\xe6\xa6\x5a\x29\xc0\x00\x44\x58\x91\xab\x96\x0b\x10\x00\x00\xff\xff\x74\x2b\x3e\x91\x68\x00\x00\x00")

func inputCreatestudyGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputCreatestudyGql,
		"input/createStudy.gql",
	)
}

func inputCreatestudyGql() (*asset, error) {
	bytes, err := inputCreatestudyGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/createStudy.gql", size: 104, mode: os.FileMode(420), modTime: time.Unix(1525796319, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputCreateuserGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x2e\x4a\x4d\x2c\x49\x0d\x2d\x4e\x2d\xd2\xe3\xca\x04\x4b\x21\x44\x20\x4a\xab\xb9\x14\x14\x52\x73\x13\x33\x73\xac\x14\x40\x20\xb8\xa4\x28\x33\x2f\x5d\x91\x4b\x41\x21\x27\x3f\x3d\x33\x0f\x5d\xb0\x20\xb1\xb8\xb8\x3c\xbf\x28\xc5\x0a\x21\x58\xcb\x05\x08\x00\x00\xff\xff\x39\x13\x48\x49\x76\x00\x00\x00")

func inputCreateuserGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputCreateuserGql,
		"input/createUser.gql",
	)
}

func inputCreateuserGql() (*asset, error) {
	bytes, err := inputCreateuserGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/createUser.gql", size: 118, mode: os.FileMode(420), modTime: time.Unix(1525788715, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputDeletelessonGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x49\xcd\x49\x2d\x49\xf5\x49\x2d\x2e\xce\xcf\xd3\xe3\xca\x04\x4b\x22\x8b\x41\x94\x57\x73\x29\x28\xe4\x40\xf8\x29\x56\x0a\x9e\x2e\x8a\x5c\xb5\x5c\x80\x00\x00\x00\xff\xff\xfe\x57\x79\x43\x4b\x00\x00\x00")

func inputDeletelessonGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputDeletelessonGql,
		"input/deleteLesson.gql",
	)
}

func inputDeletelessonGql() (*asset, error) {
	bytes, err := inputDeletelessonGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/deleteLesson.gql", size: 75, mode: os.FileMode(420), modTime: time.Unix(1525788731, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputDeletestudyGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x49\xcd\x49\x2d\x49\x0d\x2e\x29\x4d\xa9\xd4\xe3\xca\x04\xcb\x21\x09\x41\x14\x57\x73\x29\x28\x14\x83\xb9\x29\x56\x0a\x9e\x2e\x8a\x5c\xb5\x5c\x80\x00\x00\x00\xff\xff\x05\x0c\x8f\x92\x48\x00\x00\x00")

func inputDeletestudyGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputDeletestudyGql,
		"input/deleteStudy.gql",
	)
}

func inputDeletestudyGql() (*asset, error) {
	bytes, err := inputDeletestudyGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/deleteStudy.gql", size: 72, mode: os.FileMode(420), modTime: time.Unix(1525788734, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputDeleteuserGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x70\x49\xcd\x49\x2d\x49\x0d\x2d\x4e\x2d\xd2\xe3\xca\x04\x4b\x21\x44\x20\x4a\xab\xb9\x14\x14\x4a\x41\xbc\x14\x2b\x05\x4f\x17\x45\xae\x5a\x2e\x40\x00\x00\x00\xff\xff\x3b\x26\xa7\x97\x45\x00\x00\x00")

func inputDeleteuserGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputDeleteuserGql,
		"input/deleteUser.gql",
	)
}

func inputDeleteuserGql() (*asset, error) {
	bytes, err := inputDeleteuserGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/deleteUser.gql", size: 69, mode: os.FileMode(420), modTime: time.Unix(1525788737, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputUpdatelessonGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xcb\x31\x0a\x03\x21\x14\x04\xd0\xfe\x9f\x62\x42\xfa\x1c\xc0\x3a\x8d\x90\x2e\xe4\x00\x11\xff\x2e\x82\xfb\x15\xfd\x16\xb2\xec\xdd\x17\xb4\x71\xca\x99\x37\x4f\x58\xc9\x4d\xa1\x3d\x33\xb6\x54\xf0\xcb\xfe\xaf\xfc\xe1\x5a\x93\xbc\x28\x8c\x71\xed\x26\x3f\x09\x70\xc9\x77\x83\x91\xaf\x96\x20\x3b\x01\x71\x1a\x6f\x00\xfb\x7e\x10\x20\xed\x70\x5c\x06\xb3\xa2\x04\x68\xd0\xc8\x66\x7d\x5d\x74\x07\x00\x00\xff\xff\x0a\xa2\xe4\xfa\x85\x00\x00\x00")

func inputUpdatelessonGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputUpdatelessonGql,
		"input/updateLesson.gql",
	)
}

func inputUpdatelessonGql() (*asset, error) {
	bytes, err := inputUpdatelessonGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/updateLesson.gql", size: 133, mode: os.FileMode(420), modTime: time.Unix(1525788746, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputUpdatestudyGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x08\x2d\x48\x49\x2c\x49\x0d\x2e\x29\x4d\xa9\xd4\xe3\xca\x04\xcb\x21\x09\x41\x14\x57\x73\x29\x28\xa4\xa4\x16\x27\x17\x65\x16\x94\x64\xe6\xe7\x59\x29\x28\x04\x97\x14\x65\xe6\xa5\x73\x29\x28\x14\x83\x55\xa5\x58\x29\x80\x81\xa7\x8b\x22\x57\x2d\x17\x20\x00\x00\xff\xff\x90\xd8\x0c\xbd\x64\x00\x00\x00")

func inputUpdatestudyGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputUpdatestudyGql,
		"input/updateStudy.gql",
	)
}

func inputUpdatestudyGql() (*asset, error) {
	bytes, err := inputUpdatestudyGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/updateStudy.gql", size: 100, mode: os.FileMode(420), modTime: time.Unix(1525788873, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _inputUpdateuserGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xf0\xcc\x2b\x28\x2d\x51\x28\xa9\x2c\x48\x55\x48\xcb\x2f\x52\x08\x2d\x48\x49\x2c\x49\x0d\x2d\x4e\x2d\xd2\xe3\xca\x04\x4b\x21\x44\x20\x4a\xab\xb9\x14\x14\x92\x32\xf3\xad\x14\x40\x20\xb8\xa4\x28\x33\x2f\x9d\x4b\x41\xa1\x14\x24\x9f\x62\xa5\xa0\xe0\xe9\xa2\xc8\xa5\xa0\x90\x93\x9f\x9e\x99\x67\x85\xac\x20\x2f\x31\x37\xd5\x0a\x49\x4b\x2d\x17\x20\x00\x00\xff\xff\xd0\x54\xf5\x43\x7c\x00\x00\x00")

func inputUpdateuserGqlBytes() ([]byte, error) {
	return bindataRead(
		_inputUpdateuserGql,
		"input/updateUser.gql",
	)
}

func inputUpdateuserGql() (*asset, error) {
	bytes, err := inputUpdateuserGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "input/updateUser.gql", size: 124, mode: os.FileMode(420), modTime: time.Unix(1525788751, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _interfacesDeletableGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2c\xcb\x41\x0a\xc2\x40\x10\x44\xd1\x7d\x9f\xa2\x24\xfb\x1c\xc0\xa5\xd1\x83\xf4\x4c\x4a\xa6\x65\xe8\x81\xa4\xd0\x85\x78\x77\x51\xb2\xfd\x9f\x37\xe1\x96\x0a\x05\x77\xa8\xb9\x50\x3d\x51\x88\x95\x9d\xe2\x3a\x5b\xa4\xb8\xdd\xbd\x12\xd7\x5f\xf2\xd2\x89\xb7\x01\x13\x16\x4f\xa8\x11\xcf\xe0\x8b\xdb\x21\xa0\x16\x3b\x46\x79\xb0\x6a\x36\x1c\x73\xf1\xfc\x6b\x9e\x71\x19\xa3\xd3\xf3\x64\x1f\xfb\x06\x00\x00\xff\xff\x0a\xa9\x35\xe6\x7b\x00\x00\x00")

func interfacesDeletableGqlBytes() ([]byte, error) {
	return bindataRead(
		_interfacesDeletableGql,
		"interfaces/deletable.gql",
	)
}

func interfacesDeletableGql() (*asset, error) {
	bytes, err := interfacesDeletableGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "interfaces/deletable.gql", size: 123, mode: os.FileMode(420), modTime: time.Unix(1525789508, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _interfacesNodeGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\xcc\x2b\x49\x2d\x4a\x4b\x4c\x4e\x55\xf0\xcb\x4f\x49\x55\xa8\xe6\x52\x50\xc8\x4c\xb1\x52\xf0\x74\x51\xe4\xaa\xe5\x02\x04\x00\x00\xff\xff\x6f\x40\x52\x35\x1d\x00\x00\x00")

func interfacesNodeGqlBytes() ([]byte, error) {
	return bindataRead(
		_interfacesNodeGql,
		"interfaces/node.gql",
	)
}

func interfacesNodeGql() (*asset, error) {
	bytes, err := interfacesNodeGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "interfaces/node.gql", size: 29, mode: os.FileMode(420), modTime: time.Unix(1523374392, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _interfacesStudy_nodeGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\xcc\x2b\x49\x2d\x4a\x4b\x4c\x4e\x55\x08\x2e\x29\x4d\xa9\xf4\xcb\x4f\x49\x55\xa8\xe6\x52\x50\x50\x56\x08\xc9\x48\x55\x28\x06\x89\x29\x24\x16\x17\xe7\x27\x67\x26\x96\xa4\xa6\x28\x94\x67\x96\x64\x28\x94\x64\x64\x16\x2b\xe4\xe5\xa7\xa4\xea\x71\x29\x40\x94\x58\x41\x74\x2b\x72\xd5\x72\x01\x02\x00\x00\xff\xff\xdd\x43\xe1\x5f\x51\x00\x00\x00")

func interfacesStudy_nodeGqlBytes() ([]byte, error) {
	return bindataRead(
		_interfacesStudy_nodeGql,
		"interfaces/study_node.gql",
	)
}

func interfacesStudy_nodeGql() (*asset, error) {
	bytes, err := interfacesStudy_nodeGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "interfaces/study_node.gql", size: 81, mode: os.FileMode(420), modTime: time.Unix(1525789404, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _interfacesUniform_resource_locatableGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\xcb\x41\xaa\x83\x30\x14\x85\xe1\x79\x56\x71\x1e\xce\x5d\xc0\x5b\x41\x0b\x0e\x24\x98\x05\x5c\xd3\x23\x09\xd8\x44\x6e\xae\x05\x29\xdd\x7b\x27\x52\x28\x74\xf6\x0f\xfe\xaf\x83\xe7\xa6\x6c\x2c\xd6\x20\xb0\x63\x23\x2c\x89\x21\x4a\xc1\x4c\x28\x4d\x33\x1f\xbc\x61\x3e\x20\x08\x7e\xe8\x5d\x2e\x46\x5d\x24\x12\xa1\xe4\xa5\xea\xdd\xb3\xd5\x5d\x23\x87\x1a\xc5\x64\x5e\x89\xa7\x03\x3a\x4c\x89\xb8\x4c\xd3\x88\x4d\x2c\xc1\x2a\x2c\xe5\x06\x3d\xef\xde\xe1\xd3\xa3\x58\xfa\x47\xf0\xd7\x3f\xf7\x2d\x83\x1f\x7e\xc2\x5d\xd7\xf3\x7f\xb9\x77\x00\x00\x00\xff\xff\xe9\x60\xa2\xe7\xc4\x00\x00\x00")

func interfacesUniform_resource_locatableGqlBytes() ([]byte, error) {
	return bindataRead(
		_interfacesUniform_resource_locatableGql,
		"interfaces/uniform_resource_locatable.gql",
	)
}

func interfacesUniform_resource_locatableGql() (*asset, error) {
	bytes, err := interfacesUniform_resource_locatableGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "interfaces/uniform_resource_locatable.gql", size: 196, mode: os.FileMode(420), modTime: time.Unix(1523803810, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _interfacesUpdateableGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2c\xcb\x41\x0a\xc2\x40\x10\x44\xd1\x7d\x9f\xa2\x24\xfb\x1c\xc0\xa5\xc1\x23\x78\x80\x9e\x49\xc9\xb4\x84\x1e\x49\x4a\x5d\x88\x77\x17\x34\xdb\xff\x79\x03\xce\xa9\x50\x70\x83\x9a\x0b\xd5\x13\x85\x78\xdc\x67\x17\xe7\xd1\x22\xc5\xf5\xea\x95\xb8\xfc\x92\x97\x85\x78\x1b\x30\x60\xf2\x84\x1a\xf1\x0c\xbe\xb8\xee\x04\x6a\xb1\xa1\x97\x1b\xab\x46\xc3\x3e\x27\xcf\x3f\x3f\xe2\xd4\xfb\x42\xcf\x83\x7d\xec\x1b\x00\x00\xff\xff\x9b\x38\x05\xaa\x7c\x00\x00\x00")

func interfacesUpdateableGqlBytes() ([]byte, error) {
	return bindataRead(
		_interfacesUpdateableGql,
		"interfaces/updateable.gql",
	)
}

func interfacesUpdateableGql() (*asset, error) {
	bytes, err := interfacesUpdateableGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "interfaces/updateable.gql", size: 124, mode: os.FileMode(420), modTime: time.Unix(1525789514, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _scalarsHtmlGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\x70\x54\x28\x2e\x29\xca\xcc\x4b\x57\x48\xce\xcf\x2b\x49\xcc\xcc\x03\x31\x3d\x42\x7c\x7d\x14\x92\xf3\x53\x52\xf5\xb8\x8a\x93\x13\x73\x12\x8b\xc0\x22\x5c\x80\x00\x00\x00\xff\xff\x69\x83\x95\x22\x2d\x00\x00\x00")

func scalarsHtmlGqlBytes() ([]byte, error) {
	return bindataRead(
		_scalarsHtmlGql,
		"scalars/html.gql",
	)
}

func scalarsHtmlGql() (*asset, error) {
	bytes, err := scalarsHtmlGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "scalars/html.gql", size: 45, mode: os.FileMode(420), modTime: time.Unix(1523745255, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _scalarsTimeGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\x08\xc9\xcc\x4d\x55\xc8\x2c\x56\x48\xcc\x53\x08\x72\x73\x36\x36\x36\xb6\x54\x28\xc9\xcc\x4d\x2d\x2e\x49\xcc\x2d\xd0\xe3\x2a\x4e\x4e\xcc\x49\x2c\x02\x2b\xe2\x02\x04\x00\x00\xff\xff\x3c\x59\x30\xf9\x2c\x00\x00\x00")

func scalarsTimeGqlBytes() ([]byte, error) {
	return bindataRead(
		_scalarsTimeGql,
		"scalars/time.gql",
	)
}

func scalarsTimeGql() (*asset, error) {
	bytes, err := scalarsTimeGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "scalars/time.gql", size: 44, mode: os.FileMode(420), modTime: time.Unix(1523805330, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _scalarsUriGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\x70\xcc\x53\x08\x72\x73\x56\x30\xb6\xb4\x30\xd3\x81\xb1\xcc\x75\x14\x12\xf3\x52\xc0\x3c\x33\x53\x73\x03\x05\x8d\x9c\xd4\xb2\xd4\x1c\x05\x13\x4d\x85\xe4\xfc\xdc\x82\x9c\xcc\xc4\xbc\x12\x85\xd0\x20\x4f\x85\xe2\x92\xa2\xcc\xbc\x74\x3d\xae\xe2\xe4\xc4\x9c\xc4\x22\x90\x10\x17\x20\x00\x00\xff\xff\x12\xa3\x29\x58\x51\x00\x00\x00")

func scalarsUriGqlBytes() ([]byte, error) {
	return bindataRead(
		_scalarsUriGql,
		"scalars/uri.gql",
	)
}

func scalarsUriGql() (*asset, error) {
	bytes, err := scalarsUriGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "scalars/uri.gql", size: 81, mode: os.FileMode(420), modTime: time.Unix(1523745591, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x91\xcf\x4a\xc4\x30\x10\xc6\xef\x79\x8a\x09\x5e\xea\x2b\xe4\xa8\xb9\x14\x54\x90\xa5\xa7\x65\x0f\x65\x33\xac\x81\x6e\x52\xf3\x47\x29\xe2\xbb\xcb\x24\xdd\x34\xa9\xb0\xa7\x76\x7e\x33\xdf\x7c\x1f\x19\x7f\xfe\xc0\xeb\x08\x3f\x0c\xe0\x33\xa2\x5b\x04\xbc\xd3\x87\x01\x5c\x63\x18\x83\xb6\x46\xc0\xeb\xfa\xc7\x7e\x59\x58\x66\xcc\x13\x49\x62\xac\xc2\x4e\x2b\x01\xbd\xe4\x8f\x02\xde\xac\xc2\x95\xfa\x4e\x2b\x2f\xe0\xd8\x4b\x7e\xa2\xd6\x91\x7a\x27\xce\x00\xa2\x47\xd7\x4d\xf6\xa2\x8d\x80\x43\x70\xda\x5c\xa8\x3f\x78\x74\x0c\xe0\x4b\xe3\x37\xba\x5c\xf2\x9b\xdf\xcd\x3f\x59\x9e\x1d\x8e\x01\x5f\xd0\x7b\x6b\x3a\x6d\xe6\x18\x04\x3c\x57\xac\x27\x44\x1b\x73\xc9\x00\x1e\x40\xe1\x84\x7b\x8d\xac\x58\xd1\x3c\x59\x3b\xe1\x68\x78\x52\xc5\x59\xfd\x73\x1a\x2a\xb6\x73\xe2\x25\xdc\x21\x44\xb5\xb4\xd9\x12\x2a\x82\x54\x55\xc9\x1a\x81\xdc\xd0\xbd\x5c\x8d\x66\xd8\x50\x6b\xb2\x85\xa2\x17\x6d\x33\x11\x29\xd3\xeb\xfb\xe7\x3c\xf5\xac\x2c\xa4\xda\x4c\x57\xa3\x53\x26\xdb\x7a\x7a\x28\xa4\xd9\x4c\xa7\xfc\x0b\x00\x00\xff\xff\x60\x86\x54\x2a\x6a\x02\x00\x00")

func schemaGqlBytes() ([]byte, error) {
	return bindataRead(
		_schemaGql,
		"schema.gql",
	)
}

func schemaGql() (*asset, error) {
	bytes, err := schemaGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema.gql", size: 618, mode: os.FileMode(420), modTime: time.Unix(1525797670, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _typeLessonGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x52\xcd\x8e\xd3\x30\x10\xbe\xf7\x29\x3e\xb4\x57\xc4\x03\xe4\xb6\x6c\x91\xb6\x52\x41\xab\x92\x3e\xc0\x24\x9e\xe2\x41\x8e\x1d\xd9\x63\xaa\x0a\xf1\xee\xc8\xae\x89\x4a\x53\x71\xe0\x36\x33\xfe\x7e\x6c\xcf\xf7\x84\x03\xcf\x91\x13\x7b\x4d\x20\x38\x4e\x29\xf8\x0f\x1b\xbd\xcc\x8c\x7d\x6d\x20\xd3\xec\x78\xaa\x80\x2f\xc1\xf0\x7b\x1c\xbd\x9c\x42\x9c\x0e\x9c\x42\x8e\x23\xef\xc3\x48\x4a\x83\x2b\x27\xb3\x21\xe5\x52\xe3\xe7\x06\x78\x42\x6f\x19\x94\xd5\x86\x88\x70\x82\x5a\x5e\x1c\xd0\xe6\x1d\x8e\x89\xe3\x66\x41\x0f\xc1\x5c\xd6\xd8\x32\xed\xf0\x55\xa3\xf8\x6f\xef\xfe\x05\x46\x64\x6f\x38\xb2\x81\x06\xbc\xf6\x9f\xf7\x7f\xd8\xa5\xee\xea\xa4\xf1\x77\x86\xbd\xca\x49\x38\x55\x7a\xb9\x38\xc8\x1b\xa8\x4c\x8c\xb3\x65\x5f\xc7\x61\xf8\xce\xa3\xe2\x4c\x09\x63\x64\x52\x36\x45\xb0\x95\xcf\xda\xa1\x97\x89\xab\xa2\x98\x0e\xbb\xed\xff\x8a\x3b\x4a\x0a\x36\xd2\x0c\x4a\xfb\xa9\x76\x7f\x79\xac\x84\xdb\xa3\x7d\x9e\x06\x8e\x85\x78\xad\x3a\xec\xbc\xae\x19\x8b\x71\xa3\x15\xe3\x39\x0f\x4e\x92\x65\x03\xd2\x22\xb0\xf4\x77\xc6\xe5\xbb\x93\x66\x73\x01\xa5\x14\x46\x29\xef\xc7\x59\xd4\x42\xad\xa4\x9b\x55\x55\x50\xd9\x55\x36\x97\x1b\xae\x8a\x3a\x5e\x2f\xb6\x8e\x1f\x6c\xf6\xb5\xef\xdf\x30\x93\x5a\x9c\x42\xbc\xb7\x88\x2d\x79\x6f\xa4\xb6\xc3\xf1\xb0\xbb\x27\x1e\x0f\xfb\x47\xbc\x1c\xdd\x2d\xfc\x85\xae\xbf\xf1\x43\xf8\xcc\x11\xb9\x86\xf7\x9e\x73\x3d\x7c\x21\x7f\xcd\x76\x87\x8f\x21\x38\x26\xdf\x34\xb6\x62\x6e\x35\x5a\xd6\x1f\x6a\x6c\xc5\x3c\xb7\xc8\x2f\x1a\xbf\x36\xbf\x03\x00\x00\xff\xff\xd9\x28\x2f\x96\x81\x03\x00\x00")

func typeLessonGqlBytes() ([]byte, error) {
	return bindataRead(
		_typeLessonGql,
		"type/lesson.gql",
	)
}

func typeLessonGql() (*asset, error) {
	bytes, err := typeLessonGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "type/lesson.gql", size: 897, mode: os.FileMode(420), modTime: time.Unix(1525789578, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _typePage_infoGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\xcf\xbd\x4a\xc7\x40\x10\x04\xf0\xfe\x9e\x62\xe4\xdf\x4a\x1e\x20\x8d\xa0\x95\x8d\x04\x2c\xac\x37\x97\xcd\xdd\xa1\xd9\x0d\x7b\x7b\x7e\x20\xbe\xbb\x9c\x82\x82\x44\xc4\x72\x77\x98\x1f\xcc\x09\xd7\xb2\xaa\x6d\xe4\x45\x05\x34\x6b\x73\xec\x94\x8a\x7c\x3e\x8a\x80\x10\x55\x84\x63\xbf\x87\xe0\x2f\x3b\x63\xa2\xc4\xbd\x86\xd7\x00\x9c\x70\x97\x59\xbe\x4a\x92\xb0\xaa\x3d\x91\x2d\xf5\x1c\x64\x0c\xcf\x6c\x8c\x4d\x8d\x51\x9c\xb7\x7a\x11\x80\x4c\xf5\x86\x9f\xbd\x3b\x23\x2e\x55\x1f\x98\xe4\x2c\x1c\x62\x33\xc5\xfb\xbf\xb5\xc9\xf8\xb1\x68\xab\xff\x14\x3d\x33\x62\xb3\xaa\x06\xd7\xbe\xd3\x8b\x34\x1e\x02\x50\x9d\xcc\xaf\x3e\xa2\x11\xb7\x6e\x45\xd2\xb1\xf6\x3d\xf6\x77\x8c\x65\xf9\x41\xbd\x85\xf7\x00\x00\x00\xff\xff\x92\xd2\xc8\xaf\x79\x01\x00\x00")

func typePage_infoGqlBytes() ([]byte, error) {
	return bindataRead(
		_typePage_infoGql,
		"type/page_info.gql",
	)
}

func typePage_infoGql() (*asset, error) {
	bytes, err := typePage_infoGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "type/page_info.gql", size: 377, mode: os.FileMode(420), modTime: time.Unix(1523742707, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _typeStudyGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x52\xcb\x6a\xdc\x40\x10\xbc\xeb\x2b\xca\xf8\x90\x04\x82\x3f\x40\xb7\x90\x1c\x76\x61\x93\x18\x59\x4b\xce\x23\x4d\xcb\x9a\x20\xf5\x88\x9e\x16\x62\x09\xf9\xf7\x30\x8f\xb5\x97\x78\x0d\x7b\x6b\x75\x57\x57\xb5\x6a\xea\x1e\x0d\x2d\x42\x81\x58\x03\x0c\x82\xae\xf6\xf4\x50\xe9\x69\x21\x3c\xc5\x1a\x6e\x5e\x26\x9a\xd3\xf8\x87\xb7\xf4\x19\x47\x76\x83\x97\xb9\xa1\xe0\x57\xe9\xe9\xe0\x7b\xa3\xa6\x9b\x08\x7f\x2a\xe0\x1e\x7b\x4b\xac\x6e\x70\x14\xa0\x23\xc1\x1a\x25\x18\xb6\x50\x37\x13\xb6\x91\x38\xb5\x7d\xf7\x9b\x7a\xc5\x66\x02\x7a\x21\xa3\x64\x1f\x2a\x9c\xcb\x2f\x5a\xa3\x75\x33\xdd\x55\x89\xb1\x8d\x34\x14\x7a\x71\x8b\x3a\xcf\xf0\x43\xa2\x28\xa7\xe2\x72\x56\xe3\x49\xc5\xf1\xf3\x0d\x9b\x10\x62\x4b\x42\x16\xea\xb1\x6b\xbf\x1f\xfe\xa3\x8a\xad\x3a\x0d\x12\x99\xb3\x35\xf6\xdf\x0a\x6f\x43\xba\x0a\x27\xbf\x1c\x3f\x4f\x84\x89\x42\xf0\x8c\x41\xfc\x9c\x14\xfa\x55\x84\x58\x8b\x52\x77\x02\xaf\x73\x47\x12\x25\x32\xf4\x63\x05\x9c\x2f\xcc\x33\x0c\x5e\xd2\x6e\xe1\x52\x8f\x8e\x20\x49\x29\xbb\x83\x82\xac\xb1\x67\xbd\x03\x2a\xe0\x53\x8d\x43\x82\xe7\xbb\x76\x7e\xc3\x6c\xf8\x54\x38\xd2\x0b\x08\xc1\x08\xc1\x45\xdf\x5d\x78\x75\x2d\x43\xbe\xfa\x95\x35\x13\xbe\x5a\xc6\x66\xa6\x37\x2e\xc7\xe6\x15\x7b\xd3\xfc\x43\xc8\x3b\x9b\xd3\x11\x7e\xe3\xfc\xa7\xb1\xf5\xcb\xe9\xf8\x33\x36\xae\xac\x26\xe0\x1b\x1d\x9f\xd1\xc7\x40\x52\xb0\x17\x89\x7a\xc9\x4f\x36\x36\xc6\x67\x59\xbb\xc9\x85\x91\x2c\x8c\xc6\xfd\x97\xef\x2b\x29\xda\xb5\xed\x23\x16\xa3\x63\x31\xfb\xd2\x0f\x29\x79\x7e\x34\x3a\xd6\x38\x36\xfb\x9b\xd4\x27\x13\x14\xeb\x62\xcf\x09\x2e\xe5\x7b\xda\xc7\xe6\x70\x45\x7a\x95\xa9\x28\xfe\xad\xfe\x05\x00\x00\xff\xff\x80\xc1\x57\xe0\x90\x03\x00\x00")

func typeStudyGqlBytes() ([]byte, error) {
	return bindataRead(
		_typeStudyGql,
		"type/study.gql",
	)
}

func typeStudyGql() (*asset, error) {
	bytes, err := typeStudyGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "type/study.gql", size: 912, mode: os.FileMode(420), modTime: time.Unix(1525794756, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _typeUserGql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x92\x4f\x8b\xd4\x40\x10\xc5\xef\xf9\x14\x4f\xf6\xe0\x45\xf6\x03\xe4\xb6\xe2\x61\x07\x46\x59\xb2\x89\x9e\x3b\xe9\xca\x74\x49\xff\x09\xd5\x15\x83\x88\xdf\x5d\xba\xe9\x45\xd4\x85\x19\x3c\xd5\xa3\xc2\xfb\xf1\x2a\xaf\xef\x30\xd0\x26\x94\x29\x6a\x86\xc1\x9e\x49\xee\x3b\xfd\xbe\x11\xa6\x4c\x02\x0e\x9b\xa7\x50\x3f\x7e\x4a\x96\xde\x61\x8a\xbc\x26\x09\x03\xe5\xb4\xcb\x42\xe7\xb4\x18\x35\xb3\x27\xfc\xe8\x80\x3b\x8c\x8e\x2a\xe3\x6d\xc6\xb6\xcf\x9e\x17\x6c\x92\x56\xf6\x84\x99\xd3\x7d\x87\x32\x7a\x3c\xab\x70\xbc\xbc\xe9\xae\x5b\x60\x32\x1e\xc7\x8f\xe7\x66\x2d\xb2\xaf\x8b\x66\x3e\x59\x8a\xca\x2b\x53\x86\x3a\x82\x35\x4a\x30\xd1\x42\x39\x10\x0e\x47\xb1\xae\xd3\xfc\x95\x16\xc5\x61\x32\x16\x21\xa3\x64\x0b\xaf\xc9\x07\xed\x31\x72\xa0\xab\x71\x28\x18\xf6\xc5\x58\xc5\x1f\x57\xb0\xed\x71\xfa\xd0\x08\x5f\x1c\xa9\x23\x41\x12\xc4\xa4\x50\xc7\xb9\x12\xc1\xe5\x0f\x67\x2e\x11\x6d\xe0\xc8\x59\xc5\x68\x92\x82\xe4\xfc\xcc\x4a\x0f\x65\xdd\xe3\x7d\x4a\x9e\x4c\xbc\x01\x57\x8e\xfb\xc6\x74\x70\xbc\xb4\xea\x0a\xea\x33\xd3\x41\xf2\x37\xe7\xe5\xb0\x68\x42\x15\x16\x9a\xe0\xd3\x85\x63\x71\x55\x71\x73\x33\x85\x51\x5c\x65\xbe\x62\x7a\x1c\xc7\x27\x6c\x46\x1d\xd6\x24\xbf\x13\x17\x87\xb4\x87\xf3\x64\xd4\xf5\x98\x86\xd3\xff\x16\xe9\x4d\x56\xec\x9b\x7d\x69\xb3\xc9\x57\xda\xac\x69\xa6\xe1\xfc\x6f\x98\x5d\x7c\xcb\xf0\xb3\xfb\x15\x00\x00\xff\xff\x8a\xcc\xcf\x55\x09\x03\x00\x00")

func typeUserGqlBytes() ([]byte, error) {
	return bindataRead(
		_typeUserGql,
		"type/user.gql",
	)
}

func typeUserGql() (*asset, error) {
	bytes, err := typeUserGqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "type/user.gql", size: 777, mode: os.FileMode(420), modTime: time.Unix(1525462513, 0)}
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
	"input/createLesson.gql": inputCreatelessonGql,
	"input/createStudy.gql": inputCreatestudyGql,
	"input/createUser.gql": inputCreateuserGql,
	"input/deleteLesson.gql": inputDeletelessonGql,
	"input/deleteStudy.gql": inputDeletestudyGql,
	"input/deleteUser.gql": inputDeleteuserGql,
	"input/updateLesson.gql": inputUpdatelessonGql,
	"input/updateStudy.gql": inputUpdatestudyGql,
	"input/updateUser.gql": inputUpdateuserGql,
	"interfaces/deletable.gql": interfacesDeletableGql,
	"interfaces/node.gql": interfacesNodeGql,
	"interfaces/study_node.gql": interfacesStudy_nodeGql,
	"interfaces/uniform_resource_locatable.gql": interfacesUniform_resource_locatableGql,
	"interfaces/updateable.gql": interfacesUpdateableGql,
	"scalars/html.gql": scalarsHtmlGql,
	"scalars/time.gql": scalarsTimeGql,
	"scalars/uri.gql": scalarsUriGql,
	"schema.gql": schemaGql,
	"type/lesson.gql": typeLessonGql,
	"type/page_info.gql": typePage_infoGql,
	"type/study.gql": typeStudyGql,
	"type/user.gql": typeUserGql,
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
	"input": &bintree{nil, map[string]*bintree{
		"createLesson.gql": &bintree{inputCreatelessonGql, map[string]*bintree{}},
		"createStudy.gql": &bintree{inputCreatestudyGql, map[string]*bintree{}},
		"createUser.gql": &bintree{inputCreateuserGql, map[string]*bintree{}},
		"deleteLesson.gql": &bintree{inputDeletelessonGql, map[string]*bintree{}},
		"deleteStudy.gql": &bintree{inputDeletestudyGql, map[string]*bintree{}},
		"deleteUser.gql": &bintree{inputDeleteuserGql, map[string]*bintree{}},
		"updateLesson.gql": &bintree{inputUpdatelessonGql, map[string]*bintree{}},
		"updateStudy.gql": &bintree{inputUpdatestudyGql, map[string]*bintree{}},
		"updateUser.gql": &bintree{inputUpdateuserGql, map[string]*bintree{}},
	}},
	"interfaces": &bintree{nil, map[string]*bintree{
		"deletable.gql": &bintree{interfacesDeletableGql, map[string]*bintree{}},
		"node.gql": &bintree{interfacesNodeGql, map[string]*bintree{}},
		"study_node.gql": &bintree{interfacesStudy_nodeGql, map[string]*bintree{}},
		"uniform_resource_locatable.gql": &bintree{interfacesUniform_resource_locatableGql, map[string]*bintree{}},
		"updateable.gql": &bintree{interfacesUpdateableGql, map[string]*bintree{}},
	}},
	"scalars": &bintree{nil, map[string]*bintree{
		"html.gql": &bintree{scalarsHtmlGql, map[string]*bintree{}},
		"time.gql": &bintree{scalarsTimeGql, map[string]*bintree{}},
		"uri.gql": &bintree{scalarsUriGql, map[string]*bintree{}},
	}},
	"schema.gql": &bintree{schemaGql, map[string]*bintree{}},
	"type": &bintree{nil, map[string]*bintree{
		"lesson.gql": &bintree{typeLessonGql, map[string]*bintree{}},
		"page_info.gql": &bintree{typePage_infoGql, map[string]*bintree{}},
		"study.gql": &bintree{typeStudyGql, map[string]*bintree{}},
		"user.gql": &bintree{typeUserGql, map[string]*bintree{}},
	}},
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

