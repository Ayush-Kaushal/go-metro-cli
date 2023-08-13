package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MetroRoute struct {
	Status      int      `json:"status"`
	Line1       []string `json:"line1"`
	Line2       []string `json:"line2"`
	Interchange []string `json:"interchange"`
	LineEnds    []string `json:"lineEnds"`
	Path        []string `json:"path"`
	Time        float64  `json:"time"`
}

var caser = cases.Title(language.English)

func checkStatus(code int) {
	switch code {
	case 200:
		return
	case 204:
		panic("Same source and destination")
	case 400:
		panic("Undefined source or destination")
	case 406:
		panic("Invalid source and destination")
	case 4061:
		panic("Invalid source")
	case 4062:
		panic("Invalid destination")
	}
}

func getColorOrder(line1, line2 []string) []string {
	var colourOrder []string
	colourOrder = line1[:]

	if len(line2) > 0 {
		colourOrder = append(colourOrder, line2[len(line2)-1])
	}

	for i, c := range colourOrder {
		if strings.Contains(c, "branch") {
			idx := strings.Index(c, "branch")
			colourOrder[i] = c[:idx]
		}
	}

	return colourOrder
}

func printStationInColor(line, station string) {
	// line = caser.String(line)
	station = caser.String(station)

	reset := "\033[0m"

	switch line {
	case "blue":
		color.Blue(station)

	case "red":
		color.Red(station)

	case "green":
		color.Green(station)

	case "yellow":
		yellow := "\033[38;5;226m"
		fmt.Printf("%s%s%s\n", yellow, station, reset)

	case "magenta":
		color.Magenta(station)

	case "violet":
		violet := "\033[38;5;162m"
		fmt.Printf("%s%s%s\n", violet, station, reset)

	case "pink":
		pink := "\033[38;5;213m"
		fmt.Printf("%s%s%s\n", pink, station, reset)

	case "orange":
		orange := color.New(color.FgYellow)
		orange.Println(station)

	case "grey":
		grey := color.New(color.FgBlack, color.FgWhite)
		grey.Println(station)
	}
}

func printInterchangeMessage(lineEnd, line string) {
	lineEnd = caser.String(lineEnd)
	line = caser.String(line)

	mssg := fmt.Sprintf("Change train >> %s Line | Towards %s", line, lineEnd)

	br := strings.Repeat("-", len(mssg))

	bg := color.New(color.FgRed, color.Bold).Add(color.BgCyan)

	color.Cyan(br)
	bg.Printf("%s", mssg)
	fmt.Println()
	color.Cyan(br)
}

func main() {
	source := flag.String("start", "", "Source")
	destination := flag.String("end", "", "Destination")
	time := flag.Bool("t", false, "Display total time")

	flag.Parse()

	values := url.Values{}
	values.Add("from", *source)
	values.Add("to", *destination)

	url := "https://us-central1-delhimetroapi.cloudfunctions.net/route-get?" + values.Encode()

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var metroRoute MetroRoute
	err = json.Unmarshal(body, &metroRoute)
	if err != nil {
		panic(err)
	}

	checkStatus(metroRoute.Status)

	colorOrder := getColorOrder(metroRoute.Line1, metroRoute.Line2)

	startStation := caser.String(metroRoute.Path[0])
	startLine := caser.String(colorOrder[0])

	mssg := fmt.Sprintf("Board train >> %s Line | Towards %s", startLine, startStation)
	br := strings.Repeat("-", len(mssg))
	bg := color.New(color.FgRed, color.Bold).Add(color.BgCyan)
	color.Cyan(br)
	bg.Printf("%s", mssg)
	fmt.Println()
	color.Cyan(br)

	var changeIndex, endIndex, colorIndex int

	for _, station := range metroRoute.Path {

		printStationInColor(colorOrder[colorIndex], station)

		if changeIndex < len(metroRoute.Interchange) && station == metroRoute.Interchange[changeIndex] {

			changeIndex++
			endIndex++
			colorIndex++

			printInterchangeMessage(metroRoute.LineEnds[endIndex], colorOrder[colorIndex])
		}
	}

	hours := int(metroRoute.Time / 60)
	minutes := int(math.Mod(metroRoute.Time, 60))

	if *time {
		color.Cyan(br)
		color.Red("Travelling Time >> %v hours %v minutes", hours, minutes)
		color.Cyan(br)
	}
}
