package dmarc

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

// AlignmentMode represents the alignment mode for DKIM and SPF.
type AlignmentMode string

const (
	AlignmentModeRelaxed AlignmentMode = "r"
	AlignmentModeStrict  AlignmentMode = "s"
)

// FailureOption represents the failure reporting options.
type FailureOption string

const (
	FailureOptionAll  FailureOption = "0"
	FailureOptionAny  FailureOption = "1"
	FailureOptionDKIM FailureOption = "d"
	FailureOptionSPF  FailureOption = "s"
)

// Policy represents the DMARC policy.
type Policy string

const (
	PolicyNone       Policy = "none"
	PolicyQuarantine Policy = "quarantine"
	PolicyReject     Policy = "reject"
)

// ReportFormat represents the format for reports.
type ReportFormat string

const (
	ReportFormatAFRF ReportFormat = "afrf"
)

// Record represents a DMARC record.
type Record struct {
	Version             string          // 'v'
	DKIMAlignment       AlignmentMode   // 'adkim'
	SPFAlignment        AlignmentMode   // 'aspf'
	FailureOptions      []FailureOption // 'fo'
	Policy                              // 'p'
	Percent             int             // 'pct'
	ReportFormats       []ReportFormat  // 'rf'
	ReportInterval      time.Duration   // 'ri'
	ReportURIsAggregate []string        // 'rua'
	ReportURIsFailure   []string        // 'ruf'
	SubdomainPolicy     Policy          // 'sp'
}

func ParseRecord(txt string) (*Record, error) {
	params, err := parseParams(txt)
	if err != nil {
		return nil, err
	}

	rec := new(Record)

	v, ok := params["v"]
	if !ok {
		return nil, errors.New("dmarc: record is missing a 'v' parameter")
	}
	rec.Version = v

	p, ok := params["p"]
	if !ok {
		return nil, errors.New("dmarc: record is missing a 'p' parameter")
	}
	rec.Policy, err = parsePolicy(p, "p")
	if err != nil {
		return nil, errors.New("dmarc: could not parse 'p' parameter")
	}

	if adkim, ok := params["adkim"]; ok {
		rec.DKIMAlignment, err = parseAlignmentMode(adkim, "adkim")
		if err != nil {
			return nil, err
		}
	}

	if aspf, ok := params["aspf"]; ok {
		rec.SPFAlignment, err = parseAlignmentMode(aspf, "aspf")
		if err != nil {
			return nil, err
		}
	}

	if fo, ok := params["fo"]; ok {
		rec.FailureOptions, err = parseFailureOptions(fo)
		if err != nil {
			return nil, err
		}
	}

	if pct, ok := params["pct"]; ok {
		i, err := strconv.Atoi(pct)
		if err != nil {
			return nil, fmt.Errorf("dmarc: invalid parameter 'pct': %v", err)
		}
		if i < 0 || i > 100 {
			return nil, fmt.Errorf("dmarc: invalid parameter 'pct': value %v out of bounds", i)
		}
		rec.Percent = i
	}

	if rf, ok := params["rf"]; ok {
		l := strings.Split(rf, ":")
		rec.ReportFormats = make([]ReportFormat, len(l))
		for i, f := range l {
			switch f {
			case string(ReportFormatAFRF):
				rec.ReportFormats[i] = ReportFormat(f)
			default:
				return nil, errors.New("dmarc: invalid parameter 'rf'")
			}
		}
	}

	if ri, ok := params["ri"]; ok {
		i, err := strconv.Atoi(ri)
		if err != nil {
			return nil, fmt.Errorf("dmarc: invalid parameter 'ri': %v", err)
		}
		if i <= 0 {
			return nil, fmt.Errorf("dmarc: invalid parameter 'ri': negative or zero duration")
		}
		rec.ReportInterval = time.Duration(i) * time.Second
	}

	if rua, ok := params["rua"]; ok {
		rec.ReportURIsAggregate, err = parseURIList(rua)
		if err != nil {
			return nil, err
		}
	}

	if ruf, ok := params["ruf"]; ok {
		rec.ReportURIsFailure, err = parseURIList(ruf)
		if err != nil {
			return nil, err
		}
	}

	if sp, ok := params["sp"]; ok {
		rec.SubdomainPolicy, err = parsePolicy(sp, "sp")
		if err != nil {
			return nil, err
		}
	}

	return rec, nil
}

func parseParams(s string) (map[string]string, error) {
	pairs := strings.Split(s, ";")
	params := make(map[string]string)
	for _, s := range pairs {
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			if strings.TrimSpace(s) == "" {
				continue
			}
			return params, errors.New("dmarc: malformed params")
		}

		params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return params, nil
}

func parsePolicy(s, param string) (Policy, error) {
	switch s {
	case string(PolicyNone), string(PolicyQuarantine), string(PolicyReject):
		return Policy(s), nil
	default:
		return "", fmt.Errorf("dmarc: invalid policy for parameter '%v'", param)
	}
}

func parseAlignmentMode(s, param string) (AlignmentMode, error) {
	switch s {
	case string(AlignmentModeRelaxed), string(AlignmentModeStrict):
		return AlignmentMode(s), nil
	default:
		return "", fmt.Errorf("dmarc: invalid alignment mode for parameter '%v'", param)
	}
}

func parseFailureOptions(s string) ([]FailureOption, error) {
	l := strings.Split(s, ":")
	var opts []FailureOption
	for _, o := range l {
		switch strings.TrimSpace(o) {
		case string(FailureOptionAll):
			opts = append(opts, FailureOptionAll)
		case string(FailureOptionAny):
			opts = append(opts, FailureOptionAny)
		case string(FailureOptionDKIM):
			opts = append(opts, FailureOptionDKIM)
		case string(FailureOptionSPF):
			opts = append(opts, FailureOptionSPF)
		default:
			return []FailureOption{FailureOptionAll}, errors.New("dmarc: invalid failure option")
		}
	}
	return opts, nil
}

func parseURIList(s string) ([]string, error) {
	l := strings.Split(s, ",")

	var addressParser mail.AddressParser

	for i, u := range l {
		// Remove mailto: prefix and error if not present
		if strings.HasPrefix(u, "mailto:") {
			u = u[7:]
		} else {
			fmt.Printf("dmarc: invalid URI: %v", u)
			return nil, fmt.Errorf("dmarc: invalid URI at index %v: missing mailto: prefix", i)
		}

		addr, err := addressParser.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("dmarc: invalid URI at index %v: %v", i, err)
		}

		// Add the address (including the mailto: prefix) back to the list
		l[i] = "mailto:" + addr.Address
	}

	return l, nil
}
