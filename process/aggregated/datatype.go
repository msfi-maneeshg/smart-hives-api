package aggregated

type GetDBData struct {
	TotalRows int64        `json:"total_rows"`
	Rows      []DBDataRows `json:"rows"`
}

type DBDataRows struct {
	Doc DBDataRowsDoc `json:"doc"`
}
type DBDataRowsDoc struct {
	Data       DBDataRowsDocData `json:"data"`
	DeviceID   string            `json:"deviceId"`
	DeviceType string            `json:"deviceType"`
	Timestamp  string            `json:"timestamp"`
}

type DBDataRowsDocData struct {
	Timestamp   string `json:"timestamp"`
	Humidity    int64  `json:"humidity"`
	Temperature int64  `json:"temperature"`
	Weight      int64  `json:"weight"`
}

type HiveDataSet struct {
	TotalTemperature, TotalRecords, AvgTemperature int64
	MinTemperature, MaxTemperature                 *int64
	TotalHumidity, AvgHumidity                     int64
	MinHumidity, MaxHumidity                       *int64
	TotalWeight, AvgWeight                         int64
	MinWeight, MaxWeight                           *int64
	EventName                                      string
	HiveEventData                                  []DBDataRowsDocData
}

type DeviceTypeResultSet struct {
	Result []DeviceTypeDetail `json:"results"`
}

type DeviceTypeDetail struct {
	ID string `json:"id"`
}
