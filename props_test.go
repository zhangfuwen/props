// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	"github.com/cevaris/ordered_map"
)

func TestNewProps(t *testing.T) {
	p := NewProperties()
	if p.values.Len() > 0 {
		t.Errorf("want: 0 elements; got: %d", p.values.Len())
	}
}

var comments = `
# line 1
! line 2
   # line 3
   ! line 4
  # line 5
  ! line 6
`

func TestReadComments(t *testing.T) {
	p, err := Read(bytes.NewBufferString(comments))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	if p.values.Len() > 0 {
		t.Errorf("want: 0 elements; got: %d", p.values.Len())
	}
}

var simple = `
key1=a
key2=b
key3=c
`

func DeepEqual(m1 map[string]string, m2 *ordered_map.OrderedMap) bool {
	iter := m2.IterFunc()
	for pair, ok := iter(); ok; pair, ok = iter() {
		if m1[pair.Key.(string)] != pair.Value.(Element).Value {
			return false
		}
	}
	return true
}
func TestReadSimple(t *testing.T) {
	p, err := Read(bytes.NewBufferString(simple))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1": "a",
		"key2": "b",
		"key3": "c",
	}

	if !DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

var continued = `
key1=abc\
    	def
key\
	2\
	3 = ghi\
	j\
	k\
	l
`

func TestReadContinued(t *testing.T) {
	p, err := Read(bytes.NewBufferString(continued))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1":  "abcdef",
		"key23": "ghijkl",
	}

	if !DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

var keys = `
key1=a
key2:b
key3 c
key4 = d
key5 : e
key6   f
key7
`

func TestReadKeys(t *testing.T) {
	p, err := Read(bytes.NewBufferString(keys))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key1": "a",
		"key2": "b",
		"key3": "c",
		"key4": "d",
		"key5": "e",
		"key6": "f",
		"key7": "",
	}

	if !DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}

	if _, ok := p.values.Get("key7"); !ok {
		t.Error("want: key7; got none")
	}
}

var escapes = `
key\n1=a\nb\n
key\t2:c\td
key\f3 e\ff
key\\4=g\\h
key\r5:i\rj
key\z6 k\3l
key\u005a7=m\u2126n
key\uuu00478=o\uzp
key\uD834\uDD1E9=q\uD800\uDC00r
key\
    \f10=s\
	\ft
`

func TestReadEscapes(t *testing.T) {
	p, err := Read(bytes.NewBufferString(escapes))

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	want := map[string]string{
		"key\n1":  "a\nb\n",
		"key\t2":  "c\td",
		"key\f3":  "e\ff",
		"key\\4":  "g\\h",
		"key\r5":  "i\rj",
		"keyz6":   "k3l",
		"keyZ7":   "mΩn",
		"keyG8":   "o\uFFFDp",
		"key𝄞9":   "q𐀀r",
		"key\f10": "s\ft",
	}

	if !DeepEqual(want, p.values) {
		t.Errorf("want: %#v; got: %#v", want, p.values)
	}
}

func TestGet(t *testing.T) {
	p := NewProperties()
	p.values.Set("key1", Element{
		Value: "foo",
	})

	if p.Get("key1") != "foo" {
		t.Errorf("want: foo; got: %q", p.Get("key1"))
	}

	if p.Get("key2") != "" {
		t.Errorf("want: \"\"; got: %q", p.Get("key2"))
	}

	if p.GetDefault("key2", "bar") != "bar" {
		t.Errorf("want: bar; got: %q", p.GetDefault("key2", "bar"))
	}
}

func TestSet(t *testing.T) {
	p := NewProperties()
	p.Set("key1", "foo")
	p.Set("key2", "bar")

	ele, ok := p.values.Get("key2")
	if !ok || ele.(Element).Value != "bar" {
		t.Errorf("want: bar; got %q", ele.(Element).Value)
	}
}

func TestNames(t *testing.T) {
	p := NewProperties()
	p.values.Set("key1", Element{Value: "foo"})
	p.values.Set("key2", Element{Value: "bar"})

	want := []string{"key1", "key2"}
	got := p.Names()

	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %#v, got: %#v", want, got)
	}
}

type writeTest struct {
	key  string
	val  string
	want string
}

var writeTests = []writeTest{
	{"key", "val", "key=val\n"},
	{"key", "  foo bar baz", "key=\\ \\ foo bar baz\n"},
	{"key:=#!", ":=#!foo bar baz",
		"key\\:\\=\\#\\!=\\:\\=\\#\\!foo bar baz\n"},
	{"key foo", "bar", "key\\ foo=bar\n"},
	{"key\nfoo", "bar\nbaz", "key\\nfoo=bar\\nbaz\n"},
	{"key\rfoo", "bar\rbaz", "key\\rfoo=bar\\rbaz\n"},
	{"key\ffoo", "bar\fbaz", "key\\ffoo=bar\\fbaz\n"},
	{"key\tfoo", "bar\tbaz", "key\\tfoo=bar\\tbaz\n"},
	{"key\u00A0foo", "bar\u00A9baz", "key\\u00a0foo=bar\\u00a9baz\n"},
}

func TestWrite(t *testing.T) {
	for _, test := range writeTests {
		p := NewProperties()
		p.values.Set(test.key, Element{Value: test.val})

		buf := new(bytes.Buffer)
		err := p.Write(buf)
		if err != nil {
			t.Errorf("got err: %v", err)
		}

		got := buf.String()
		if got != test.want {
			t.Errorf("want: %q; got: %q", test.want, got)
		}
	}
}

func TestProperties_GetMap(t *testing.T) {
	p, err := Read(bytes.NewBufferString(keys))
	if err != nil {
		t.Fail()
	}
	if p.GetMap()["key1"] != "a" {
		t.Fail()
	}
}
