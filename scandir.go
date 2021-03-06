package ScanDir

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Djoulzy/MovieDB"
	"github.com/Djoulzy/ScanDir/stripchar"
	"github.com/Djoulzy/Tools/clog"
	tmdb "github.com/ryanbradynd05/go-tmdb"
)

type DataSource interface {
	GetHTTPAddr() string
	GetPrefixDir() string
	GetCacheDir() string
}

var globalConf DataSource
var myDB *MovieDB.MDB

type items []fileInfos

func (i items) String() string {
	var ret string
	for _, f := range i {
		ret = fmt.Sprintf("%s%s", ret, f)
	}
	return ret
}

type pagination struct {
	totalFiles int
	totalPages int
	actualPage int
	nbPerPage  int
}

type fileInfos struct {
	FileName      string
	TMDBID        string
	Path          string
	Name          string
	Type          string
	Ext           string
	ArtworkUrl    string
	Year          string
	Langues       string
	Origine       string
	Qualite       string
	Size          int64
	ModTime       time.Time
	NBItems       int
	NBPages       int
	DisplayedPage int
	ItemsPerPage  int
	Items         items
}

func (f fileInfos) String() string {
	return fmt.Sprintf("%s [%s]: %s Size: %d items: %s\n", f.Name, f.FileName, f.Type, f.Size, f.Items)
}

//////////////////////////////////////// SORT //////////////////////////////////

type ByTitle items

func (s ByTitle) Len() int {
	return len(s)
}
func (s ByTitle) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTitle) Less(i, j int) bool {
	return []byte(s[i].Name)[0] < []byte(s[j].Name)[0]
}

type ByDate items

func (s ByDate) Len() int {
	return len(s)
}
func (s ByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByDate) Less(i, j int) bool {
	return s[i].ModTime.Before(s[j].ModTime)
}

type ByYear items

func (s ByYear) Len() int {
	return len(s)
}
func (s ByYear) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByYear) Less(i, j int) bool {
	yi, _ := strconv.Atoi(s[i].Year)
	yj, _ := strconv.Atoi(s[j].Year)
	return yi < yj
}

////////////////////////////////////////////////////////////////////////////////

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
}

func fullList(root string) {
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)
}

