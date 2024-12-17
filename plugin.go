// SPDX-License-Identifier: 0BSD

package libstoragemgmt

import (
	"fmt"
	"net"
	"os"
	"strconv"

	errors "github.com/libstorage/libstoragemgmt-golang/errors"
)

// TmoSetCb used to register timeout value for plugin
type TmoSetCb func(timeout uint32) error

// TmoGetCb used to register timeout value for plugin
type TmoGetCb func() uint32

// CapabilitiesCb returns what the plugin is capable of
type CapabilitiesCb func(system *System) (*Capabilities, error)

// JobInfo is the information about a job
type JobInfo struct {
	Status  JobStatusType
	Percent uint8
	Item    interface{}
}

// JobStatusCb callback returns the job status for the specified job
type JobStatusCb func(jobId string) (*JobInfo, error)

// JobFreeCb callback for freeing job resources
type JobFreeCb func(jobID string) error

// PoolsCb callback for pools
type PoolsCb func(search ...string) ([]Pool, error)

// PluginRegisterCb callback to register needed information
type PluginRegisterCb func(p *PluginRegister) error

// PluginUnregisterCb callback to allow plugin to cleanup resources
type PluginUnregisterCb func() error

// SystemsCb callback to retrieve systems
type SystemsCb func() ([]System, error)

// DisksCb callback to retrieve disks
type DisksCb func() ([]Disk, error)

// VolumesCb callback to retrieve volumes
type VolumesCb func(search ...string) ([]Volume, error)

// VolumeCreateCb callback is for creating a volume
type VolumeCreateCb func(pool *Pool,
	volumeName string,
	size uint64,
	provisioning VolumeProvisionType) (*Volume, *string, error)

// VolumeDeleteCb callback is for deleting a volume
type VolumeDeleteCb func(vol *Volume) (*string, error)

// VolumeReplicateCb returns volume, job id, error.
type VolumeReplicateCb func(optionalPool *Pool, repType VolumeReplicateType,
	sourceVolume *Volume, name string) (*Volume, *string, error)

// VolumeReplicateRangeCb returns job id, error
type VolumeReplicateRangeCb func(repType VolumeReplicateType, srcVol *Volume, dstVol *Volume,
	ranges []BlockRange) (*string, error)

// VolumeRepRangeBlkSizeCb returns blocksize, error
type VolumeRepRangeBlkSizeCb func(system *System) (uint32, error)

// VolumeResizeCb returns volume, job id, error
type VolumeResizeCb func(vol *Volume, newSizeBytes uint64) (*Volume, *string, error)

// VolumeEnableCb enables a volume
type VolumeEnableCb func(vol *Volume) error

// VolumeDisableCb enables a volume
type VolumeDisableCb func(vol *Volume) error

// VolumeMaskCb maskes a volume to the associated access group
type VolumeMaskCb func(vol *Volume, ag *AccessGroup) error

// VolumeUnMaskCb unmaskes a volume from the associated access group
type VolumeUnMaskCb func(vol *Volume, ag *AccessGroup) error

// VolsMaskedToAgCb returns those volumes accessible from specified access group
type VolsMaskedToAgCb func(ag *AccessGroup) ([]Volume, error)

// AgsGrantedToVolCb returns access group(s) which have access to specified volume
type AgsGrantedToVolCb func(vol *Volume) ([]AccessGroup, error)

// AccessGroupsCb returns all the access groups
type AccessGroupsCb func() ([]AccessGroup, error)

// AccessGroupCreateCb creates an access group
type AccessGroupCreateCb func(name string, initID string, initType InitiatorType, system *System) (*AccessGroup, error)

// AccessGroupDeleteCb deletes an access group
type AccessGroupDeleteCb func(ag *AccessGroup) error

// AccessGroupInitAddCb adds an initiator to an AccessGroup
type AccessGroupInitAddCb func(ag *AccessGroup,
	initID string, initType InitiatorType) (*AccessGroup, error)

