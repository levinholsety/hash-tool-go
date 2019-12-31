package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"os"
	"path/filepath"

	"github.com/levinholsety/common-go/commio"
	"github.com/levinholsety/console-go/console"
)

var (
	prtFilePath = console.NewColorPrinter(console.DefaultColor, console.LightAqua)
	prtError    = console.NewColorPrinter(console.DefaultColor, console.LightRed)
	prtEmphasis = console.NewColorPrinter(console.DefaultColor, console.LightWhite)
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
	if len(filePaths) == 0 {
		return
	}
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
		if os.IsNotExist(err) {
			prtError.Println("File does not exist.")
			err = nil
		}
		return
	}
	if fileInfo.IsDir() {
		prtError.Println("Directory is not applicable.")
		return
	}
	fmt.Printf("Size: %d\n", fileInfo.Size())
	fmt.Printf("Modified: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
	algs := []*hashAlg{
		&hashAlg{"CRC32", crc32.NewIEEE()},
		&hashAlg{"MD5", md5.New()},
		&hashAlg{"SHA1", sha1.New()},
		&hashAlg{"SHA256", sha256.New()},
		&hashAlg{"SHA512", sha512.New()},
	}
	pb := console.NewProgressBar(int(fileInfo.Size()))
	n := 0
	err = commio.OpenRead(filePath, func(file *os.File) error {
		return commio.ReadBlocks(file, 0x10000, func(block []byte) (err error) {
			n += len(block)
			for _, alg := range algs {
				_, err = alg.h.Write(block)
				if err != nil {
					return
				}
			}
			pb.Progress(n)
			return
		})
	})
	if err != nil {
		return
	}
	fmt.Println()
	for _, alg := range algs {
		fmt.Printf("%s: ", alg.name)
		prtEmphasis.Printf("%s", hex.EncodeToString(alg.h.Sum(nil)))
		fmt.Println()
	}
	return
}

type hashAlg struct {
	name string
	h    hash.Hash
}
