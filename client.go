// SPDX-License-Identifier: 0BSD

package libstoragemgmt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	errors "github.com/libstorage/libstoragemgmt-golang/errors"
)

// ClientConnection is the structure that encomposes the needed data for the plugin connection.
type ClientConnection struct {
	tp         transPort
	PluginName string
	timeout    uint32
}

// Client establishes a connection to a plugin as specified in the URI.
func Client(uri string, password string, timeout uint32) (*ClientConnection, error) {

	p, parseError := url.Parse(uri)
	if parseError != nil {
		return nil, &errors.LsmError{
			Code:    errors.InvalidArgument,
			Message: fmt.Sprintf("invalid uri: %s", parseError)}
	}

	pluginName := p.Scheme
	pluginIpcPath := getPluginIpcPath(pluginName)

	transport, transPortError := newTransport(pluginIpcPath, true)
	if transPortError != nil {
		return nil, transPortError
	}

	args := map[string]interface{}{"password": password, "uri": uri, "timeout": timeout}
	if libError := transport.invoke("plugin_register", args, nil); libError != nil {
		return nil, libError
	}

	return &ClientConnection{tp: *transport, PluginName: pluginName, timeout: timeout}, nil
}

// PluginInfo information about the current plugin
func (c *ClientConnection) PluginInfo() (*PluginInfo, error) {
	args := make(map[string]interface{})
	var info []string
	if invokeError := c.tp.invoke("plugin_info", args, &info); invokeError != nil {
		return nil, invokeError
	}
	return &PluginInfo{Description: info[0], Version: info[1], Name: c.PluginName}, nil
}

// AvailablePlugins retrieves all the available plugins.
func AvailablePlugins() ([]PluginInfo, error) {
	var udsPath = udsPath()

	if _, err := os.Stat(udsPath); os.IsNotExist(err) {
		return make([]PluginInfo, 0), &errors.LsmError{
			Code:    errors.DameonNotRunning,
			Message: fmt.Sprintf("LibStorageMgmt daemon is not running for socket folder %s", udsPath),
			Data:    ""}
	}

	var pluginInfos []PluginInfo
	for _, pluginPath := range getPlugins(udsPath) {

		var trans, transError = newTransport(pluginPath, true)
		if transError != nil {
			return nil, transError
		}

		args := make(map[string]interface{})
		var info []string
		invokeError := trans.invoke("plugin_info", args, &info)

		trans.close()

		if invokeError != nil {
			return pluginInfos, invokeError
		}
		pluginInfos = append(pluginInfos, PluginInfo{
			Description: info[0],
			Version:     info[1],
			Name:        pluginPath})
	}

	return pluginInfos, nil
}

// Close instructs the plugin to shutdown and exist.
func (c *ClientConnection) Close() error {
	args := make(map[string]interface{})
	ourError := c.tp.invoke("plugin_unregister", args, nil)
	c.tp.close()
	return ourError
}

// Systems returns systems information
func (c *ClientConnection) Systems() ([]System, error) {
	args := make(map[string]interface{})
	var systems []System
	return systems, c.tp.invoke("systems", args, &systems)
}

// Volumes returns block device information
func (c *ClientConnection) Volumes(search ...string) ([]Volume, error) {
	args := make(map[string]interface{})
	var volumes []Volume

	if !handleSearch(args, search) {
		return make([]Volume, 0), &errors.LsmError{
			Code: errors.InvalidArgument,
			Message: fmt.Sprintf(
				"volume supports 0 or 2 search parameters (key, value), provide %d", len(search)),
			Data: ""}
	}

	return volumes, c.tp.invoke("volumes", args, &volumes)
}

// Pools returns the units of storage that block devices and FS
// can be created from.
func (c *ClientConnection) Pools(search ...string) ([]Pool, error) {
	args := make(map[string]interface{})

	if !handleSearch(args, search) {
		return make([]Pool, 0), &errors.LsmError{
			Code: errors.InvalidArgument,
			Message: fmt.Sprintf(
				"pool supports 0 or 2 search parameters (key, value), provide %d", len(search)),
			Data: ""}
	}

	var pools []Pool
	return pools, c.tp.invoke("pools", args, &pools)
}

