package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/noppawitt/pammy/youtube"
)

func main() {
	var filename string
	var resultType string

	flag.StringVar(&filename, "f", "", "Input file")
	flag.StringVar(&resultType, "t", "search", "Result type [ search | suggested ]")
	flag.Parse()

	var err error

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var infos []youtube.VideoInfo

	switch resultType {
	case "search":
		infos, err = youtube.ExtractSearchResult(f)
	case "suggested":
		infos, err = youtube.ExtractSuggestedVideos(f)
	default:
		log.Fatal("invalid type")
	}

	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "ID\tTitle\tLength\t")
	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", info.ID, info.Title, info.Duration)
	}
	w.Flush()
}
