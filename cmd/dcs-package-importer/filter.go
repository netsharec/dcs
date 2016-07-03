package main

import (
	"flag"
	"os"
	"strings"
)

// TODO: filter /debian/api/ (in every linux package), e.g.
// linux_3.14.4-1/debian/abi/3.14-1/armel_none_orion5x
var (
	ignoredDirnamesList = flag.String("ignored_dirnames",
		".pc,po,.git,libtool.m4",
		"(comma-separated list of) names of directories that will be deleted from packages when importing")

	// NB: we don’t skip "configure" since that might be a custom shell-script
	// NB: we actually skip some autotools files because they blow up our index otherwise
	ignoredFilenamesList = flag.String("ignored_filenames",
		"NEWS,COPYING,LICENSE,CHANGES,Makefile.in,ltmain.sh,config.guess,config.sub,depcomp,aclocal.m4,libtool.m4,.gitignore,57710_init_values.c,57711_init_values.c,57712_init_values.c",
		"(comma-separated list of) names of files that will be deleted from packages when importing")

	ignoredSuffixesList = flag.String("ignored_suffixes",
		"conf,dic,cfg,man,xml,xsl,html,sgml,pod,po,txt,tex,rtf,docbook,symbols",
		"(comma-separated list of) suffixes of files that will be deleted from packages when importing")

	onlySmallFilesSuffixesList = flag.String("only_small_files_suffixes",
		"ref,result,S,out,rst,def,afm,ps,pao,tom,ovp,UPF,map,ucm,json,svg,ppd,acc,ipp,eps,sym,pass,F90,tei,stl,tmp,dmp,vtk,csv,stp,decTest,test,lla,pamphlet",
		"(comma-separated list of) suffixes of files that will not be indexed if their size is more than 64 KB")

	ignoredDirnames        = make(map[string]bool)
	ignoredFilenames       = make(map[string]bool)
	ignoredSuffixes        = make(map[string]bool)
	onlySmallFilesSuffixes = make(map[string]bool)
)

func setupFilters() {
	for _, entry := range strings.Split(*ignoredDirnamesList, ",") {
		ignoredDirnames[entry] = true
	}
	for _, entry := range strings.Split(*ignoredFilenamesList, ",") {
		ignoredFilenames[entry] = true
	}
	for _, entry := range strings.Split(*ignoredSuffixesList, ",") {
		ignoredSuffixes[entry] = true
	}
	for _, entry := range strings.Split(*onlySmallFilesSuffixesList, ",") {
		onlySmallFilesSuffixes[entry] = true
	}
}

// Returns true when the file matches .[0-9]$ (cheaper than a regular
// expression).
func hasManpageSuffix(filename string) bool {
	return len(filename) > 2 &&
		filename[len(filename)-2] == '.' &&
		filename[len(filename)-1] >= '0' &&
		filename[len(filename)-1] <= '9'
}

// Returns true for files that should not be indexed for various reasons:
// • generated files
// • non-source (but text) files, e.g. .doc, .svg, …
func ignored(info os.FileInfo, dir, filename string) bool {
	if info.IsDir() {
		if ignoredDirnames[filename] {
			return true
		}
	} else {
		size := info.Size()
		// index/write.go will skip the file if it’s too big, so we might as
		// well skip it here and save the disk space.
		if size > (1 << 30) {
			return true
		}

		// TODO: peek inside the files (we’d have to read them anyways) and
		// check for messages that indicate that the file is generated. either
		// by autoconf or by bison for example.
		if ignoredFilenames[filename] ||
			// Don’t match /debian/changelog or /debian/README, but
			// exclude changelog and readme files generally.
			(!strings.HasSuffix(dir, "/debian/") &&
				strings.HasPrefix(strings.ToLower(filename), "changelog") ||
				strings.HasPrefix(strings.ToLower(filename), "readme")) ||
			hasManpageSuffix(filename) {
			return true
		}
		idx := strings.LastIndex(filename, ".")
		if idx > -1 {
			if ignoredSuffixes[filename[idx+1:]] &&
				!strings.HasPrefix(strings.ToLower(filename), "cmakelists.txt") {
				return true
			}
			if size > 65*1024 && onlySmallFilesSuffixes[filename[idx+1:]] {
				return true
			}
		}
	}

	return false
}