// Disks returns disks that are present.
func (c *ClientConnection) Disks() ([]Disk, error) {
	args := make(map[string]interface{})
	var disks []Disk
	return disks, c.tp.invoke("disks", args, &disks)
}

// FileSystems returns pools that are present.
func (c *ClientConnection) FileSystems(search ...string) ([]FileSystem, error) {
	args := make(map[string]interface{})
	var fileSystems []FileSystem

	if !handleSearch(args, search) {
		return make([]FileSystem, 0), &errors.LsmError{
			Code: errors.InvalidArgument,
			Message: fmt.Sprintf(
				"fs supports 0 or 2 search parameters (key, value), provide %d", len(search)),
			Data: ""}
	}

	return fileSystems, c.tp.invoke("fs", args, &fileSystems)
}

// NfsExports returns nfs exports  that are present.
func (c *ClientConnection) NfsExports(search ...string) ([]NfsExport, error) {
	args := make(map[string]interface{})

	if !handleSearch(args, search) {
		return make([]NfsExport, 0), &errors.LsmError{
			Code: errors.InvalidArgument,
			Message: fmt.Sprintf(
				"NfsExports supports 0 or 2 search parameters (key, value), provide %d", len(search)),
			Data: ""}
	}

	var nfsExports []NfsExport
	return nfsExports, c.tp.invoke("exports", args, &nfsExports)
}

// NfsExportAuthTypes returns list of support authentication types
func (c *ClientConnection) NfsExportAuthTypes() ([]string, error) {
	var authTypes []string
	return authTypes, c.tp.invoke("export_auth", make(map[string]interface{}), &authTypes)
}

// FsExport creates or modifies a NFS export.
func (c *ClientConnection) FsExport(fs *FileSystem, exportPath *string,
	access *NfsAccess, authType *string, options *string) (*NfsExport, error) {

	if len(access.Ro) == 0 && len(access.Rw) == 0 {
		return nil, &errors.LsmError{
			Code:    errors.InvalidArgument,
			Message: "at least 1 host should exist in access.ro or access.rw",
			Data:    ""}
	}

	for _, i := range access.Root {
		if !contains(access.Rw, i) && !contains(access.Ro, i) {
			return nil, &errors.LsmError{
				Code: errors.InvalidArgument,
				Message: fmt.Sprintf(
					"host '%s' contained in access.root should also be in access.rw or access.ro", i),
				Data: ""}
		}
	}

	for _, i := range access.Rw {
		if contains(access.Ro, i) {
			return nil, &errors.LsmError{
				Code:    errors.InvalidArgument,
				Message: fmt.Sprintf("host '%s' in both access.ro and access.rw", i),
				Data:    ""}
		}
	}

	args := map[string]interface{}{
		"fs_id":       fs.ID,
		"export_path": exportPath,
		"root_list":   emptySliceIfNil(access.Ro),
		"rw_list":     emptySliceIfNil(access.Rw),
		"ro_list":     emptySliceIfNil(access.Ro),
		"anon_uid":    access.AnonUID,
		"anon_gid":    access.AnonGID,
		"auth_type":   authType,
		"options":     options,
	}
	var nfsExport NfsExport
	if err := c.tp.invoke("export_fs", args, &nfsExport); err != nil {
		return nil, err
	}
	return &nfsExport, nil
}

// FsUnExport removes a file system export.
func (c *ClientConnection) FsUnExport(export *NfsExport) error {
	args := map[string]interface{}{"export": *export}
	return c.tp.invoke("export_remove", args, nil)
}

// AccessGroups returns access groups  that are present.
// TODO: Add search arguments
func (c *ClientConnection) AccessGroups() ([]AccessGroup, error) {
	args := make(map[string]interface{})
	var accessGroups []AccessGroup
	return accessGroups, c.tp.invoke("access_groups", args, &accessGroups)
}

// TargetPorts returns target ports that are present.
func (c *ClientConnection) TargetPorts() ([]TargetPort, error) {
	args := make(map[string]interface{})
	var targetPorts []TargetPort
	return targetPorts, c.tp.invoke("target_ports", args, &targetPorts)
}

