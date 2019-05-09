/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package blog

/*
 * This file provides some template/views related functions and methods.
 */

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

type (
	// TDataList is a list of values to be injected into a template.
	TDataList map[string]interface{}
)

// NewDataList returns a new (empty) TDataList instance.
func NewDataList() *TDataList {
	result := make(TDataList, 32)

	return &result
} // NewDatalist()

// Set inserts `aValue` identified by `aKey` to the list.
//
// If there's already a list entry with `aKey` its current value
// gets replaced by `aValue`.
//
// `aKey` is the values's identifier (as used as placeholder in the template).
//
// `aValue` contains the data entry's value.
func (dl *TDataList) Set(aKey string, aValue interface{}) *TDataList {
	(*dl)[aKey] = aValue

	return dl
} // Set()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type (
	// Internal type to track changes in certain template vars.
	tChange struct {
		current string
	}
)

// Changed returns whether `aNext` is the same as the last value.
func (c *tChange) Changed(aNext string) bool {
	if c.current == aNext {
		return false
	}
	c.current = aNext

	return true
} // Changed()

// `newChange()` returns a new change structure.
func newChange() *tChange {
	return &tChange{
		"{{$}}", // ensure that first change is recognised
	}
} // newChange()

// `dateNow()` returns the current date.
func dateNow() string {
	y, m, d := time.Now().Date()

	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
} // dateNow()

// `htmlSafe()` returns `aText` as template.HTML.
func htmlSafe(aText string) template.HTML {
	return template.HTML(aText)
} // htmlSafe()

func int2post(aPost interface{}) *TPosting {
	if p, ok := aPost.(*TPosting); ok {
		return p
	} else if p, ok := aPost.(TPosting); ok {
		return &p
	}

	return nil
} // int2post()

// `isPost()` checks whether `aPost` is an instance of `TPost` or `TPosting`.
func isPost(aPost interface{}) bool {
	p := int2post(aPost)

	return (nil != p)
} // isPost()

// isPostEmpty() checks whether the text of `aPost` is empty or not.
func isPostEmpty(aPost interface{}) bool {
	p := int2post(aPost)
	if nil != p {
		return (0 == p.Len())
	}

	return true
} // isPostEmpty()

// `isPostlist()` checks whether `aPostlist` is an instance of `TPostList`.
func isPostlist(aPostlist interface{}) (rOK bool) {
	if _, rOK = aPostlist.(TPostList); rOK {
		return
	}
	_, rOK = aPostlist.(*TPostList)

	return
} // isPostlist()

// `postID()` returns the ID (i.e. filename) of `aPost`.
func postID(aPost interface{}) string {
	p := int2post(aPost)
	if nil != p {
		return p.ID()
	}

	return ""
} // postID()

// `postMonthURL()` returns the month URL of `aPost`.
func postMonthURL(aPost interface{}) string {
	p := int2post(aPost)
	if nil == p {
		return ""
	}
	y, m, d := timeID(p.ID()).Date()

	return fmt.Sprintf("/m/%d%02d%02d", y, m, d)
} // postMonthURL()

// `postText()` returns the safe HTML of `aPost`.
func postText(aPost interface{}) template.HTML {
	p := int2post(aPost)
	if nil != p {
		return p.Post()
	}

	return ""
} // postText()

// `postWeekURL()` returns the week URL of `aPost`.
func postWeekURL(aPost interface{}) string {
	p := int2post(aPost)
	if nil == p {
		return ""
	}
	y, m, d := timeID(p.ID()).Date()

	return fmt.Sprintf("/w/%d%02d%02d", y, m, d)
} // postWeekURL()

// `monthURL()` returns an URL for the current month.
func monthURL() string {
	y, m, d := time.Now().Date()

	return fmt.Sprintf("/m/%d%02d%02d", y, m, d)
} // monthURL()

// `weekURL()` returns an URL for the current week.
func weekURL() string {
	y, m, d := time.Now().Date()

	return fmt.Sprintf("/w/%d%02d%02d", y, m, d)
} // weekURL()

