package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/levinholsety/common-go/comm"
	"github.com/levinholsety/console-go/console"
)

var (
	prtFilePath = console.NewColorPrinter(console.DefaultColor, console.LightAqua)
	prtError    = console.NewColorPrinter(console.DefaultColor, console.LightRed)
	prtEmphasis = console.NewColorPrinter(console.DefaultColor, console.LightWhite)
)

var (
	errNotAFile = errors.New("not a file")
)

func main() {
	if len(os.Args) > 1 {
		hashFiles(os.Args[1:])
		return
	}
	printHelp()
}

func printHelp() {
	fmt.Printf("%s [files...]\n", filepath.Base(os.Args[0]))
}

func hashFiles(filePaths []string) {
	for _, filePath := range filePaths {
		prtFilePath.Println(filePath)
		err := hashFile(filePath)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}
	return
}

func hashFile(filePath string) (err error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		err = errNotAFile
		return
	}
	fmt.Printf("Size: %d\n", fileInfo.Size())
	fmt.Printf("Modified: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
	algGroup := hashAlgGroup{
		{name: "CRC32", h: crc32.NewIEEE()},
		{name: "MD5", h: md5.New()},
		{name: "SHA1", h: sha1.New()},
		{name: "SHA256", h: sha256.New()},
		{name: "SHA512", h: sha512.New()},
	}
	err = console.ExecuteWithProgressBar(func(bar *console.ProgressBar) error {
		bar.SetSpeedCalculator(func(n int64, elapsed time.Duration) string {
			if elapsed == 0 {
				return ""
			}
			return "@ " + comm.FormatIOSpeed(comm.CalculateIOSpeed(n, elapsed), 0)
		})
		return comm.OpenRead(filePath, func(file *os.File) error {
			return comm.ReadStream(file, 0x20000, func(buf []byte) (err error) {
				algGroup.write(buf)
				bar.AddProgress(int64(len(buf)))
				return
			})
		})
	}, fileInfo.Size())
	if err != nil {
		return
	}
	algGroup.print()
	return
}

type hashAlg struct {
	name string
	h    hash.Hash
}

func (p *hashAlg) write(buf []byte, wg *sync.WaitGroup) {
	p.h.Write(buf)
	wg.Done()
}

func (p *hashAlg) print() {
	fmt.Printf("\n%s: ", p.name)
	prtEmphasis.Printf("%s", hex.EncodeToString(p.h.Sum(nil)))
}

type hashAlgGroup []*hashAlg

func (p hashAlgGroup) write(buf []byte) {
	wg := new(sync.WaitGroup)
	for _, alg := range p {
		wg.Add(1)
		go alg.write(buf, wg)
	}
	wg.Wait()
}

func (p hashAlgGroup) print() {
	for _, alg := range p {
		alg.print()
	}
}
