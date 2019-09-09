/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package nele

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mwat56/ini"
)

type (
	// tAguments is the list structure for the cmdline argument values
	// merged with the key-value pairs from the INI file.
	tAguments struct {
		ini.TSection // embedded INI section
	}
)

var (
	// AppArguments is the list for the cmdline arguments and INI values.
	AppArguments tAguments
)

// `set()` adds/sets another key-value pair.
//
// If `aValue` is empty then `aKey` gets removed.
func (al *tAguments) set(aKey, aValue string) {
	if 0 < len(aValue) {
		al.AddKey(aKey, aValue)
	} else {
		al.RemoveKey(aKey)
	}
} // set()

// Get returns the value associated with `aKey` and `nil` if found,
// or an empty string and an error.
//
//	`aKey` The key to lookup in the list.
func (al *tAguments) Get(aKey string) (string, error) {
	if result, ok := al.AsString(aKey); ok {
		return result, nil
	}

	//lint:ignore ST1005 - capitalisation wanted
	return "", fmt.Errorf("Missing config value: %s", aKey)
} // Get()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `absolute()` return `aDir` as an absolute path
func absolute(aBaseDir, aDir string) string {
	if 0 == len(aDir) {
		return aDir
	}
	if '/' == aDir[0] {
		s, _ := filepath.Abs(aDir)
		return s
	}

	s, _ := filepath.Abs(filepath.Join(aBaseDir, aDir))
	return s
} // absolute()

// `readIniData()` returns the config values read from INI file(s).
//
//	The steps here are:
//	(1) read the local `./.nele.ini`,
//	(2) read the global `/etc/nele.ini`,
//	(3) read the user-local `~/.nele.ini`,
//	(4) read the user-local `~/.config/nele.ini`,
//	(5) read the `-ini` commandline argument.
func readIniData() {
	// (1) ./
	fName, _ := filepath.Abs("./nele.ini")
	ini1, err := ini.New(fName)
	if nil == err {
		ini1.AddSectionKey("", "inifile", fName)
	}

	// (2) /etc/
	fName = "/etc/nele.ini"
	if ini2, err := ini.New(fName); nil == err {
		ini1.Merge(ini2)
		ini1.AddSectionKey("", "inifile", fName)
	}

	// (3) ~user/
	fName, _ = os.UserHomeDir()
	if 0 < len(fName) {
		fName, _ = filepath.Abs(filepath.Join(fName, ".nele.ini"))
		if ini2, err := ini.New(fName); nil == err {
			ini1.Merge(ini2)
			ini1.AddSectionKey("", "inifile", fName)
		}
	}

	// (4) ~/.config/
	if confDir, err := os.UserConfigDir(); nil == err {
		fName, _ = filepath.Abs(filepath.Join(confDir, "nele.ini"))
		if ini2, err := ini.New(fName); nil == err {
			ini1.Merge(ini2)
			ini1.AddSectionKey("", "inifile", fName)
		}
	}

	// (5) cmdline
	aLen := len(os.Args)
	for i := 1; i < aLen; i++ {
		if `-ini` == os.Args[i] {
			//XXX Note that this works only if `-ini` and
			// filename are two separate arguments. It will
			// fail if it's given in the form `-ini=filename`.
			i++
			if i < aLen {
				fName, _ = filepath.Abs(os.Args[i])
				if ini2, _ := ini.New(fName); nil == err {
					ini1.Merge(ini2)
					ini1.AddSectionKey("", "inifile", fName)
				}
			}
			break
		}
	}

	AppArguments = tAguments{*ini1.GetSection("")}
} // readIniData()

/*
func init() {
	// // see: https://github.com/microsoft/vscode-go/issues/2734
	// testing.Init() // workaround for Go 1.13
	InitConfig()
} // init()
*/

