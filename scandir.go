package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type item map[int]fileInfos

func (i item) String() string {
	var ret string
	for _, f := range i {
		ret = fmt.Sprintf("%s%s", ret, f)
	}
	return ret
}

type fileInfos struct {
	FileName string
	Name     string
	Type     string
	Path     string
	Size     int64
	Items    item
}

func (f fileInfos) String() string {
	return fmt.Sprintf("%s [%s]: %s Size: %d items: %s\n", f.Name, f.FileName, f.Type, f.Size, f.Items)
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
}

func fullList(root string) {
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)
}

func MakePrettyName(UglyName string) string {
	return UglyName
}

func simpleList(root string, base string) item {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	zeFilez := make(item)
	for index, f := range files {
		if filepath.HasPrefix(f.Name(), ".") {
			continue
		}
		tmp := fileInfos{
			FileName: f.Name(),
			Name:     MakePrettyName(f.Name()),
			Path:     fmt.Sprintf("%s/%s", base, f.Name()),
		}
		if f.IsDir() {
			tmp.Type = "folder"
			tmp.Items = simpleList(fmt.Sprintf("%s/%s", root, f.Name()), fmt.Sprintf("%s/%s", base, f.Name()))
		} else {
			tmp.Type = "file"
			tmp.Size = f.Size()
		}
		zeFilez[index] = tmp
	}

	return zeFilez
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	base := filepath.Base(root)

	rootFiles := fileInfos{
		FileName: base,
		Name:     base,
		Path:     base,
		Type:     "folder",
		Items:    simpleList(root, base),
	}

	json, _ := json.MarshalIndent(rootFiles, "", "    ")
	fmt.Printf("%s", json)
}
