package main

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
)

type FloatParams []float32
type Int64Params []int64

var RecalculatingEvents []int64

func (a *FloatParams) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	s := string(b)
	s = strings.Replace(s, "{", "[", 1)
	s = strings.Replace(s, "}", "]", 1)

	return json.Unmarshal([]byte(s), &a)
}

func (a *Int64Params) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	s := string(b)
	s = strings.Replace(s, "{", "[", 1)
	s = strings.Replace(s, "}", "]", 1)

	return json.Unmarshal([]byte(s), &a)
}

// Строка с набором кэшированных параметров для отчета. Не содержит расчетных параметров, ид игрока и т.п. Исходные данные для расчета
type ReportMinimalRecord struct {
	SumLength                 float32     `db:"sum_length" json:"sum_length"`
	LpsSeconds                float32     `db:"lps_seconds" json:"lps_seconds"`
	DopplerSumLength          float32     `db:"doppler_len" json:"doppler_len"`
	DopplerMaxSpeed           float32     `db:"max_speed" json:"max_speed"`
	DopplerMaxAcceleration    float32     `db:"max_acceleration" json:"max_acceleration"`
	LengthInSpeedZones        FloatParams `db:"len_in_speed_zones" json:"len_in_speed_zones"`
	JumpCount                 int64       `db:"jump_count" json:"jump_count"`
	CountLoadData             int64       `db:"count_load_data" json:"count_load_data"`
	MaxLoad                   int32       `db:"max_load" json:"max_load"`
	SumLoad                   int32       `db:"sum_load" json:"sum_load"`
	TimeInSpeedZones          FloatParams `db:"time_in_speed_zones" json:"time_in_speed_zones"`
	CountPulseValues          int64       `db:"count_pulse_values" json:"count_pulse_values"`
	SumPulseValues            int64       `db:"sum_pulse_values" json:"sum_pulse_values"`
	MaxPulse                  int16       `db:"max_pulse" json:"max_pulse"`
	TimeInHRZones             FloatParams `db:"time_in_hr_zones" json:"time_in_hr_zones"`
	ImpactCnt                 int64       `db:"impact_cnt" json:"impact_cnt"`
	AccelCnt                  int64       `db:"accel_cnt" json:"accel_cnt"`
	StopCnt                   int64       `db:"stop_cnt" json:"stop_cnt"`
	MaxAccel_pow              float32     `db:"max_accel_pow" json:"max_accel_pow"`
	MaxStop_pow               float32     `db:"max_stop_pow" json:"max_stop_pow"`
	AccelerationCntByZones    Int64Params `db:"acceleration_cnt_by_zones" json:"acceleration_cnt_by_zones"`
	AccelerationLengthByZones FloatParams `db:"acceleration_length_by_zones" json:"acceleration_length_by_zones"`
	StopCountByZones          Int64Params `db:"stop_count_by_zones" json:"stop_count_by_zones"`
	PlayerHeight              *float32    `db:"player_height" json:"player_height"`
	PlayerWeight              *float32    `db:"player_weight" json:"player_weight"`
	PlayerMaxPulse            *int16      `db:"player_max_pulse" json:"player_max_pulse"`
	SplitsDuration            int64       `db:"splits_duration" json:"splits_duration"`
	EventDuration             int64       `db:"event_duration" json:"event_duration"`
	TimeInAccelerationZones   FloatParams `db:"time_in_acc_zones" json:"time_in_acc_zones"`
	ShiftsLeft                int64       `db:"shift_left" json:"shift_left"`
	ShiftsRight               int64       `db:"shift_right" json:"shift_right"`
}

// Строка с кэшем отчета в базе данных
type DBReportRecord struct {
	ReportMinimalRecord
	SplitID    string          `db:"split_id" json:"split_id"`
	PlayerID   int32           `db:"player_id" json:"player_id"`
	PlayerInfo json.RawMessage `db:"player_info" json:"player_info"`
	EventID    string          `db:"event_id" json:"event_id"`
	EventInfo  json.RawMessage `db:"event_info" json:"event_info"`
	SplitInfo  json.RawMessage `db:"split_info" json:"split_info"`
}

// Строка с набором параметров для отчета. Содержит расчетные параметры.
type ReportCalculatedRecord struct {
	ReportMinimalRecord
	AveragePulse    int16   `json:"average_pulse"`
	AverageSpeed    float32 `json:"average_speed"`
	Implodes        int64   `json:"implodes"`
	LoadPerMin      float32 `json:"load_per_min"`
	ActiveTime      float32 `json:"active_time"`
	ExcentricIndex  float32 `json:"excentric_index"`
	ExcentricShifts float32 `json:"excentric_shifts"`
	Energy          float32 `json:"energy"`
}

