package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func printTable(headers []interface{}, columns [][]interface{}) {
	tbl := table.New(headers...)
	tbl.WithHeaderFormatter(color.New(color.FgHiGreen, color.Bold, color.Underline).SprintfFunc())

	for _, col := range columns {
		tbl.AddRow(col...)
	}
	tbl.Print()
}

// Returns the HTTP address from the raft address.
// HTTP port = Raft port - 100
func getHTTPAddressFromRaft(hostandport string) string {
	// host:port
	ss := strings.Split(hostandport, ":")
	if len(ss) != 2 {
		return ""
	}
	port, _ := strconv.Atoi(ss[1])
	port -= 100
	return fmt.Sprintf("%s:%d", ss[0], port)
}