// Batteries returns batteries that are present
func (c *ClientConnection) Batteries() ([]Battery, error) {
	args := make(map[string]interface{})
	var batteries []Battery
	return batteries, c.tp.invoke("batteries", args, &batteries)
}

// JobFree instructs the plugin to release resources for the job that was returned.
func (c *ClientConnection) JobFree(jobID string) error {
	args := map[string]interface{}{"job_id": jobID}
	return c.tp.invoke("job_free", args, nil)
}

// JobStatus instructs the plugin to return the status of the specified job.  The returned values are
// the current job status, percent complete, and any errors that occured.  Always check error first as if it's
// set the other two are meaningless.  If checking on the status of an operation that doesn't return a result
// or you are not wanting the result, pass nil.
func (c *ClientConnection) JobStatus(jobID string, returnedResult interface{}) (JobStatusType, uint8, error) {
	args := map[string]interface{}{"job_id": jobID}

	var result [3]json.RawMessage
	if jobError := c.tp.invoke("job_status", args, &result); jobError != nil {
		return JobStatusError, 0, jobError
	}

	var status JobStatusType
	if statusMe := json.Unmarshal(result[0], &status); statusMe != nil {
		return JobStatusError, 0, statusMe
	}

	switch status {
	case JobStatusInprogress:
		var percent uint8
		if percentError := json.Unmarshal(result[1], &percent); percentError != nil {
			return JobStatusError, 0, percentError
		}

		return status, percent, nil
	case JobStatusComplete:
		// Some RPC calls with jobs do not return a value, thus the third item is
		// "null"
		if string(result[2]) != "null" && returnedResult != nil {
			return status, 100, json.Unmarshal(result[2], returnedResult)
		}
		return status, 100, nil
	case JobStatusError:
		// Error
		var error errors.LsmError
		if checkErrorE := json.Unmarshal(result[2], &error); checkErrorE != nil {
			return JobStatusError, 0, checkErrorE
		}

		return JobStatusError, 0, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: "job_status returned error status with no error information"}
	default:
		return JobStatusError, 0, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Invalid status type returned %v", status)}
	}
}

func (c *ClientConnection) getJobOrResult(err error, returned [2]json.RawMessage, sync bool, result interface{}) (*string, error) {
	if err != nil {
		return nil, err
	}

	var job string
	var um = json.Unmarshal(returned[0], &job)
	if um == nil {
		// We have a job, but want to wait for result, so do so.
		if sync {
			return nil, c.JobWait(job, result)
		}

		return &job, nil
	}
	// We have the result
	return nil, json.Unmarshal(returned[1], result)
}

func (c *ClientConnection) getJobOrNone(err error, returned json.RawMessage, sync bool) (*string, error) {
	if err != nil {
		return nil, err
	}

	var job string
	var um = json.Unmarshal(returned, &job)
	if um == nil {
		// We have a job, but want to wait for result, so do so.
		if sync {
			return nil, c.JobWait(job, nil)
		}

		return &job, nil
	}
	return nil, um
}

// JobWait waits for the job to finish and retrieves the end result in "returnedResult".
func (c *ClientConnection) JobWait(jobID string, returnedResult interface{}) error {

	for true {
		var status, _, err = c.JobStatus(jobID, returnedResult)
		if err != nil {
			return err
		}

		if status == JobStatusInprogress {
			time.Sleep(time.Millisecond * 250)
			continue
		} else if status == JobStatusComplete {
			if freeError := c.JobFree(jobID); freeError != nil {
				return &errors.LsmError{
					Code: errors.PluginBug,
					Message: fmt.Sprintf(
						"We successfully waited for job %s, but got an error freeing it: %s", jobID, freeError)}
			}
			break
		}
	}
	return nil
}

// Capabilities retrieve capabilities
func (c *ClientConnection) Capabilities(system *System) (*Capabilities, error) {
	args := map[string]interface{}{"system": *system}
	var cap Capabilities
	return &cap, c.tp.invoke("capabilities", args, &cap)
}