// Строка из базы отчетом опросника
type DBReportSurveyRecord struct {
	PlayerID     int32           `db:"player_id" json:"player_id"`
	PlayerInfo   json.RawMessage `db:"player_info" json:"player_info"`
	PlayerRating int32           `db:"player_rating" json:"player_rating"`
	EventRating  int32           `db:"event_rating" json:"event_rating"`
	Length21     float32         `db:"sum_length_21" json:"sum_length_21"`
	Length3      float32         `db:"sum_length_3" json:"sum_length_3"`
}

// Строка отчетом опросника
type ReportSurveyRecord struct {
	DBReportSurveyRecord
	ReportCalculatedRecord
	Imbalance float32 `db:"imbalance_coeff" json:"imbalance_coeff"`
	Acute     float32 `db:"acute_coeff" json:"acute_coeff"`
}

const (
	queryReportGetData         string = `select * from api_sight."reportGetData"($1, $2, $3);`
	queryReporCalcCache        string = `select * from api_sight."prcReporCalcCache"($1, $2, $3);`
	queryReportEventSurveyData string = `select * from api_sight."reportSurvey1"($1, $2);`
)

/*
Дополняет кэшированные параметры расчетными
*/
func MakeCalculatedParams(rec ReportCalculatedRecord) ReportCalculatedRecord {
	rec.AveragePulse = 0
	if rec.CountPulseValues != 0 {
		rec.AveragePulse = int16(math.Round(float64(rec.SumPulseValues) / float64(rec.CountPulseValues)))
	}
	rec.AverageSpeed = 0
	if rec.LpsSeconds != 0 {
		rec.AverageSpeed = float32(rec.DopplerSumLength) / float32(rec.LpsSeconds)
	}

	rec.Implodes = rec.JumpCount + rec.AccelCnt + rec.StopCnt

	rec.SumLoad = rec.SumLoad / 100 / 40

	rec.LoadPerMin = 0
	if rec.LpsSeconds != 0 {
		rec.LoadPerMin = float32(rec.SumLoad) / float32(float32(rec.LpsSeconds)/float32(60))
	}

	rec.ActiveTime = rec.LpsSeconds

	rec.ExcentricIndex = 0
	if (rec.AccelCnt + rec.StopCnt) != 0 {
		rec.ExcentricIndex = float32(rec.AccelCnt-rec.StopCnt) / float32(rec.AccelCnt+rec.StopCnt)
	}

	rec.ExcentricShifts = 0
	if (rec.ShiftsLeft + rec.ShiftsRight) != 0 {
		rec.ExcentricShifts = float32(rec.ShiftsRight-rec.ShiftsLeft) / float32(rec.ShiftsRight+rec.ShiftsLeft)
	}

	if rec.AveragePulse > 60 {
		var weight float32
		if rec.PlayerWeight != nil {
			weight = *rec.PlayerWeight
		}
		rec.Energy = 0.014 * weight * rec.LpsSeconds / 60 * (0.12*float32(rec.AveragePulse) - 7)
	}

	return rec
}

