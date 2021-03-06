package ripper

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

// Default plugins used in AutoRun
var (
	DefaultPlugins     []*Plugin
	DefaultPreFilters  []*PreFilter
	DefaultPostFilters []*PostFilter
)

// flags
var (
	input    = ""
	output   = ""
	column   = ""
	dic      = ""
	quote    = ""
	replace  = false
	show     = false
	debug    = false
	progress = 30
)

// AutoRun creates *Ripper from CLI flags and run it
func AutoRun() {
	err := InitFlags()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	r, err := newDefaultRipper()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer r.Close()

	Run(r)
}

// Run runs text processing
func Run(r *Ripper) {

	go func() {
		tick := time.Tick(time.Duration(progress) * time.Second)
		prev := 0
		for {
			select {
			case <-tick:
				cur := r.GetCurrentPosition()
				fmt.Printf("[%s] line: %d, tps: %d\n", time.Now().Format("2006-01-02 15:04:05"), cur, (cur-prev)/progress)
				prev = cur
			}
		}
	}()

	fmt.Println("read and write lines...")

	err := r.ReadAndWriteLines()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("finish process")
}

// InitFlags initializes flag args
func InitFlags() error {
	err := parseFlags()
	if err != nil {
		return err
	}

	return checkFlags()
}

func parseFlags() error {
	flag.StringVar(&input, "input", "", "read file")
	flag.StringVar(&output, "output", "", "output file")
	flag.StringVar(&column, "column", "", "target column name")
	flag.StringVar(&dic, "dic", "", "custom dictionaly path(ipa dictionaly)")
	flag.StringVar(&quote, "quote", "", "columns to add double-quotes (separated by comma)")
	flag.BoolVar(&replace, "replace", false, "replace text column")
	flag.BoolVar(&show, "show", false, "print separated words to console")
	flag.BoolVar(&debug, "debug", false, "print debug result to console")
	flag.IntVar(&progress, "progress", 30, "print current progress (sec)")

	flag.Parse()
	return nil
}

func checkFlags() error {
	switch {
	case input == "":
		return fmt.Errorf("no input file\nuse -input <input file path>\n")
	case output == "" && !show && !debug:
		return fmt.Errorf("no output file\nuse -output <output file path>\n")
	case column == "":
		return fmt.Errorf("no column name\nuse -column <column name>\n")
	}

	return nil
}

func newDefaultRipper() (*Ripper, error) {
	var r *Ripper
	var err error

	switch {
	case output == "":
		// read only
		r, err = NewWithReaderFromFile(input, column)
	default:
		r, err = NewFromFiles(input, output, column)
	}
	if err != nil {
		return nil, err
	}

	// set output options
	if show {
		r.ShowResult = true
	}
	if debug {
		r.ShowDebug = true
	}
	if quote != "" {
		r.SetQuoteColumns(strings.Split(quote, ","))
	}

	// set original dictionary
	if dic != "" {
		err = r.SetDictionary(dic)
		if err != nil {
			r.Close()
			return nil, err
		}
	}

	for _, p := range DefaultPreFilters {
		r.AddPreFilter(p)
	}
	for _, p := range DefaultPlugins {
		r.AddPlugin(p)
	}
	for _, p := range DefaultPostFilters {
		r.AddPostFilter(p)
	}

	switch {
	case replace:
		err = r.WriteHeaderWithReplace()
	default:
		err = r.WriteHeader()
	}
	if err != nil {
		r.Close()
		return nil, err
	}

	return r, nil
}