var (
	fMap = template.FuncMap{
		"change":       newChange,    // a new change structure
		"dateNow":      dateNow,      // the current date
		"htmlSafe":     htmlSafe,     // returns `aText` as template.HTML
		"isPost":       isPost,       // whether `aPost` is `TPost`/`TPosting`
		"isPostEmpty":  isPostEmpty,  // whether the text of `aPost` is empty
		"isPostlist":   isPostlist,   // whether `aPostlist` is `TPostList`
		"postID":       postID,       // the ID (i.e. filename) of `aPost`
		"postText":     postText,     // the safe HTML of `aPost`
		"postWeekURL":  postWeekURL,  // the week URL of `aPost`
		"postMonthURL": postMonthURL, // the month URL of `aPost
		"monthURL":     monthURL,     // URL for the current month
		"weekURL":      weekURL,      // URL for the current week
	}
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// TView combines a template and its logical name.
type TView struct {
	// The view's symbolic name.
	name string
	// The template as returned by a `NewView()` function call.
	tpl *template.Template
}

// NewView returns a new `TView` with `aName`.
//
// `aBaseDir` is the path to the directory storing the template files.
//
// `aName` is the name of the template file providing the page's main
// body without the filename extension (i.e. w/o ".gohtml"). `aName`
// serves as both the main template's name as well as the view's name.
func NewView(aBaseDir, aName string) (*TView, error) {
	bd, err := filepath.Abs(aBaseDir)
	if nil != err {
		return nil, err
	}
	files, err := filepath.Glob(fmt.Sprintf("%s/layout/*.gohtml", bd))
	if nil != err {
		return nil, err
	}
	files = append(files, fmt.Sprintf("%s/%s.gohtml", bd, aName))

	templ, err := template.New(aName).
		Funcs(fMap).
		ParseFiles(files...)
	if nil != err {
		return nil, err
	}

	return &TView{
		name: aName,
		tpl:  templ,
	}, nil
} // NewView()

// `render()` is the core of `Render()` with a slightly different API
// (`io.Writer` instead of `http.ResponseWriter`) for easier testing.
func (v *TView) render(aWriter io.Writer, aData *TDataList) (rErr error) {
	var page []byte

	if page, rErr = v.RenderedPage(aData); nil != rErr {
		return
	}

	// if _, rErr := aWriter.Write(page); nil != rErr {
	if _, rErr := aWriter.Write(RemoveWhiteSpace(page)); nil != rErr {
		return rErr
	}

	return
} // render()

// Render executes the template using the TView's properties.
//
// `aWriter` is a http.ResponseWriter, or e.g. `os.Stdout` in console apps.
//
// `aData` is a list of data to be injected into the template.
//
// If an error occurs executing the template or writing its output,
// execution stops, and the method returns without writing anything
// to the output `aWriter`.
func (v *TView) Render(aWriter http.ResponseWriter, aData *TDataList) error {
	return v.render(aWriter, aData)
} // Render()

// RenderedPage returns the rendered template/page and a possible Error
// executing the template.
//
// `aData` is a list of data to be injected into the template.
func (v *TView) RenderedPage(aData *TDataList) (rBytes []byte, rErr error) {
	buf := &bytes.Buffer{}

	if rErr = v.tpl.ExecuteTemplate(buf, v.name, aData); nil != rErr {
		return
	}

	return buf.Bytes(), nil
} // RenderedPage()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type (
	tViewList map[string]*TView

	// TViewList is a list of `TView` instances (to be used as a template pool).
	TViewList tViewList
)

// NewViewList returns a new (empty) `TViewList` instance.
func NewViewList() *TViewList {
	result := make(TViewList, 8)

	return &result
} // NewViewlist()

// Add appends `aView` to the list.
//
// `aView` is the view to add to this list.
//
// The view's name (as specified in the `NewView()` function call)
// is used as the view's key in this list.
func (vl *TViewList) Add(aView *TView) *TViewList {
	(*vl)[aView.name] = aView

	return vl
} // Add()

// Get returns the view with `aName`.
//
// `aName` is the name (key) of the `TView` object to retrieve.
//
// If `aName` doesn't exist, the return value is `nil`.
// The second value (ok) is a `bool` that is `true` if `aName`
// exists in the list, and `false` if not.
func (vl *TViewList) Get(aName string) (*TView, bool) {
	if result, ok := (*vl)[aName]; ok {
		return result, true
	}

	return nil, false
} // Get()

// `render()` is the core of `Render()` with a slightly different API
// (`io.Writer` instead of `http.ResponseWriter`) for easier testing.
func (vl *TViewList) render(aName string, aWriter io.Writer, aData *TDataList) error {
	if view, ok := (*vl)[aName]; ok {
		return view.render(aWriter, aData)
	}

	return fmt.Errorf("template/view '%s' not found", aName)
} // render()

// Render executes the template with the key `aName`.
//
// `aName` is the name of the template/view to use.
//
// `aWriter` is a `http.ResponseWriter` to handle the executed template.
//
// `aData` is a list of data to be injected into the template.
//
// If an error occurs executing the template or writing its output,
// execution stops, and the method returns without writing anything
// to the output `aWriter`.
func (vl *TViewList) Render(aName string, aWriter http.ResponseWriter, aData *TDataList) error {
	return vl.render(aName, aWriter, aData)
} // Render()

// RenderedPage returns the rendered template/page with the key `aName`.
//
// `aName` is the name of the template/view to use.
//
// `aData` is a list of data to be injected into the template.
func (vl *TViewList) RenderedPage(aName string, aData *TDataList) (rBytes []byte, rErr error) {

	if view, ok := (*vl)[aName]; ok {
		return view.RenderedPage(aData)
	}

	return rBytes, fmt.Errorf("template/view '%s' not found", aName)
} // RenderedPage()

/* _EoF_ */
