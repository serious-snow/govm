package models

import (
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	V1    int
	V2    int
	V3    int
	Beta  bool
	VBeta int
	RC    bool
	VRC   int
}

func (v Version) MarshalText() (text []byte, err error) {
	text = []byte(strconv.Quote(v.String()))
	return
}

func (v *Version) UnmarshalText(text []byte) error {
	s, err := strconv.Unquote(string(text))
	if err != nil {
		s = string(text)
	}
	v.Parse(s)
	return nil
}

func (v Version) MarshalJSON() (text []byte, err error) {
	return v.MarshalText()
}

func (v *Version) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		return nil
	}
	return v.UnmarshalText(text)
}

var (
	rcReg   = regexp.MustCompile("rc(\\d)$")
	betaReg = regexp.MustCompile("beta(\\d)$")
)

func str2int(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}

func NewVInfo(version string) *Version {
	res := new(Version)
	res.Parse(version)
	return res
}

func (v *Version) Parse(version string) {

	if rcReg.MatchString(version) {
		v.RC = true
		numStr := rcReg.FindStringSubmatch(version)[1]
		v.VRC = str2int(numStr)
		version = rcReg.ReplaceAllString(version, "")
	}

	if betaReg.MatchString(version) {
		v.Beta = true
		numStr := betaReg.FindStringSubmatch(version)[1]
		v.VBeta = str2int(numStr)
		version = betaReg.ReplaceAllString(version, "")
	}

	vs := strings.Split(version, ".")
	const maxL = 3
	nums := make([]int, maxL)
	for i, s := range vs {
		if i == maxL {
			break
		}
		nums[i] = str2int(s)
	}

	v.V1 = nums[0]
	v.V2 = nums[1]
	v.V3 = nums[2]

	return
}
func (v Version) String() string {
	sb := strings.Builder{}
	sb.WriteString(strconv.Itoa(v.V1))
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(v.V2))
	if v.V3 != 0 {
		sb.WriteString(".")
		sb.WriteString(strconv.Itoa(v.V3))
	}
	if v.RC {
		sb.WriteString("rc")
		sb.WriteString(strconv.Itoa(v.VRC))
	}
	if v.Beta {
		sb.WriteString("beta")
		sb.WriteString(strconv.Itoa(v.VBeta))
	}
	result := sb.String()
	sb.Reset()
	return result
}

func (v Version) Valid() bool {
	return v.V1+v.V2+v.V3 != 0
}

func (v Version) Compare(b Version) int {
	return Compare(v, b)
}
func Compare(a, b Version) int {
	if a == b {
		return 0
	}
	if a.V1 != b.V1 {
		return a.V1 - b.V1

	}
	if a.V2 != b.V2 {
		return a.V2 - b.V2
	}
	if a.V3 != b.V3 {
		return a.V3 - b.V3
	}
	if a.RC {
		if b.RC {
			return a.VRC - b.VRC
		}
		if b.Beta {
			return 1
		}
		return -1
	}
	if a.Beta {
		if b.RC {
			return -1
		}
		if b.Beta {
			return a.VBeta - b.VBeta
		}
		return -1
	}
	return 1
}

type SortV []*Version

func (s SortV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortV) Len() int {
	return len(s)
}

func (s SortV) Less(i, j int) bool {
	return s[i].Compare(*s[j]) < 0
}

func (s SortV) Sort() {
	sort.Sort(s)
}

func (s SortV) Reverse() {
	sort.Sort(sort.Reverse(s))
	runtime.LockOSThread()
}
