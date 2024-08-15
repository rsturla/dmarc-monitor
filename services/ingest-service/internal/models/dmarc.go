package models

// DmarcReportMetadataItem represents a DMARC report item in the DynamoDB table.  This item
// contains the metadata for a DMARC report.
type DmarcReportMetadataItem struct {
	ID               string `dynamodbav:"id"`
	ReportId         string `dynamodbav:"reportId"`
	OrgName          string `dynamodbav:"orgName"`
	Email            string `dynamodbav:"email"`
	ExtraContactInfo string `dynamodbav:"extraContactInfo"`
	DateRangeBegin   int64  `dynamodbav:"dateRangeBegin"`
	DateRangeEnd     int64  `dynamodbav:"dateRangeEnd"`
	Domain           string `dynamodbav:"domain"`
	Adkim            string `dynamodbav:"adkim"`
	Aspf             string `dynamodbav:"aspf"`
	P                string `dynamodbav:"p"`
	Sp               string `dynamodbav:"sp"`
	Pct              int    `dynamodbav:"pct"`
	Np               string `dynamodbav:"np"`
}

// DmarcRecordItem represents a DMARC record item in the DynamoDB table.  This item
// contains the details of a DMARC record.  Each record is associated with a single DMARC report.
type DmarcRecordItem struct {
	ID                         string                           `dynamodbav:"id"`
	ReportId                   string                           `dynamodbav:"reportId"`
	SourceIp                   string                           `dynamodbav:"sourceIp"`
	Count                      int                              `dynamodbav:"count"`
	PolicyEvaluatedDisposition string                           `dynamodbav:"policyEvaluatedDisposition"`
	PolicyEvaluatedDkim        string                           `dynamodbav:"policyEvaluatedDkim"`
	PolicyEvaluatedSpf         string                           `dynamodbav:"policyEvaluatedSpf"`
	HeaderFrom                 string                           `dynamodbav:"headerFrom"`
	AuthResultsDkim            []DmarcAuthResultNestedAttribute `dynamodbav:"authResultsDkim"`
	AuthResultsSpf             DmarcAuthResultNestedAttribute   `dynamodbav:"authResultsSpf"`
}

// DmarcAuthResultNestedAttribute represents a nested attribute for the DMARC record item in the DynamoDB table.
// This attribute contains the details of the authentication results for a specific domain.
type DmarcAuthResultNestedAttribute struct {
	Domain   string `dynamodbav:"domain"`
	Result   string `dynamodbav:"result"`
	Selector string `dynamodbav:"selector"`
}
