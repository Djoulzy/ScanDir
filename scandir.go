package ScanDir

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

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
	// var res = it.replaceAll("_", " ")
	regex1 := regexp.MustCompile(`(?iU)^(.+?)[.( \t]*(19\d{2}|20(?:0\d|1[0-9])).*(\d+p).*(bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip).*[.](mkv|avi|mpe?g|mp4)$`)
	// var regex2 = regexp.MustCompile(`^(.+?)(bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)(?:.+?)$`)
	// var regex3 = regexp.MustCompile(`^(.+?)(?:[^\d](\d+p)[^p])(?:.*?)$`)
	// var regex4 = regexp.MustCompile(`^(?:.+?)(multi|vf(?:\w*)|(?:\w*)french)(?:.+?)$`)

	clog.Trace("", "", "%q", regex1.FindStringSubmatch(UglyName))
	//
	//    var out = new Array(5);
	//
	//    if ( m = regex.exec(res) ) {
	// 	   out[0] = titlelize(m[1]) || '-'; // Title
	// 	   out[1] = m[3] || '-';	// Year
	// 	   tmp = m[2].replaceAll("\\.", " ");
	// 	//    tmp = tmp.replaceAll("-", " ");
	// 	   if ( n = regex2.exec(tmp) ) {
	// 		   out[2] = n[2] || '-'; // Origine
	// 		   if ( o = regex3.exec(n[1]) ) {
	// 			   out[3] = o[2] || '-'; // Qualite
	// 		   }
	// 	   }
	// 	   if ( p = regex4.exec(tmp) ) out[4] = p[1] || '-'; // Langue
	// 	   else out[4] = '-';
	//    } else {
	// 	   out[0] = '<font color="red">No match</font>';
	//    }
	//    //the replace is an hack to remove html in live input text
	// //    return (html) ? out : out.replace(/<[^>]+>|&[^;]+;/g,'');
	// return out;
	return ""
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
			// infos := MakePrettyName(f.Name())
			tmp.Name = tmp.FileName
			tmp.Type = "file"
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
