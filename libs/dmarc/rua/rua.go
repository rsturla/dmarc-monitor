package rua

import (
	"encoding/xml"
	"net/netip"
)

func (f *RUA) ParseXML(data []byte) error {
	return xml.Unmarshal(data, f)
}

// RUA represents the top-level structure of a DMARC RUA report
type RUA struct {
	ReportMetadata  ReportMetadata  `xml:"report_metadata"`
	PolicyPublished PolicyPublished `xml:"policy_published"`
	Records         []Record        `xml:"record"`
}

// ReportMetadata contains metadata about the DMARC report
type ReportMetadata struct {
	OrgName          string    `xml:"org_name"`
	Email            string    `xml:"email"`
	ExtraContactInfo string    `xml:"extra_contact_info"`
	ReportID         string    `xml:"report_id"`
	DateRange        DateRange `xml:"date_range"`
}

// DateRange represents the date range for the DMARC report (in Unix time)
type DateRange struct {
	Begin int64 `xml:"begin"`
	End   int64 `xml:"end"`
}

// PolicyPublished contains information about the DMARC policy published by the domain owner
type PolicyPublished struct {
	Domain string `xml:"domain"`
	Adkim  string `xml:"adkim"`
	Aspf   string `xml:"aspf"`
	P      string `xml:"p"`
	Sp     string `xml:"sp"`
	Pct    int    `xml:"pct"`
	Np     string `xml:"np"`
}

// Record represents a single record in the DMARC feedback report
type Record struct {
	Row         Row         `xml:"row"`
	Identifiers Identifiers `xml:"identifiers"`
	AuthResults AuthResults `xml:"auth_results"`
}

// Row represents detailed information about the DMARC evaluation for a specific source IP
type Row struct {
	SourceIp        netip.Addr      `xml:"source_ip"`
	Count           int             `xml:"count"`
	PolicyEvaluated PolicyEvaluated `xml:"policy_evaluated"`
}

// PolicyEvaluated contains the result of the DMARC policy evaluation
type PolicyEvaluated struct {
	Disposition string `xml:"disposition"`
	Dkim        string `xml:"dkim"`
	Spf         string `xml:"spf"`
}

// Identifiers represent the identifiers used for the DMARC evaluation
type Identifiers struct {
	HeaderFrom string `xml:"header_from"`
}

// AuthResults contains authentication results for DKIM and SPF
type AuthResults struct {
	Dkim []DKIM `xml:"dkim"`
	Spf  SPF    `xml:"spf"`
}

// DKIM represents a single DKIM authentication result
type DKIM struct {
	Domain   string `xml:"domain"`
	Result   string `xml:"result"`
	Selector string `xml:"selector"`
}

// SPF represents a single SPF authentication result
type SPF struct {
	Domain string `xml:"domain"`
	Result string `xml:"result"`
}
