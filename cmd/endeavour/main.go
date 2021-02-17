package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/brennan-macaig/endeavour"
	"log"
	"os"
	"strings"
	"time"
)

const (
	app             = "endeavour"
	versionFlag     = "version"
	userEnvFlag     = "user-var"
	passEnvFlag     = "pass-var"
	urlFlag         = "U"
	userEnv         = "REPO_USERNAME"
	passEnv         = "REPO_PASSWORD"
	repoFlag        = "r"
	pathFlag        = "P"
	suppressArtFlag = "no-art"
	verboseFlag     = "v"
	helpFlag        = "h"

	usageFormat = app + ` (the last NASA-produced manned vehicle to bring stuff to space) - %s

Uploads files or directories to nexus, for use in CI/CD.

usage: ` + app + ` [options] file-or-dir0 ... file-or-dirN

` + app + ` by default will look at the ` + userEnv + ` and ` + passEnv + ` 
environment variables to get Nexus login information. 
It then attempts to upload provided artifacts to Nexus.

options:
%s
`
	art = `
                .                                            .
     *   .                  .              .        .   *          .
  .         .                     .       .           .      .        .
        o                             .                   .
         .              .                  .           .
          0     .
                 .          .                 ,                ,    ,
 .          \          .                         .
      .      \   ,
   .          o     .                 .                   .            .
     .         \                 ,             .                .
               #\##\#      .                              .        .
             #  #O##\###                .                        .
   .        #*#  #\##\###                       .                     ,
        .   ##*#  #\##\##               .                     .
      .      ##*#  #o##\#         .                             ,       .
          .     *#  #\#     .                    .             .          ,
                      \          .                         .
____^/\___^--____/\____O______________/\/\---/\___________---______________
   /\^   ^  ^    ^                  ^^ ^  '\ ^          ^       ---
         --           -            --  -      -         ---  __       ^
   --  __                      ___--  ^  ^                         --  __
`
)

var (
	version    string
	defaulturl string
)

func main() {
	userEnvPtr := flag.String(userEnvFlag, userEnv, "Environment variable for Nexus username")
	passEnvPtr := flag.String(passEnvFlag, passEnv, "Environment variable for Nexus password")
	urlPtr := flag.String(urlFlag, defaulturl, "Nexus URL to upload to")
	printVersion := flag.Bool(versionFlag, false, "Print the version and exit")
	repoPtr := flag.String(repoFlag, "", "Nexus repository to upload to")
	pathPtr := flag.String(pathFlag, "", "Path to publish to inside repository")
	suppressArtPtr := flag.Bool(suppressArtFlag, false, "If set, art will not be displayed on a success")
	verbosePtr := flag.Bool(verboseFlag, false, "If set, extra log messages will be displayed that do not contain secrets")
	help := flag.Bool(helpFlag, false, "Display this help message")

	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(1)
	}

	if *help {
		_, _ = fmt.Fprintf(os.Stderr, usageFormat, version, allDefaultsToString(flag.CommandLine))
		os.Exit(1)
	}
	nex := endeavour.Nexus{}

	if val, ok := os.LookupEnv(*userEnvPtr); ok {
		nex.Username = val
	} else {
		log.Fatalf("The variable %s must be set and must be non-empty", *userEnvPtr)
	}

	if val, ok := os.LookupEnv(*passEnvPtr); ok {
		nex.Password = val
	} else {
		log.Fatalf("The variable %s must be set and must be non-empty", *passEnvPtr)
	}

	nex.Url = *urlPtr
	nex.Path = *pathPtr
	nex.Repo = *repoPtr
	nex.Verbose = *verbosePtr
	nex.Files = flag.Args()
	start := time.Now()
	err := nex.Upload()
	if err != nil {
		log.Fatalf("Upload failed. Error: %s", err.Error())
	}
	fmt.Printf("All done! Completed in %s", time.Since(start).String())
	if !*suppressArtPtr {
		fmt.Print(art)
	}
	os.Exit(0)
}

func allDefaultsToString(set *flag.FlagSet) string {
	orig := set.Output()
	out := bytes.NewBuffer(nil)
	set.SetOutput(out)
	set.PrintDefaults()
	set.SetOutput(orig)
	return strings.TrimRight(out.String(), "\n")
}