// InitConfig reads the commandline arguments into a list
// structure merging it with key-value pairs read from INI file(s).
//
// The steps here are:
// (1) read the INI file(s),
// (2) merge the commandline arguments the INI values
// into the global `AppArguments` variable.
//
// This function is meant to be called first thing in the program's `main()`.
func InitConfig() {
	readIniData()

	bnStr, _ := AppArguments.Get("blogname")
	flag.StringVar(&bnStr, "blogname", bnStr,
		"Name of this Blog (shown on every page)\n")

	s, _ := AppArguments.Get("datadir")
	dataStr, _ := filepath.Abs(s)
	flag.StringVar(&dataStr, "datadir", dataStr,
		"<dirName> the directory with CSS, IMG, JS, POSTINGS, STATIC, VIEWS sub-directories\n")

	s, _ = AppArguments.Get("certKey")
	ckStr := absolute(dataStr, s)
	flag.StringVar(&ckStr, "certKey", ckStr,
		"<fileName> the name of the TLS certificate key\n")

	s, _ = AppArguments.Get("certPem")
	cpStr := absolute(dataStr, s)
	flag.StringVar(&cpStr, "certPem", cpStr,
		"<fileName> the name of the TLS certificate PEM\n")

	s, _ = AppArguments.Get("hashfile")
	hashStr := absolute(dataStr, s)
	flag.StringVar(&hashStr, "hashfile", hashStr,
		"<fileName> (optional) the name of a file storing #hashtags and @mentions\n")

	/*
		s, _ = AppArguments.Get("intl")
		intlStr := absolute(dataStr, s)
		flag.StringVar(&intlStr, "intl", intlStr,
			"<fileName> the path/filename of the localisation file\n")
	*/

	iniStr, _ := AppArguments.Get("inifile")
	flag.StringVar(&iniStr, "ini", iniStr,
		"<fileName> the path/filename of the INI file to use\n")

	langStr, _ := AppArguments.Get("lang")
	flag.StringVar(&langStr, "lang", langStr,
		"(optional) the default language to use ")

	listenStr, _ := AppArguments.Get("listen")
	flag.StringVar(&listenStr, "listen", listenStr,
		"the host's IP to listen at ")

	s, _ = AppArguments.Get("logfile")
	logStr := absolute(dataStr, s)
	flag.StringVar(&logStr, "log", logStr,
		"(optional) name of the logfile to write to\n")

	mfsStr, _ := AppArguments.Get("maxfilesize")
	flag.StringVar(&mfsStr, "maxfilesize", mfsStr,
		"max. accepted size of uploaded files")

	portInt, _ := AppArguments.AsInt("port")
	flag.IntVar(&portInt, "port", portInt,
		"<portNumber> the IP port to listen to ")

	paBool := false
	flag.BoolVar(&paBool, "pa", paBool,
		"(optional) posting add: write a posting from the commandline")

	pfStr := ""
	flag.StringVar(&pfStr, "pf", pfStr,
		"<fileName> (optional) post file: name of a file to add as new posting")

	realStr, _ := AppArguments.Get("realm")
	flag.StringVar(&realStr, "realm", realStr,
		"(optional) <hostName> name of host/domain to secure by BasicAuth\n")

	themStr, _ := AppArguments.Get("theme")
	flag.StringVar(&themStr, "theme", themStr,
		"<name> the display theme to use ('light' or 'dark')\n")

	uaStr := ""
	flag.StringVar(&uaStr, "ua", uaStr,
		"<userName> (optional) user add: add a username to the password file")

	ucStr := ""
	flag.StringVar(&ucStr, "uc", ucStr,
		"<userName> (optional) user check: check a username in the password file")

	udStr := ""
	flag.StringVar(&udStr, "ud", udStr,
		"<userName> (optional) user delete: remove a username from the password file")

	s, _ = AppArguments.Get("passfile")
	ufStr := absolute(dataStr, s)
	flag.StringVar(&ufStr, "uf", ufStr,
		"<fileName> (optional) user passwords file storing user/passwords for BasicAuth\n")

	ulBool := false
	flag.BoolVar(&ulBool, "ul", ulBool,
		"(optional) user list: show all users in the password file")

	uuStr := ""
	flag.StringVar(&uuStr, "uu", uuStr,
		"<userName> (optional) user update: update a username in the password file")

	flag.Usage = ShowHelp
	flag.Parse() // // // // // // // // // // // // // // // // // // //

	if 0 == len(bnStr) {
		bnStr = time.Now().Format("2006:01:02:15:04:05")
	}
	AppArguments.set("blogname", bnStr)

	if 0 < len(dataStr) {
		dataStr, _ = filepath.Abs(dataStr)
	}
	if f, err := os.Stat(dataStr); nil != err {
		log.Fatalf("datadir == %s` problem: %v", dataStr, err)
	} else if !f.IsDir() {
		log.Fatalf("Error: Not a directory `%s`", dataStr)
	}
	AppArguments.set("datadir", dataStr)

	// `postingBaseDirectory` defined in `posting.go`:
	SetPostingBaseDirectory(filepath.Join(dataStr, "./postings"))

	if 0 < len(ckStr) {
		ckStr = absolute(dataStr, ckStr)
		if fi, err := os.Stat(ckStr); (nil != err) || (0 >= fi.Size()) {
			ckStr = ""
		}
	}
	AppArguments.set("certKey", ckStr)

	if 0 < len(cpStr) {
		cpStr = absolute(dataStr, cpStr)
		if fi, err := os.Stat(cpStr); (nil != err) || (0 >= fi.Size()) {
			cpStr = ""
		}
	}
	AppArguments.set("certPem", cpStr)

	if 0 < len(hashStr) {
		hashStr = absolute(dataStr, hashStr)
	}
	AppArguments.set("hashfile", hashStr)

	/*
		if 0 <len(intlStr) {
			intlStr = absolute(dataStr, intlStr)
		}
		AppArguments.set("intl", intlStr)
	*/

	if 0 == len(langStr) {
		langStr = "en"
	}
	AppArguments.set("lang", langStr)

	if "0" == listenStr {
		listenStr = ""
	}
	AppArguments.set("listen", listenStr)

	if 0 < len(logStr) {
		logStr = absolute(dataStr, logStr)
	}
	AppArguments.set("logfile", logStr)

	if 0 == len(mfsStr) {
		mfsStr = "10485760" // 10 MB
	} else {
		mfs := kmg2Num(mfsStr)
		mfsStr = fmt.Sprintf("%d", mfs)
	}
	AppArguments.set("mfs", mfsStr)

	AppArguments.set("port", fmt.Sprintf("%d", portInt))

	if paBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("pa", s)

	if 0 < len(pfStr) {
		pfStr = absolute(dataStr, pfStr)
	}
	AppArguments.set("pf", pfStr)
	AppArguments.set("realm", realStr)
	AppArguments.set("theme", themStr)
	AppArguments.set("ua", uaStr)
	AppArguments.set("uc", ucStr)
	AppArguments.set("ud", udStr)

	if 0 < len(ufStr) {
		ufStr = absolute(dataStr, ufStr)
	}
	AppArguments.set("uf", ufStr)

	if ulBool {
		s = "true"
	} else {
		s = ""
	}
	AppArguments.set("ul", s)
	AppArguments.set("uu", uuStr)
} // InitConfig()

var (
	// RegEx to match a size value (xxx)
	cfKmgRE = regexp.MustCompile(`(?i)\s*(\d+)\s*([bgkm]+)?`)
)

// `kmg2Num()` returns a 'KB|MB|GB` string as an integer.
func kmg2Num(aString string) (rInt int) {
	matches := cfKmgRE.FindStringSubmatch(aString)
	if 2 < len(matches) {
		// The RegEx only matches digits so we can
		// safely ignore all Atoi() errors.
		rInt, _ = strconv.Atoi(matches[1])
		switch strings.ToLower(matches[2]) {
		case "", "b":
			return
		case "kb":
			rInt *= 1024
		case "mb":
			rInt *= 1024 * 1024
		case "gb":
			rInt *= 1024 * 1024 * 1024
		}
	}

	return
} // kmg2Num()

// ShowHelp lists the commandline options to `Stderr`.
func ShowHelp() {
	fmt.Fprintf(os.Stderr, "\n  Usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\n  Most options can be set in an INI file to keep the command-line short ;-)")
} // ShowHelp()

/* _EoF_ */
