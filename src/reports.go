package main

import (
	"encoding/json"

	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
)

// Строка отчета "по тренировке" во внутреннем формате
type CommonReportFullRecord struct {
	ReportCalculatedRecord
	PlayerID   int32           `json:"player_id"`
	PlayerInfo json.RawMessage `json:"player_info"`
}

// Строка "матч"-отчета для таблицы во внутреннем формате
type MatchReportFullRecord struct {
	EventID   string                      `json:"event_id"`
	EventInfo json.RawMessage             `json:"event_info"`
	Data      []MatchReportFullDataRecord `json:"data"`
}

// Строка отчета "по тренировке" во внутреннем формате
type MatchReportFullDataRecord struct {
	ReportCalculatedRecord
	SplitID    string          `json:"split_id"`
	SplitInfo  json.RawMessage `json:"split_info"`
	PlayerID   int32           `json:"player_id"`
	PlayerInfo json.RawMessage `json:"player_info"`
}

// Строка "матч"-отчета для графика во внутреннем формате
type MatchGrapgFullRecord struct {
	PlayerID     int32                       `json:"player_id"`
	PlayerInfo   json.RawMessage             `json:"player_info"`
	Data         []MatchReportFullDataRecord `json:"splits"`
	SumLength    float32                     `json:"sum_length"`
	SumLoad      int32                       `json:"sum_load"`
	SumDuration  int64                       `json:"duration"`
	LengthPerMin float32                     `json:"length_per_min"`
	LoadPerMin   float32                     `json:"load_per_min"`
	AccelCnt     int64                       `json:"accel_cnt"`
	StopCnt      int64                       `json:"stop_cnt"`
}

// Строка индивидуального отчета для таблицы во внутреннем формате
type IndividualReportFullRecord struct {
	PlayerID   int32                            `json:"id"`
	PlayerInfo json.RawMessage                  `json:"player_info"`
	Data       []IndividualReportFullDataRecord `json:"data"`
}

// Строка отчета "по тренировке" во внутреннем формате
type IndividualReportFullDataRecord struct {
	ReportCalculatedRecord
	EventID   string          `json:"event_id"`
	EventInfo json.RawMessage `json:"event_info"`
}

