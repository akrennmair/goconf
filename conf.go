// This package implements a parser for configuration files.
// This allows easy reading and writing of structured configuration files.
//
// Given a sample configuration file:
//
//	[default]
//	host=www.example.com
//	protocol=http://
//	base-url=%(protocol)s%(host)s
//
//	[service-1]
//	url=%(base-url)s/some/path
//	delegation : on
//	maxclients=200 # do not set this higher
//	comments=This is a multi-line
//		entry	; And this is a comment
//
// To read this configuration file, do:
//
//	c, err := configfile.ReadConfigFile("config.cfg");
//	c.GetString("service-1", "url"); // result is string :http://www.example.com/some/path"
//	c.GetInt("service-1", "maxclients"); // result is int 200
//	c.GetBool("service-1", "delegation"); // result is bool true
//	c.GetString("service-1", "comments"); // result is string "This is a multi-line\nentry"
//
// Note the support for unfolding variables (such as %(base-url)s), which are read from the special
// (reserved) section name [default].
//
// A new configuration file can also be created with:
//
//	c := configfile.NewConfigFile();
//	c.AddSection("section");
//	c.AddOption("section", "option", "value");
//	c.WriteConfigFile("config.cfg", 0644, "A header for this file"); // use 0644 as file permission
//
// This results in the file:
//
//	# A header for this file
//	[section]
//	option=value
//
// Note that sections and options are case-insensitive (values are case-sensitive)
// and are converted to lowercase when saved to a file.
//
// The functionality and workflow is loosely based on the configparser.py package
// of the Python Standard Library.
package conf

import (
	"regexp"
	"strings"
	"fmt"
)


// ConfigFile is the representation of configuration settings.
// The public interface is entirely through methods.
type ConfigFile struct {
	data map[string]map[string]string;	// Maps sections to options to values.
}

const (
	// Get Errors
	SectionNotFound = iota
	OptionNotFound
	MaxDepthReached

	// Read Errors
	BlankSection

	// Get and Read Errors
	CouldNotParse
)

var (
	DefaultSection	= "default";	// Default section name (must be lower-case).
	DepthValues	= 200;		// Maximum allowed depth when recursively substituing variable names.

	// Strings accepted as bool.
	BoolStrings	= map[string]bool{
		"t": true,
		"true": true,
		"y": true,
		"yes": true,
		"on": true,
		"1": true,
		"f": false,
		"false": false,
		"n": false,
		"no": false,
		"off": false,
		"0": false,
	};

	varRegExp	= regexp.MustCompile(`%\(([a-zA-Z0-9_.\-]+)\)s`);
)


// AddSection adds a new section to the configuration.
// It returns true if the new section was inserted, and false if the section already existed.
func (c *ConfigFile) AddSection(section string) bool {
	section = strings.ToLower(section);

	if _, ok := c.data[section]; ok {
		return false
	}
	c.data[section] = make(map[string]string);

	return true;
}


// RemoveSection removes a section from the configuration.
// It returns true if the section was removed, and false if section did not exist.
func (c *ConfigFile) RemoveSection(section string) bool {
	section = strings.ToLower(section);

	switch _, ok := c.data[section]; {
	case !ok:
		return false
	case section == DefaultSection:
		return false	// default section cannot be removed
	default:
		for o, _ := range c.data[section] {
			c.data[section][o] = "", false
		}
		c.data[section] = nil, false;
	}

	return true;
}


// AddOption adds a new option and value to the configuration.
// It returns true if the option and value were inserted, and false if the value was overwritten.
// If the section does not exist in advance, it is created.
func (c *ConfigFile) AddOption(section string, option string, value string) bool {
	c.AddSection(section);	// make sure section exists

	section = strings.ToLower(section);
	option = strings.ToLower(option);

	_, ok := c.data[section][option];
	c.data[section][option] = value;

	return !ok;
}


// RemoveOption removes a option and value from the configuration.
// It returns true if the option and value were removed, and false otherwise,
// including if the section did not exist.
func (c *ConfigFile) RemoveOption(section string, option string) bool {
	section = strings.ToLower(section);
	option = strings.ToLower(option);

	if _, ok := c.data[section]; !ok {
		return false
	}

	_, ok := c.data[section][option];
	c.data[section][option] = "", false;

	return ok;
}


// NewConfigFile creates an empty configuration representation.
// This representation can be filled with AddSection and AddOption and then
// saved to a file using WriteConfigFile.
func NewConfigFile() *ConfigFile {
	c := new(ConfigFile);
	c.data = make(map[string]map[string]string);

	c.AddSection(DefaultSection);	// default section always exists

	return c;
}


func stripComments(l string) string {
	// comments are preceded by space or TAB
	for _, c := range []string{" ;", "\t;", " #", "\t#"} {
		if i := strings.Index(l, c); i != -1 {
			l = l[0:i]
		}
	}
	return l;
}


func firstIndex(s string, delim []byte) int {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(delim); j++ {
			if s[i] == delim[j] {
				return i
			}
		}
	}
	return -1;
}

type GetError struct {
	Reason int
	ValueType string
	Value string
	Section string
	Option string
}

func (err GetError) String() string {
	switch err.Reason {
		case SectionNotFound:
			return fmt.Sprintf("section '%s' not found", err.Section)
		case OptionNotFound:
			return fmt.Sprintf("option '%s' not found in section '%s'", err.Option, err.Section)
		case CouldNotParse:
			return fmt.Sprintf("could not parse %s value '%s'", err.ValueType, err.Value)
		case MaxDepthReached:
			return fmt.Sprintf("possible cycle while unfolding variables: max depth of %d reached", DepthValues)
	}
	
	return "invalid get error"
}

type ReadError struct {
	Reason int
	Line string
}

func (err ReadError) String() string {
	switch err.Reason {
		case BlankSection:
			return "empty section name not allowed"
		case CouldNotParse:
			return fmt.Sprintf("could not parse line: %s", err.Line)
	}
	
	return "invalid read error"
}