// AccessGroupInitDeleteCb removes an initiator from an AccessGroup
type AccessGroupInitDeleteCb func(ag *AccessGroup,
	initID string, initType InitiatorType) (*AccessGroup, error)

// IscsiChapAuthSetCb iSCSI CHAP authentication
type IscsiChapAuthSetCb func(initID string, inUser *string, inPassword *string, outUser *string, outPassword *string) error

// VolHasChildDepCb returns boolean on if specified volume has child dependencies
type VolHasChildDepCb func(vol *Volume) (bool, error)

// VolChildDepRmCb removes any child dependencies
type VolChildDepRmCb func(vol *Volume) (*string, error)

// TargetPortsCb returns target ports
type TargetPortsCb func() ([]TargetPort, error)

// VolIdentLedOnCb turn identification led on
type VolIdentLedOnCb func(volume *Volume) error

// VolIdentLedOffCb turn identification led off
type VolIdentLedOffCb func(volume *Volume) error

// ManagementOps are the callbacks that plugins must implement
type ManagementOps struct {
	TimeOutSet       TmoSetCb
	TimeOutGet       TmoGetCb
	JobStatus        JobStatusCb
	JobFree          JobFreeCb
	Capabilities     CapabilitiesCb
	Systems          SystemsCb
	Pools            PoolsCb
	PluginRegister   PluginRegisterCb
	PluginUnregister PluginUnregisterCb
}

// SanOps are storage area network callbacks
type SanOps struct {
	Volumes               VolumesCb
	VolumeCreate          VolumeCreateCb
	VolumeDelete          VolumeDeleteCb
	Disks                 DisksCb
	VolumeReplicate       VolumeReplicateCb
	VolumeReplicateRange  VolumeReplicateRangeCb
	VolumeRepRangeBlkSize VolumeRepRangeBlkSizeCb
	VolumeResize          VolumeResizeCb
	VolumeEnable          VolumeEnableCb
	VolumeDisable         VolumeDisableCb
	VolumeMask            VolumeMaskCb
	VolumeUnMask          VolumeUnMaskCb
	VolsMaskedToAg        VolsMaskedToAgCb
	VolHasChildDep        VolHasChildDepCb
	VolChildDepRm         VolChildDepRmCb
	AccessGroups          AccessGroupsCb
	AccessGroupCreate     AccessGroupCreateCb
	AccessGroupDelete     AccessGroupDeleteCb
	AccessGroupInitAdd    AccessGroupInitAddCb
	AccessGroupInitDelete AccessGroupInitDeleteCb
	AgsGrantedToVol       AgsGrantedToVolCb
	IscsiChapAuthSet      IscsiChapAuthSetCb
	TargetPorts           TargetPortsCb
	VolIdentLedOn         VolIdentLedOnCb
	VolIdentLedOff        VolIdentLedOffCb
}

// FsCb callback returns filesystems
type FsCb func(search ...string) ([]FileSystem, error)

// FsCreateCb callback creates a file system
type FsCreateCb func(pool *Pool, name string, size uint64) (*FileSystem, *string, error)

// FsDeleteCb callback deletes a file system
type FsDeleteCb func(fs *FileSystem) (*string, error)

// FsResizeCb callback resizes a file system
type FsResizeCb func(fs *FileSystem, newSizeBytes uint64) (*FileSystem, *string, error)

// FsCloneCb callback clones a file system
type FsCloneCb func(srcFs *FileSystem,
	destName string,
	optionalSnapShot *FileSystemSnapShot) (*FileSystem, *string, error)

// FsFileCloneCb callback snap shots files on a file system
type FsFileCloneCb func(fs *FileSystem,
	srcFileName string,
	dstFileName string,
	optionalSnapShot *FileSystemSnapShot) (*string, error)

