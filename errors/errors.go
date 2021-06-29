// SPDX-License-Identifier: 0BSD

package errors

import "fmt"

// LsmError returned from JSON API
type LsmError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *LsmError) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("code = %d, message = %s, data = %s", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("code = %d, message = %s", e.Code, e.Message)
}

const (

	// Ok ... No errors encountered, in this case should likely never be seen
	Ok int32 = 0

	// LibBug ... Library bug
	LibBug int32 = 1

	// PluginBug ... Bug found in plugin
	PluginBug int32 = 2

	// JobStarted ... Job has been started
	JobStarted int32 = 7

	// TimeOut ... Plugin timeout
	TimeOut int32 = 11

	// DameonNotRunning ... lsmd does not appear to be running
	DameonNotRunning int32 = 12

	// PermissionDenied Insufficient permission
	PermissionDenied int32 = 13

	// NameConflict Request has a duplicate named resource
	NameConflict int32 = 50

	// ExistsInitiator ... Initiator already exists in group
	ExistsInitiator int32 = 52

	// InvalidArgument ... provided argument is incorrect
	InvalidArgument int32 = 101

	// NoStateChange ... Request resulted in no change of storage state
	NoStateChange int32 = 125

	// NetworkConnRefused ... Network connection was refused
	NetworkConnRefused int32 = 140

	// NetworkHostDown ... Networked host is not reachable on network
	NetworkHostDown int32 = 141

	// NetworkError ... Generic network error
	NetworkError int32 = 142

	// NoMemory ... Library encountered an out of memory condition
	NoMemory int32 = 152

	// NoSupport operation not supported
	NoSupport int32 = 153

	// IsMasked ... Volume is masked to an access group
	IsMasked int32 = 160

	// HasChildDependency ... Volume/File system has a child dependency
	HasChildDependency int32 = 161

	// NotFoundAccessGroup ... The specified access group was not found
	NotFoundAccessGroup int32 = 200

	// NotFoundFs ... The specified file system was not found
	NotFoundFs int32 = 201

	// NotFoundJob ... The specified job was not found
	NotFoundJob int32 = 202

	// NotFoundPool ... The specified pool was not found
	NotFoundPool int32 = 203

	// NotFoundFsSs ... The specfified file system/snap shot was not found
	NotFoundFsSS int32 = 204

	// NotFoundVolume ... The specified volume was not found
	NotFoundVolume int32 = 205

	// NotFoundNfsExport ... The specified NFS export was not found
	NotFoundNfsExport int32 = 206

	// NotFoundSystem ... The specified system was not found
	NotFoundSystem int32 = 208

	// NotFoundDisk ... The specified disk was not found
	NotFoundDisk int32 = 209

	// NotLicensed ... The required functionality is not licensed
	NotLicensed int32 = 226

	// NoSupportOnlineChange ... The specified operation requires offline
	NoSupportOnlineChange int32 = 250

	// NoSupportOfflineChange ... The specified operation requires online
	NoSupportOfflineChange int32 = 251

	// PluginAuthFailed Plugin failed to authenticate
	PluginAuthFailed int32 = 300

	// PluginSocketPermission ... Incorrect permission on UNIX domain socket used for IPC
	PluginSocketPermission int32 = 307

	// PluginNotExist ... Plugin doesn't apprear to exist
	PluginNotExist int32 = 311

	// NotEnoughSpace ... Insufficient space to complete the request
	NotEnoughSpace int32 = 350

	//TransPortComunication ... Issue reading/writing to plugin
	TransPortComunication int32 = 400

	// TransPortSerialization ... Issue with serializing the payload of a request
	TransPortSerialization int32 = 401

	// TransPortInvalidArg parameter transported over IPC is invalid
	TransPortInvalidArg int32 = 402

	// LastInitInAccessGroup ... refuse to remove the last initiator from access group
	LastInitInAccessGroup int32 = 502

	// UnsupportedSearchKey ... The specified search key is not supported
	UnsupportedSearchKey int32 = 510

	// EmptyAccessGroup ... volume_mask() will fail if access group has no member/initiator
	EmptyAccessGroup int32 = 511

	// PoolNotReady ... Pool is not ready for create/resize/etc
	PoolNotReady int32 = 512

	// DiskNotFree ... Disk is not in DiskStatusFree status
	DiskNotFree int32 = 513
)
