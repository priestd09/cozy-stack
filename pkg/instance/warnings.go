package instance

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cozy/cozy-stack/web/jsonapi"
)

// Warnings returns a list of possible warnings associated with the instance.
func (i *Instance) Warnings() (warnings []*jsonapi.Error) {
	notSigned, _ := i.CheckTOSSigned()
	if notSigned {
		tosLink, _ := i.ManagerURL(ManagerTOSURL)
		warnings = append(warnings, &jsonapi.Error{
			Status: http.StatusPaymentRequired,
			Title:  "TOS Updated",
			Code:   "tos-updated",
			Detail: "Terms of services have been updated",
			Links:  &jsonapi.LinksList{Self: tosLink},
		})
	}
	return
}

// TOSDeadline represent the state for reaching the TOS deadline.
type TOSDeadline int

const (
	// TOSNone when no deadline is reached.
	TOSNone TOSDeadline = iota
	// TOSWarning when the warning deadline is reached, 2 weeks before the actual
	// activation of the CGU.
	TOSWarning
	// TOSBlocked when the deadline is reached and the access should be blocked.
	TOSBlocked
)

// CheckTOSSigned checks whether or not the current Term of Services have been
// signed by the user.
func (i *Instance) CheckTOSSigned(args ...string) (notSigned bool, deadline TOSDeadline) {
	tosLatest := i.TOSLatest
	if len(args) > 0 {
		tosLatest = args[0]
	}
	latest, latestDate, ok := parseTOSVersion(tosLatest)
	if !ok {
		return
	}
	signed, _, ok := parseTOSVersion(i.TOSSigned)
	if !ok {
		return
	}
	if signed >= latest {
		return
	}
	notSigned = true
	now := time.Now()
	if now.After(latestDate) {
		deadline = TOSBlocked
	} else if now.After(latestDate.Add(-15 * 24 * time.Hour)) {
		deadline = TOSWarning
	} else {
		deadline = TOSNone
	}
	return
}

// parseTOSVersion returns the "major" and the date encoded in a version string
// with the following format:
//    parseTOSVersion(A.B.C-YYYYMMDD) -> (A, YYYY, true)
func parseTOSVersion(v string) (major int, date time.Time, ok bool) {
	if v == "" {
		return
	}
	if len(v) == 8 {
		var err error
		major = 1
		date, err = time.Parse("20060102", v)
		ok = err == nil
		return
	}
	if v[0] == 'v' {
		v = v[1:]
	}
	a := strings.SplitN(v, ".", 3)
	if len(a) != 3 {
		return
	}
	major, err := strconv.Atoi(a[0])
	if err != nil {
		return
	}
	suffix := strings.SplitN(a[2], "-", 2)
	if len(suffix) < 2 {
		return
	}
	date, err = time.Parse("20060102", suffix[1])
	ok = err == nil
	return
}