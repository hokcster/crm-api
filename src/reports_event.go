package main

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
)

// Строка индивидуального отчета для таблицы во внутреннем формате
type EventGraphResponseRecord struct {
	Date  time.Time `json:"date" db:"rep_date"`
	Len21 float32   `json:"sum_length_21" db:"sum_length_21"`
	Len3  float32   `json:"sum_length_3" db:"sum_length_3"`
	Coeff float32   `json:"injury_ratio" db:"coeff"`
}

/*
Отчет по тренировке
*/
func (h *handler) reportEventSurvey(c jrpc.Context) error {
	survey_data, err := h.reportFetchEventSurvey(c)
	if err != nil {
		return errors.Wrap(err, "reportEventSurvey FetchData error")
	}

	split_data, err := h.reportFetchDataEvent(c)
	if err != nil {
		return errors.Wrap(err, "reportEventSurvey FetchData error")
	}

	var imploded_data []CommonReportFullRecord

	for _, element := range split_data {
		var found bool = false
		for idx, imploded := range imploded_data {
			if imploded.PlayerID == element.PlayerID {
				found = true
				imploded_data[idx].ReportMinimalRecord = MegreReportMinimalRecords(imploded_data[idx].ReportMinimalRecord, element.ReportMinimalRecord)
			}
		}

		if !found {
			var rec CommonReportFullRecord
			rec.ReportMinimalRecord = element.ReportMinimalRecord
			rec.PlayerID = element.PlayerID
			rec.PlayerInfo = element.PlayerInfo

			imploded_data = append(imploded_data, rec)
		}
	}

	//	for idx := range imploded_data {
	//		imploded_data[idx].ReportCalculatedRecord = MakeCalculatedParams(imploded_data[idx].ReportCalculatedRecord)
	//	}

	var report_data []ReportSurveyRecord

	for _, survey_element := range survey_data {
		var rec ReportSurveyRecord
		rec.PlayerID = survey_element.PlayerID
		rec.PlayerInfo = survey_element.PlayerInfo
		rec.PlayerRating = survey_element.PlayerRating
		rec.EventRating = survey_element.EventRating
		rec.Length21 = survey_element.Length21
		rec.Length3 = survey_element.Length3
		for _, element := range imploded_data {
			if element.PlayerID == survey_element.PlayerID {
				rec.ReportMinimalRecord = MegreReportMinimalRecords(rec.ReportMinimalRecord, element.ReportMinimalRecord)
			}
		}
		report_data = append(report_data, rec)
	}

	for idx := range report_data {
		report_data[idx].ReportCalculatedRecord = MakeCalculatedParams(report_data[idx].ReportCalculatedRecord)
	}

	for idx := range report_data {
		if report_data[idx].SumLoad != 0 {
			report_data[idx].Imbalance = float32(report_data[idx].PlayerRating) / float32(report_data[idx].SumLoad)
		}
		if report_data[idx].Length21 != 0 {
			report_data[idx].Acute = float32(report_data[idx].Length3) / float32(report_data[idx].Length21)
		}
	}

	return c.Result(report_data)
}

/*
График по тренировке для injure prevention
*/
func (h *handler) reportEventGraph(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	var params string

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "reportEventGraph Bind error")
	}

	var reportData []struct {
		Date  time.Time `json:"date" db:"rep_date"`
		Len21 float32   `json:"sum_length_21" db:"sum_length_21"`
		Len3  float32   `json:"sum_length_3" db:"sum_length_3"`
	}

	if err := h.DB.Select(&reportData, `select * from api_sight."reportSurveyGraph"($1, $2)`, club_id, params); err != nil {
		return errors.Wrap(err, "reportEventGraph SQL error")
	}

	var responseData []EventGraphResponseRecord

	for _, element := range reportData {
		var rec EventGraphResponseRecord

		rec.Date = element.Date
		rec.Len21 = element.Len21
		rec.Len3 = element.Len3
		rec.Coeff = 0
		if element.Len21 != 0 {
			rec.Coeff = element.Len3 / element.Len21
		}

		responseData = append(responseData, rec)
	}

	return c.Result(responseData)
}