// TimeOutSet sets the connection timeout with the storage device.
func (c *ClientConnection) TimeOutSet(milliSeconds uint32) error {
	args := map[string]interface{}{"ms": milliSeconds}
	var err = c.tp.invoke("time_out_set", args, nil)
	if err == nil {
		c.timeout = milliSeconds
	}
	return err
}

// TimeOutGet sets the connection timeout with the storage device.
func (c *ClientConnection) TimeOutGet() uint32 {
	return c.timeout
}

// SysReadCachePctSet changes the read cache percentage for the specified system.
func (c *ClientConnection) SysReadCachePctSet(system *System, readPercent uint32) error {

	if readPercent > 100 {
		return &errors.LsmError{
			Code: errors.InvalidArgument,
			Message: fmt.Sprintf(
				"Invalid readPercent %d, valid range 0-100", readPercent)}
	}

	args := map[string]interface{}{"system": *system, "read_pct": readPercent}
	return c.tp.invoke("system_read_cache_pct_update", args, nil)
}

// IscsiChapAuthSet iSCSI CHAP authentication.
func (c *ClientConnection) IscsiChapAuthSet(initID string, inUser *string, inPassword *string,
	outUser *string, outPassword *string) error {

	args := map[string]interface{}{
		"init_id":      initID,
		"in_user":      inUser,
		"in_password":  inPassword,
		"out_user":     outUser,
		"out_password": outPassword,
	}

	return c.tp.invoke("iscsi_chap_auth", args, nil)
}

// VolumeCreate creates a block device, returns job id, error.
// If job id and error are nil, then returnedVolume has newly created volume.
func (c *ClientConnection) VolumeCreate(
	pool *Pool,
	volumeName string,
	size uint64,
	provisioning VolumeProvisionType,
	sync bool) (*Volume, *string, error) {
	args := map[string]interface{}{
		"pool":         *pool,
		"volume_name":  volumeName,
		"size_bytes":   size,
		"provisioning": provisioning,
	}

	var returnedVolume Volume
	var result [2]json.RawMessage
	jobID, err := c.getJobOrResult(c.tp.invoke("volume_create", args, &result), result, sync, &returnedVolume)
	return ensureExclusiveVol(&returnedVolume, jobID, err)
}

// VolumeDelete deletes a block device.
func (c *ClientConnection) VolumeDelete(vol *Volume, sync bool) (*string, error) {
	args := map[string]interface{}{"volume": *vol}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("volume_delete", args, &result), result, sync)
}

// VolumeResize resizes an existing volume, data loss may occur depending on storage implementation.
func (c *ClientConnection) VolumeResize(vol *Volume, newSizeBytes uint64, sync bool) (*Volume, *string, error) {
	args := map[string]interface{}{"volume": *vol, "new_size_bytes": newSizeBytes}
	var returnedVolume Volume
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("volume_resize", args, &result), result, sync, &returnedVolume)
	return ensureExclusiveVol(&returnedVolume, job, err)
}

// VolumeReplicate makes a replicated image of existing Volume
func (c *ClientConnection) VolumeReplicate(
	optionalPool *Pool, repType VolumeReplicateType, sourceVolume *Volume, name string,
	sync bool) (*Volume, *string, error) {

	args := map[string]interface{}{
		"volume_src": *sourceVolume,
		"rep_type":   repType,
		"name":       name,
	}
	if optionalPool != nil {
		args["pool"] = *optionalPool
	} else {
		args["pool"] = nil
	}

	var returnedVolume Volume
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("volume_replicate", args, &result), result, sync, &returnedVolume)
	return ensureExclusiveVol(&returnedVolume, job, err)
}

// VolumeRepRangeBlkSize block size for replicating a range of blocks
func (c *ClientConnection) VolumeRepRangeBlkSize(system *System) (uint32, error) {
	args := map[string]interface{}{"system": *system}
	var blkSize uint32
	return blkSize, c.tp.invoke("volume_replicate_range_block_size", args, &blkSize)
}

// VolumeReplicateRange replicates a range of blocks on the same or different Volume
func (c *ClientConnection) VolumeReplicateRange(
	repType VolumeReplicateType, srcVol *Volume, dstVol *Volume,
	ranges []BlockRange, sync bool) (*string, error) {

	args := map[string]interface{}{
		"rep_type":    repType,
		"ranges":      ranges,
		"volume_src":  *srcVol,
		"volume_dest": *dstVol,
	}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("volume_replicate_range", args, &result), result, sync)
}