// FsSnapShotCreateCb callback creates a snapshot
type FsSnapShotCreateCb func(s *FileSystem, name string) (*FileSystemSnapShot, *string, error)

// FsSnapShotDeleteCb callback deletes a snapshot
type FsSnapShotDeleteCb func(fs *FileSystem, snapShot *FileSystemSnapShot) (*string, error)

// FsSnapShotsCb callback returns array of file system snapshots
type FsSnapShotsCb func(fs *FileSystem) ([]FileSystemSnapShot, error)

// FsSnapShotRestoreCb callback restores a file system from a snapshot
type FsSnapShotRestoreCb func(
	fs *FileSystem, snapShot *FileSystemSnapShot, allFiles bool,
	files []string, restoreFiles []string) (*string, error)

// FsHasChildDepCb callback returns boolean indicating if filesystem has child dependencies
type FsHasChildDepCb func(fs *FileSystem, files []string) (bool, error)

// FsChildDepRmCb callback removes child filesystem dependecies by replicating as needed
type FsChildDepRmCb func(fs *FileSystem, files []string) (*string, error)

// FsOps file system callbacks
type FsOps struct {
	FileSystems       FsCb
	FsCreate          FsCreateCb
	FsDelete          FsDeleteCb
	FsResize          FsResizeCb
	FsClone           FsCloneCb
	FsFileClone       FsFileCloneCb
	FsSnapShotCreate  FsSnapShotCreateCb
	FsSnapShotDelete  FsSnapShotDeleteCb
	FsSnapShots       FsSnapShotsCb
	FsSnapShotRestore FsSnapShotRestoreCb
	FsHasChildDep     FsHasChildDepCb
	FsChildDepRm      FsChildDepRmCb
}

// ExportsCb returns all exported file systems
type ExportsCb func(search ...string) ([]NfsExport, error)

// ExportAuthTypesCb returns array of strings that state what authentication types are supported
type ExportAuthTypesCb func() ([]string, error)

// FsExportCb exports a file system over NFS
type FsExportCb func(fs *FileSystem, exportPath *string,
	access *NfsAccess, authType *string, options *string) (*NfsExport, error)

// FsUnExportCb removes a NFS export
type FsUnExportCb func(export *NfsExport) error

// NfsOps orientated callbacks
type NfsOps struct {
	Exports         ExportsCb
	ExportAuthTypes ExportAuthTypesCb
	FsExport        FsExportCb
	FsUnExport      FsUnExportCb
}

// VolRaidInfoCb callback for volume raid info
type VolRaidInfoCb func(vol *Volume) (*VolumeRaidInfo, error)

// PoolMemberInfoCb callback for pool member info
type PoolMemberInfoCb func(pool *Pool) (*PoolMemberInfo, error)

// VolRaidCreateCapGetCb callback for getting raid capacity
type VolRaidCreateCapGetCb func(system *System) (*SupportedRaidCapability, error)

// VolRaidCreateCb callback for creating a volume on HBA raid
type VolRaidCreateCb func(name string,
	raidType RaidType, disks []Disk, stripSize uint32) (*Volume, error)

// BatteriesCb returns array of batteries
type BatteriesCb func() ([]Battery, error)

// HbaRaidOps callbacks for HBA raid
type HbaRaidOps struct {
	VolRaidInfo         VolRaidInfoCb
	PoolMemberInfo      PoolMemberInfoCb
	VolRaidCreateCapGet VolRaidCreateCapGetCb
	VolRaidCreate       VolRaidCreateCb
	Batteries           BatteriesCb
}

// SysReadCachePctSetCb callback for changing the read cache percentage for the specified system
type SysReadCachePctSetCb func(system *System, readPercent uint32) error

// VolCacheInfoCb callback for cache information for specified volume
type VolCacheInfoCb func(volume *Volume) (*VolumeCacheInfo, error)

// VolPhyDiskCacheSetCb callback for setting the physcial disk cache policy
type VolPhyDiskCacheSetCb func(volume *Volume, pdc PhysicalDiskCache) error

