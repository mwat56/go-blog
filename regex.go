/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
              EMail : <support@mwat.de>
*/

package nele

/*
 * This files provides a few RegEx based functions.
 */

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"time"

	// bf "github.com/russross/blackfriday/v2"
	bf "gopkg.in/russross/blackfriday.v2"
)

var (
	// RegEx to HREF= tag attributes
	reHrefRE = regexp.MustCompile(` (href="http)`)
)

const (
	// replacement text for `hrefRE`
	reHrefReplace = ` target="_extern" $1`
)

// `addExternURLtagets()` adds a TARGET attribute to HREFs.
func addExternURLtagets(aPage []byte) []byte {
	return reHrefRE.ReplaceAll(aPage, []byte(reHrefReplace))
} // addExternURLtagets()

var (
	// RegEx to match hh:mm:ss
	reHmsRE = regexp.MustCompile(`^(([01]?[0-9])|(2[0-3]))[^0-9](([0-5]?[0-9])([^0-9]([0-5]?[0-9]))?)?[^0-9]?|$`)
)

// `getHMS()` splits up `aTime` into `rHour`, `rMinute`, and `rSecond`.
func getHMS(aTime string) (rHour, rMinute, rSecond int) {
	matches := reHmsRE.FindStringSubmatch(aTime)
	if 1 < len(matches) {
		// The RegEx only matches digits so we can
		// safely ignore all Atoi() errors.
		rHour, _ = strconv.Atoi(matches[1])
		if 0 < len(matches[5]) {
			rMinute, _ = strconv.Atoi(matches[5])
			if 0 < len(matches[7]) {
				rSecond, _ = strconv.Atoi(matches[7])
			}
		}
	}

	return
} // getHMS()

var (
	// RegEx to match YYYY(MM)(DD)
	// Invalid values for month or day result in a `0` result.
	// This is just a pattern test, it doesn't check whether the date is valid.
	// reYmdRE = regexp.MustCompile("^([0-9]{4})[^0-9]?(((0?[0-9])|(1[0-2]))[^0-9]?((0?[0-9])?|([12][0-9])?|(3[01])?)?)?$")
	reYmdRE = regexp.MustCompile(`^([0-9]{4})([^0-9]?(0[1-9]|1[012])([^0-9]?(0[1-9]|[12][0-9]|3[01])?)?)?[^0-9]?`)
)

// `getYMD()` splits up `aDate` into `rYear`, `rMonth`, and `rDay`.
//
// This is just a pattern test: the function doesn't check whether
// the date as such is a valid date.
func getYMD(aDate string) (rYear int, rMonth time.Month, rDay int) {
	matches := reYmdRE.FindStringSubmatch(aDate)
	if 1 < len(matches) {
		// The RegEx only matches digits so we can
		// safely ignore all Atoi() errors.
		rYear, _ = strconv.Atoi(matches[1])
		if 0 < len(matches[3]) {
			m, _ := strconv.Atoi(matches[3])
			rMonth = time.Month(m)
			if 0 < len(matches[5]) {
				rDay, _ = strconv.Atoi(matches[5])
			}
		}
	}

	return
} // getYMD()

func init() {
	initWSre()
} // init()

// Initialise the `whitespaceREs` list.
func initWSre() int {
	result := 0
	for idx, re := range reWhitespaceREs {
		reWhitespaceREs[idx].regEx = regexp.MustCompile(re.search)
		result++
	}
	result++

	return result
} // initWSre()

var (
	// RegEx to correct wrong markup created by 'bf';
	// see MDtoHTML()
	bfPreCodeRE = regexp.MustCompile(`(?s)\s*(<pre>)<code>(.*?)\s*</code>(</pre>)\s*`)

	bfPreCodeRE2 = regexp.MustCompile(`(?s)\s*(<pre)><code (class="language-\w+")>(.*?)\s*</code>(</pre>)\s*`)
)

// `handlePreCode()` tries to fix the Pre/Code markup
func handlePreCode(aMarkdown []byte) (rHTML []byte) {
	rHTML = bfPreCodeRE.ReplaceAll(aMarkdown, []byte("$1\n$2\n$3"))
	if i := bytes.Index(rHTML, []byte("<pre><code ")); 0 > i {
		// no need for the second RegEx execution
		return
	}
	rHTML = bfPreCodeRE2.ReplaceAll(rHTML, []byte("$1 $2>\n$3\n$4"))

	return
} // handlePreCode()

// MDtoHTML converts the `aMarkdown` data returning HTML data.
//
// `aMarkdown` the raw Markdown text to convert.
func MDtoHTML(aMarkdown []byte) []byte {
	extensions := bf.WithExtensions(
		bf.Autolink |
			bf.BackslashLineBreak |
			bf.DefinitionLists |
			bf.FencedCode |
			bf.Footnotes |
			bf.HeadingIDs |
			bf.NoIntraEmphasis |
			bf.SpaceHeadings |
			bf.Strikethrough |
			bf.Tables)
	r := bf.NewHTMLRenderer(bf.HTMLRendererParameters{
		Flags: bf.FootnoteReturnLinks |
			bf.Smartypants |
			bf.SmartypantsFractions |
			bf.SmartypantsDashes |
			bf.SmartypantsLatexDashes,
	})
	result := bf.Run(aMarkdown, bf.WithRenderer(r), extensions)

	if i := bytes.Index(result, []byte("</pre>")); 0 > i {
		// no need for RegEx execution
		return result
	}
	// Testing for PRE makes this implementation twice as fast
	// if there's no PRE in the generated HTML and about the
	// same speed if there actually is a PRE part.

	return handlePreCode(result)
} // MDtoHTML()

