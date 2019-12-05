package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	fileSym     string = "├───"
	lastFileSym string = "└───"
)

func getFileNames(path string, printFiles bool) (fNames []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if printFiles {
		fNames, err = f.Readdirnames(-1)
	} else {
		namesPath, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}
		for _, val := range namesPath {
			if val.IsDir() {
				fNames = append(fNames, val.Name())
			}
		}
	}
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(fNames)
	return
}

func recDirTree(output io.Writer, path string, printFiles bool, s string) error {
	fNames, err := getFileNames(path, printFiles)
	if err != nil {
		return err
	}

	for _, name := range fNames {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			news := printDir(output, s, fNames, fileInfo)
			recDirTree(output, filename, printFiles, news)
		}
		if !fileInfo.IsDir() && printFiles {
			printDirFile(output, s, fNames, fileInfo)
		}
	}
	return nil
}

func printDir(output io.Writer, s string, fNames []string, fileInfo os.FileInfo) string {
	if fNames[len(fNames)-1] == fileInfo.Name() {
		fmt.Fprint(output, s, lastFileSym, fileInfo.Name(), "\n")
		return s + "\t"
	}
	fmt.Fprint(output, s, fileSym, fileInfo.Name(), "\n")
	return s + "│\t"
}

func printDirFile(output io.Writer, s string, fNames []string, fileInfo os.FileInfo) {
	if fNames[len(fNames)-1] == fileInfo.Name() {
		s = fmt.Sprint(s, lastFileSym, fileInfo.Name(), " (")
	} else {
		s = fmt.Sprint(s, fileSym, fileInfo.Name(), " (")
	}
	if fileInfo.Size() == 0 {
		s = fmt.Sprint(s, "empty)\n")
	} else {
		s = fmt.Sprint(s, fileInfo.Size(), "b)\n")
	}
	fmt.Fprintf(output, s)
}

func dirTree(output io.Writer, path string, printFiles bool) error {
	recDirTree(output, path, printFiles, "")
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
