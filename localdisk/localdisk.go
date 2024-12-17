// SPDX-License-Identifier: 0BSD

package localdisk

// #cgo pkg-config: libstoragemgmt
// #include <stdio.h>
// #include <libstoragemgmt/libstoragemgmt.h>
// #include <stdlib.h>
// #include <string.h>
import "C"
import (
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

// Vpd83Seach searches local disks for vpd
func Vpd83Seach(vpd string) ([]string, error) {

	cs := C.CString(vpd)
	defer C.free(unsafe.Pointer(cs))

	var deviceList []string

	var slist *C.lsm_string_list
	var lsmError *C.lsm_error

	var err = C.lsm_local_disk_vpd83_search(cs, &slist, &lsmError)

	if err == 0 {
		deviceList = getStrings(slist, true)
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
