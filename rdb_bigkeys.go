package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/cupcake/rdb"
)

func main() {
	var (
		sizeThreshold, processThreads uint
		resultFile, rdbFile           string
		resultSep                     int
		sortResult                    bool
		sortLimit                     uint
	)
	flag.Usage = func() {
		fmt.Println("rdb_scanner v1.0 by laijunshou@gmail.com\nie: rdb_scanner -b 1024 -o 6379.csv -S -l 50 -t 3 dump6379.rdb")
		flag.PrintDefaults()
	}
	flag.UintVar(&sizeThreshold, "b", 1024, "only output keys used memory equal or greater than this size(in byte), default 1024")
	flag.StringVar(&resultFile, "o", "", "the file the result write to, default sys stdout")
	flag.UintVar(&processThreads, "t", 2, "threads to parsing rdb file, default 2")
	flag.IntVar(&resultSep, "s", 0, "seperator of result, 1: space, otherelse: comma, default 0")
	flag.BoolVar(&sortResult, "S", false, "sort keys in descending order by memory, default false")
	flag.UintVar(&sortLimit, "l", 100, "works with -S, only output top N biggest keys, default 100, max 1000")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("rdb file missing")
		flag.Usage()
		os.Exit(1)
	}
	if sortLimit > 1000 {
		fmt.Printf("-l %d > 1000, set it to 1000\n", sortLimit)
		sortLimit = 1000
	}
	rdbFile = flag.Args()[0]

	//fmt.Printf("%v %v %v %v\n",sizeThreshold, processThreads,  resultFile, rdbFile)

	rdbFh, err := os.Open(rdbFile)
	if err != nil {
		panic(fmt.Sprintf("Fail to open rdb file %s: %v", rdbFile, err))
	} else {
		defer rdbFh.Close()
	}

	var resultFh *os.File
	if resultFile == "" {
		resultFh = os.Stdout
	} else {
		resultFh, err = os.OpenFile(resultFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(fmt.Sprintf("Fail to open  file %s: %v", resultFile, err))
		}
		defer resultFh.Close()
	}
	var resultTitle string
	if resultSep == 1 {
		resultTitle = strings.Join(ResultTitleColumns, "	")
	} else {
		resultTitle = strings.Join(ResultTitleColumns, ",")
	}
	resultTitle += "\n"
	resultFh.WriteString(resultTitle)

	printChan := make(chan printStruct, 256)

	memCalback := MemCallback{}
	memCalback.Init()

	var wgCal, wgPrint sync.WaitGroup

	wgPrint.Add(1)
	if sortResult {
		go printResultSorted(&memCalback, bufio.NewWriter(resultFh), printChan, int(sizeThreshold), int(sortLimit), &wgPrint, resultSep)
	} else {
		go printResult(&memCalback, bufio.NewWriter(resultFh), printChan, int(sizeThreshold), &wgPrint, resultSep)
	}

	for i := uint(0); i < processThreads; i++ {
		wgCal.Add(1)
		go countMemory(&memCalback, printChan, int(sizeThreshold), &wgCal)
	}

	err = rdb.Decode(rdbFh, &memCalback)
	if err != nil {
		panic(err)
	}
	wgCal.Wait()
	close(printChan)

	wgPrint.Wait()
	//fmt.Println("Exit!")
}
