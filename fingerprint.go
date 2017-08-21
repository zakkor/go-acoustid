package acoustid

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Fingerprint struct {
	fingerprint string
	duration    int
}

func NewFingerprint(file string) Fingerprint {
	fp := Fingerprint{}

	fpcalc := "./fpcalc"
	if os.Getenv("FPCALC_BINARY_PATH") == "" {
	} else {
		fpcalc = os.Getenv("FPCALC_BINARY_PATH")
	}
	out, err := exec.Command(fpcalc, file).Output()
	if err != nil {
		panic(err)
	}
	outstrs := strings.Split(string(out), "\n")

	for _, s := range outstrs {
		if strings.Index(s, "DURATION=") == 0 {
			ds := strings.Split(s, "=")[1]
			fp.duration, _ = strconv.Atoi(ds)
		} else if strings.Index(s, "FINGERPRINT=") == 0 {
			fp.fingerprint = strings.Split(s, "=")[1]
		}
	}

	return fp
}
