package dmarc

import (
	"fmt"

	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/dmarc/rua"
	"github.com/rsturla/dmarc-monitor/services/ingest-service/internal/models"
)

func ParseRUAReport(body []byte) (*rua.RUA, error) {
	var ruaReport rua.RUA
	if err := ruaReport.ParseXML(body); err != nil {
		return nil, fmt.Errorf("error parsing RUA report: %w", err)
	}
	return &ruaReport, nil
}

func CreateDmarcReportItem(sqsMessage models.IngestMessage, ruaReport *rua.RUA) models.DmarcReportMetadataItem {
	return models.DmarcReportMetadataItem{
		ID:               fmt.Sprintf("%s#%s", sqsMessage.TenantID, ruaReport.ReportMetadata.ReportID),
		ReportId:         ruaReport.ReportMetadata.ReportID,
		OrgName:          ruaReport.ReportMetadata.OrgName,
		Email:            ruaReport.ReportMetadata.Email,
		ExtraContactInfo: ruaReport.ReportMetadata.ExtraContactInfo,
		DateRangeBegin:   ruaReport.ReportMetadata.DateRange.Begin,
		DateRangeEnd:     ruaReport.ReportMetadata.DateRange.End,
		Domain:           ruaReport.PolicyPublished.Domain,
		Adkim:            ruaReport.PolicyPublished.Adkim,
		Aspf:             ruaReport.PolicyPublished.Aspf,
		P:                ruaReport.PolicyPublished.P,
		Sp:               ruaReport.PolicyPublished.Sp,
		Pct:              ruaReport.PolicyPublished.Pct,
		Np:               ruaReport.PolicyPublished.Np,
	}
}

func CreateDmarcRecordItems(dmarcReportItem models.DmarcReportMetadataItem, ruaReport *rua.RUA) []models.DmarcRecordItem {
	var dmarcRecordItems []models.DmarcRecordItem
	for i, record := range ruaReport.Records {
		var authResultsDkim []models.DmarcAuthResultNestedAttribute
		for _, dkim := range record.AuthResults.Dkim {
			authResultsDkim = append(authResultsDkim, models.DmarcAuthResultNestedAttribute{
				Domain:   dkim.Domain,
				Result:   dkim.Result,
				Selector: dkim.Selector,
			})
		}

		dmarcRecordItems = append(dmarcRecordItems, models.DmarcRecordItem{
			ID:                         fmt.Sprintf("%s#%d", dmarcReportItem.ID, i),
			ReportId:                   dmarcReportItem.ReportId,
			SourceIp:                   record.Row.SourceIp.String(),
			Count:                      record.Row.Count,
			PolicyEvaluatedDisposition: record.Row.PolicyEvaluated.Disposition,
			PolicyEvaluatedDkim:        record.Row.PolicyEvaluated.Dkim,
			PolicyEvaluatedSpf:         record.Row.PolicyEvaluated.Spf,
			HeaderFrom:                 record.Identifiers.HeaderFrom,
			AuthResultsDkim:            authResultsDkim,
			AuthResultsSpf: models.DmarcAuthResultNestedAttribute{
				Domain: record.AuthResults.Spf.Domain,
				Result: record.AuthResults.Spf.Result,
			},
		})
	}
	return dmarcRecordItems
}
