// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
	"github.com/cevaris/ordered_map"
)

// Properties represents a set of key-value pairs.
type Properties struct {
	values *ordered_map.OrderedMap
}

type Element struct {
	Key string
	Value string
	Comment string
}

// NewProperties creates a new, empty property set.
func NewProperties() *Properties {
	p := &Properties{
		values: ordered_map.NewOrderedMap(),
	}
	return p
}

// Read creates a new property set and fills it with the contents of a file.
// See Load for the supported file format.
func Read(r io.Reader) (*Properties, error) {
	p := NewProperties()
	err := p.Load(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Get retrieves the value of a property. If the property does not exist, an
// empty string will be returned.
func (p *Properties) Get(key string) string {
	ele, ok := p.values.Get(key)
	if !ok {
		return ""
	}
	return ele.(Element).Value
}

// GetMap returns a map
func (p *Properties) GetMap() map[string]string {
	m := make(map[string]string)
	iter := p.values.IterFunc()
	for pair,ok := iter();ok; pair,ok= iter() {
		m[pair.Key.(string)] = m[pair.Value.(Element).Value]
	}
	return m
}

// GetDefault retrieves the value of a property. If the property does not
// exist, then the default value will be returned.
func (p *Properties) GetDefault(key, defVal string) string {
	if val := p.Get(key); val!="" {
		return val
	}
	return defVal
}

// Set adds or changes the value of a property.
func (p *Properties) Set(key, val string) {
	if ele, ok :=p.values.Get(key); ok {
		p.values.Set(key,Element{
			Key:key,
			Value:ele.(Element).Value,
			Comment:ele.(Element).Comment,
		})
	}else {
		p.values.Set(key, Element{
			Value:val,
		})
	}
}

// Clear removes all key-value pairs.
func (p *Properties) Clear() {
	p.values = ordered_map.NewOrderedMap()
}

/*
Load reads the contents of a property file. Existing properties will be
retained. The contents of the file will override any existing properties with
matching keys.

File Format

The supported property file format follows the Java conventions. Each line of
the file represents a key-value pair. Keys and values may be separated by '=',
':', or whitespace. Comments are indicated by a leading '#' or '!' character.

Encoding

Java property files require an ISO 8859-1 encoding, but this package will also
accept files encoded in UTF-8.

Escapes

The escape character is '\'; valid escapes are '\f', '\n', '\r', '\t', and
UTF-16 escapes in the format "\uXXXX" where each "X" is a hexadecimal digit.
Invalid escapes are replaced with the escaped character only, so '\A' will
result in 'A'. (This is useful for escaping the key separator or comment
characters.) Invalid UTF-16 escapes will be replaced with the Unicode
replacement character U+FFFD.

Spanning Lines

To create a key or value that spans multiple lines, end the line with '\'
followed by a newline. All leading whitespace on the next line will be ignored
and not included in the key or value, allowing for indentation of continued
lines.

Sample File

This is a sample property file:
	# env.properties
	! for dev environment
	site.url = http://localhost:8180/

	# database
	db.host:localhost
	db.port:5432
	db.user:devdb

	# email
	email.from dev@example.com
	email.to   me@example.org

	email.welcome  Subject: Welcome! \
				  Thank you. Now: \
				  \t Feat 1 \
				  \t Feat 2 \
				  Enjoy!

	# reporting
	rpt\ newline=\u000a
	rpt\ list\ bullet=\u2022

Loading this file would result in the following properties:
	"site.url":        "http://localhost:8180/"
	"db.host":         "localhost"
	"db.port":         "5432"
	"db.user":         "devdb"
	"email.from":      "dev@example.com"
	"email.to":        "me@example.org"
	"email.welcome":   "Subject: Welcome! \nThank you. Now: \n\tFeat 1 \n..."
	"rpt newline":     "\n"
	"rpt list bullet": "•"
*/
func (p *Properties) Load(r io.Reader) error {
	state := stateNone
	s := &scanner{p: p}

	buf := bufio.NewReader(r)
	for {
		ch, _, err := buf.ReadRune()
		if err != nil {
			if err == io.EOF {
				s.done()
				return nil
			} else {
				return err
			}
		}
		state = state(s, ch)
	}

	return nil
}

// Names returns the keys for all properties in the set.
func (p *Properties) Names() []string {
	names := make([]string, 0, p.values.Len())
	iter := p.values.IterFunc()
	for pair,ok := iter();ok; pair,ok = iter() {
		names = append(names, pair.Key.(string))
	}
	return names
}

// Write saves the property set to a file. The output will be in "key=value"
// format, with appropriate characters escaped. See Load for more details on
// the file format.
//
// Note: if the property set was loaded from a file, the formatting and
// comments from the original file will not be retained in the output file.
func (p *Properties) Write(w io.Writer) error {
	iter := p.values.IterFunc()
	for pair,ok := iter();ok;pair,ok = iter() {
		_, err:= io.WriteString(w,pair.Value.(Element).Comment)
		if err!=nil {
			return err
		}
		line := fmt.Sprintf("%s=%s\n", escape(pair.Key.(string), true),
			escape(pair.Value.(Element).Value, false))
		_, err = io.WriteString(w, line)
		if err != nil {
			return err
		}
	}
	return nil
}

// escape returns a string that is safe to use as either a key or value in a
// property file. Whitespace characters, key separators, and comment markers
// should always be escaped.
func escape(s string, key bool) string {

	leading := true
	var buf bytes.Buffer
	for _, ch := range s {
		wasSpace := false
		if ch == '\t' {
			buf.WriteString(`\t`)
		} else if ch == '\n' {
			buf.WriteString(`\n`)
		} else if ch == '\r' {
			buf.WriteString(`\r`)
		} else if ch == '\f' {
			buf.WriteString(`\f`)
		} else if ch == ' ' {
			if key || leading {
				buf.WriteString(`\ `)
				wasSpace = true
			} else {
				buf.WriteRune(ch)
			}
		} else if ch == ':' {
			buf.WriteString(`\:`)
		} else if ch == '=' {
			buf.WriteString(`\=`)
		} else if ch == '#' {
			buf.WriteString(`\#`)
		} else if ch == '!' {
			buf.WriteString(`\!`)
		} else if !unicode.IsPrint(ch) || ch > 126 {
			buf.WriteString(fmt.Sprintf(`\u%04x`, ch))
		} else {
			buf.WriteRune(ch)
		}

		if !wasSpace {
			leading = false
		}
	}
	return buf.String()
}
