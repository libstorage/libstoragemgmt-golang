// SPDX-License-Identifier: 0BSD

package localdisk

// #cgo pkg-config: libstoragemgmt
// #include <stdio.h>
// #include <libstoragemgmt/libstoragemgmt.h>
// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"unsafe"

	lsm "github.com/libstorage/libstoragemgmt-golang"
	"github.com/libstorage/libstoragemgmt-golang/errors"
)

func processError(errorNum int, e *C.lsm_error) error {
	if e != nil {
		// Make sure we only free e if e is not nil
		defer C.lsm_error_free(e)
		return &errors.LsmError{
			Code:    int32(C.lsm_error_number_get(e)),
			Message: C.GoString(C.lsm_error_message_get(e))}
	}
	if errorNum != 0 {
		return &errors.LsmError{
			Code: int32(errorNum)}
	}
	return nil
}

func getStrings(lsmStrings *C.lsm_string_list, free bool) []string {
	var rc []string

	var num = C.lsm_string_list_size(lsmStrings)

	var i C.uint
	for i = 0; i < num; i++ {
		var item = C.GoString(C.lsm_string_list_elem_get(lsmStrings, i))
		rc = append(rc, item)
	}

	if free {
		C.lsm_string_list_free(lsmStrings)
	}
	return rc
}

// List returns local disk path(s)
func List() ([]string, error) {
	var disks []string

	var diskPaths *C.lsm_string_list
	var lsmError *C.lsm_error

	var e = C.lsm_local_disk_list(&diskPaths, &lsmError)
	if e == 0 {
		disks = getStrings(diskPaths, true)
	} else {
		return disks, processError(int(e), lsmError)
	}
	return disks, nil
}

// Vpd83Search searches local disks for vpd
func Vpd83Search(vpd string) ([]string, error) {

	cs := C.CString(vpd)
	defer C.free(unsafe.Pointer(cs))

	var deviceList []string

	var str_list *C.lsm_string_list
	var lsmError *C.lsm_error

	var err = C.lsm_local_disk_vpd83_search(cs, &str_list, &lsmError)

	if err == 0 {
		deviceList = getStrings(str_list, true)
	} else {
		return deviceList, processError(int(err), lsmError)
	}

	return deviceList, nil
}

// SerialNumGet retrieves the serial number for the local
// disk with the specified path
func SerialNumGet(diskPath string) (string, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var sn *C.char
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_serial_num_get(dp, &sn, &lsmError)
	if rc == 0 {
		var serialNum = C.GoString(sn)
		C.free(unsafe.Pointer(sn))
		return serialNum, nil
	}
	return "", processError(int(rc), lsmError)
}

// Vpd83Get retrieves vpd83 for the specified local disk path
func Vpd83Get(diskPath string) (string, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var vpd *C.char
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_vpd83_get(dp, &vpd, &lsmError)
	if rc == 0 {
		var vpdNum = C.GoString(vpd)
		C.free(unsafe.Pointer(vpd))
		return vpdNum, nil
	}
	return "", processError(int(rc), lsmError)
}

// HealthStatusGet retrieves health status for the specified local disk path
func HealthStatusGet(diskPath string) (lsm.DiskHealthStatus, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var healthStatus C.int32_t
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_health_status_get(dp, &healthStatus, &lsmError)
	if rc == 0 {
		return lsm.DiskHealthStatus(healthStatus), nil
	}
	return -1, processError(int(rc), lsmError)
}

// RpmGet retrieves health RPM for the specified local disk path
func RpmGet(diskPath string) (int32, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var rpm C.int32_t
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_rpm_get(dp, &rpm, &lsmError)
	if rc == 0 {
		return int32(rpm), nil
	}
	return -1, processError(int(rc), lsmError)
}

// LinkTypeGet retrieves link type for the specified local disk path
func LinkTypeGet(diskPath string) (lsm.DiskLinkType, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var linkType C.lsm_disk_link_type
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_link_type_get(dp, &linkType, &lsmError)
	if rc == 0 {
		return lsm.DiskLinkType(linkType), nil
	}
	return -1, processError(int(rc), lsmError)
}

// IndentLedOff turns off the identification LED for the specified disk
func IndentLedOff(diskPath string) error {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var lsmError *C.lsm_error
	var rc = C.lsm_local_disk_ident_led_off(dp, &lsmError)
	return processError(int(rc), lsmError)
}

// IndentLedOn turns on the identification LED for the specified disk
func IndentLedOn(diskPath string) error {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var lsmError *C.lsm_error
	var rc = C.lsm_local_disk_ident_led_on(dp, &lsmError)
	return processError(int(rc), lsmError)
}

// FaultLedOn turns on the fault LED for the specified disk
func FaultLedOn(diskPath string) error {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var lsmError *C.lsm_error
	var rc = C.lsm_local_disk_fault_led_on(dp, &lsmError)
	return processError(int(rc), lsmError)
}

// FaultLedOff turns on the fault LED for the specified disk
func FaultLedOff(diskPath string) error {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var lsmError *C.lsm_error
	var rc = C.lsm_local_disk_fault_led_off(dp, &lsmError)
	return processError(int(rc), lsmError)
}