/*
Отчет по тренировке
*/
func (h *handler) reportWorkout(c jrpc.Context) error {
	split_data, err := h.reportFetchData(c)
	if err != nil {
		return errors.Wrap(err, "reportCommon FetchData error")
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

	for idx := range imploded_data {
		imploded_data[idx].ReportCalculatedRecord = MakeCalculatedParams(imploded_data[idx].ReportCalculatedRecord)
	}

	var report_data []map[string]interface{}

	for _, element := range imploded_data {
		rec := map[string]interface{}{}
		rec["player_id"] = element.PlayerID
		rec["player_info"] = element.PlayerInfo
		rec["sum_length"] = element.SumLength
		rec["duration"] = element.SplitsDuration

		rec["len_spd_1"] = element.LengthInSpeedZones[0]
		rec["len_spd_2"] = element.LengthInSpeedZones[1]
		rec["len_spd_3"] = element.LengthInSpeedZones[2]
		rec["len_spd_4"] = element.LengthInSpeedZones[3]
		rec["len_spd_5"] = element.LengthInSpeedZones[4]

		rec["impact_cnt"] = element.ImpactCnt
		rec["average_pulse"] = element.AveragePulse
		rec["max_pulse"] = element.MaxPulse

		rec["hr_time1"] = element.TimeInHRZones[0]
		rec["hr_time2"] = element.TimeInHRZones[1]
		rec["hr_time3"] = element.TimeInHRZones[2]
		rec["hr_time4"] = element.TimeInHRZones[3]
		rec["hr_time5"] = element.TimeInHRZones[4]

		rec["sum_load"] = element.SumLoad
		rec["accel_cnt"] = element.AccelCnt
		rec["stop_cnt"] = element.StopCnt

		rec["jump_count"] = element.JumpCount
		rec["implodes"] = element.Implodes
		rec["load_per_min"] = element.LoadPerMin
		rec["active_time"] = element.ActiveTime
		rec["max_speed"] = element.DopplerMaxSpeed
		rec["excentric_index"] = element.ExcentricIndex
		rec["excentric_shifts"] = element.ExcentricShifts
		rec["shift_left"] = element.ShiftsLeft
		rec["shift_right"] = element.ShiftsRight

		rec["energy"] = element.Energy

		report_data = append(report_data, rec)
	}

	return c.Result(report_data)
}

/*
Матч отчет для таблицы
*/
func (h *handler) reportMatchTable(c jrpc.Context) error {
	split_data, err := h.reportFetchData(c)
	if err != nil {
		return errors.Wrap(err, "reportMatchTable FetchData error")
	}

	var imploded_data []MatchReportFullRecord

	for _, element := range split_data {
		var event_found bool = false
		for event_idx, event := range imploded_data {
			if event.EventID == element.EventID {
				event_found = true
				var found bool = false
				for idx, imploded := range event.Data {
					if imploded.SplitID == element.SplitID && imploded.PlayerID == element.PlayerID {
						found = true

						imploded_data[event_idx].Data[idx].ReportMinimalRecord = MegreReportMinimalRecords(imploded_data[event_idx].Data[idx].ReportMinimalRecord, element.ReportMinimalRecord)
					}
				}
				if !found {
					var rec MatchReportFullDataRecord
					rec.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
					rec.SplitID = element.SplitID
					rec.SplitInfo = element.SplitInfo
					rec.PlayerID = element.PlayerID
					rec.PlayerInfo = element.PlayerInfo

					imploded_data[event_idx].Data = append(imploded_data[event_idx].Data, rec)
				}

			}
		}

		if !event_found {
			var rec MatchReportFullRecord
			rec.EventID = element.EventID
			rec.EventInfo = element.EventInfo

			var rec1 MatchReportFullDataRecord
			rec1.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
			rec1.SplitID = element.SplitID
			rec1.SplitInfo = element.SplitInfo
			rec1.PlayerID = element.PlayerID
			rec1.PlayerInfo = element.PlayerInfo

			rec.Data = append(rec.Data, rec1)

			imploded_data = append(imploded_data, rec)
		}
	}

	for event_idx, event := range imploded_data {
		for idx := range event.Data {
			imploded_data[event_idx].Data[idx].ReportCalculatedRecord = MakeCalculatedParams(imploded_data[event_idx].Data[idx].ReportCalculatedRecord)
		}
	}

	var report_data []map[string]interface{}

	for _, event := range imploded_data {
		rec := make(map[string]interface{})
		rec["id"] = event.EventID
		rec["event_info"] = event.EventInfo

		recdata := make([]map[string]interface{}, 0)
		for _, data := range event.Data {
			event_data := make(map[string]interface{})
			event_data["split_info"] = data.SplitInfo
			event_data["player_id"] = data.PlayerID
			event_data["player_info"] = data.PlayerInfo

			event_data["duration"] = data.SplitsDuration
			event_data["sum_length"] = data.SumLength

			event_data["len_spd_1"] = data.LengthInSpeedZones[0]
			event_data["len_spd_2"] = data.LengthInSpeedZones[1]
			event_data["len_spd_3"] = data.LengthInSpeedZones[2]
			event_data["len_spd_4"] = data.LengthInSpeedZones[3]
			event_data["len_spd_5"] = data.LengthInSpeedZones[4]

			event_data["impacts"] = data.ImpactCnt
			event_data["average_pulse"] = data.AveragePulse

			event_data["max_pulse"] = data.MaxPulse

			event_data["hr_time1"] = data.TimeInHRZones[0]
			event_data["hr_time2"] = data.TimeInHRZones[1]
			event_data["hr_time3"] = data.TimeInHRZones[2]
			event_data["hr_time4"] = data.TimeInHRZones[3]
			event_data["hr_time5"] = data.TimeInHRZones[4]

			event_data["sum_load"] = data.SumLoad
			event_data["accel_cnt"] = data.AccelCnt
			event_data["stop_cnt"] = data.StopCnt

			event_data["jump_count"] = data.JumpCount
			event_data["implodes"] = data.Implodes
			event_data["load_per_min"] = data.LoadPerMin
			event_data["active_time"] = data.ActiveTime
			event_data["max_speed"] = data.DopplerMaxSpeed
			event_data["excentric_index"] = data.ExcentricIndex
			event_data["excentric_shifts"] = data.ExcentricShifts

			event_data["energy"] = data.Energy

			recdata = append(recdata, event_data)
		}

		rec["data"] = recdata

		report_data = append(report_data, rec)
	}

	return c.Result(report_data)

}

/*
Матч отчет для графика
*/
func (h *handler) reportMatchGraph(c jrpc.Context) error {
	split_data, err := h.reportFetchData(c)
	if err != nil {
		return errors.Wrap(err, "reportMatchGraph FetchData error")
	}

	var imploded_data []MatchGrapgFullRecord

	for _, element := range split_data {
		var player_found bool = false
		for player_idx, player := range imploded_data {
			if player.PlayerID == element.PlayerID {
				player_found = true
				var found bool = false
				for idx, imploded := range player.Data {
					if imploded.SplitID == element.SplitID && player.PlayerID == element.PlayerID {
						found = true

						imploded_data[player_idx].Data[idx].ReportMinimalRecord = MegreReportMinimalRecords(imploded_data[player_idx].Data[idx].ReportMinimalRecord, element.ReportMinimalRecord)
					}
				}
				if !found {
					var rec MatchReportFullDataRecord
					rec.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
					rec.SplitID = element.SplitID
					rec.SplitInfo = element.SplitInfo
					rec.PlayerID = element.PlayerID
					rec.PlayerInfo = element.PlayerInfo

					imploded_data[player_idx].Data = append(imploded_data[player_idx].Data, rec)
				}

			}
		}

		if !player_found {
			var rec MatchGrapgFullRecord
			rec.PlayerID = element.PlayerID
			rec.PlayerInfo = element.PlayerInfo

			var rec1 MatchReportFullDataRecord
			rec1.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
			rec1.SplitID = element.SplitID
			rec1.SplitInfo = element.SplitInfo
			rec1.PlayerID = element.PlayerID
			rec1.PlayerInfo = element.PlayerInfo

			rec.Data = append(rec.Data, rec1)

			imploded_data = append(imploded_data, rec)
		}
	}

	for player_idx, player := range imploded_data {
		imploded_data[player_idx].SumLoad = 0
		imploded_data[player_idx].SumLength = 0
		imploded_data[player_idx].SumDuration = 0
		imploded_data[player_idx].LengthPerMin = 0
		imploded_data[player_idx].LoadPerMin = 0
		imploded_data[player_idx].AccelCnt = 0
		imploded_data[player_idx].StopCnt = 0

		for idx := range player.Data {
			imploded_data[player_idx].Data[idx].ReportCalculatedRecord = MakeCalculatedParams(imploded_data[player_idx].Data[idx].ReportCalculatedRecord)
			imploded_data[player_idx].SumLoad = imploded_data[player_idx].SumLoad + imploded_data[player_idx].Data[idx].SumLoad
			imploded_data[player_idx].SumLength = imploded_data[player_idx].SumLength + imploded_data[player_idx].Data[idx].SumLength

			imploded_data[player_idx].SumDuration = imploded_data[player_idx].SumDuration + int64(imploded_data[player_idx].Data[idx].LpsSeconds)
			imploded_data[player_idx].AccelCnt = imploded_data[player_idx].AccelCnt + imploded_data[player_idx].Data[idx].AccelCnt
			imploded_data[player_idx].StopCnt = imploded_data[player_idx].StopCnt + imploded_data[player_idx].Data[idx].StopCnt
		}

		if imploded_data[player_idx].SumDuration != 0 {
			imploded_data[player_idx].LengthPerMin = float32(imploded_data[player_idx].SumLength) / (float32(imploded_data[player_idx].SumDuration) / 60)
			imploded_data[player_idx].LoadPerMin = float32(imploded_data[player_idx].SumLoad) / (float32(imploded_data[player_idx].SumDuration) / 60)
		}
	}

	var report_data []map[string]interface{}

	for _, event := range imploded_data {
		rec := map[string]interface{}{}
		rec["id"] = event.PlayerID
		rec["player_info"] = event.PlayerInfo
		rec["sum_load"] = event.SumLoad
		rec["sum_length"] = event.SumLength
		rec["duration"] = event.SumDuration
		rec["length_per_min"] = event.LengthPerMin
		rec["load_per_min"] = event.LoadPerMin
		rec["accel_cnt"] = event.AccelCnt
		rec["stop_cnt"] = event.StopCnt

		recdata := make([]map[string]interface{}, 0)

		for _, data := range event.Data {
			event_data := make(map[string]interface{})
			event_data["split_info"] = data.SplitInfo

			event_data["duration"] = data.SplitsDuration
			event_data["sum_length"] = data.SumLength

			event_data["len_spd_1"] = data.LengthInSpeedZones[0]
			event_data["len_spd_2"] = data.LengthInSpeedZones[1]
			event_data["len_spd_3"] = data.LengthInSpeedZones[2]
			event_data["len_spd_4"] = data.LengthInSpeedZones[3]
			event_data["len_spd_5"] = data.LengthInSpeedZones[4]

			event_data["impacts"] = data.ImpactCnt
			event_data["average_pulse"] = data.AveragePulse

			event_data["max_pulse"] = data.MaxPulse

			event_data["hr_time1"] = data.TimeInHRZones[0]
			event_data["hr_time2"] = data.TimeInHRZones[1]
			event_data["hr_time3"] = data.TimeInHRZones[2]
			event_data["hr_time4"] = data.TimeInHRZones[3]
			event_data["hr_time5"] = data.TimeInHRZones[4]

			event_data["sum_load"] = data.SumLoad
			event_data["accel_cnt"] = data.AccelCnt
			event_data["stop_cnt"] = data.StopCnt

			event_data["jump_count"] = data.JumpCount
			event_data["implodes"] = data.Implodes
			event_data["load_per_min"] = data.LoadPerMin
			event_data["active_time"] = data.ActiveTime
			event_data["max_speed"] = data.DopplerMaxSpeed
			event_data["excentric_index"] = data.ExcentricIndex

			event_data["energy"] = data.Energy

			recdata = append(recdata, event_data)
		}

		rec["splits"] = recdata

		report_data = append(report_data, rec)
	}

	return c.Result(report_data)

}

/*
Индивидуальный отчет
*/
func (h *handler) reportPersonal(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	split_data, err := h.reportFetchData(c)
	if err != nil {
		return errors.Wrap(err, "reportPersonal FetchData error")
	}

	var imploded_data []IndividualReportFullRecord

	for _, element := range split_data {
		var player_found bool = false
		for player_idx, player := range imploded_data {
			if player.PlayerID == element.PlayerID {
				player_found = true
				var found bool = false
				for idx, imploded := range player.Data {
					if imploded.EventID == element.EventID && player.PlayerID == element.PlayerID {
						found = true

						imploded_data[player_idx].Data[idx].ReportMinimalRecord = MegreReportMinimalRecords(imploded_data[player_idx].Data[idx].ReportMinimalRecord, element.ReportMinimalRecord)
					}
				}
				if !found {
					var rec IndividualReportFullDataRecord
					rec.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
					rec.EventID = element.EventID
					rec.EventInfo = element.EventInfo

					imploded_data[player_idx].Data = append(imploded_data[player_idx].Data, rec)
				}

			}
		}

		if !player_found {
			var rec IndividualReportFullRecord
			rec.PlayerID = element.PlayerID
			rec.PlayerInfo = element.PlayerInfo

			var rec1 IndividualReportFullDataRecord
			rec1.ReportCalculatedRecord.ReportMinimalRecord = element.ReportMinimalRecord
			rec1.EventID = element.EventID
			rec1.EventInfo = element.EventInfo

			rec.Data = append(rec.Data, rec1)

			imploded_data = append(imploded_data, rec)
		}
	}

	for player_idx, player := range imploded_data {
		for idx := range player.Data {
			imploded_data[player_idx].Data[idx].ReportCalculatedRecord = MakeCalculatedParams(imploded_data[player_idx].Data[idx].ReportCalculatedRecord)
		}
	}

	var report_data []map[string]interface{}

	for _, event := range imploded_data {
		rec := map[string]interface{}{}
		rec["id"] = event.PlayerID
		rec["player_info"] = event.PlayerInfo

		recdata := make([]map[string]interface{}, 0)
		rec_totals := map[string]interface{}{}
		rec_totals["event_count"] = 0
		rec_totals["sum_length"] = float32(0)
		rec_totals["duration"] = int64(0)
		rec_totals["sum_load"] = 0
		rec_totals["sum_implodes"] = int64(0)
		rec_totals["avg_excentric_shifts"] = float32(0)
		rec_totals["avg_excentric_index"] = float32(0)
		rec_totals["avg_imbalance_coeff"] = float32(0)
		rec_totals["last_acute_coeff"] = float32(0)
		rec_totals["accel_cnt"] = int64(0)
		rec_totals["stop_cnt"] = int64(0)
		rec_totals["shift_left"] = int64(0)
		rec_totals["shift_right"] = int64(0)
		rec_totals["max_pulse"] = int16(0)
		rec_totals["energy"] = float32(0)

		for _, data := range event.Data {
			event_data := make(map[string]interface{})
			event_data["event_id"] = data.EventID
			event_data["event_info"] = data.EventInfo

			event_data["duration"] = data.SplitsDuration
			event_data["sum_length"] = data.SumLength

			event_data["len_spd_1"] = data.LengthInSpeedZones[0]
			event_data["len_spd_2"] = data.LengthInSpeedZones[1]
			event_data["len_spd_3"] = data.LengthInSpeedZones[2]
			event_data["len_spd_4"] = data.LengthInSpeedZones[3]
			event_data["len_spd_5"] = data.LengthInSpeedZones[4]

			event_data["impacts"] = data.ImpactCnt
			event_data["average_pulse"] = data.AveragePulse

			event_data["max_pulse"] = data.MaxPulse

			event_data["hr_time1"] = data.TimeInHRZones[0]
			event_data["hr_time2"] = data.TimeInHRZones[1]
			event_data["hr_time3"] = data.TimeInHRZones[2]
			event_data["hr_time4"] = data.TimeInHRZones[3]
			event_data["hr_time5"] = data.TimeInHRZones[4]

			event_data["sum_load"] = data.SumLoad
			event_data["accel_cnt"] = data.AccelCnt
			event_data["stop_cnt"] = data.StopCnt

			event_data["jump_count"] = data.JumpCount
			event_data["implodes"] = data.Implodes
			event_data["load_per_min"] = data.LoadPerMin
			event_data["active_time"] = data.ActiveTime
			event_data["max_speed"] = data.DopplerMaxSpeed
			event_data["excentric_index"] = data.ExcentricIndex

			event_data["shift_left"] = data.ShiftsLeft
			event_data["shift_right"] = data.ShiftsRight
			event_data["excentric_shifts"] = data.ExcentricShifts

			event_data["accel_cnt"] = data.AccelCnt
			event_data["stop_cnt"] = data.StopCnt

			event_data["energy"] = data.Energy

			//recdata = append(recdata, event_data)

			rec_totals["event_count"] = rec_totals["event_count"].(int) + 1
			rec_totals["sum_length"] = rec_totals["sum_length"].(float32) + data.SumLength
			rec_totals["duration"] = rec_totals["duration"].(int64) + int64(data.LpsSeconds)
			rec_totals["sum_load"] = rec_totals["sum_load"].(int) + int(data.SumLoad)
			rec_totals["sum_implodes"] = rec_totals["sum_implodes"].(int64) + data.Implodes
			rec_totals["avg_excentric_shifts"] = rec_totals["avg_excentric_shifts"].(float32) + data.ExcentricShifts
			rec_totals["avg_excentric_index"] = rec_totals["avg_excentric_index"].(float32) + data.ExcentricIndex
			rec_totals["accel_cnt"] = rec_totals["accel_cnt"].(int64) + data.AccelCnt
			rec_totals["stop_cnt"] = rec_totals["stop_cnt"].(int64) + data.StopCnt
			rec_totals["shift_left"] = rec_totals["shift_left"].(int64) + data.ShiftsLeft
			rec_totals["shift_right"] = rec_totals["shift_right"].(int64) + data.ShiftsRight
			rec_totals["energy"] = rec_totals["energy"].(float32) + data.Energy
			if data.MaxPulse > rec_totals["max_pulse"].(int16) {
				rec_totals["max_pulse"] = data.MaxPulse
			}

			var survey_data []DBReportSurveyRecord
			if err := h.DB.Select(&survey_data, queryReportEventSurveyData, club_id, data.EventID); err != nil {
				return errors.Wrap(err, "reportFetchEventSurvey SQL error")
			}

			event_data["sum_length_21"] = 0
			event_data["sum_length_3"] = 0
			event_data["injury_ratio"] = 0
			for _, survey_element := range survey_data {
				if survey_element.PlayerID == event.PlayerID {
					if data.SumLoad != 0 {
						rec_totals["avg_imbalance_coeff"] = rec_totals["avg_imbalance_coeff"].(float32) + float32(survey_element.PlayerRating)/float32(data.SumLoad)
					}

					if survey_element.Length21 != 0 {
						rec_totals["last_acute_coeff"] = float32(survey_element.Length3) / float32(survey_element.Length21)
					}

					event_data["sum_length_21"] = survey_element.Length21
					event_data["sum_length_3"] = survey_element.Length3
					if survey_element.Length21 != 0 {
						event_data["injury_ratio"] = survey_element.Length3 / survey_element.Length21
					}

				}
			}

			recdata = append(recdata, event_data)
		}
		if rec_totals["event_count"].(int) != 0 {
			rec_totals["avg_excentric_shifts"] = rec_totals["avg_excentric_shifts"].(float32) / float32(rec_totals["event_count"].(int))
			rec_totals["avg_excentric_index"] = rec_totals["avg_excentric_index"].(float32) / float32(rec_totals["event_count"].(int))
			rec_totals["avg_imbalance_coeff"] = rec_totals["avg_imbalance_coeff"].(float32) / float32(rec_totals["event_count"].(int))
		} else {
			rec_totals["avg_excentric_shifts"] = float32(0)
			rec_totals["avg_excentric_index"] = float32(0)
			rec_totals["avg_imbalance_coeff"] = float32(0)
		}

		rec["data"] = recdata
		rec["totals"] = rec_totals

		report_data = append(report_data, rec)
	}

	return c.Result(report_data)

}
