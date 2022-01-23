package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update .golden files")

//go:embed testdata
var testdataFS embed.FS

func testFile(p string) io.Reader {
	r, err := testdataFS.Open(p)
	if err != nil {
		panic(err)
	}

	return r
}

func testString(p string) string {
	r := testFile(p)
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func TestMain(t *testing.T) {
	type Test struct {
		args []string
	}

	tests := []Test{
		0: {[]string{"#footer"}},
		1: {[]string{"#footer li"}},
		2: {[]string{"#footer ul + div"}},
		3: {[]string{"#footer ul + div attr{style}"}},
		4: {[]string{"#toc li > a"}},
		5: {[]string{"table li"}},
		6: {[]string{"table li:first-child"}},
		7: {[]string{"table li:first-of-type"}},
		8: {[]string{"table li:last-child"}},
		9: {[]string{"table li:last-of-type"}},
		10: {[]string{`table a[title="The Practice of Programming"]`}},
		11: {[]string{`table a[title="The Practice of Programming"] text{}`}},
		12: {[]string{"json{}"}},
		13: {[]string{"text{}"}},
		14: {[]string{".after-portlet"}},
		15: {[]string{".organiser"}},
		16: {[]string{":empty"}},
		17: {[]string{"td:empty"}},
		18: {[]string{".navbox-list li:nth-child(1)"}},
		19: {[]string{".navbox-list li:nth-child(2)"}},
		20: {[]string{".navbox-list li:nth-child(3)"}},
		21: {[]string{".navbox-list li:nth-last-child(1)"}},
		22: {[]string{".navbox-list li:nth-last-child(2)"}},
		23: {[]string{".navbox-list li:nth-last-child(3)"}},
		24: {[]string{".navbox-list li:nth-child(n+1)"}},
		25: {[]string{".navbox-list li:nth-child(3n+1)"}},
		26: {[]string{".navbox-list li:nth-last-child(n+1)"}},
		27: {[]string{".navbox-list li:nth-last-child(3n+1)"}},
		28: {[]string{":only-child"}},
		29: {[]string{".navbox-list li:only-child"}},
		30: {[]string{".summary"}},
		31: {[]string{"[class=summary]"}},
		32: {[]string{`[class="summary"]`}},
		33: {[]string{"#toc"}},
		34: {[]string{"#toc div + ul"}},
		35: {[]string{"#toc div + ul li:nth-of-type(1) a span:nth-of-type(1) text{}"}},
		36: {[]string{"#toc div + ul li:nth-of-type(1) a span:nth-of-type(1) json{}"}},
		37: {[]string{"#toc div + ul span + span"}},
		38: {[]string{"span + a"}},
		39: {[]string{"#footer a > img"}},
		40: {[]string{"li a:not([rel])"}},
		41: {[]string{"link, a"}},
		42: {[]string{"link ,a"}},
		43: {[]string{"link , a"}},
		44: {[]string{"link , a sup"}},
		45: {[]string{"link , a:parent-of(sup)"}},
		46: {[]string{"link , a:parent-of(sup) sup"}},
		47: {[]string{"li", "--number"}},
		48: {[]string{"li", "-n"}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// Reset globals
			pupCharset = ""
			pupMaxPrintLevel = -1
			pupPreformatted = false
			pupPrintColor = false
			pupEscapeHTML = true
			pupIndentString = " "
			pupDisplayer = TreeDisplayer{}

			t.Logf("testing args: %v", tt.args)

			a := []string{"pup"}
			a = append(a, tt.args...)
			os.Args = a

			cmds, err := ParseArgs()
			if err != nil {
				panic(err)
			}

			f := testFile("testdata/index.html")

			root, err := ParseHTML(f, pupCharset)
			if err != nil {
				panic(err)
			}

			buf := &bytes.Buffer{}
			if err := runSelectors(buf, cmds, root); err != nil {
				panic(err)
			}
			gp := fmt.Sprintf("testdata/%03d.golden", i)
			if *update {
				if err := ioutil.WriteFile(gp, buf.Bytes(), 0644); err != nil {
					t.Fatalf("failed to update golden file: %s", err)
				}
				return
			}

			if len(buf.Bytes()) == 0 {
				t.Error("test has no output")
			}


			if diff := cmp.Diff(testString(gp), buf.String()); diff != ""  {
				t.Errorf("test does not match golden data (-got, +expect): %s", diff)
			}

		})
	}

}