// VolumeEnable sets a volume to online.
func (c *ClientConnection) VolumeEnable(vol *Volume) error {
	args := map[string]interface{}{"volume": *vol}
	return c.tp.invoke("volume_enable", args, nil)
}

// VolumeDisable sets a volume to offline.
func (c *ClientConnection) VolumeDisable(vol *Volume) error {
	args := map[string]interface{}{"volume": *vol}
	return c.tp.invoke("volume_disable", args, nil)
}

// VolumeMask grants access to a volume for the specified access group.
func (c *ClientConnection) VolumeMask(vol *Volume, ag *AccessGroup) error {
	args := map[string]interface{}{"volume": *vol, "access_group": *ag}
	return c.tp.invoke("volume_mask", args, nil)
}

// VolumeUnMask removes access to a volume for the specified access group.
func (c *ClientConnection) VolumeUnMask(vol *Volume, ag *AccessGroup) error {
	args := map[string]interface{}{"volume": *vol, "access_group": *ag}
	return c.tp.invoke("volume_unmask", args, nil)
}

// VolsMaskedToAg returns the volumes accessible to access group
func (c *ClientConnection) VolsMaskedToAg(ag *AccessGroup) ([]Volume, error) {
	args := map[string]interface{}{"access_group": *ag}
	var volumes []Volume
	return volumes, c.tp.invoke("volumes_accessible_by_access_group", args, &volumes)
}

// AgsGrantedToVol returns access group(s) which have access to specified volume
func (c *ClientConnection) AgsGrantedToVol(vol *Volume) ([]AccessGroup, error) {
	args := map[string]interface{}{"volume": *vol}
	var accessGroups []AccessGroup
	return accessGroups, c.tp.invoke("access_groups_granted_to_volume", args, &accessGroups)
}

// VolHasChildDep returns true|false if volume has child dependency
func (c *ClientConnection) VolHasChildDep(vol *Volume) (bool, error) {
	args := map[string]interface{}{"volume": *vol}
	var deps bool
	return deps, c.tp.invoke("volume_child_dependency", args, &deps)
}

// VolChildDepRm removes any child dependencies
func (c *ClientConnection) VolChildDepRm(vol *Volume, sync bool) (*string, error) {
	args := map[string]interface{}{"volume": *vol}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("volume_child_dependency_rm", args, &result), result, sync)
}

// FsCreate creates a file system, returns job id, error.
// If job id and error are nil, then returnedFs has newly created filesystem.
func (c *ClientConnection) FsCreate(
	pool *Pool,
	name string,
	size uint64,
	sync bool) (*FileSystem, *string, error) {
	args := map[string]interface{}{
		"pool":       *pool,
		"name":       name,
		"size_bytes": size,
	}
	var returnedFs FileSystem
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("fs_create", args, &result), result, sync, &returnedFs)
	return ensureExclusiveFs(&returnedFs, job, err)
}

// FsResize resizes an existing file system
func (c *ClientConnection) FsResize(
	fs *FileSystem, newSizeBytes uint64, sync bool) (*FileSystem, *string, error) {
	args := map[string]interface{}{"fs": *fs, "new_size_bytes": newSizeBytes}
	var returnedFs FileSystem
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("fs_resize", args, &result), result, sync, &returnedFs)
	return ensureExclusiveFs(&returnedFs, job, err)
}

// FsDelete deletes a file system.
func (c *ClientConnection) FsDelete(fs *FileSystem, sync bool) (*string, error) {
	args := map[string]interface{}{"fs": *fs}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("fs_delete", args, &result), result, sync)
}

// FsClone makes a clone of an existing file system
func (c *ClientConnection) FsClone(
	srcFs *FileSystem,
	destName string,
	optionalSnapShot *FileSystemSnapShot,
	sync bool) (*FileSystem, *string, error) {
	args := map[string]interface{}{"src_fs": *srcFs, "dest_fs_name": destName}

	handleSnapshotOptArg(args, optionalSnapShot)

	var returnedFs FileSystem
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("fs_clone", args, &result), result, sync, &returnedFs)
	return ensureExclusiveFs(&returnedFs, job, err)
}

