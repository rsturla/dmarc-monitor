package models

type IngestSQSMessage struct {
	TenantID     string `json:"tenantID"`
	S3ObjectPath string `json:"s3ObjectPath"`
	Timestamp    string `json:"timestamp"`
	MessageID    string `json:"messageID"`
}

type DmarcReportEntry struct {
	ID               string `dynamodbav:"id"`
	ReportID         string `dynamodbav:"reportID"`
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

type DmarcRecordEntry struct {
	ID                         string `dynamodbav:"id"`
	ReportID                   string `dynamodbav:"reportID"`
	SourceIp                   string `dynamodbav:"sourceIp"`
	Count                      int    `dynamodbav:"count"`
	PolicyEvaluatedDisposition string `dynamodbav:"policyEvaluatedDisposition"`
	PolicyEvaluatedDkim        string `dynamodbav:"policyEvaluatedDkim"`
	PolicyEvaluatedSpf         string `dynamodbav:"policyEvaluatedSpf"`
	HeaderFrom                 string `dynamodbav:"headerFrom"`
	AuthResultsDkim            string `dynamodbav:"authResultsDkim"`
	AuthResultsSpf             string `dynamodbav:"authResultsSpf"`
}
