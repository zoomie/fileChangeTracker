package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const seperator = " "

type store map[string][]byte

func (s store) add(path string, info os.FileInfo, err error) error {
	if info.IsDir() || strings.Contains(path, "freezeVersions") {
		return nil
	}
	content, _ := ioutil.ReadFile(path)
	hash := sha1.Sum(content)
	s[path] = hash[:]
	return nil
}

func getFromCurrent(path string) store {
	s := store{}
	filepath.Walk(path, s.add)
	return s
}

func saveToFile(data store, dotFreezeVersions string) {
	savePath := filepath.Join(dotFreezeVersions, time.Now().Format(time.RFC3339))
	file, _ := os.Create(savePath)
	defer file.Close()
	for path, hash := range data {
		file.WriteString(path)
		file.WriteString(seperator)
		file.Write(hash[:])
		file.WriteString("\n")
	}
}

func openFromFile(path string) store {
	file, _ := os.Open(path)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	s := store{}
	for scanner.Scan() {
		line := scanner.Text()
		lineSplit := strings.Split(line, seperator)
		s[lineSplit[0]] = []byte(lineSplit[1])
	}
	return s
}

func openMostRecent(path string) store {
	files, _ := ioutil.ReadDir(path)
	if len(files) == 0 {
		return store{}
	}
	var latestTime time.Time
	var latestFileInfo os.FileInfo
	for _, fileInfo := range files {
		curTime, _ := time.Parse(time.RFC3339, fileInfo.Name())
		if curTime.After(latestTime) {
			latestTime = curTime
			latestFileInfo = fileInfo
		}
	}
	lastSavePoint := filepath.Join(path, latestFileInfo.Name())
	return openFromFile(lastSavePoint)
}

func checkAndDisplayDiff(before, after store) {
	// colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	colorYellow := "\033[33m"
	for beforePath, beforeHash := range before {
		afterHash, ok := after[beforePath]
		if ok {
			if !bytes.Equal(beforeHash, afterHash) {
				fmt.Println(string(colorYellow), beforePath, "has been changed")
			}
		} else {
			fmt.Println(string(colorRed), beforePath, "has been deleted")
		}
	}
	for afterPath := range after {
		_, ok := before[afterPath]
		if !ok {
			fmt.Println(string(colorGreen), afterPath, "has been added")
		}
	}
}

func main() {
	path := os.Args[1]
	freezeVersions := filepath.Join(path, ".freezeVersions")
	os.MkdirAll(freezeVersions, 0777)
	before := openMostRecent(freezeVersions)
	after := getFromCurrent(path)
	checkAndDisplayDiff(before, after)
	saveToFile(after, freezeVersions)
}
