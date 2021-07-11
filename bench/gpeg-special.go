//+build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zyedidia/gpeg/bench"
	"github.com/zyedidia/gpeg/input/linerope"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

var nedits = flag.Int("n", 100, "number of major edits to perform")
var mthreshold = flag.Int("mthreshold", 512, "memoization entry size threshold")

func main() {
	flag.Parse()

	if len(flag.Args()) <= 0 {
		fmt.Println("Usage: bench [OPTIONS]... FILE")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(0)
	}

	inputf := flag.Args()[0]
	if !strings.HasSuffix(inputf, ".java") {
		log.Fatal("input file must end with .java")
	}
	data, err := ioutil.ReadFile(inputf)
	if err != nil {
		log.Fatal(err)
	}

	gdat, err := ioutil.ReadFile("grammars/java_memo.peg")
	if err != nil {
		log.Fatal(err)
	}

	code := vm.Encode(pattern.MustCompile(re.MustCompile(string(gdat))))

	tbl := memo.NewTreeTable(*mthreshold)

	r := linerope.New(data)

	edits := []bench.Edit{
		bench.Edit{
			Start: 0,
			End:   0,
			Text:  []byte("/*"),
		},
		bench.Edit{
			Start: 0,
			End:   2,
			Text:  nil,
		},
		bench.Edit{
			Start: 0,
			End:   0,
			Text:  []byte("/*"),
		},
	}

	st := time.Now()
	code.Exec(r, tbl)
	fmt.Println(time.Since(st).Microseconds())

	for _, e := range edits {
		start := e.Start
		end := e.End

		r.Remove(start, end)
		r.Insert(start, []byte(e.Text))

		st := time.Now()
		tbl.ApplyEdit(memo.Edit{
			Start: start,
			End:   end,
			Len:   len(e.Text),
		})

		code.Exec(r, tbl)
		fmt.Println(time.Since(st).Microseconds())
	}
}
