package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	// "path/filepath"
	"strings"
)

func usage() {
    fmt.Fprint(os.Stderr,
        "sort files by type into subdirectories\n dirsort <directory>\n")
    flag.PrintDefaults()
}

func contains[K comparable](slice []K, target K) bool {
    for _, item := range slice {
        if item == target { return true}
    }
    return false
}

type fileinfo struct {
    path string
    tpe string
    basename string
}

func getFileType(path string) string {
    stdout, err := exec.Command("file", "-b", path).Output()
    if err != nil { log.Fatal(err) }
    t := strings.ReplaceAll(strings.Split(string(stdout), " ")[0], "\n", "")
    return string(t)
}

func getFiles(dirpath string) (files []fileinfo) {
    dir, err := os.ReadDir(dirpath)
    if err != nil {log.Fatal(err)}

    for _, file := range dir {

        if !file.Type().IsDir() {
            fpath := fmt.Sprintf("%s/%s", dirpath, file.Name())
            files = append(files, fileinfo{
                path: fpath,
                tpe: getFileType(fpath),
                basename: file.Name(),
            })
        }
    }
    return files
}

func types(files []fileinfo) map[string]int {
    types := make(map[string]int)
    for _, f := range files {
        types[f.tpe] += 1
    }
    return types
}

func main() {
    var dir string
    var dry bool
    var help bool
    var yes bool
    flag.BoolVar(&dry, "dry",  false,
        "list changes to be made without making any modifications")
    flag.BoolVar(&help, "help", false, "Show help")
    flag.BoolVar(&yes, "yes", false, "assume yes to all questions")
    flag.Parse()

    fmt.Println(dry)
    fmt.Println(yes)
    fmt.Println(help)

    dir = flag.Arg(0)
    if dir == "" || help == true {
        usage()
        os.Exit(0)
    }
    
    files := getFiles(dir)
    types := types(files)

    for k, v := range types {
        if v > 1 && k != "empty" {
            fmt.Printf("%s/\n", k)
            for _, f := range files {
                if f.tpe == k {
                    fmt.Printf("\t%s\n", f.basename)
                }
            }
        }
    }

    if types["empty"] > 0 {
        fmt.Println("Empty files (to be removed):")
        for _, f := range files {
            if f.tpe == "empty" {
                fmt.Printf("\t%s\n", f.basename)
            }
        }
    }

    if !yes {
        var imput string
        fmt.Print("Make changes to files? (y)/n: ")
        fmt.Scanln(&imput)
        if imput == "n" {
            fmt.Println("aborting.")
            os.Exit(0)
        }
    }

    for k, v := range types {
        if v > 1 && k != "empty" {
            os.Mkdir(fmt.Sprintf("%s/%s", dir, k), 755)
            for _, f := range files {
                if f.tpe == k {
                    dest := fmt.Sprintf("%s/%s/%s", dir, k, f.basename)
                    switch dry {
                    case true:
                        log.Printf("%s -> %s\n", f.path, dest)
                    case false:
                        os.Rename(f.path, dest)
                    }
                }
            }
        }
    }

    for _, f := range files {
        if f.tpe == "empty" {
            switch dry {
            case true:
                log.Printf("rm %s", f.path)
                
            case false:
                os.Remove(f.path)
            }
        }
    }
}
