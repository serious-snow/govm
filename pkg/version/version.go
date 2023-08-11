package version

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch *int

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
	rcReg   = regexp.MustCompile(`rc(\d)$`)
	betaReg = regexp.MustCompile(`beta(\d)$`)
)

func str2int(str string) int {
	res, _ := strconv.Atoi(str)
	return res
}

func (v *Version) Parse(version string) {
	version = strings.ToLower(version)
	version = strings.TrimLeft(version, "v")
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
	nums := make([]*int, maxL)
	for i, s := range vs {
		if i == maxL {
			break
		}
		v := str2int(s)
		nums[i] = &v
	}

	if nums[0] != nil {
		v.Major = *nums[0]
	}
	if nums[1] != nil {
		v.Minor = *nums[1]
	}
	if nums[2] != nil {
		temp := *nums[2]
		v.Patch = &temp
	}
}

func (v Version) String() string {
	sb := strings.Builder{}
	//sb.WriteString("v")
	sb.WriteString(strconv.Itoa(v.Major))
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(v.Minor))
	if v.Patch != nil {
		sb.WriteString(".")
		sb.WriteString(strconv.Itoa(*v.Patch))
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

func (v Version) GetPatch() int {
	if v.Patch != nil {
		return *v.Patch
	}
	return 0
}

func (v Version) Valid() bool {
	x := v.Major + v.Minor + v.GetPatch()

	return x != 0
}

func (v Version) Compare(b Version) int {
	return Compare(v, b)
}

func Equal(a, b Version) bool {
	return Compare(a, b) == 0
}

func Less(a, b Version) bool {
	return Compare(a, b) < 0
}

func Greater(a, b Version) bool {
	return Compare(a, b) > 0
}

func compareSegment(v, o int) int {
	if v < o {
		return -1
	}
	if v > o {
		return 1
	}

	return 0
}

func Compare(a, b Version) int {
	if d := compareSegment(a.Major, b.Major); d != 0 {
		return d
	}
	if d := compareSegment(a.Minor, b.Minor); d != 0 {
		return d
	}
	if d := compareSegment(a.GetPatch(), b.GetPatch()); d != 0 {
		return d
	}
	if a.RC {
		if b.RC {
			return compareSegment(a.VRC, b.VRC)
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
			return compareSegment(a.VBeta, b.VBeta)
		}
		return -1
	}
	return 0
}

func New(version string) (v *Version) {
	version = strings.TrimPrefix(version, "v")

	v = new(Version)
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
	nums := make([]*int, maxL)
	for i, s := range vs {
		if i == maxL {
			break
		}
		v := str2int(s)
		nums[i] = &v
	}
	if nums[0] != nil {
		v.Major = *nums[0]
	}
	if nums[1] != nil {
		v.Minor = *nums[1]
	}
	if nums[2] != nil {
		temp := *nums[2]
		v.Patch = &temp
	}
	return
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
}
