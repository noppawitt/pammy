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

	flag.StringVar(&filename, "f", "", "Input file")
	flag.Parse()

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	infos, err := youtube.ExtractSearchResult(f)
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "ID\tTitle\tLength\t")
	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", info.ID, info.Title, info.LengthText)
	}
	w.Flush()
}