// FsFileClone makes a clone of an existing file system
func (c *ClientConnection) FsFileClone(
	fs *FileSystem,
	srcFileName string,
	dstFileName string,
	optionalSnapShot *FileSystemSnapShot,
	sync bool,
) (*string, error) {
	args := map[string]interface{}{
		"fs":             *fs,
		"src_file_name":  srcFileName,
		"dest_file_name": dstFileName,
	}

	handleSnapshotOptArg(args, optionalSnapShot)

	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("fs_file_clone", args, &result), result, sync)
}

// FsSnapShotCreate creates a file system snapshot for the supplied snapshot
// If job id and error are nil, then returnedFs has newly created filesystem.
func (c *ClientConnection) FsSnapShotCreate(fs *FileSystem, name string, sync bool) (*FileSystemSnapShot, *string, error) {
	args := map[string]interface{}{"fs": *fs, "snapshot_name": name}
	var returnedSnapshot FileSystemSnapShot
	var result [2]json.RawMessage
	job, err := c.getJobOrResult(c.tp.invoke("fs_snapshot_create", args, &result), result, sync, &returnedSnapshot)
	return ensureExclusiveSs(&returnedSnapshot, job, err)
}

// FsSnapShotDelete deletes a file system snapshot.
func (c *ClientConnection) FsSnapShotDelete(fs *FileSystem, snapShot *FileSystemSnapShot, sync bool) (*string, error) {
	args := map[string]interface{}{"fs": *fs, "snapshot": *snapShot}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("fs_snapshot_delete", args, &result), result, sync)
}

// FsSnapShots returns list of file system snapsthos for specified file system.
// can be created from.
func (c *ClientConnection) FsSnapShots(fs *FileSystem) ([]FileSystemSnapShot, error) {
	args := map[string]interface{}{"fs": *fs}
	var snapShots []FileSystemSnapShot
	return snapShots, c.tp.invoke("fs_snapshots", args, &snapShots)
}

// FsSnapShotRestore restores all the files for a file systems or specific files.
func (c *ClientConnection) FsSnapShotRestore(
	fs *FileSystem, snapShot *FileSystemSnapShot, allFiles bool,
	files []string, restoreFiles []string, sync bool) (*string, error) {

	if !allFiles {
		if len(files) == 0 {
			return nil, &errors.LsmError{
				Code:    errors.InvalidArgument,
				Message: "'files' is empty and 'all_files' is false!"}
		}

		if len(files) != len(restoreFiles) {
			return nil, &errors.LsmError{
				Code:    errors.InvalidArgument,
				Message: "'files' and 'restoreFiles' have different lengths!"}
		}
	}

	args := map[string]interface{}{
		"fs":            *fs,
		"snapshot":      *snapShot,
		"files":         files,
		"restore_files": restoreFiles,
		"all_files":     allFiles,
	}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("fs_snapshot_restore", args, &result), result, sync)
}

// FsHasChildDep checks whether file system has a child dependency.
func (c *ClientConnection) FsHasChildDep(fs *FileSystem, files []string) (bool, error) {
	args := map[string]interface{}{"fs": *fs, "files": files}
	var result bool
	return result, c.tp.invoke("fs_child_dependency", args, &result)
}

// FsChildDepRm remove dependencies for specified file system.
func (c *ClientConnection) FsChildDepRm(
	fs *FileSystem, files []string, sync bool) (*string, error) {
	args := map[string]interface{}{"fs": *fs, "files": files}
	var result json.RawMessage
	return c.getJobOrNone(c.tp.invoke("fs_child_dependency_rm", args, &result), result, sync)
}

// AccessGroupCreate creates an access group.
func (c *ClientConnection) AccessGroupCreate(name string, initID string,
	initType InitiatorType, system *System) (*AccessGroup, error) {

	if check := validateInitID(initID, initType); check != nil {
		return nil, check
	}

	args := map[string]interface{}{
		"name":      name,
		"init_id":   initID,
		"init_type": initType,
		"system":    *system,
	}
	var accessGroup AccessGroup
	if err := c.tp.invoke("access_group_create", args, &accessGroup); err != nil {
		return nil, err
	}
	return &accessGroup, nil
}

