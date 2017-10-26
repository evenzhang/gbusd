package gbusd

import (
	"bufio"
	"fmt"
	"github.com/evenzhang/gbusd/gbusd/res"
	"github.com/evenzhang/gbusd/gserv/log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var cliUsageStr = `
Help: > user-cmd

    setmeta  <file>:<pos>   Set pos info   eg: setmeta edu-mysql-bin.000001:21499
    meta    / m             Show meta info
    help    / h             Show this message
    version / v             Show version
    quit    / q             Exit
`

func (this *BusdDaemon) cliShowMetaInfo() {
	fmt.Printf("*****************\n  Enter cli mode\n*****************\n")

	masterFirstPos, err := this.master.GetFirstPos()

	masterCurPos, err := this.master.GetPos()
	fmt.Printf("Master Info:\n\tcurrent pos => %s:%d \terr => %v\tsql:show master status\n", masterCurPos.FileName, masterCurPos.Pos, err)

	fmt.Printf("\t first pos  => %s:%d\t\terr => %v\tsql:show binlog events limit 1\n", masterFirstPos.FileName, masterFirstPos.EndPos, err)

	metaExist, err := this.meta.IsExist()
	fmt.Printf("Redis Meta Info:\n\texist\t    => %t\t\t\t\terr => %v\n", metaExist, err)
	if metaExist {
		data, e := this.meta.Get()
		fmt.Printf("\tcurrent-pos => %s:%d\terr => %v\n", data.BinLogFileName, data.BinLogPos, e)
	}
}
func (this *BusdDaemon) cliSetMeta(userCmd string, items []string) {
	if len(items) < 2 {
		fmt.Println("param num error", items)
		return
	}

	paramstr := strings.TrimSpace(userCmd[len("setmeta"):len(userCmd)])

	params := strings.Split(paramstr, ":")
	if len(params) != 2 {
		fmt.Println("param format error", params)
		return
	}

	pos, _ := strconv.Atoi(params[1])
	log.Infoln("cli-setmeta", params[0], uint32(pos), 0)
	this.meta.Save(params[0], uint32(pos), 0)

	fmt.Println("    set succ")
}
func (this *BusdDaemon) enterCliMode() {
	fmt.Println(cliUsageStr)

	match, _ := regexp.MatchString("setmeta  [a-z][0-9]:", "setmeta   	    edu-mysql-bin.000001:214923")
	fmt.Println(match)

	for {
		fmt.Printf("> ")
		userCmd := cliReadCmdLine()

		items := strings.Split(userCmd, " ")
		if len(items) < 1 {
			continue
		}

		if len(items[0]) == 0 {
			continue
		}

		switch items[0] {
		case "h":
			fallthrough
		case "help":
			fmt.Println(cliUsageStr)

		case "setmeta":
			this.cliSetMeta(userCmd, items)
		case "m":
			fallthrough
		case "meta":
			this.cliShowMetaInfo()

		case "v":
			fallthrough
		case "version":
			fmt.Println("  version:", res.RES_PROJ_VERSION)

		case "q":
			fallthrough
		case "quit":
			return

		default:
			fmt.Println("-busd-cli: ", userCmd, ": command not found")
			fmt.Println(cliUsageStr)
		}
	}
}
func cliReadCmdLine() string {
	reader := bufio.NewReader(os.Stdin)
	data, _, _ := reader.ReadLine()
	return strings.TrimSpace(string(data))
}
