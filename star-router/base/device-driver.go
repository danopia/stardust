package base

// Manufactors virtual devices which can exist in more than one place at once
type DeviceDriver interface {
  Open(params map[string]interface{}) *Device
}
