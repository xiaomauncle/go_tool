package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func GitPull(path string, wg *sync.WaitGroup) error {
	defer func() {
		if p := recover(); p != nil {
			_ = fmt.Errorf("Error: %v\n", p)
		}
	}()

	defer wg.Done()
	fmt.Println("------------> ", path)

	chdirErr := os.Chdir(path)
	if chdirErr != nil {
		fmt.Println(chdirErr)
	}
	cmd := exec.Command("git", "pull")

	//fmt.Println(cmd.Args)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		return err
	}

	cmdErr := cmd.Start()
	if cmdErr != nil {
		fmt.Println(cmdErr)
	}
	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	for {
		line, e := reader.ReadString('\n')
		if e != nil || io.EOF == e {
			break
		}

		if "Already up to date.\n" == line {
			fmt.Println(" ==============================================     ", path)
		} else {
			fmt.Println(line)
		}
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		fmt.Println(waitErr)
		return waitErr
	}
	return nil
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func init() {
	maxProcess := runtime.NumCPU() / 2
	if maxProcess <= 0 {
		maxProcess = 1
	}
	runtime.GOMAXPROCS(maxProcess)
}

func main() {
	var targetDir string
	flag.StringVar(&targetDir, "d", getCurrentDirectory(), "target Directory")
	flag.Parse()
	fmt.Println("selectedDirectory: ", targetDir)
	fmt.Println()

	DotGit := ".git"

	var wg sync.WaitGroup

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			if strings.EqualFold(info.Name(), DotGit) {
				dir := strings.TrimRight(path, DotGit)
				fmt.Println("=====> ", dir)
				wg.Add(1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Println(r)
						}
					}()
					GitPull(dir, &wg)
				}()
			}
			return nil
		}
		return nil
	})

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	wg.Wait()

	fmt.Println()
	fmt.Println("=================================== END ")
	os.Exit(0)

}
