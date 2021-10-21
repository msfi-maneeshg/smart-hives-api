package common

import "time"

const MY_KEY = "smarthives"
const REFERESH_KEY = "referesh-smarthives"
const EXPIRE_TIME time.Duration = 10 // in minutes

const (
	IOT_USERNAME  = "a-8l173e-otjztnyacu"
	IOT_PASSWORD  = "ChLq7u0pO+*hl7JER_"
	IOT_URL       = "https://" + IOT_USERNAME + ":" + IOT_PASSWORD + "@8l173e.internetofthings.ibmcloud.com/api/v0002/"
	EVENT_TYPE_ID = "615d3165cf7abe0fa1cabe73"
	SCHEMA_ID     = "615d3164cf7abe0fa1cabe72"
	ACTION_ID     = "615d33202086e476fbb9b550"
)

const (
	PROFILES = "profiles"
)
