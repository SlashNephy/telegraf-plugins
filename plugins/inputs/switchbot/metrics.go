package switchbot

import "github.com/nasa9084/go-switchbot/v5"

type MetricSource struct {
	Key   string
	Value func(status *switchbot.DeviceStatus) any
}

var (
	AmbientBrightness = &MetricSource{
		Key: "ambient_brightness",
		Value: func(status *switchbot.DeviceStatus) any {
			value, _ := status.Brightness.AmbientBrightness()
			return value
		},
	}
	Battery = &MetricSource{
		Key: "battery",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.Battery
		},
	}
	Brightness = &MetricSource{
		Key: "brightness",
		Value: func(status *switchbot.DeviceStatus) any {
			value, _ := status.Brightness.Int()
			return value
		},
	}
	CO2 = &MetricSource{
		Key: "co2",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.CO2
		},
	}
	ColorTemperature = &MetricSource{
		Key: "color_temperature",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.ColorTemperature
		},
	}
	ElectricCurrent = &MetricSource{
		Key: "electric_current",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.ElectricCurrent
		},
	}
	ElectricityOfDay = &MetricSource{
		Key: "electricity_of_day",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.ElectricityOfDay
		},
	}
	FanSpeed = &MetricSource{
		Key: "fan_speed",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.FanSpeed
		},
	}
	Humidity = &MetricSource{
		Key: "humidity",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.Humidity
		},
	}
	IsAuto = &MetricSource{
		Key: "is_auto",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsAuto
		},
	}
	IsCalibrated = &MetricSource{
		Key: "is_calibrated",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsCalibrated
		},
	}
	IsChildLock = &MetricSource{
		Key: "is_child_lock",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsChildLock
		},
	}
	IsGrouped = &MetricSource{
		Key: "is_grouped",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsGrouped
		},
	}
	IsLackWater = &MetricSource{
		Key: "is_lack_water",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsLackWater
		},
	}
	IsMoveDetected = &MetricSource{
		Key: "is_move_detected",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsMoveDetected
		},
	}
	IsMoving = &MetricSource{
		Key: "is_moving",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsMoving
		},
	}
	IsSound = &MetricSource{
		Key: "is_sound",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.IsSound
		},
	}
	LightLevel = &MetricSource{
		Key: "light_level",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.LightLevel
		},
	}
	NebulizationEfficiency = &MetricSource{
		Key: "nebulization_efficiency",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.NebulizationEfficiency
		},
	}
	OnlineStatus = &MetricSource{
		Key: "online_status",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.OnlineStatus
		},
	}
	SlidePosition = &MetricSource{
		Key: "slide_position",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.SlidePosition
		},
	}
	Temperature = &MetricSource{
		Key: "temperature",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.Temperature
		},
	}
	Voltage = &MetricSource{
		Key: "voltage",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.Voltage
		},
	}
	Weight = &MetricSource{
		Key: "weight",
		Value: func(status *switchbot.DeviceStatus) any {
			return status.Weight
		},
	}
)

var SupportedMetrics = map[switchbot.PhysicalDeviceType][]*MetricSource{
	// https://github.com/OpenWonderLabs/SwitchBotAPI/blob/main/README.md#responses-1
	switchbot.Bot:                      {Battery},
	switchbot.Curtain:                  {IsCalibrated, IsGrouped, IsMoving, Battery, SlidePosition},
	"Curtain3":                         {IsCalibrated, IsGrouped, IsMoving, Battery, SlidePosition},
	switchbot.Meter:                    {Temperature, Battery, Humidity},
	switchbot.MeterPlus:                {Battery, Temperature, Humidity},
	"MeterPro(CO2)":                    {Battery, Temperature, Humidity, CO2},
	switchbot.WoIOSensor:               {Battery, Temperature, Humidity},
	switchbot.Lock:                     {Battery /* lockState, doorState */, IsCalibrated},
	"Smart Lock Pro":                   {Battery /* lockState, doorState */, IsCalibrated},
	switchbot.KeyPad:                   {},
	switchbot.KeyPadTouch:              {},
	switchbot.MotionSensor:             {Battery, IsMoveDetected, AmbientBrightness},
	switchbot.ContactSensor:            {Battery, IsMoveDetected /* openState */, AmbientBrightness},
	"Water Detector":                   {Battery /* status */},
	switchbot.CeilingLight:             {Brightness, ColorTemperature},
	switchbot.CeilingLightPro:          {Brightness, ColorTemperature},
	switchbot.PlugMiniUS:               {Voltage, Weight, ElectricityOfDay, ElectricCurrent},
	switchbot.PlugMiniJP:               {Voltage, Weight, ElectricityOfDay, ElectricCurrent},
	switchbot.Plug:                     {},
	switchbot.StripLight:               {Brightness},
	switchbot.ColorBulb:                {Brightness, ColorTemperature},
	switchbot.RobotVacuumCleanerS1:     { /* workingStatus */ OnlineStatus, Battery},
	switchbot.RobotVacuumCleanerS1Plus: { /* workingStatus */ OnlineStatus, Battery},
	"K10+":                             { /* workingStatus */ OnlineStatus, Battery},
	"K10+ Pro":                         { /* workingStatus */ OnlineStatus, Battery},
	"Robot Vacuum Cleaner S10":         { /* workingStatus */ OnlineStatus, Battery /* waterBaseBatterym, taskType */},
	switchbot.Humidifier:               {Humidity, Temperature, NebulizationEfficiency, IsAuto, IsChildLock, IsSound, IsLackWater},
	switchbot.BlindTilt:                {IsCalibrated, IsGrouped, IsMoving, SlidePosition},
	switchbot.Hub2:                     {Temperature, LightLevel, Humidity},
	"Battery Circulator Fan":           {Battery /* nightStatus */, FanSpeed},
}
