package ScanDir

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Djoulzy/Tools/clog"
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
	NBItems  int
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

func simpleList(prefix string, root string, base string) item {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s%s", prefix, root))
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
			Path:     fmt.Sprintf("%s/%s", root, f.Name()),
			// Path: root,
		}
		if f.IsDir() {
			tmp.Type = "folder"
			tmpfiles, _ := ioutil.ReadDir(fmt.Sprintf("%s%s", prefix, root))
			tmp.NBItems = len(tmpfiles)
			// tmp.Items = simpleList(prefix, fmt.Sprintf("%s/%s", root, f.Name()), fmt.Sprintf("%s/%s", base, f.Name()))
		} else {
			tmp.Type = "file"
			tmp.Size = f.Size()
		}
		zeFilez[index] = tmp
	}

	return zeFilez
}

func Start(prefix string, root string) []byte {
	clog.Info("ScanDir", "Start", "Prefix: %s, Dir: %s", prefix, root)
	base := filepath.Base(root)

	list := simpleList(prefix, root, base)
	rootFiles := fileInfos{
		FileName: base,
		Name:     base,
		Path:     root,
		Type:     "folder",
		Items:    list,
		NBItems:  len(list),
	}

	// json, _ := json.MarshalIndent(rootFiles, "", "    ")
	json, _ := json.Marshal(rootFiles)
	return json
}
