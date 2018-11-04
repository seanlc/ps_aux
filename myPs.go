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

func get_secs_since_midnight() int {
    secsSince1970,err := exec.Command("date", "+%s").Output()
    check(err)
    totalSecsInt,err := strconv.Atoi(strings.TrimSuffix(string(secsSince1970), "\n"))
    check(err)
    midnightSecs,err := exec.Command("date", "-d", "today 00:00:00", "+%s").Output()
    check(err)
    midnightSecsInt,err := strconv.Atoi(strings.TrimSuffix(string(midnightSecs), "\n"))
    check(err)
    return totalSecsInt - midnightSecsInt
}

func get_current_month() string {
    month, err := exec.Command("date", "+%b").Output()
    check(err)
    return strings.TrimSuffix(string(month), "\n")
}

func get_current_day() int {
    day, err := exec.Command("date", "+%d").Output()
    check(err)
    dayInt,err := strconv.Atoi(strings.TrimSuffix(string(day), "\n"))
    check(err)
    return dayInt
}

func get_prev_month(month string) string {
    lastMonth := ""
    switch month {
    case "Jan":
        lastMonth = "Dec"
    case "Feb":
        lastMonth = "Jan"
    case "Mar":
        lastMonth = "Feb"
    case "Apr":
        lastMonth = "Mar"
    case "May":
        lastMonth = "Apr"
    case "Jun":
        lastMonth = "May"
    case "Jul":
        lastMonth = "Jun"
    case "Aug":
        lastMonth = "Jul"
    case "Sep":
        lastMonth = "Aug"
    case "Oct":
        lastMonth = "Sep"
    case "Nov":
        lastMonth = "Oct"
    case "Dec":
        lastMonth = "Jan"
    default:
	lastMonth = "Err"
    }

    return lastMonth
}

func dec_day(month *string, day *int) {
    if *day == 1 {
        if *month == "May" || *month == "Jul" || *month == "Oct"  || *month == "Dec" {
	    *day = 30
	} else if  *month == "Mar" {
	    *day = 28
        } else {
	    *day = 31
        }
	*month = get_prev_month(*month)
    } else {
        *day = *day - 1
    }
}

func get_user(uid_line string, uids map[string] string) string {
    var out bytes.Buffer
    var err error
    words := strings.Fields(uid_line)
    uid, prs := uids[words[1]]
    if prs == false {
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
        uid = strings.TrimSuffix(out.String(), "\n")
	uids[words[1]] = uid
    }
    return uid
}

func parse_stat(pid string, statContents *[]string) {
    var pNameParts []string
    var pName string
    var numCloseParen int

    stat, err := os.Open("/proc/" + pid + "/stat")
    check(err)
    defer stat.Close()

    reader := bufio.NewReader(stat)

    // read PID and add to statContents
    first,err := reader.ReadString(' ')
    check(err)
    *statContents = append(*statContents, first[:len(first)-1])

    // find number of closing parenthesis
    for {
	    str, err := reader.ReadString(')')
	if err != nil {
	    break
	}
	pNameParts = append(pNameParts, str)
        numCloseParen += 1
    }

    // compose process name and add to statContents
    for _,tok := range pNameParts {
        pName += tok
    }
    *statContents = append(*statContents, pName)

    // rewind file to beginning
    stat.Seek(0, 0)

    // read past last closing paren
    for i := 0; i < numCloseParen; i++ {
	_,err := reader.ReadString(')')
	check(err)
    }

    // parse remaining fields with space as delim and store in statContent slice
    _, err = reader.ReadString(' ')
    check(err)
    for {
	str,err := reader.ReadString(' ')
	if err != nil {
	    break
	}
	*statContents = append(*statContents, str[:len(str)-1])
    }
}

func get_parm_from_file(fName string, pIndex int) float64 {
    file,err := os.Open(fName)
    check(err)
    defer file.Close()
    scanner := bufio.NewScanner(file)
    scanner.Scan()
    firstLine := scanner.Text()
    toks := strings.Fields(firstLine)
    parm,err := strconv.ParseFloat(toks[pIndex], 64)
    check(err)
    return parm
}

func get_command(PID string, fName string) string {
    cmdline := ""
    cmd,err := os.Open("/proc/" + PID + "/cmdline")
    check(err)
    defer cmd.Close()
    scanner := bufio.NewScanner(cmd)
    scanner.Scan()
    firstLine := scanner.Text()
    if firstLine != "" {
	if len(firstLine) > 40 {
	    cmdline = firstLine[0: 39]
        } else {
	    cmdline = firstLine
        }
    } else {
        cmdline = fName
	cmdline = strings.Replace(cmdline, "(", "[", -1)
	cmdline = strings.Replace(cmdline, ")", "]", -1)
    }
    return cmdline
}

