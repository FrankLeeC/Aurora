package test

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var path = "*********/app.log"
var dateRegexpStr = `_\d{8}_`
var dateRexexp = regexp.MustCompile(dateRegexpStr)

func TestCompressLog(t *testing.T) {
	deleteFile := make([]string, 0)
	files := getEarlyFile()
	hasError := false
	if files != nil {
		for k, v := range files {
			deleteFile = append(deleteFile, v...)
			fileName := filepath.Base(path)
			dir := filepath.Dir(path)
			files := make([]*os.File, 0)
			for _, s := range v {
				f, err := os.Open(s)
				hasError = (err != nil) || hasError
				files = append(files, f)
			}
			namePrefix := fileName
			if strings.Contains(fileName, ".") {
				i := strings.Index(fileName, ".")
				namePrefix = fileName[0:i]
			}
			zipName := dir + string(filepath.Separator) + namePrefix + k + ".zip"
			d, err := os.Create(zipName)
			hasError = (err != nil) || hasError
			if d != nil {
				defer d.Close()
			}
			w := zip.NewWriter(d)
			if w != nil {
				defer w.Close()
			}
			for _, file := range files {
				if file != nil {
					err = compress(file, "", w)
					hasError = (err != nil) || hasError
				}
			}
		}
		if !hasError {
			t.Log("====no error")
			deleteLog(deleteFile)
		}
	}
}

func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		if prefix != "" {
			header.Name = prefix + "/" + header.Name
		}
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func getEarlyFile() map[string][]string {
	result := make(map[string][]string)
	dayStr := time.Now().AddDate(0, 0, -2).Format("20060102") + " 00:00:00"
	pointTime, _ := time.ParseInLocation("20060102 15:04:05", dayStr, time.Local)
	point := pointTime.Unix()
	dir := filepath.Dir(path)
	fileName := filepath.Base(path)
	filePrefix := fileName
	fileSuffix := ""
	if strings.Contains(fileName, ".") {
		i := strings.Index(fileName, ".")
		filePrefix = fileName[0:i]
		fileSuffix = `\.` + fileName[i+1:]
	}
	fl, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	var fileRegexpStr = `^` + filePrefix + `_\d{8}_\d+` + fileSuffix + `$`
	var logRegex = regexp.MustCompile(fileRegexpStr)
	for _, f := range fl {
		if strings.HasPrefix(f.Name(), filePrefix) && logRegex.MatchString(f.Name()) {
			loc := dateRexexp.FindIndex([]byte(f.Name()))
			day := f.Name()[loc[0]+1 : loc[1]-1]
			fileTime, _ := time.ParseInLocation("20060102", day, time.Local)
			t := fileTime.Unix()
			if t < point {
				if _, contains := result[day]; contains {
					v := result[day]
					v = append(v, dir+string(filepath.Separator)+f.Name())
					result[day] = v
				} else {
					s := make([]string, 0)
					s = append(s, dir+string(filepath.Separator)+f.Name())
					result[day] = s
				}
			}
		}
	}
	return result
}

func deleteLog(file []string) {
	for _, s := range file {
		os.Remove(s)
	}
}
