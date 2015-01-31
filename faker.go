package faker

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/enodata/faker/locales"
)

const (
	Digits           = "0123456789"
	ULetters         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DLetters         = "abcdefghijklmnopqrstuvwxyz"
	Letters          = ULetters + DLetters
	DigitsAndLetters = Digits + Letters
)

var (
	aryDigits   []string = strings.Split(Digits, "")
	aryULetters []string = strings.Split(ULetters, "")
	aryDLetters []string = strings.Split(DLetters, "")
	aryLetters  []string = strings.Split(Letters, "")
)

// Default locale.
var Locale = locales.En

// RandonmInt returns random int in [min, max] range.
func RandomInt(min, max int) int {
	if max <= min {
		// degenerate case, return min
		return min
	}
	return min + rand.Intn(max-min+1)
}

// RandonmInt64 returns random int64 in [min, max] range.
func RandomInt64(min, max int64) int64 {
	if max <= min {
		// degenerate case, return min
		return min
	}
	return min + rand.Int63n(max-min+1)
}

// RandomString returns a random alphanumeric string with length n.
func RandomString(n int) string {
	bytes := make([]byte, n)
	crand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = DigitsAndLetters[b%byte(len(DigitsAndLetters))]
	}
	return string(bytes)
}

// RandomRepeat returns a new string consisting of random number of copies of the string s.
func RandomRepeat(s string, min, max int) string {
	return strings.Repeat(s, RandomInt(min, max))
}

// RandomChoice returns random string from slice of strings.
func RandomChoice(ss []string) string {
	return ss[rand.Intn(len(ss))]
}

func includesString(ss []string, s string) (res bool) {
	res = false
	for _, v := range ss {
		if v == s {
			res = true
		}
	}
	return
}

// Numerify replaces pattern like '##-###' with randomly generated digits.
func Numerify(s string) string {
	var first bool = true
	for _, sm := range regexp.MustCompile(`#`).FindAllStringSubmatch(s, -1) {
		if first {
			// make sure result does not start with zero
			s = strings.Replace(s, sm[0], RandomChoice(aryDigits[1:]), 1)
			first = false
		} else {
			s = strings.Replace(s, sm[0], RandomChoice(aryDigits), 1)
		}
	}
	return s
}

// Letterify replaces pattern like '??? ??' with randomly generated uppercase letters
func Letterify(s string) string {
	for _, sm := range regexp.MustCompile(`\?`).FindAllStringSubmatch(s, -1) {
		s = strings.Replace(s, sm[0], RandomChoice(aryULetters), 1)
	}
	return s
}

// NumerifyAndLetterify both numerifies and letterifies s.
func NumerifyAndLetterify(s string) string {
	return Letterify(Numerify(s))
}

