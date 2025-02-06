package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/re"
	"github.com/zyedidia/gpeg/vm"
)

var peg = flag.String("peg", "", "file containing PEG to use for parsing")

func main() {
	flag.Parse()

	args := flag.Args()

	var in io.Reader
	if len(args) <= 0 {
		in = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		in = f
	}

	b, err := io.ReadAll(in)
	if err != nil {
		log.Fatal(err)
	}

	peg, err := os.ReadFile(*peg)
	if err != nil {
		log.Fatal(err)
	}
	ids := make(map[string]int)
	patt, err := re.CompileCap(string(peg), ids)
	if err != nil {
		log.Fatal(err)
	}
	prog, err := pattern.Compile(patt)
	if err != nil {
		log.Fatal(err)
	}
	code := vm.Encode(prog)
	match, _, ast, _ := code.Exec(bytes.NewReader(b), memo.NoneTable{})
	if !match {
		log.Fatal("parse failed")
	}

	// We can now access AST nodes via ast.Child(n) and use ids["..."] to look
	// for IDs.

	// Dump the AST as a graphviz dot graph
	fmt.Print(GraphAST(ast, b, ids))
}

func text(n *memo.Capture, data []byte) string {
	str := string(data[n.Start():n.End()])
	str = strings.ReplaceAll(str, ">", "&gt;")
	str = strings.ReplaceAll(str, "<", "&lt;")
	return strconv.Quote(strconv.QuoteToASCII(str))
}

func uniqueID(n *memo.Capture) string {
	return fmt.Sprintf("%p", n)[2:]
}

func exploreNode(n *memo.Capture, data []byte, ids map[int]string, graph *gographviz.Graph) {
	graph.AddNode("AST", uniqueID(n), map[string]string{
		"label": fmt.Sprintf("%s", ids[n.Id()]),
		"shape": "Mrecord",
		"color": "black",
	})

	it := n.ChildIterator(0)

	for c := it(); c != nil; c = it() {
		exploreNode(c, data, ids, graph)
		graph.AddEdge(uniqueID(n), uniqueID(c), false, map[string]string{
			"color": "black",
		})
	}

	if n.NumChildren() == 0 {
		textID := uniqueID(n) + "_text"
		graph.AddNode("AST", textID, map[string]string{
			"label": text(n, data),
			"shape": "Mrecord",
			"color": "black",
		})
		graph.AddEdge(uniqueID(n), textID, false, map[string]string{
			"color": "black",
		})
	}
}

// GraphAST renders the given AST to a graphviz dot graph, returned as a
// string.
func GraphAST(root *memo.Capture, data []byte, ids map[string]int) string {
	graph := gographviz.NewGraph()
	graph.SetName("AST")
	graph.SetDir(false)

	revids := make(map[int]string)
	for k, v := range ids {
		revids[v] = k
	}

	exploreNode(root, data, revids, graph)

	return graph.String()
}
