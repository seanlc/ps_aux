package main

import (
    "fmt"
    "io/ioutil"
    "strconv"
    "sort"
    "text/tabwriter"
    "os"
    "bufio"
    "strings"
    "bytes"
    "os/exec"
)

func check(e error){
    if e != nil {
        panic(e)
    }
}

func get_user(uid_line string) string {
    var out bytes.Buffer
    var err error
    words := strings.Fields(uid_line)
    getentCmd := exec.Command("getent", "passwd", words[1])
    cutCmd := exec.Command("cut", "-d:", "-f1")
    cutCmd.Stdin, err = getentCmd.StdoutPipe()
    check(err)
    cutCmd.Stdout = &out
    err = cutCmd.Start()
    check(err)
    err = getentCmd.Run()
    check(err)
    err = cutCmd.Wait()
    return strings.TrimSuffix(out.String(), "\n")
}

func main(){
    // slices for storing /proc PID dirs
    procDir := make([]int64,0)
    procStr := make([]string, 0)

    // tab writer
    w := new(tabwriter.Writer)
    w.Init(os.Stdout, 8, 8, 0, '\t', 0)

    // gets all files in directory /proc
    files, err := ioutil.ReadDir("/proc")
    check(err)

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
	fileHandle, err := os.Open("/proc/" + PID + "/status")
	check(err)
	scanner := bufio.NewScanner(fileHandle)
	var user string
	for scanner.Scan() {
	    if strings.HasPrefix(scanner.Text(), "Uid:") {
		// get uid from line and uses getent to get user id
		user = get_user(scanner.Text())
	    }
	}

        fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
                    user, PID, "", "", "", "", "", "", "", "", "")
    }

}
