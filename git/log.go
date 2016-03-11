package git

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

const l = `commit 878df59ace4d844cd3702284231d6271a4a1aff6
Author: Jan Halfar <jan@bestbytes.com>
Date:   Fri Apr 24 10:15:27 2015 +0200

    docs tests with pandoc

186	0	docs/CA-Proxy.html
327	0	docs/caproxy.8
6	1	server/server.go
 create mode 100644 docs/CA-Proxy.html
 create mode 100644 docs/caproxy.8


 2       2       pkg/rpm/etc/caproxy.conf.dist
 4       6       sbca/client.go
 1       1       server/bindata.go
`

type Change struct {
	File    string
	Added   int
	Removed int
}

type LogItem struct {
	ID      string
	Author  string
	Date    time.Time
	Changes []*Change
	Message string
}

func (li *LogItem) stats() (added int, removed int, files []string) {
	files = []string{}
	for _, c := range li.Changes {
		removed += c.Removed
		added += c.Added
		files = append(files, c.File)
	}
	return
}

func trim(str string) string {
	return strings.Trim(str, "  ")
}
func numChanges(str string) int {
	if str == "-" {
		return 0
	}
	n, _ := strconv.Atoi(str)
	return n
}

func FindRepos(p string, repos []string) error {
	listing, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}
	for _, f := range listing {
		pathname := path.Join(p, f.Name())
		if f.IsDir() && f.Name() == ".git" {
			repos = append(repos, pathname)
		} else if f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			err := FindRepos(pathname, repos)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CSV(dir string) error {
	log, err := Log(dir)
	if err != nil {
		return err
	}

	w := csv.NewWriter(os.Stdout)

	writeLine := func(line []string) error {
		if err := w.Write(line); err != nil {
			return err
		}
		return nil
	}
	writeLine([]string{
		"Date", "Author", "Message", "Added", "Removed", "Files affected",
	})
	for _, item := range log {
		added, removed, files := item.stats()
		data := []interface{}{
			item.Date, item.Author, item.Message, added, removed, len(files),
		}
		record := []string{}
		for _, d := range data {
			record = append(record, fmt.Sprint(d))
		}
		err = writeLine(record)
		if err != nil {
			return err
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func Log(dir string) (log []*LogItem, err error) {
	cmd := exec.Command("git", "log", "--numstat", "--pretty=medium", "--summary", "--date=local")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	log = []*LogItem{}
	var item *LogItem
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "commit") && len(line) == 47:
			if item != nil {
				item.Message = strings.Trim(item.Message, "\n")
				log = append(log, item)
			}
			item = &LogItem{
				ID:      strings.TrimPrefix(line, "commit "),
				Changes: []*Change{},
			}
		case strings.HasPrefix(line, "Author:"):
			item.Author = trim(strings.TrimPrefix(line, "Author:"))
		case strings.HasPrefix(line, "Date:"):
			item.Date, err = time.Parse("Mon Jan 2 15:04:05 2006", trim(strings.TrimPrefix(line, "Date:")))
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, "    "):
			item.Message += strings.Trim(line, "    ") + "\n"

		case len(strings.Split(line, string(9))) == 3:
			parts := strings.Split(line, string(9))
			for pi, part := range parts {
				parts[pi] = trim(part)
			}
			item.Changes = append(item.Changes, &Change{
				Added:   numChanges(parts[0]),
				Removed: numChanges(parts[1]),
				File:    parts[2],
			})
		}
	}
	return log, nil
}
