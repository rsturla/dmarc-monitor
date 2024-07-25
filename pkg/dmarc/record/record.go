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

// Custom error types
type DMARCError struct {
	Parameter string
	Message   string
}

func (e *DMARCError) Error() string {
	return fmt.Sprintf("dmarc: error with parameter '%s': %s", e.Parameter, e.Message)
}

var (
	ErrMissingParam     = errors.New("dmarc: missing required parameter")
	ErrInvalidParam     = errors.New("dmarc: invalid parameter")
	ErrInvalidURI       = errors.New("dmarc: invalid URI")
	ErrNegativeDuration = errors.New("dmarc: negative or zero duration")
	ErrOutOfBounds      = errors.New("dmarc: value out of bounds")
)

func ParseRecord(txt string) (*Record, error) {
	params, err := parseParams(txt)
	if err != nil {
		return nil, err
	}

	rec := new(Record)

	v, ok := params["v"]
	if !ok {
		return nil, &DMARCError{Parameter: "v", Message: ErrMissingParam.Error()}
	}
	rec.Version = v

	p, ok := params["p"]
	if !ok {
		return nil, &DMARCError{Parameter: "p", Message: ErrMissingParam.Error()}
	}
	rec.Policy, err = parsePolicy(p, "p")
	if err != nil {
		return nil, err
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
			return nil, &DMARCError{Parameter: "pct", Message: ErrInvalidParam.Error()}
		}
		if i < 0 || i > 100 {
			return nil, &DMARCError{Parameter: "pct", Message: ErrOutOfBounds.Error()}
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
				return nil, &DMARCError{Parameter: "rf", Message: ErrInvalidParam.Error()}
			}
		}
	}

	if ri, ok := params["ri"]; ok {
		i, err := strconv.Atoi(ri)
		if err != nil {
			return nil, &DMARCError{Parameter: "ri", Message: ErrInvalidParam.Error()}
		}
		if i <= 0 {
			return nil, &DMARCError{Parameter: "ri", Message: ErrNegativeDuration.Error()}
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
			return params, &DMARCError{Parameter: s, Message: "malformed parameter"}
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
		return "", &DMARCError{Parameter: param, Message: ErrInvalidParam.Error()}
	}
}

func parseAlignmentMode(s, param string) (AlignmentMode, error) {
	switch s {
	case string(AlignmentModeRelaxed), string(AlignmentModeStrict):
		return AlignmentMode(s), nil
	default:
		return "", &DMARCError{Parameter: param, Message: ErrInvalidParam.Error()}
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
			return []FailureOption{FailureOptionAll}, &DMARCError{Parameter: "fo", Message: ErrInvalidParam.Error()}
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
			return nil, &DMARCError{Parameter: "rua/ruf", Message: ErrInvalidURI.Error()}
		}

		addr, err := addressParser.Parse(u)
		if err != nil {
			return nil, &DMARCError{Parameter: fmt.Sprintf("uri index %d", i), Message: err.Error()}
		}

		// Add the address (including the mailto: prefix) back to the list
		l[i] = "mailto:" + addr.Address
	}

	return l, nil
}
