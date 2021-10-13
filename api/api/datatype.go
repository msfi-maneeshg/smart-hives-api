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
	Metadata    DeviceMetadata `json:"metadata"`
}

type DeviceMetadata struct {
	MaximumTemperature *int64 `json:"maximumTemperature,omitempty"`
	MinimumTemperature *int64 `json:"minimumTemperature,omitempty"`
	MinimumHumidity    *int64 `json:"minimumHumidity,omitempty"`
	MaximumHumidity    *int64 `json:"maximumHumidity,omitempty"`
	MinimumWeight      *int64 `json:"minimumWeight,omitempty"`
	MaximumWeight      *int64 `json:"maximumWeight,omitempty"`
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

//NewDeviceInfo:
type NewDeviceInfo struct {
	DeviceId   string          `json:"deviceId,omitempty"`
	DeviceInfo DeviceOtherInfo `json:"deviceInfo,omitempty"`
	Metadata   DeviceMetadata  `json:"metadata,omitempty"`
}

//DeviceOtherInfo:
type DeviceOtherInfo struct {
	SserialNumber       string `json:"serialNumber,omitempty"`
	Description         string `json:"description,omitempty"`
	DescriptiveLocation string `json:"descriptiveLocation,omitempty"`
	DeviceClass         string `json:"deviceClass,omitempty"`
	FirmwareVersion     string `json:"firmwareVersion,omitempty"`
	HardwareVersion     string `json:"hardwareVersion,omitempty"`
	Manufacturer        string `json:"manufacturer,omitempty"`
	Model               string `json:"model,omitempty"`
}

//CreateInterface:
type CreateInterface struct {
	ID                   string            `json:"id,omitempty"`
	Name                 string            `json:"name,omitempty"`
	SchemaId             string            `json:"schemaId,omitempty"`
	EventId              string            `json:"eventId,omitempty"`
	EventTypeId          string            `json:"eventTypeId,omitempty"`
	Alias                string            `json:"alias,omitempty"`
	LogicalInterfaceId   string            `json:"logicalInterfaceId,omitempty"`
	NotificationStrategy string            `json:"notificationStrategy,omitempty"`
	PropertyMappings     *PropertyMappings `json:"propertyMappings,omitempty"`
}

//PropertyMappings:
type PropertyMappings struct {
	HiveEvent HiveEvent `json:"HiveEvent,omitempty"`
}

//HiveEvent:
type HiveEvent struct {
	Temperature string `json:"temperature,omitempty"`
	Humidity    string `json:"humidity,omitempty"`
	Weight      string `json:"weight,omitempty"`
}

//OutputInterfaceInfo:
type OutputInterfaceInfo struct {
	ID string `json:"id,omitempty"`
}

//ActivateInterface:
type ActivateInterface struct {
	Operation string `json:"operation,omitempty"`
}

//NotificationRules:
type NotificationRules struct {
	Name                 string               `json:"name,omitempty"`
	Condition            string               `json:"condition,omitempty"`
	NotificationStrategy NotificationStrategy `json:"notificationStrategy,omitempty"`
}

//NotificationStrategy:
type NotificationStrategy struct {
	When       string `json:"when,omitempty"`
	Count      int    `json:"count,omitempty"`
	TimePeriod int    `json:"timePeriod,omitempty"`
}

type ActionTrigger struct {
	Name             string                  `json:"name,omitempty"`
	Description      string                  `json:"description,omitempty"`
	Type             string                  `json:"type,omitempty"`
	Enabled          string                  `json:"enabled,omitempty"`
	Configuration    TriggerConfiguration    `json:"configuration,omitempty"`
	VariableMappings TriggerVariableMappings `json:"variableMappings,omitempty"`
}

type TriggerConfiguration struct {
	LogicalInterfaceId string `json:"logicalInterfaceId,omitempty"`
	RuleId             string `json:"ruleId,omitempty"`
	Type               string `json:"type,omitempty"`
	TypeId             string `json:"typeId,omitempty"`
	InstanceId         string `json:"instanceId,omitempty"`
}

type TriggerVariableMappings struct {
	DeviceType  string `json:"deviceType,omitempty"`
	Temperature string `json:"temperature,omitempty"`
	Humidity    string `json:"humidity,omitempty"`
	Weight      string `json:"weight,omitempty"`
	InterfaceId string `json:"InterfaceId,omitempty"`
	DeviceId    string `json:"deviceId,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}
