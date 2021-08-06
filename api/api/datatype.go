package api

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
}

type DBDataRowsDocData struct {
	Timestamp   string `json:"timestamp"`
	Humidity    int64  `json:"humidity"`
	Temperature int64  `json:"temperature"`
	Weight      int64  `json:"weight"`
}

type GetHiveData struct {
	ID             string              `json:"_id"`
	Date           string              `json:"date"`
	DeviceID       string              `json:"deviceID"`
	MaxWeight      int                 `json:"maxWeight"`
	MinWeight      int                 `json:"minWeight"`
	AvgHumidity    int                 `json:"avgHumidity"`
	MaxHumidity    int                 `json:"maxHumidity"`
	AvgWeight      int                 `json:"avgWeight"`
	MinTemperature int                 `json:"minTemperature"`
	MaxTemperature int                 `json:"maxTemperature"`
	AvgTemperature int                 `json:"avgTemperature"`
	MinHumidity    int                 `json:"minHumidity"`
	Data           []DBDataRowsDocData `json:"data"`
}

type DeviceLastEventInfo struct {
	Timestamp string     `json:"timestamp"`
	Payload   string     `json:"payload"`
	Data      DeviceData `json:"data"`
}

type DeviceData struct {
	Temperature int64 `json:"temperature"`
	Humidity    int64 `json:"humidity"`
	Weight      int64 `json:"weight"`
}