// Regexify attempts to generate a string that would match given regular expression.
// It does not handle ., *, unbounded ranges such as {1,},
// extensions such as (?=), character classes, some abbreviations for character classes,
// and nested parentheses.
func Regexify(s string) (string, error) {
	// ditch the anchors
	res := regexp.MustCompile(`^\/?\^?`).ReplaceAllString(s, "")
	res = regexp.MustCompile(`\$?\/?$`).ReplaceAllString(res, "")

	// all {2} become {2,2} and ? become {0,1}
	res = regexp.MustCompile(`\{(\d+)\}`).ReplaceAllString(res, `{$1,$1}`)
	res = regexp.MustCompile(`\?`).ReplaceAllString(res, `{0,1}`)

	// [12]{1,2} becomes [12] or [12][12]
	for _, sm := range regexp.MustCompile(`(\[[^\]]+\])\{(\d+),(\d+)\}`).FindAllStringSubmatch(res, -1) {
		min, _ := strconv.Atoi(sm[2])
		max, _ := strconv.Atoi(sm[3])
		res = strings.Replace(res, sm[0], RandomRepeat(sm[1], min, max), 1)
	}

	// (12|34){1,2} becomes (12|34) or (12|34)(12|34)
	for _, sm := range regexp.MustCompile(`(\([^\)]+\))\{(\d+),(\d+)\}`).FindAllStringSubmatch(res, -1) {
		min, _ := strconv.Atoi(sm[2])
		max, _ := strconv.Atoi(sm[3])
		res = strings.Replace(res, sm[0], RandomRepeat(sm[1], min, max), 1)
	}

	// A{1,2} becomes A or AA or \d{3} becomes \d\d\d
	for _, sm := range regexp.MustCompile(`(\\?.)\{(\d+),(\d+)\}`).FindAllStringSubmatch(res, -1) {
		min, _ := strconv.Atoi(sm[2])
		max, _ := strconv.Atoi(sm[3])
		res = strings.Replace(res, sm[0], RandomRepeat(sm[1], min, max), 1)
	}

	// (this|that) becomes 'this' or 'that'
	for _, sm := range regexp.MustCompile(`\((.*?)\)`).FindAllStringSubmatch(res, -1) {
		res = strings.Replace(res, sm[0], RandomChoice(strings.Split(sm[1], "|")), 1)
	}

	// all A-Z inside of [] become C (or X, or whatever)
	for _, sm := range regexp.MustCompile(`\[([^\]]+)\]`).FindAllStringSubmatch(res, -1) {
		cls := sm[1]
		// find and replace all ranges in character class cls
		for _, subsm := range regexp.MustCompile(`(\w\-\w)`).FindAllStringSubmatch(cls, -1) {
			rng := strings.Split(subsm[1], "-")
			repl := string(RandomInt(int(rng[0][0]), int(rng[1][0])))
			cls = strings.Replace(cls, subsm[0], repl, 1)
		}
		res = strings.Replace(res, sm[1], cls, 1)
	}

	// all [ABC] become B (or A or C)
	for _, sm := range regexp.MustCompile(`\[([^\]]+)\]`).FindAllStringSubmatch(res, -1) {
		res = strings.Replace(res, sm[0], RandomChoice(strings.Split(sm[1], "")), 1)
	}

	// all \d become random digits
	res = regexp.MustCompile(`\\d`).ReplaceAllStringFunc(res, func(s string) string {
		return RandomChoice(aryDigits)
	})

	// all \w become random letters
	res = regexp.MustCompile(`\\d`).ReplaceAllStringFunc(res, func(s string) string {
		return RandomChoice(aryLetters)
	})

	return res, nil
}

func localeValueAt(path string, locale map[string]interface{}) (interface{}, bool) {
	var val interface{} = locale
	for _, key := range strings.Split(path, ".") {
		v, ok := val.(map[string]interface{})
		if !ok {
			// all nodes are expected to be of map[string]interface{}
			panic(fmt.Sprintf("%v: invalid value type %v", path, reflect.TypeOf(val)))
		}
		val, ok = v[key]
		if !ok {
			// given path does not exists in given locale
			return nil, false
		}
	}
	return val, true
}

func valueAt(path string) interface{} {
	val, ok := localeValueAt(path, Locale)
	if !ok {
		// path does not exist in given locale, fallback to En
		val, ok = localeValueAt(path, locales.En)
		if !ok {
			// not in En either, give up
			panic(fmt.Sprintf("%v: invalid path", path))
		}
	}
	return val
}

// Fetch returns a value at given key path in default locale. If key path holds an array,
// it returns random array element. If value looks like a regex, it attempts to regexify it.
func Fetch(path string) string {
	var res string

	switch val := valueAt(path).(type) {
	case [][]string:
		// slice of string slices - select random element and join
		choices := make([]string, len(val))
		for i, slice := range val {
			choices[i] = RandomChoice(slice)
		}
		res = strings.Join(choices, " ")
	case []string:
		// slice of strings - select random element
		res = RandomChoice(val)
	case string:
		// plain string
		res = val
	default:
		// not supported
		panic(fmt.Sprintf("%v: invalid value type %v", path, reflect.TypeOf(val)))
	}

	// recursively substitute #{...} value references
	for _, sm := range regexp.MustCompile(`#\{([A-Za-z]+\.[^\}]+)\}`).FindAllStringSubmatch(res, -1) {
		res = strings.Replace(res, sm[0], Fetch(sm[1]), 1)
	}

	// if res looks like regex, regexify
	if strings.HasPrefix(res, "/") && strings.HasSuffix(res, "/") {
		res, err := Regexify(res)
		if err != nil {
			panic(fmt.Sprintf("failed to regexify %v: %v", res, err))
		}
		return res
	}

	return res
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}