// AccessGroupDelete deletes an access group.
func (c *ClientConnection) AccessGroupDelete(ag *AccessGroup) error {
	args := map[string]interface{}{"access_group": *ag}
	return c.tp.invoke("access_group_delete", args, nil)
}

func initSetup(initID string,
	initType InitiatorType, accessGroup *AccessGroup) (map[string]interface{}, error) {
	args := map[string]interface{}{"access_group": *accessGroup, "init_id": initID, "init_type": initType}
	return args, validateInitID(initID, initType)
}

// AccessGroupInitAdd adds an initiator to an access group.
func (c *ClientConnection) AccessGroupInitAdd(ag *AccessGroup,
	initID string, initType InitiatorType) (*AccessGroup, error) {

	var args, setupErr = initSetup(initID, initType, ag)
	if setupErr != nil {
		return nil, setupErr
	}

	var accessGroup AccessGroup
	if err := c.tp.invoke("access_group_initiator_add", args, &accessGroup); err != nil {
		return nil, err
	}
	return &accessGroup, nil
}

// AccessGroupInitDelete deletes an initiator from an access group.
func (c *ClientConnection) AccessGroupInitDelete(ag *AccessGroup,
	initID string, initType InitiatorType) (*AccessGroup, error) {
	var args, setupErr = initSetup(initID, initType, ag)
	if setupErr != nil {
		return nil, setupErr
	}
	var accessGroup AccessGroup
	if err := c.tp.invoke("access_group_initiator_delete", args, &accessGroup); err != nil {
		return nil, err
	}
	return &accessGroup, nil
}

// VolRaidInfo retrieves RAID information about specified volume.
func (c *ClientConnection) VolRaidInfo(vol *Volume) (*VolumeRaidInfo, error) {
	args := map[string]interface{}{"volume": *vol}

	var ret [5]int32
	if err := c.tp.invoke("volume_raid_info", args, &ret); err != nil {
		return nil, err
	}
	var info VolumeRaidInfo
	info.Type = RaidType(ret[0])
	info.StripSize = uint32(ret[1])
	info.DiskCount = uint32(ret[2])
	info.MinIOSize = uint32(ret[3])
	info.OptIOSize = uint32(ret[4])
	return &info, nil
}

// PoolMemberInfo retrieves RAID information about specified volume.
func (c *ClientConnection) PoolMemberInfo(pool *Pool) (*PoolMemberInfo, error) {
	args := map[string]interface{}{"pool": *pool}

	var ret [3]json.RawMessage
	if err := c.tp.invoke("pool_member_info", args, &ret); err != nil {
		return nil, err
	}

	var info PoolMemberInfo

	// JSON is [number, number, [string,] ]

	if uE := json.Unmarshal(ret[0], &info.Raid); uE != nil {
		return nil, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("First array item not a raid type %s", ret[0])}
	}

	if uE := json.Unmarshal(ret[1], &info.Member); uE != nil {
		return nil, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Second array item not a pool member type %s", ret[1])}
	}

	if uE := json.Unmarshal(ret[2], &info.ID); uE != nil {
		return nil, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Third array item not array of strings %s", ret[2])}
	}

	return &info, nil
}

// VolRaidCreateCapGet returns supported RAID types and strip sizes for hardware raid.
func (c *ClientConnection) VolRaidCreateCapGet(system *System) (*SupportedRaidCapability, error) {
	args := map[string]interface{}{"system": *system}
	var ret []json.RawMessage
	if err := c.tp.invoke("volume_raid_create_cap_get", args, &ret); err != nil {
		return nil, err
	}

	var info SupportedRaidCapability
	if uE := json.Unmarshal(ret[0], &info.Types); uE != nil {
		return nil, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("First array item not array of raid types %s", ret[0])}
	}

	if uE := json.Unmarshal(ret[1], &info.StripeSizes); uE != nil {
		return nil, &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Second array item not array of stripe sizes %s", ret[1])}
	}

	return &info, nil
}

