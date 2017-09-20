package ScanDir

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Djoulzy/Tools/clog"
)

type DataSource interface {
	GetPrefixDir() string
	GetCacheDir() string
}

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
	Ext      string
	Year     string
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

func MakePrettyName(UglyName string) map[string]string {
	regex := regexp.MustCompile(`(?iU)^(.+?)[.( \t](?:19\d{2}|20(?:0\d|1[0-9])).*[.](mkv|avi|mpe?g|mp4)$`)
	// regex := regexp.MustCompile(`(?iU)^(.+?)[.( \t]*((19\d{2}|20(?:0\d|1[0-9])).*|(?:(?=\d+p|bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)..*)?[.](mkv|avi|mpe?g|mp4)$)`)
	infosBase := regex.FindStringSubmatch(UglyName)
	regex = regexp.MustCompile(`(?iU)^(?:.+?)(19\d{2}|20(?:0\d|1[0-9]))(?:.+?)$`)
	year := regex.FindStringSubmatch(UglyName)
	regex = regexp.MustCompile(`(?iU)^(?:.+?)(bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)(?:.+?)$`)
	origine := regex.FindStringSubmatch(UglyName)
	regex = regexp.MustCompile(`(?iU)^(?:.+?)(?:[^\d](\d+p)[^p])(?:.*?)$`)
	qualite := regex.FindStringSubmatch(UglyName)
	regex = regexp.MustCompile(`(?iU)^(?:.+?)(multi|vf(?:\w*)|(?:\w*)french)(?:.+?)$`)
	langue := regex.FindStringSubmatch(UglyName)

	results := make(map[string]string)

	results["titre"] = infosBase[1]
	results["ext"] = infosBase[2]
	if len(year) == 2 {
		results["year"] = origine[1]
	}
	if len(origine) == 2 {
		results["origine"] = origine[1]
	}
	if len(qualite) == 2 {
		results["qualite"] = qualite[1]
	}
	if len(langue) == 2 {
		results["langue"] = langue[1]
	}

	NBsep := strings.Count(UglyName, " ")
	sep := " "
	tmp := strings.Count(UglyName, ".")
	if tmp > NBsep {
		NBsep = tmp
		sep = "."
	}
	tmp = strings.Count(UglyName, "_")
	if tmp > NBsep {
		NBsep = tmp
		sep = "_"
	}
	tmp = strings.Count(UglyName, "-")
	if tmp > NBsep {
		NBsep = tmp
		sep = "-"
	}
	if sep != " " {
		results["titre"] = strings.Replace(results["titre"], sep, " ", -1)
	}

	clog.Trace("", "", "%s", results)
	return results
}

func simpleList(prefix string, root string, base string) item {
	theDir := fmt.Sprintf("%s%s", prefix, root)
	files, err := ioutil.ReadDir(theDir)
	if err != nil {
		log.Fatal(err)
	}
	stat, _ := os.Stat(theDir)
	clog.Trace("", "", "%s", stat.ModTime())
	zeFilez := make(item)
	for index, f := range files {
		if filepath.HasPrefix(f.Name(), ".") || filepath.HasPrefix(f.Name(), "@") || filepath.HasPrefix(f.Name(), "thumbs") {
			continue
		}
		tmp := fileInfos{
			FileName: f.Name(),
			Path:     fmt.Sprintf("%s/%s", root, f.Name()),
			// Path: root,
		}
		if f.IsDir() {
			tmp.Type = "folder"
			tmp.Name = tmp.FileName
			tmpfiles, _ := ioutil.ReadDir(fmt.Sprintf("%s/%s", theDir, f.Name()))
			tmp.NBItems = len(tmpfiles)
			// tmp.Items = simpleList(prefix, fmt.Sprintf("%s/%s", root, f.Name()), fmt.Sprintf("%s/%s", base, f.Name()))
		} else {
			clog.Trace("", "", "%s", f.Name())
			infos := MakePrettyName(f.Name())
			tmp.Name = infos["titre"]
			tmp.Type = "file"
			tmp.Ext = infos["ext"]
			tmp.Year = infos["year"]
			tmp.Size = f.Size()
		}
		zeFilez[index] = tmp
	}

	return zeFilez
}

func Start(appConf DataSource, root string) []byte {
	clog.Info("ScanDir", "Start", "Prefix: %s, Dir: %s", appConf.GetPrefixDir(), root)
	base := filepath.Base(root)

	list := simpleList(appConf.GetPrefixDir(), root, base)
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
