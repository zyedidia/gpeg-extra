//+build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	sitter "github.com/zyedidia/gpeg-extra/bench/go-tree-sitter"
	"github.com/zyedidia/gpeg-extra/bench/go-tree-sitter/java"
	"github.com/zyedidia/gpeg/bench"
	"github.com/zyedidia/gpeg/input/linerope"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	file := os.Args[1]
	if !strings.HasSuffix(file, ".java") {
		log.Fatal("input file must be java")
	}

	text, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	r := linerope.New(text)

	var b [4096]byte
	ropeRead := func(offset uint32, position sitter.Point) []byte {
		n, _ := r.ReadAt(b[:], int64(offset))
		return b[:n]
	}

	input := sitter.Input{
		Read:     ropeRead,
		Encoding: sitter.InputEncodingUTF8,
	}

	st := time.Now()
	tree := parser.ParseInput(nil, input)
	fmt.Println(time.Since(st).Microseconds())

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

	for _, e := range edits {
		startl, startc := r.LineColAt(e.Start)
		oldendl, oldendc := r.LineColAt(e.End)
		r.Remove(e.Start, e.End)
		r.Insert(e.Start, e.Text)
		newendl, newendc := r.LineColAt(e.Start + len(e.Text))
		st := time.Now()
		tree.Edit(sitter.EditInput{
			StartIndex:  uint32(e.Start),
			OldEndIndex: uint32(e.End),
			NewEndIndex: uint32(e.Start + len(e.Text)),
			StartPoint: sitter.Point{
				Row:    uint32(startl),
				Column: uint32(startc),
			},
			OldEndPoint: sitter.Point{
				Row:    uint32(oldendl),
				Column: uint32(oldendc),
			},
			NewEndPoint: sitter.Point{
				Row:    uint32(newendl),
				Column: uint32(newendc),
			},
		})

		tree = parser.ParseInput(tree, input)
		fmt.Println(time.Since(st).Microseconds())
	}
}
