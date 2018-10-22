package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "strconv"
    "sort"
    "text/tabwriter"
    "os"
)

func main(){
    procDir := make([]int64,0)
    procStr := make([]string, 0)
    w := new(tabwriter.Writer)
    w.Init(os.Stdout, 8, 8, 0, '\t', 0)
    // gets all files in directory /proc
    files, err := ioutil.ReadDir("/proc")
    if err != nil {
        log.Fatal(err)
    }

    // adds all files that are numbers in files to procDir slice
    for _,file := range(files) {
	    if n,err := strconv.ParseInt(file.Name(),10,64); err == nil {
                procDir = append(procDir, n)
	}
    }

    // sorts procDir in ascending order
    sort.Slice(procDir, func(i,j int) bool {return procDir[i] < procDir[j]})

    // converts all int64 in procDir to string and stores in procStr slice
    for _, name := range(procDir) {
        procStr = append(procStr, strconv.Itoa(int(name)))
    }

    defer w.Flush()

    fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
                "USER", "PID", "%CPU", "%MEM", "VSZ", "RSS", "TTY", "STAT", "START", "TIME", "COMMAND")

    for _, PID := range(procStr) {
        fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
                    "", PID, "", "", "", "", "", "", "", "", "")
    }

}