var (
	// RegEx to extract number and start of articles shown
	reNumStartRE = regexp.MustCompile(`^(\d*)(\D*(\d*)?)?`)
)

// `numStart()` extracts two numbers from `aString`.
func numStart(aString string) (rNum, rStart int) {
	matches := reNumStartRE.FindStringSubmatch(aString)
	if 3 < len(matches) {
		if 0 < len(matches[1]) {
			rNum, _ = strconv.Atoi(matches[1])
		}
		if 0 < len(matches[3]) {
			rStart, _ = strconv.Atoi(matches[3])
		}
	}

	return
} // numStart()

// `trimPREmatches()` removes leading/trailing whitespace from list entries.
func trimPREmatches(aList [][]byte) [][]byte {
	for idx, hit := range aList {
		aList[idx] = bytes.TrimSpace(hit)
	}

	return aList
} // trimPREmatches()

// Internal list of regular expressions used by
// the `RemoveWhiteSpace()` function.
type (
	tReItem struct {
		search  string
		replace string
		regEx   *regexp.Regexp
	}
	tReList []tReItem
)

var (
	// RegEx to find PREformatted parts in an HTML page.
	rePreRE = regexp.MustCompile(`(?si)\s*<pre[^>]*>.*?</pre>\s*`)

	// List of regular expressions matching different sets of HTML whitespace.
	reWhitespaceREs = tReList{
		// comments
		{`(?s)<!--.*?-->`, ``, nil},
		// HTML and HEAD elements:
		{`(?i)\s*(</?(body|\!DOCTYPE|head|html|link|meta|script|style|title)[^>]*>)\s*`, `$1`, nil},
		// block elements:
		{`(?i)\s*(</?(article|blockquote|div|footer|h[1-6]|header|nav|p|section)[^>]*>)\s*`, `$1`, nil},
		// lists:
		{`(?i)\s*(</?([dou]l|li|d[dt])[^>]*>)\s*`, `$1`, nil},
		// table:
		{`(?i)\s*(</?(col|t(able|body|foot|head|[dhr]))[^>]*>)\s*`, `$1`, nil},
		// forms:
		{`(?i)\s*(</?(form|fieldset|legend|opt(group|ion))[^>]*>)\s*`, `$1`, nil},
		// BR / HR:
		{`(?i)\s*(<[bh]r[^>]*>)\s*`, `$1`, nil},
		// whitespace after opened anchor:
		{`(?i)(<a\s+[^>]*>)\s+`, `$1`, nil},
		// preserve empty table cells:
		{`(?i)(<td(\s+[^>]*)?>)\s+(</td>)`, `$1&#160;$3`, nil},
		// remove empty paragraphs:
		{`(?i)<(p)(\s+[^>]*)?>\s*</$1>`, ``, nil},
		// whitespace before closing GT:
		{`\s+>`, `>`, nil},
	}
)

// RemoveWhiteSpace removes HTML comments and unnecessary whitespace.
//
// This function removes all unneeded/redundant whitespace
// and HTML comments from the given <tt>aPage</tt>.
// This can reduce significantly the amount of data to send to
// the remote user agent thus saving bandwidth.
func RemoveWhiteSpace(aPage []byte) []byte {
	var repl, search string

	// fmt.Println("Page0:", string(aPage))
	// (0) Check whether there are PREformatted parts:
	preMatches := rePreRE.FindAll(aPage, -1)
	if (nil == preMatches) || (0 >= len(preMatches)) {
		// no PRE hence only the other REs to perform
		for _, reEntry := range reWhitespaceREs {
			aPage = reEntry.regEx.ReplaceAll(aPage, []byte(reEntry.replace))
		}
		return aPage
	}
	preMatches = trimPREmatches(preMatches)

	// Make sure PREformatted parts remain as-is.
	// (1) replace the PRE parts with a dummy text:
	for l, cnt := len(preMatches), 0; cnt < l; cnt++ {
		search = fmt.Sprintf("\\s*%s\\s*", regexp.QuoteMeta(string(preMatches[cnt])))
		if re, err := regexp.Compile(search); nil == err {
			repl = fmt.Sprintf("</-%d-%d-%d-%d-/>", cnt, cnt, cnt, cnt)
			aPage = re.ReplaceAllLiteral(aPage, []byte(repl))
		}
	}
	// fmt.Println("Page1:", string(aPage))

	// (2) traverse through all the whitespace REs:
	for _, re := range reWhitespaceREs {
		aPage = re.regEx.ReplaceAll(aPage, []byte(re.replace))
	}
	// fmt.Println("Page2:", string(aPage))

	// (3) replace the PRE dummies with the real markup:
	for l, cnt := len(preMatches), 0; cnt < l; cnt++ {
		search = fmt.Sprintf("\\s*</-%d-%d-%d-%d-/>\\s*", cnt, cnt, cnt, cnt)
		if re, err := regexp.Compile(search); nil == err {
			aPage = re.ReplaceAllLiteral(aPage, preMatches[cnt])
		}
	}
	// fmt.Println("Page3:", string(aPage))

	return aPage
} // RemoveWhiteSpace()

/* _EoF_ */