/*
Метод для схлопывания кэшированных параметров. Объединяет две строки с кэшированными ппрпметрами.
*/
func MegreReportMinimalRecords(src1 ReportMinimalRecord, src2 ReportMinimalRecord) ReportMinimalRecord {
	src1.SumLength = src1.SumLength + src2.SumLength
	src1.LpsSeconds = src1.LpsSeconds + src2.LpsSeconds
	src1.DopplerSumLength = src1.DopplerSumLength + src2.DopplerSumLength

	src1.SplitsDuration = src1.SplitsDuration + src2.SplitsDuration

	src1.DopplerMaxSpeed = float32(math.Max(float64(src1.DopplerMaxSpeed), float64(src2.DopplerMaxSpeed)))
	src1.DopplerMaxAcceleration = float32(math.Max(float64(src1.DopplerMaxAcceleration), float64(src2.DopplerMaxAcceleration)))

	for idx := range src1.LengthInSpeedZones {
		src1.LengthInSpeedZones[idx] = src1.LengthInSpeedZones[idx] + src2.LengthInSpeedZones[idx]
	}

	src1.JumpCount = src1.JumpCount + src2.JumpCount
	src1.CountLoadData = src1.CountLoadData + src2.CountLoadData
	src1.MaxLoad = int32(math.Max(float64(src1.MaxLoad), float64(src2.MaxLoad)))
	src1.SumLoad = src1.SumLoad + src2.SumLoad

	for idx := range src1.TimeInSpeedZones {
		src1.TimeInSpeedZones[idx] = src1.TimeInSpeedZones[idx] + src2.TimeInSpeedZones[idx]
	}

	src1.CountPulseValues = src1.CountPulseValues + src2.CountPulseValues
	src1.SumPulseValues = src1.SumPulseValues + src2.SumPulseValues
	src1.MaxPulse = int16(math.Max(float64(src1.MaxPulse), float64(src2.MaxPulse)))

	for idx := range src1.TimeInHRZones {
		src1.TimeInHRZones[idx] = src1.TimeInHRZones[idx] + src2.TimeInHRZones[idx]
	}
	src1.ImpactCnt = src1.ImpactCnt + src2.ImpactCnt
	src1.AccelCnt = src1.AccelCnt + src2.AccelCnt
	src1.StopCnt = src1.StopCnt + src2.StopCnt
	src1.MaxAccel_pow = float32(math.Max(float64(src1.MaxAccel_pow), float64(src2.MaxAccel_pow)))
	src1.MaxStop_pow = float32(math.Max(float64(src1.MaxStop_pow), float64(src2.MaxStop_pow)))

	for idx := range src1.AccelerationCntByZones {
		src1.AccelerationCntByZones[idx] = src1.AccelerationCntByZones[idx] + src2.AccelerationCntByZones[idx]
	}
	for idx := range src1.AccelerationLengthByZones {
		src1.AccelerationLengthByZones[idx] = src1.AccelerationLengthByZones[idx] + src2.AccelerationLengthByZones[idx]
	}
	for idx := range src1.StopCountByZones {
		src1.StopCountByZones[idx] = src1.StopCountByZones[idx] + src2.StopCountByZones[idx]
	}

	for idx := range src1.TimeInHRZones {
		src1.TimeInAccelerationZones[idx] = src1.TimeInAccelerationZones[idx] + src2.TimeInAccelerationZones[idx]
	}
	src1.ShiftsLeft = src1.ShiftsLeft + src2.ShiftsLeft
	src1.ShiftsRight = src1.ShiftsRight + src2.ShiftsRight

	return src1
}

func inArray(val interface{}, array interface{}) (index int) {
	values := reflect.ValueOf(array)

	if reflect.TypeOf(array).Kind() == reflect.Slice || values.Len() > 0 {
		for i := 0; i < values.Len(); i++ {
			if reflect.DeepEqual(val, values.Index(i).Interface()) {
				return i
			}
		}
	}

	return -1
}

// вернуть данные по сплитам для переданных сплитов или эвентов
func (h *handler) reportFetchData(c jrpc.Context) (split_data []DBReportRecord, err error) {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	var params struct {
		EventIds []string `json:"event_ids"`
		SplitIds []string `json:"split_ids"`
	}

	if err := c.Bind(&params); err != nil {
		return nil, errors.Wrap(err, "reportFetchData Bind error")
	}

	if err := h.DB.Select(&split_data, queryReportGetData, club_id, pq.StringArray(params.EventIds), pq.StringArray(params.SplitIds)); err != nil {
		return nil, errors.Wrap(err, "reportFetchData SQL error")
	}
	return split_data, nil
}

// вернуть данные по сплитам для переданного евента
func (h *handler) reportFetchDataEvent(c jrpc.Context) (split_data []DBReportRecord, err error) {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	var params string

	if err := c.Bind(&params); err != nil {
		return nil, errors.Wrap(err, "reportFetchDataEvent Bind error")
	}

	var parr []string
	parr = append(parr, params)
	var sarr []string

	if err := h.DB.Select(&split_data, queryReportGetData, club_id, pq.StringArray(parr), pq.StringArray(sarr)); err != nil {
		return nil, errors.Wrap(err, "reportFetchDataEvent SQL error")
	}
	return split_data, nil
}

// вернуть данные по опроснику эвента для переданного евента
func (h *handler) reportFetchEventSurvey(c jrpc.Context) (survey_data []DBReportSurveyRecord, err error) {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	var params string

	if err := c.Bind(&params); err != nil {
		return nil, errors.Wrap(err, "reportFetchEventSurvey Bind error")
	}

	if err := h.DB.Select(&survey_data, queryReportEventSurveyData, club_id, params); err != nil {
		return nil, errors.Wrap(err, "reportFetchEventSurvey SQL error")
	}
	return survey_data, nil
}