func isPrettyName(UglyName string) (map[string]string, bool) {
	regex := regexp.MustCompile(`(?iU)^([^_]+?)_+\((19\d{2}|20(?:0\d|1[0-9]))\)_+(multi(?:-vf[f|q])?|vf(?:[f|q])?|(?:[a-z]*)french)?_+(\d+p)?_+(bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)?_+\[(\d+)?\]\.(mkv|avi|mpe?g|mp4)$`)
	globalRule := regex.FindStringSubmatch(UglyName)

	results := make(map[string]string)
	if len(globalRule) == 8 {
		results["titre"] = globalRule[1]
		results["year"] = globalRule[2]
		results["langue"] = globalRule[3]
		results["qualite"] = globalRule[4]
		results["origine"] = globalRule[5]
		results["tmdbid"] = globalRule[6]
		results["ext"] = globalRule[7]
		return results, true
	} else {
		regex = regexp.MustCompile(`(?iU)^(.+)[.( _\t]+(?:19\d{2}|20(?:0\d|1[0-9])).*[.](.+)$`)
		// regex := regexp.MustCompile(`(?iU)^(.+?)[.( \t]*((19\d{2}|20(?:0\d|1[0-9])).*|(?:(?=\d+p|bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)..*)?[.](mkv|avi|mpe?g|mp4)$)`)
		infosBase := regex.FindStringSubmatch(UglyName)
		regex = regexp.MustCompile(`(?iU)^(?:.+?)(19\d{2}|20(?:0\d|1[0-9]))(?:.+?)$`)
		year := regex.FindStringSubmatch(UglyName)
		regex = regexp.MustCompile(`(?iU)^(?:.+?)(bluray|brrip|webrip|hdlight|dvdrip|web-dl|hdrip)(?:.+?)$`)
		origine := regex.FindStringSubmatch(UglyName)
		regex = regexp.MustCompile(`(?iU)^(?:.+?)(?:[^\d](\d+p)[^p])(?:.*?)$`)
		qualite := regex.FindStringSubmatch(UglyName)
		regex = regexp.MustCompile(`(?i)^(?:.+?)(multi(?:-vf[f|q])?|vf(?:[f|q])?|(?:\w*)french)(?:.+?)$`)
		langue := regex.FindStringSubmatch(UglyName)

		if len(infosBase) >= 2 {
			results["titre"] = infosBase[1]
		} else {
			infosBase := strings.Split(UglyName, ".")
			results["titre"] = infosBase[0]
		}
		results["ext"] = filepath.Ext(UglyName)
		if len(results["ext"]) > 0 {
			results["ext"] = strings.ToLower(results["ext"][1:])
		}

		if len(year) == 2 {
			results["year"] = year[1]
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
		results["titre"] = strings.Trim(results["titre"], " ")
		return results, false
	}
}

func renameFile(path string, from string, with map[string]string) (os.FileInfo, error) {
	titre := stripchar.StripCtlAndExtFromUnicode(with["titre"])
	id, err := myDB.GetMovieID(string(titre), with["year"])
	if err == nil {
		raw, err := myDB.GetMovieInfos(id)
		var dbMovieInfos = &tmdb.Movie{}
		err = json.Unmarshal(raw, dbMovieInfos)
		if err == nil {
			with["titre"] = dbMovieInfos.Title
			with["year"] = strings.Split(dbMovieInfos.ReleaseDate, "-")[0]
		}
		newFileName := fmt.Sprintf("%s_(%s)_%s_%s_%s_[%s].%s", with["titre"], with["year"], with["langue"], with["qualite"], with["origine"], id, with["ext"])
		clog.Info("scandir", "renameFile", "FROM: %s TO: %s", from, newFileName)
		destFile := fmt.Sprintf("%s%s", path, newFileName)
		os.Rename(fmt.Sprintf("%s%s", path, from), destFile)
		return os.Stat(destFile)
	} else {
		return nil, err
	}
}

func makeCorrectFileList(theDir string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(theDir)
	if err != nil {
		return nil, err
	}

	doTMDBCheck := strings.Contains(theDir, "FILM") || strings.Contains(theDir, "ANIME")

	var tmp []os.FileInfo
	for _, f := range files {
		fileName := f.Name()
		if filepath.HasPrefix(fileName, ".") || filepath.HasPrefix(fileName, "@") || filepath.HasPrefix(fileName, "_") || filepath.HasPrefix(fileName, "thumbs") || (filepath.Ext(fileName) == ".part") {
			continue
		}

		if !f.IsDir() {
			newInfos := f
			if doTMDBCheck {
				infos, ok := isPrettyName(fileName)
				if !ok {
					newf, err := renameFile(theDir, fileName, infos)
					if err == nil {
						newInfos = newf
					}
				}
			}
			tmp = append(tmp, newInfos)
		} else {
			tmp = append(tmp, f)
		}
	}
	return tmp, nil
}

func simpleList(prefix string, root string, base string) items {
	theDir := fmt.Sprintf("%s%s/", prefix, root)
	files, err := makeCorrectFileList(theDir)
	if err != nil {
		log.Fatal(err)
	}
	stat, _ := os.Stat(theDir)

	var zeFilez items
	for _, f := range files {
		fileName := f.Name()
		fileFullPath := fmt.Sprintf("%s%s", theDir, fileName)
		stat, _ = os.Stat(fileFullPath)
		modTime := stat.ModTime()
		tmp := fileInfos{}
		if f.IsDir() {
			tmp.Type = "folder"
			tmp.Name = fileName
			tmpfiles, _ := ioutil.ReadDir(fileFullPath)
			tmp.NBItems = len(tmpfiles)
			// tmp.Items = simpleList(prefix, fmt.Sprintf("%s/%s", root, f.Name()), fmt.Sprintf("%s/%s", base, f.Name()))
		} else {
			infos, _ := isPrettyName(fileName)
			tmp.TMDBID = infos["tmdbid"]
			tmp.Name = infos["titre"]
			tmp.Type = "file"
			tmp.Ext = infos["ext"]
			switch infos["ext"] {
			case "mkv":
				fallthrough
			case "avi":
				fallthrough
			case "mp4":
				tmp.ArtworkUrl = fmt.Sprintf("http://%s/art/%s/w185", globalConf.GetHTTPAddr(), infos["tmdbid"])
			case "epub":
				tmp.ArtworkUrl = fmt.Sprintf("http://%s/ico/epub.png", globalConf.GetHTTPAddr())
			case "jpg":
				tmp.ArtworkUrl = fmt.Sprintf("http://%s/static%s", globalConf.GetHTTPAddr(), fileFullPath)
			default:
				tmp.ArtworkUrl = fmt.Sprintf("http://%s/ico/default.png", globalConf.GetHTTPAddr())
			}
			tmp.Year = infos["year"]
			tmp.Langues = infos["langue"]
			tmp.Origine = infos["origine"]
			tmp.Qualite = infos["qualite"]
			tmp.Size = f.Size()
		}
		tmp.FileName = fileName
		tmp.Path = fmt.Sprintf("%s/%s", root, fileName)
		tmp.ModTime = modTime
		zeFilez = append(zeFilez, tmp)
	}
	return zeFilez
}

func Start(appConf DataSource, TMDB *MovieDB.MDB, root string, orderby string, asc bool, pagenum int, nbperpage int) []byte {
	globalConf = appConf
	myDB = TMDB
	base := filepath.Base(root)

	clog.Info("ScanDir", "Start", "Prefix: %s, Dir: %s, OrderBy: %s (ASC:%v) Page: %d, NbPP: %d", appConf.GetPrefixDir(), root, orderby, asc, pagenum, nbperpage)
	list := simpleList(appConf.GetPrefixDir(), root, base)

	if pagenum == 0 {
		pagenum = 1
	}
	if nbperpage == 0 {
		nbperpage = len(list)
	}

	index := nbperpage * (pagenum - 1)
	stop := index + nbperpage
	if index+nbperpage > len(list) {
		stop = len(list)
	}

	switch orderby {
	case "title":
		if asc {
			sort.Sort(ByTitle(list))
		} else {
			sort.Sort(sort.Reverse(ByTitle(list)))
		}
	case "date":
		if asc {
			sort.Sort(ByDate(list))
		} else {
			sort.Sort(sort.Reverse(ByDate(list)))
		}
	case "year":
		if asc {
			sort.Sort(ByYear(list))
		} else {
			sort.Sort(sort.Reverse(ByYear(list)))
		}
	}

	rootFiles := fileInfos{
		FileName:      base,
		Name:          base,
		Path:          root,
		Type:          "folder",
		Items:         list[index:stop],
		NBItems:       len(list),
		NBPages:       int(math.Ceil(float64(len(list)) / float64(nbperpage))),
		DisplayedPage: pagenum,
		ItemsPerPage:  nbperpage,
	}

	// json, _ := json.MarshalIndent(rootFiles, "", "    ")
	json, _ := json.Marshal(rootFiles)
	return json
}
