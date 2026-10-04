package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ivm97/GoIrbis/irbis"
)

func main() {
	file, err := os.Open("data/test1.iso")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	for mfn := 1; ; mfn++ {
		record, err := irbis.ReadIsoRecord(file, irbis.FromAnsi)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		record.Mfn = mfn
		fmt.Println(record)
	}
}