// VolWriteCacheSetCb callback for setting the volume write cache policy
type VolWriteCacheSetCb func(volume *Volume, wcp WriteCachePolicy) error

// VolReadCacheSetCb callback for setting the read cache policy
type VolReadCacheSetCb func(volume *Volume, rcp ReadCachePolicy) error

// CacheOps callbacks for caching
type CacheOps struct {
	SysReadCachePctSet SysReadCachePctSetCb
	VolCacheInfo       VolCacheInfoCb
	VolPhyDiskCacheSet VolPhyDiskCacheSetCb
	VolWriteCacheSet   VolWriteCacheSetCb
	VolReadCacheSet    VolReadCacheSetCb
}

// PluginCallBacks callbacks for plugin to implement
type PluginCallBacks struct {
	Mgmt  ManagementOps
	San   SanOps
	File  FsOps
	Nfs   NfsOps
	Hba   HbaRaidOps
	Cache CacheOps
}

type handler func(p *Plugin, params *requestMsg) (interface{}, error)

// Plugin represents plugin
type Plugin struct {
	tp        transPort
	cb        *PluginCallBacks
	callTable map[string]handler
	desc      string
	ver       string
}

// PluginRegister data passed to PluginRegister callback
type PluginRegister struct {
	URI      string
	Password string
	Timeout  uint32
	Flags    uint32
}

// PluginInit initializes the plugin with the specified callbacks
func PluginInit(callbacks *PluginCallBacks, cmdLineArgs []string, desc string, ver string) (*Plugin, error) {

	var fd int64 = -1
	var fdStr string

	fdStr = os.Getenv("LSM_GO_FD")

	if len(fdStr) == 0 && len(cmdLineArgs) == 2 {
		fdStr = cmdLineArgs[1]
	}

	fd, err := strconv.ParseInt(fdStr, 10, 32)
	if err != nil {
		return nil, err
	}

	if fd != -1 {
		// Only information I could find which pretains to how to do this.
		// https://play.golang.org/p/0uEcuPk291
		f := os.NewFile(uintptr(fd), "client")
		s, err := net.FileConn(f)
		if err != nil {
			return nil, err
		}

		tp := transPort{uds: s, debug: false}
		return &Plugin{tp: tp, cb: callbacks, callTable: buildTable(callbacks), desc: desc, ver: ver}, nil
	}
	return nil, &errors.LsmError{
		Code:    errors.LibBug,
		Message: fmt.Sprintf("Plugin called with invalid args: %s\n", cmdLineArgs)}
}

func noSupport(tp *transPort, method string) {
	tp.sendError(&errors.LsmError{
		Code: errors.NoSupport,
		Message: fmt.Sprintf(
			"method %s not supported", method)})
}

// Run the plugin, looping processing requests and sending responses.
func (p *Plugin) Run() {
	for {
		request, err := p.tp.readRequest()
		if err != nil {
			if lsmError, ok := err.(*errors.LsmError); ok == true {

				if lsmError.Code != errors.TransPortCommunication {
					p.tp.sendError(lsmError)
					//fmt.Printf("Returned error %+v\n", lsmError)
					continue
				} else {
					fmt.Printf("Communication error: exiting! %s\n", lsmError)
				}
				return
			}
			fmt.Printf("Unexpected error, exiting! %s\n", err)
			return
		}

		var response interface{}
		if f, ok := p.callTable[request.Method]; ok == true && f != nil {
			//fmt.Printf("Executing %s(%s)\n", request.Method, string(request.Params))
			response, err = f(p, request)
			if err != nil {
				p.tp.sendError(err)
			} else {
				p.tp.sendResponse(response)
			}

			// Need to shut down the connection.
			if request.Method == "plugin_unregister" {
				p.tp.close()
				return
			}
		} else {
			noSupport(&p.tp, request.Method)
		}
	}
}
