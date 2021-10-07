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

type NewDeviceType struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	ClassId     string         `json:"classId"`
	metadata    DeviceMetadata `json:"metadata"`
}

type DeviceMetadata struct {
	MaxTemperature int64 `json:"maxTemperature"`
	MinTemperature int64 `json:"minTemperature"`
	MinWeight      int64 `json:"minWeight"`
	MaxHumidity    int64 `json:"maxHumidity"`
	MaxWeight      int64 `json:"maxWeight"`
	MinHumidity    int64 `json:"minHumidity"`
}

type CreateDestination struct {
	Type          string                   `json:"type"`
	Name          string                   `json:"name"`
	Configuration DestinationConfiguration `json:"configuration"`
}

type DestinationConfiguration struct {
	BucketInterval string `json:"bucketInterval"`
}

type CreateForwardingRule struct {
	Name            string                 `json:"name"`
	DestinationName string                 `json:"destinationName"`
	Type            string                 `json:"type"`
	Selector        ForwardingRuleSelector `json:"selector"`
}
type ForwardingRuleSelector struct {
	DeviceType string `json:"deviceType"`
	EventId    string `json:"eventId"`
}

type CreateNewDevice struct {
	DeviceId string         `json:"deviceId"`
	Metadata DeviceMetadata `json:"metadata"`
}