func pad_time_with_zeroes(num *string) {
    if len(*num) == 1 {
        *num = "0" + *num
    }
}

func main(){
    // slices for storing /proc PID dirs
    procDir := make([]int64,0)
    procStr := make([]string, 0)
    var statContents[]string
    origCurMonth := get_current_month()
    origCurDay := get_current_day()
    curTime := get_secs_since_midnight()
    const stateIndex = 2
    const ttyIndex = 6
    const utimeIndex = 13
    const stimeIndex = 14
    const startTimeIndex = 21
    const fNameIndex = 1
    parse_stat("1", &statContents)

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

    uids := map[string]string{"initial": "0"}
    totalMem := get_parm_from_file("/proc/meminfo", 1)

    // system uptime calc
    sysUptimeSecs := get_parm_from_file("/proc/uptime", 0)

    for _, PID := range(procStr) {
	user := ""
	state := ""
	major_tty_str := ""
	tty := ""
	vsz := 0.0
	rss := 0.0
	memUse := 0.0
	totalTime := ""
	procExeTime := 0
	cmdline := ""
	startTime := ""
	curDay := origCurDay
	curMonth := origCurMonth

	parse_stat(PID, &statContents)

	// state calculation
	state = statContents[stateIndex]

	// time calculation
	userTime,err := strconv.Atoi(statContents[utimeIndex])
	systemTime,err := strconv.Atoi(statContents[stimeIndex])
	procExeTime = (userTime + systemTime) / 100
	minTimeInt := procExeTime / 60
	secTimeInt := procExeTime % 60
	minTime := strconv.Itoa(minTimeInt)
	secTime := strconv.Itoa(secTimeInt)
	pad_time_with_zeroes(&minTime)
	pad_time_with_zeroes(&secTime)
	totalTime = minTime + ":" + secTime

	// tty calculation
	i64,err := strconv.ParseInt(statContents[ttyIndex], 10, 32)
	check(err)
	minor_tty := (int32(i64) >> 8) & 0xFF
	major_tty1 := (int32(i64) >> 12) & 0xFFF00
	major_tty0 := int32(i64) & 0xFF
	major_tty := major_tty1 + major_tty0
	major_tty_str = strconv.Itoa(int(major_tty))

	if minor_tty == 136 {
	    tty += "pts/"
	} else if minor_tty == 4 {
            tty += "tty" 
	} else {
	    tty += "?"
        }

	if tty != "?" || major_tty_str != "0" {
            tty += major_tty_str
        }

	// %CPU calc
	procStartTimeTicks,err := strconv.ParseFloat(statContents[startTimeIndex], 64)
	procRunTimeSecs := sysUptimeSecs - (procStartTimeTicks / 100.0)
	cpuPercentage := ((float64(userTime) + float64(systemTime)) / procRunTimeSecs)

	// get command
	cmdline = get_command(PID, statContents[fNameIndex])

	// get startTime
        timeStarted := float64(curTime) - procRunTimeSecs
	if timeStarted > 0 {
	    mins := timeStarted / 60
	    hourStr := strconv.Itoa((int(mins / 60)))
            minStr := strconv.Itoa(int(mins) % 60)
	    pad_time_with_zeroes(&hourStr)
	    pad_time_with_zeroes(&minStr)
            startTime = hourStr + ":" + minStr
	} else {
	    daysBack := int(timeStarted / -86400) + 1
	    fmt.Println(daysBack)
            for i:= 0; i < daysBack; i++ {
	        dec_day(&curMonth, &curDay)
            }
	    startTime = curMonth + strconv.Itoa(curDay)
	}

	// TODO: rewrite use info from /proc/PID/stat instead of /proc/PID/status
	status, err := os.Open("/proc/" + PID + "/status")
	defer status.Close()
	check(err)
	scanner := bufio.NewScanner(status)
	for scanner.Scan() {
	    inputLine := scanner.Text()
	    if strings.HasPrefix(inputLine, "Uid:") {
		user = get_user(inputLine, uids)
	    }
	    if strings.HasPrefix(inputLine, "VmSize:") {
		words := strings.Fields(inputLine)
		vsz, err = strconv.ParseFloat(words[1], 64)
		check(err)
	    }
	    if strings.HasPrefix(inputLine, "VmRSS:") {
	        words := strings.Fields(inputLine)
                rss, err = strconv.ParseFloat(words[1], 64)
		check(err)
	    }
	}

        memUse = (rss / totalMem) * 100

        fmt.Fprintf(w, "%s\t%s\t%.1f\t%.1f\t%.0f\t%.0f\t%s\t%s\t%s\t%s\t%s\t\n",
                    user, PID, cpuPercentage, memUse, vsz, rss, tty, state, startTime, totalTime, cmdline)
        statContents = nil
    }
}