// LedStatusGet retrieves status of LEDs for specified local disk path
func LedStatusGet(diskPath string) (lsm.DiskLedStatusBitField, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var ledStatus C.uint32_t
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_led_status_get(dp, &ledStatus, &lsmError)
	if rc == 0 {
		return lsm.DiskLedStatusBitField(ledStatus), nil
	}
	return 1, processError(int(rc), lsmError)
}

// LinkSpeedGet retrieves link speed for specified local disk path
func LinkSpeedGet(diskPath string) (uint32, error) {
	dp := C.CString(diskPath)
	defer C.free(unsafe.Pointer(dp))

	var linkSpeed C.uint32_t
	var lsmError *C.lsm_error

	var rc = C.lsm_local_disk_link_speed_get(dp, &linkSpeed, &lsmError)
	if rc == 0 {
		return uint32(linkSpeed), nil
	}
	return 0, processError(int(rc), lsmError)
}

// Opaque type for LED Slots API
type LedSlotsHandle struct {
	handle *C.lsm_led_handle
}

// Retrieve the handle to use for interacting with LED Slots, make sure to call LedSlotsHandleFree
// when done
func LedSlotsHandleGet() (*LedSlotsHandle, error) {
	var l_handle *C.lsm_led_handle
	var rc = C.lsm_led_handle_get(&l_handle, 0)

	if rc == 0 {
		return &LedSlotsHandle{handle: l_handle}, nil
	}

	return nil, &errors.LsmError{
		Code:    int32(rc),
		Message: fmt.Sprintf("Unexpected error: code = [%d]", rc)}

}

// Frees the resources used by the LED Slot API, calling this is required to prevent a memory leak
func LedSlotsHandleFree(led_slots *LedSlotsHandle) {
	C.lsm_led_handle_free(led_slots.handle)
}

// Information about a specific LED slot
type LedSlot struct {
	SlotId string // The slot identifier
	Device string // The slot device node, if it has one
}

// Retrieves all the LED slots
func (l *LedSlotsHandle) SlotsGet() ([]LedSlot, error) {
	var slots []LedSlot
	var itr *C.lsm_led_slot_itr
	var lsmError *C.lsm_error

	var rc = C.lsm_led_slot_iterator_get(l.handle, &itr, &lsmError, 0)
	if int32(rc) == errors.Ok {
		for {
			var slot = C.lsm_led_slot_next(l.handle, itr)
			if slot != nil {
				var id = C.lsm_led_slot_id(slot)
				var device = C.lsm_led_slot_device(slot)

				var slot_id = C.GoString(id)
				var slot_device string

				// The device node can be null as not every device may have a device node
				if device != nil {
					slot_device = C.GoString(device)
				}

				slots = append(slots, LedSlot{SlotId: slot_id, Device: slot_device})

			} else {
				break
			}
		}

		// Free the slot iterator
		C.lsm_led_slot_iterator_free(l.handle, itr)
	}
	return slots, processError(int(rc), lsmError)
}

// Retrieves the current status of the LED slot
func (l *LedSlotsHandle) StatusGet(slot *LedSlot) (lsm.DiskLedStatusBitField, error) {
	var itr *C.lsm_led_slot_itr
	var lsmError *C.lsm_error

	var rc = C.lsm_led_slot_iterator_get(l.handle, &itr, &lsmError, 0)
	if int32(rc) == errors.Ok {
		defer C.lsm_led_slot_iterator_free(l.handle, itr)
		for {
			var c_slot_handle = C.lsm_led_slot_next(l.handle, itr)
			if c_slot_handle != nil {
				var id = C.lsm_led_slot_id(c_slot_handle)
				var slot_id = C.GoString(id)

				if slot.SlotId == slot_id {
					var ledStatus = C.lsm_led_slot_status_get(c_slot_handle)
					return lsm.DiskLedStatusBitField(ledStatus), nil
				}

			} else {
				break
			}
		}
		return 0, &errors.LsmError{
			Code:    errors.NotFoundGeneric,
			Message: fmt.Sprintf("Slot with id = %v not found!", slot.SlotId)}
	}
	return 0, processError(int(rc), lsmError)
}

// Sets the LED slot
func (l *LedSlotsHandle) StatusSet(slot *LedSlot, led_status lsm.DiskLedStatusBitField) error {
	var itr *C.lsm_led_slot_itr
	var lsmError *C.lsm_error

	var rc = C.lsm_led_slot_iterator_get(l.handle, &itr, &lsmError, 0)
	if int32(rc) == errors.Ok {
		defer C.lsm_led_slot_iterator_free(l.handle, itr)
		for {
			var c_slot_handle = C.lsm_led_slot_next(l.handle, itr)
			if c_slot_handle != nil {
				var id = C.lsm_led_slot_id(c_slot_handle)
				var slot_id = C.GoString(id)

				if slot.SlotId == slot_id {
					var status = C.lsm_led_slot_status_set(l.handle, c_slot_handle, C.uint32_t(led_status), &lsmError, 0)
					if int32(status) == errors.Ok {
						return nil
					}
					return processError(int(status), lsmError)
				}

			} else {
				break
			}
		}
		return &errors.LsmError{
			Code:    errors.NotFoundGeneric,
			Message: fmt.Sprintf("Slot with id = %v not found!", slot.SlotId)}
	}
	return processError(int(rc), lsmError)
}
