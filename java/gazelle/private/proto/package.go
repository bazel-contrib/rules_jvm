package proto

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
)

var optionRe = regexp.MustCompile(`^option\s+(\S+)\s*=\s*(\S+);$`)

// File is a very incomplete proto file descriptor.
type File struct {
	PackageName string
	Options     map[string]string
	Enums       []string
	Messages    []string
	Services    []string
}

// ParseFile parses some elements of a proto file.
//
// Bugs:
// - it does not handle multi line options
// - it is very incomplete
//
// Todo:
// - explore using https://github.com/tallstoat/pbparser instead
func ParseFile(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := &File{
		PackageName: "",
		Options:     make(map[string]string),
		Enums:       nil,
		Messages:    nil,
	}

	reader := bufio.NewReader(f)
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line := strings.TrimSpace(l)

		if strings.HasPrefix(line, "option ") {
			m := optionRe.FindStringSubmatch(line)
			if len(m) != 3 {
				// FIXME this line parser cannot parse multiple lines options.
				continue
			}

			out.Options[m[1]] = strings.Trim(m[2], "\"")
			continue
		}

		if strings.HasPrefix(line, "package ") {
			out.PackageName = strings.TrimSuffix(strings.TrimPrefix(line, "package "), ";")
			continue
		}

		if strings.HasPrefix(line, "service ") {
			out.Services = append(out.Services, strings.Split(strings.TrimPrefix(line, "service "), " ")[0])
			continue
		}

		if strings.HasPrefix(line, "enum ") {
			out.Enums = append(out.Enums, strings.Split(strings.TrimPrefix(line, "enum "), " ")[0])
			continue
		}

		if strings.HasPrefix(line, "message ") {
			out.Messages = append(out.Messages, strings.Split(strings.TrimPrefix(line, "message "), " ")[0])
			continue
		}
	}

	return out, nil
}

func (f *File) Symbols() []string {
	var symbols []string
	symbols = append(symbols, f.Services...)
	symbols = append(symbols, f.Enums...)
	symbols = append(symbols, f.Messages...)

	// hack to include the generated class name.
	if v, found := f.Options["java_outer_classname"]; found {
		symbols = append(symbols, v)
	}

	return symbols
}