func paramError(msg string) error {
	return &errors.LsmError{
		Code:    errors.InvalidArgument,
		Message: msg}
}

// VolRaidCreate creates RAIDed volume directly from disks, only for hardware RAID.
func (c *ClientConnection) VolRaidCreate(name string,
	raidType RaidType, disks []Disk, stripSize uint32) (*Volume, error) {

	if len(disks) == 0 {
		return nil, paramError("no disks included")
	}

	if raidType == Raid1 && len(disks) != 2 {
		return nil, paramError("RAID 1 only allows 2 disks")
	}

	if raidType == Raid5 && len(disks) < 3 {
		return nil, paramError("RAID 5 requires 3 or more disks")
	}

	if raidType == Raid6 && len(disks) < 4 {
		return nil, paramError("RAID 6 requires 4 or more disks")
	}

	if raidType == Raid10 && (len(disks)%2 != 0 || len(disks) < 4) {
		return nil, paramError("RAID 10 requires even disks count and 4 or more disks")
	}

	if raidType == Raid50 && (len(disks)%2 != 0 || len(disks) < 6) {
		return nil, paramError("RAID 50 requires even disks count and 6 or more disks")
	}

	if raidType == Raid60 && (len(disks)%2 != 0 || len(disks) < 8) {
		return nil, paramError("RAID 60 requires even disks count and 8 or more disks")
	}

	args := map[string]interface{}{
		"name":       name,
		"raid_type":  raidType,
		"disks":      disks,
		"strip_size": stripSize, //stripe
	}
	var returnedVolume Volume
	if err := c.tp.invoke("volume_raid_create", args, &returnedVolume); err != nil {
		return nil, err
	}
	return &returnedVolume, nil
}

func (c *ClientConnection) identLED(volume *Volume, method string) error {
	args := map[string]interface{}{"volume": *volume}
	return c.tp.invoke(method, args, nil)
}

// VolIdentLedOn turn on the identification LED for the specified volume.
func (c *ClientConnection) VolIdentLedOn(volume *Volume) error {
	return c.identLED(volume, "volume_ident_led_on")
}

// VolIdentLedOff turn off the identification LED for the specified volume.
func (c *ClientConnection) VolIdentLedOff(volume *Volume) error {
	return c.identLED(volume, "volume_ident_led_off")
}

// VolCacheInfo returns cache information for specified volume
func (c *ClientConnection) VolCacheInfo(volume *Volume) (*VolumeCacheInfo, error) {
	args := map[string]interface{}{"volume": *volume}

	var ret [5]uint32
	if err := c.tp.invoke("volume_cache_info", args, &ret); err != nil {
		return nil, err
	}

	var info VolumeCacheInfo
	info.WriteSetting = WriteCachePolicy(ret[0])
	info.WriteStatus = WriteCacheStatus(ret[1])
	info.ReadSetting = ReadCachePolicy(ret[2])
	info.ReadStatus = ReadCacheStatus(ret[3])
	info.PhysicalDiskStatus = PhysicalDiskCache(ret[4])
	return &info, nil
}

// VolPhyDiskCacheSet set the volume physical disk cache policy
func (c *ClientConnection) VolPhyDiskCacheSet(volume *Volume, pdc PhysicalDiskCache) error {
	args := map[string]interface{}{
		"volume": *volume,
		"pdc":    pdc,
	}
	return c.tp.invoke("volume_physical_disk_cache_update", args, nil)
}

// VolWriteCacheSet sets volume write cache policy
func (c *ClientConnection) VolWriteCacheSet(volume *Volume, wcp WriteCachePolicy) error {
	args := map[string]interface{}{
		"volume": *volume,
		"wcp":    wcp,
	}
	return c.tp.invoke("volume_write_cache_policy_update", args, nil)
}

// VolReadCacheSet sets volume read cache policy
func (c *ClientConnection) VolReadCacheSet(volume *Volume, rcp ReadCachePolicy) error {
	args := map[string]interface{}{
		"volume": *volume,
		"rcp":    rcp,
	}
	return c.tp.invoke("volume_read_cache_policy_update", args, nil)
}
