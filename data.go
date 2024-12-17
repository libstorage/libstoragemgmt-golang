// SPDX-License-Identifier: 0BSD

package libstoragemgmt

// PluginInfo - Information about a specific plugin
type PluginInfo struct {
	Version     string
	Description string
	Name        string
}

// System represents a storage system.
// * A hardware RAID card
// * A storage area network (SAN)
// * A software solutions running on commodity hardware
// * A Linux system running NFS Service
type System struct {
	Class        string           `json:"class"`
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Status       SystemStatusType `json:"status"`
	StatusInfo   string           `json:"status_info"`
	PluginData   *string          `json:"plugin_data"`
	FwVersion    string           `json:"fw_version"`
	ReadCachePct int8             `json:"read_cache_pct"`
	SystemMode   SystemModeType   `json:"mode"`
}

const (
	// SystemReadCachePctNoSupport System read cache percentage not supported.
	SystemReadCachePctNoSupport int8 = -2 + iota

	// SystemReadCachePctUnknown System read cache percentage unknown.
	SystemReadCachePctUnknown
)

// SystemStatusType type representing system status.
type SystemStatusType uint32

const (
	// SystemStatusUnknown System status is unknown.
	SystemStatusUnknown SystemStatusType = 1 << iota

	// SystemStatusOk  System status is OK.
	SystemStatusOk

	// SystemStatusError System is in error state.
	SystemStatusError

	// SystemStatusDegraded System is degraded in some way
	SystemStatusDegraded

	// SystemStatusPredictiveFailure System has potential failure.
	SystemStatusPredictiveFailure

	// SystemStatusOther Vendor specific status.
	SystemStatusOther
)

// SystemModeType type representing system mode.
type SystemModeType int8

const (
	// SystemModeUnknown Plugin failed to query system mode.
	SystemModeUnknown SystemModeType = -2 + iota

	// SystemModeNoSupport Plugin does not support querying system mode.
	SystemModeNoSupport

	//SystemModeHardwareRaid The storage system is a hardware RAID card
	SystemModeHardwareRaid

	// SystemModeHba The physical disks can be exposed to OS directly without any
	// configurations.
	SystemModeHba
)

// Volume represents a storage volume, aka. a logical unit
type Volume struct {
	Class       string  `json:"class"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Enabled     LsmBool `json:"admin_state"`
	BlockSize   uint64  `json:"block_size"`
	NumOfBlocks uint64  `json:"num_of_blocks"`
	PluginData  *string `json:"plugin_data"`
	Vpd83       string  `json:"vpd83"`
	SystemID    string  `json:"system_id"`
	PoolID      string  `json:"pool_id"`
}

// JobStatusType is enumerated type returned from Job control
type JobStatusType uint32

const (

	// JobStatusInProgress indicated job is in progress
	JobStatusInProgress JobStatusType = 1 + iota

	// JobStatusComplete indicates job is complete
	JobStatusComplete

	// JobStatusError indicated job has errored
	JobStatusError
)

// VolumeReplicateType enumerated type for VolumeReplicate
type VolumeReplicateType int

const (
	// VolumeReplicateTypeUnknown plugin failed to detect volume replication type
	VolumeReplicateTypeUnknown VolumeReplicateType = -1 + iota

	// Reserved "0"
	_
	// Reserved "1"
	_

	// VolumeReplicateTypeClone Point in time read writeable space efficient copy of data
	VolumeReplicateTypeClone

	// VolumeReplicateTypeCopy Full bitwise copy of the data (occupies full space)
	VolumeReplicateTypeCopy

	// VolumeReplicateTypeMirrorSync I/O will be blocked until I/O reached
	// both source and target storage systems. There will be no data difference
	// between source and target storage systems.
	VolumeReplicateTypeMirrorSync

	// VolumeReplicateTypeMirrorAsync I/O will be blocked until I/O
	// reached source storage systems.  The source storage system will use
	// copy the changes data to target system in a predefined interval.
	// There will be a small data differences between source and target.
	VolumeReplicateTypeMirrorAsync
)

// VolumeProvisionType enumerated type for volume creation provisioning
type VolumeProvisionType int

const (
	// VolumeProvisionTypeUnknown provision type unknown
	VolumeProvisionTypeUnknown VolumeProvisionType = -1 + iota

	// Reserved "0"
	_

	// VolumeProvisionTypeThin thin provision volume
	VolumeProvisionTypeThin

	// VolumeProvisionTypeFull fully provision volume
	VolumeProvisionTypeFull

	// VolumeProvisionTypeDefault use the default for the storage provider
	VolumeProvisionTypeDefault
)

// Pool represents the unit of storage where block
// devices and/or file systems are created from.
type Pool struct {
	Class              string              `json:"class"`
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	ElementType        PoolElementType     `json:"element_type"`
	UnsupportedActions PoolUnsupportedType `json:"unsupported_actions"`
	TotalSpace         uint64              `json:"total_space"`
	FreeSpace          uint64              `json:"free_space"`
	Status             PoolStatusType      `json:"status"`
	StatusInfo         string              `json:"status_info"`
	PluginData         *string             `json:"plugin_data"`
	SystemID           string              `json:"system_id"`
}

// PoolElementType type used to describe what things can be created from pool
type PoolElementType uint64

const (

	// PoolElementPool This pool could allocate space for sub pool.
	PoolElementPool PoolElementType = 1 << (iota + 1)

	// PoolElementTypeVolume This pool can be used for volume creation.
	PoolElementTypeVolume

	// PoolElementTypeFs this pool can be used to for FS creation.
	PoolElementTypeFs

	// PoolElementTypeDelta this pool can hold delta data for snapshots.
	PoolElementTypeDelta

	// PoolElementTypeVolumeFull this pool could be used to create fully allocated volume.
	PoolElementTypeVolumeFull

	// PoolElementTypeVolumeThin this pool could be used to create thin provisioned volume.
	PoolElementTypeVolumeThin

	// Reserved 1 << 7
	_
	// Reserved 1 << 8
	_
	// Reserved 1 << 9
	_

	// PoolElementTypeSysReserved this pool is reserved for internal use.
	PoolElementTypeSysReserved
)

// PoolUnsupportedType type used to describe what actions are unsupported
type PoolUnsupportedType uint64

const (
	// PoolUnsupportedVolumeGrow this pool does not allow growing volumes
	PoolUnsupportedVolumeGrow PoolUnsupportedType = 1 << iota

	// PoolUnsupportedVolumeShink this pool does not allow shrinking volumes
	PoolUnsupportedVolumeShink
)

// PoolStatusType type used to describe the status of pool
type PoolStatusType uint64

const (

	// PoolStatusUnknown Plugin failed to query pool status.
	PoolStatusUnknown PoolStatusType = 1 << iota

	// PoolStatusOk The data of this pool is accessible with no data loss. But it might
	// be set with PoolStatusDegraded to indicate redundancy loss.
	PoolStatusOk

	// PoolStatusOther Vendor specific status, check Pool.StatusInfo for more information.
	PoolStatusOther

	// Reserved 1 << 3
	_

	// PoolStatusDegraded indicates pool has lost data redundancy.
	PoolStatusDegraded

	// PoolStatusError indicates pool data is not accessible due to some members offline.
	PoolStatusError

	// Reserved 1 << 6
	_
	// Reserved 1 << 7
	_
	// Reserved 1 << 8
	_

	// PoolStatusStopped pool is stopped by administrator.
	PoolStatusStopped

	// PoolStatusStarting is reviving from STOPPED status. Pool data is not accessible yet.
	PoolStatusStarting

	// Reserved 1 << 11
	_

	// PoolStatusReconstructing pool is be reconstructing hash or mirror data.
	PoolStatusReconstructing

	// PoolStatusVerifying indicates array is running integrity check(s).
	PoolStatusVerifying

	// PoolStatusInitializing indicates pool is not accessible and performing initialization.
	PoolStatusInitializing

	// PoolStatusGrowing indicates pool is growing in size.  PoolStatusInfo can contain more
	// information about this task.  If PoolStatusOk is set, data is still accessible.
	PoolStatusGrowing
)

// Disk represents a physical device.
type Disk struct {
	Class       string         `json:"class"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	DiskType    DiskType       `json:"disk_type"`
	BlockSize   uint64         `json:"block_size"`
	NumOfBlocks uint64         `json:"num_of_blocks"`
	Status      DiskStatusType `json:"status"`
	PluginData  *string        `json:"plugin_data"`
	SystemID    string         `json:"system_id"`
	Location    string         `json:"location"`
	Rpm         int            `json:"rpm"`
	LinkType    DiskLinkType   `json:"link_type"`
	Vpd83       string         `json:"vpd83"`
}

// DiskType is an enumerated type representing different types of disks.
type DiskType int

const (
	// DiskTypeUnknown Plugin failed to query disk type
	DiskTypeUnknown DiskType = iota

	// DiskTypeOther Vendor specific disk type
	DiskTypeOther

	// Reserved "2"
	_

	// DiskTypeAta IDE disk type.
	DiskTypeAta

	// DiskTypeSata SATA disk
	DiskTypeSata

	// DiskTypeSas SAS disk.
	DiskTypeSas

	// DiskTypeFc FC disk.
	DiskTypeFc

	// DiskTypeSop SCSI over PCI-Express.
	DiskTypeSop

	// DiskTypeScsi SCSI disk.
	DiskTypeScsi

	// DiskTypeLun Remote LUN from SAN array.
	DiskTypeLun

	// Reserved 10 - 50

	// DiskTypeNlSas Near-Line SAS, just SATA disk + SAS port
	DiskTypeNlSas DiskType = iota + 41

	// DiskTypeHdd Normal HDD, fall back value if failed to detect HDD type(SAS/SATA/etc).
	DiskTypeHdd

	// DiskTypeSsd Solid State Drive.
	DiskTypeSsd

	// DiskTypeHybrid Hybrid disk uses a combination of HDD and SSD.
	DiskTypeHybrid
)

// DiskLinkType is an enumerated type representing different types of disk connection.
type DiskLinkType int

const (
	// DiskLinkTypeNoSupport Plugin does not support querying disk link type.
	DiskLinkTypeNoSupport DiskLinkType = iota + -2

	// DiskLinkTypeUnknown Plugin failed to query disk link type
	DiskLinkTypeUnknown

	// DiskLinkTypeFc Fibre channel
	DiskLinkTypeFc

	// Skip enumerated value "1" which is unused
	_

	//DiskLinkTypeSsa Serial Storage Architecture
	DiskLinkTypeSsa

	// DiskLinkTypeSbp Serial Bus Protocol, used by IEEE 1394.
	DiskLinkTypeSbp

	// DiskLinkTypeSrp SCSI RDMA Protocol
	DiskLinkTypeSrp

	// DiskLinkTypeIscsi Internet Small Computer System Interface
	DiskLinkTypeIscsi

	// DiskLinkTypeSas Serial Attached SCSI.
	DiskLinkTypeSas

	// DiskLinkTypeAdt Automation/Drive Interface Transport. Often used by tape.
	DiskLinkTypeAdt

	// DiskLinkTypeAta PATA/IDE or SATA.
	DiskLinkTypeAta

	// DiskLinkTypeUsb USB
	DiskLinkTypeUsb

	// DiskLinkTypeSop SCSI over PCI-E.
	DiskLinkTypeSop

	// DiskLinkTypePciE PCI-E, e.g. NVMe.
	DiskLinkTypePciE
)

// DiskStatusType base type for bitfield
type DiskStatusType uint64

// These constants are bitfields, eg. more than one bit can be set at the same time.
const (
	// DiskStatusUnknown Plugin failed to query out the status of disk.
	DiskStatusUnknown DiskStatusType = 1 << iota

	// DiskStatusOk Disk is up and healthy.
	DiskStatusOk

	//DiskStatusOther Vendor specific status.
	DiskStatusOther

	//DiskStatusPredictiveFailure Disk is functional but will fail soon
	DiskStatusPredictiveFailure

	//DiskStatusError Disk is not functional
	DiskStatusError

	//DiskStatusRemoved Disk was removed by administrator
	DiskStatusRemoved

	// DiskStatusStarting Disk is in the process of becomming ready.
	DiskStatusStarting

	// DiskStatusStopping Disk is shutting down.
	DiskStatusStopping

	// DiskStatusStopped Disk is stopped by administrator.
	DiskStatusStopped

	// DiskStatusInitializing Disk is not yet functional, could be initializing eg. RAID, zeroed or scrubed etc.
	DiskStatusInitializing

	// DiskStatusMaintenanceMode In maintenance for bad sector scan, integrity check and etc
	DiskStatusMaintenanceMode

	// DiskStatusSpareDisk Disk is configured as a spare disk.
	DiskStatusSpareDisk

	// DiskStatusReconstruct Disk is reconstructing its data.
	DiskStatusReconstruct

	// DiskStatusFree Disk is not holding any data and it not designated as a spare.
	DiskStatusFree
)

// FileSystem represents a file systems information
type FileSystem struct {
	Class      string  `json:"class"`
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	TotalSpace uint64  `json:"total_space"`
	FreeSpace  uint64  `json:"free_space"`
	PluginData *string `json:"plugin_data"`
	SystemID   string  `json:"system_id"`
	PoolID     string  `json:"pool_id"`
}

// NfsExport represents exported file systems over NFS.
type NfsExport struct {
	Class      string   `json:"class"`
	ID         string   `json:"id"`
	FsID       string   `json:"fs_id"`
	ExportPath string   `json:"export_path"`
	Auth       string   `json:"auth"`
	Root       []string `json:"root"`
	Rw         []string `json:"rw"`
	Ro         []string `json:"ro"`
	AnonUID    int64    `json:"anonuid"`
	AnonGID    int64    `json:"anongid"`
	Options    string   `json:"options"`
	PluginData *string  `json:"plugin_data"`
}

// NfsAccess argument for exporting a filesystem
type NfsAccess struct {
	Root    []string
	Rw      []string
	Ro      []string
	AnonUID int64
	AnonGID int64
}

// AnonUIDGIDNotApplicable use when anonUID or anonGID not applicable for NfsAccess.
const AnonUIDGIDNotApplicable int64 = -1

// AccessGroup represents a collection of initiators.
type AccessGroup struct {
	Class         string        `json:"class"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	InitIDs       []string      `json:"init_ids"`
	InitiatorType InitiatorType `json:"init_type"`
	PluginData    *string       `json:"plugin_data"`
	SystemID      string        `json:"system_id"`
}

// InitiatorType is enumerated type of initiators
type InitiatorType int

const (
	// InitiatorTypeUnknown plugin failed to query initiator type
	InitiatorTypeUnknown InitiatorType = iota

	// InitiatorTypeOther vendor specific initiator type
	InitiatorTypeOther

	// InitiatorTypeWwpn FC or FCoE WWPN
	InitiatorTypeWwpn

	// Reserved "3"
	_
	// Reserved "4"
	_

	// InitiatorTypeIscsiIqn iSCSI IQN
	InitiatorTypeIscsiIqn

	// Reserved "6"
	_

	// InitiatorTypeMixed this access group contains more than 1 type of initiator
	InitiatorTypeMixed
)

// TargetPort represents information about target ports.
type TargetPort struct {
	Class           string   `json:"class"`
	ID              string   `json:"id"`
	PortType        PortType `json:"port_type"`
	ServiceAddress  string   `json:"service_address"`
	NetworkAddress  string   `json:"network_address"`
	PhysicalAddress string   `json:"physical_address"`
	PhysicalName    string   `json:"physical_name"`
	PluginData      *string  `json:"plugin_data"`
	SystemID        string   `json:"system_id"`
}

// PortType in enumerated type of port
type PortType int32

const (

	// PortTypeOther is a vendor specific port type
	PortTypeOther PortType = 1 + iota

	// PortTypeFc indicates FC port type
	PortTypeFc

	// PortTypeFCoE indicates FC over Ethernet type
	PortTypeFCoE

	// PortTypeIscsi indicates FC over iSCSI type
	PortTypeIscsi
)

// Battery represents a battery in the system.
type Battery struct {
	Class       string        `json:"class"`
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	BatteryType BatteryType   `json:"type"`
	PluginData  *string       `json:"plugin_data"`
	Status      BatteryStatus `json:"status"`
	SystemID    string        `json:"system_id"`
}

// BatteryType indicates enumerated type of battery
type BatteryType int32

const (
	// BatteryTypeUnknown plugin failed to detect battery type
	BatteryTypeUnknown BatteryType = 1 + iota

	// BatteryTypeOther vendor specific battery type
	BatteryTypeOther

	// BatteryTypeChemical indicates li-ion etc.
	BatteryTypeChemical

	// BatteryTypeCapacitor indicates capacitor
	BatteryTypeCapacitor
)

// BatteryStatus indicates bitfield for status of battery
type BatteryStatus uint64

const (
	// BatteryStatusUnknown plugin failed to query battery status
	BatteryStatusUnknown BatteryStatus = 1 << iota

	// BatteryStatusOther vendor specific status
	BatteryStatusOther

	// BatteryStatusOk indicated battery is healthy and operational
	BatteryStatusOk

	// BatteryStatusDischarging indicates battery is discharging
	BatteryStatusDischarging

	// BatteryStatusCharging battery is charging
	BatteryStatusCharging

	// BatteryStatusLearning indicated battery system is optimizing battery use
	BatteryStatusLearning

	// BatteryStatusDegraded indicated battery should be checked and/or replaced
	BatteryStatusDegraded

	// BatteryStatusError indicates battery is in bad state
	BatteryStatusError
)

// Capabilities representation
type Capabilities struct {
	Class string `json:"class"`
	Cap   string `json:"cap"`
}

// IsSupported used to determine if a capability is supported
func (c *Capabilities) IsSupported(cap CapabilityType) bool {
	var capIdx = int32(cap) * 2
	if c.Cap[capIdx:capIdx+2] == "01" {
		return true
	}
	return false
}

// IsSupportedSet is used to determine if 1 or more capabilities
// are supported.
func (c *Capabilities) IsSupportedSet(cap []CapabilityType) bool {
	for _, i := range cap {
		if !c.IsSupported(i) {
			return false
		}
	}
	return true
}

// CapabilityType Enumerated type for capabilities
type CapabilityType uint32

const (
	// CapVolumes supports retrieving Volumes
	CapVolumes CapabilityType = 20

	// CapVolumeCreate supports VolumeCreate
	CapVolumeCreate CapabilityType = 21

	// CapVolumeCResize supports VolumeResize
	CapVolumeCResize CapabilityType = 22

	// CapVolumeCReplicate supports VolumeReplicate
	CapVolumeCReplicate CapabilityType = 23

	// CapVolumeCReplicateClone supports volume replication via clone
	CapVolumeCReplicateClone CapabilityType = 24

	// CapVolumeCReplicateCopy supports volume replication via copy
	CapVolumeCReplicateCopy CapabilityType = 25

	// CapVolumeCReplicateMirrorAsync supports volume replication via async. mirror
	CapVolumeCReplicateMirrorAsync CapabilityType = 26

	// CapVolumeCReplicateMirrorSync supports volume replication via sync. mirror
	CapVolumeCReplicateMirrorSync CapabilityType = 27

	// CapVolumeCopyRangeBlockSize supports reporting of what block size to be used in Copy Range
	CapVolumeCopyRangeBlockSize CapabilityType = 28

	// CapVolumeCopyRange supports copying a range of a Volume
	CapVolumeCopyRange CapabilityType = 29

	// CapVolumeCopyRangeClone supports a range clone of a volume
	CapVolumeCopyRangeClone CapabilityType = 30

	// CapVolumeCopyRangeCopy supports a range copy of a volume
	CapVolumeCopyRangeCopy CapabilityType = 31

	// CapVolumeDelete supports volume deletion
	CapVolumeDelete CapabilityType = 33

	// CapVolumeEnable admin. volume enable
	CapVolumeEnable CapabilityType = 34

	// CapVolumeDisable admin. volume disable
	CapVolumeDisable CapabilityType = 35

	// CapVolumeMask support volume masking
	CapVolumeMask CapabilityType = 36

	// CapVolumeUnmask support volume unmasking
	CapVolumeUnmask CapabilityType = 37

	// CapAccessGroups supports AccessGroups
	CapAccessGroups CapabilityType = 38

	// CapAccessGroupCreateWwpn supports access group wwpn creation
	CapAccessGroupCreateWwpn CapabilityType = 39

	// CapAccessGroupDelete delete an access group
	CapAccessGroupDelete CapabilityType = 40

	// CapAccessGroupInitiatorAddWwpn support adding WWPN to an access group
	CapAccessGroupInitiatorAddWwpn CapabilityType = 41

	// CapAccessGroupInitiatorDel supports removal of an initiator from access group
	CapAccessGroupInitiatorDel CapabilityType = 42

	// CapVolumesMaskedToAg supports listing of volumes masked to access group
	CapVolumesMaskedToAg CapabilityType = 43

	// CapAgsGrantedToVol list access groups with access to volume
	CapAgsGrantedToVol CapabilityType = 44

	// CapHasChildDep indicates support for determing if volume has child dep.
	CapHasChildDep CapabilityType = 45

	// CapChildDepRm indiates support for removing child dep.
	CapChildDepRm CapabilityType = 46

	// CapAccessGroupCreateIscsiIqn supports ag creating with iSCSI initiator
	CapAccessGroupCreateIscsiIqn CapabilityType = 47

	// CapAccessGroupInitAddIscsiIqn supports adding iSCSI initiator
	CapAccessGroupInitAddIscsiIqn CapabilityType = 48

	// CapIscsiChapAuthSet support iSCSI CHAP setting
	CapIscsiChapAuthSet CapabilityType = 53

	// CapVolRaidInfo supports RAID information
	CapVolRaidInfo CapabilityType = 54

	// CapVolumeThin supports creating thinly provisioned Volumes.
	CapVolumeThin CapabilityType = 55

	// CapBatteries supports Batteries Call
	CapBatteries CapabilityType = 56

	// CapVolCacheInfo supports CacheInfo
	CapVolCacheInfo CapabilityType = 57

	// CapVolPhyDiskCacheSet support VolPhyDiskCacheSet
	CapVolPhyDiskCacheSet CapabilityType = 58

	// CapVolPhysicalDiskCacheSetSystemLevel supports VolPhyDiskCacheSet
	CapVolPhysicalDiskCacheSetSystemLevel CapabilityType = 59

	// CapVolWriteCacheSetEnable supports VolWriteCacheSet
	CapVolWriteCacheSetEnable CapabilityType = 60

	// CapVolWriteCacheSetAuto supports VolWriteCacheSet
	CapVolWriteCacheSetAuto CapabilityType = 61

	// CapVolWriteCacheSetDisabled supported VolWriteCacheSet
	CapVolWriteCacheSetDisabled CapabilityType = 62

	// CapVolWriteCacheSetImpactRead indicates the VolWriteCacheSet might also
	// impact read cache policy.
	CapVolWriteCacheSetImpactRead CapabilityType = 63

	// CapVolWriteCacheSetWbImpactOther Indicate the VolWriteCacheSet with
	// `wcp=Cache::Enabled` might impact other volumes in the same
	// system.
	CapVolWriteCacheSetWbImpactOther CapabilityType = 64

	// CapVolReadCacheSet Support VolReadCacheSet()
	CapVolReadCacheSet CapabilityType = 65

	// VolReadCacheSetImpactWrite Indicates the VolReadCacheSet might
	// also impact write cache policy.
	VolReadCacheSetImpactWrite CapabilityType = 66

	// CapFs support Fs listing.
	CapFs CapabilityType = 100

	// CapFsDelete supports  FsDelete
	CapFsDelete CapabilityType = 101

	// CapFsResize Support FsResize
	CapFsResize CapabilityType = 102

	// CapFsCreate support FsCreate
	CapFsCreate CapabilityType = 103

	// CapFsClone support FsClone
	CapFsClone CapabilityType = 104

	// CapFsFileClone support FsFileClone
	CapFsFileClone CapabilityType = 105

	// CapFsSnapshots support FsSnapshots
	CapFsSnapshots CapabilityType = 106

	// CapFsSnapshotCreate support FsSnapshotCreate
	CapFsSnapshotCreate CapabilityType = 107

	// CapFsSnapshotDelete support FfsSnapshotDelete
	CapFsSnapshotDelete CapabilityType = 109

	// CapFsSnapshotRestore support FsSnapshotRestore
	CapFsSnapshotRestore CapabilityType = 110

	// CapFsSnapshotRestoreSpecificFiles support FsSnapshotRestore with `files` argument.
	CapFsSnapshotRestoreSpecificFiles CapabilityType = 111

	// CapFsHasChildDep support FsHasChildDep
	CapFsHasChildDep CapabilityType = 112

	// CapFsChildDepRm support FsChildDepRm
	CapFsChildDepRm CapabilityType = 113

	// CapFsChildDepRmSpecificFiles support FsChildDepRm with `files` argument.
	CapFsChildDepRmSpecificFiles CapabilityType = 114

	// CapNfsExportAuthTypeList support NfsExpAuthTypeList
	CapNfsExportAuthTypeList CapabilityType = 120

	// CapNfsExports support NfsExports
	CapNfsExports CapabilityType = 121

	// CapFsExport support FsExport
	CapFsExport CapabilityType = 122

	// CapFsUnexport support FsUnexport
	CapFsUnexport CapabilityType = 123

	// CapFsExportCustomPath support FsExport
	CapFsExportCustomPath CapabilityType = 124

	// CapSysReadCachePctSet support SystemReadCachePctSet
	CapSysReadCachePctSet CapabilityType = 158

	// CapSysReadCachePctGet support Systems() `ReadCachePct` property
	CapSysReadCachePctGet CapabilityType = 159

	// CapSysFwVersionGet support Systems()  with valid `FwVersion` property.
	CapSysFwVersionGet CapabilityType = 160

	// CapSysModeGet support `Systems()` with valid `mode` property.
	CapSysModeGet CapabilityType = 161

	// CapDiskLocation support Disks with valid `Location` property.
	CapDiskLocation CapabilityType = 163

	// CapDiskRpm support `Disks()` with valid `Rpm` property.
	CapDiskRpm CapabilityType = 164

	// CapDiskLinkType support `Disks()` with valid `LinkType` property.
	CapDiskLinkType CapabilityType = 165

	// CapVolumeLed support `VolIdentLedOn()` and `VolIdentLedOff()`.
	CapVolumeLed CapabilityType = 171

	// CapTargetPorts Support TargetPorts()
	CapTargetPorts CapabilityType = 216

	// CapDisks support Disks()
	CapDisks CapabilityType = 220

	// CapPoolMemberInfo support `PoolMemberInfo()`
	CapPoolMemberInfo CapabilityType = 221

	// CapVolumeRaidCreate support `VolRaidCreateCapGet()` and
	// VolRaidCreate().
	CapVolumeRaidCreate CapabilityType = 222

	//CapDiskVpd83Get support `Disks()` with valid `Vpd83` property.
	CapDiskVpd83Get CapabilityType = 223
)

// BlockRange defines a source block, destination block and number of blocks
type BlockRange struct {
	Class      string `json:"class"`
	SrcBlkAddr uint64 `json:"src_block"`
	DstBlkAddr uint64 `json:"dest_block"`
	BlkCount   uint64 `json:"block_count"`
}

// FileSystemSnapShot defines information relating to a file system snapshot
type FileSystemSnapShot struct {
	Class      string  `json:"class"`
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Ts         uint64  `json:"ts"`
	PluginData *string `json:"plugin_data"`
}

// RaidType what type of RAID
type RaidType int

const (
	// RaidUnknown Plugin failed to detect RAID type.
	RaidUnknown RaidType = -1

	// Raid0 https://en.wikipedia.org/wiki/Standard_RAID_levels#RAID_0
	Raid0 RaidType = 0

	// Raid1 is disk mirror
	Raid1 RaidType = 1

	// Raid3 is byte level striping with dedicated parity
	Raid3 RaidType = 3

	// Raid4 Block-level striping with dedicated parity.
	Raid4 RaidType = 4

	// Raid5 Block-level striping with distributed parity.
	Raid5 RaidType = 5

	// Raid6 Block-level striping with two distributed parities. Also known as
	// RAID-DP.
	Raid6 RaidType = 6

	// Raid10 Stripe of mirrors.
	Raid10 RaidType = 10

	// Raid15 Parity of mirrors.
	Raid15 RaidType = 15

	// Raid16 Dual parity of mirrors.
	Raid16 RaidType = 16

	// Raid50 Stripe of parities.
	Raid50 RaidType = 50

	// Raid60 Stripe of dual parities.
	Raid60 RaidType = 60

	// Raid51 Mirror of parities.
	Raid51 RaidType = 51

	// Raid61 Mirror of dual parities.
	Raid61 RaidType = 61

	// RaidJbod Just bunch of disks, no parity, no striping.
	RaidJbod RaidType = 20

	// RaidMixed volume contains multiple RAID settings.
	RaidMixed RaidType = 21

	// RaidOther Vendor specific RAID type
	RaidOther RaidType = 22
)

// VolumeRaidInfo information about RAID
type VolumeRaidInfo struct {
	Type      RaidType
	StripSize uint32
	DiskCount uint32
	MinIOSize uint32
	OptIOSize uint32
}

// MemberType represents unique type for pool member type
type MemberType int

const (
	// MemberTypeUnknown plugin failed to detect the RAID member type.
	MemberTypeUnknown MemberType = iota

	// MemberTypeOther vendor specific RAID member type.
	MemberTypeOther

	// MemberTypeDisk pool is created from RAID group using whole disks.
	MemberTypeDisk

	// MemberTypePool pool is allocated from other pool.
	MemberTypePool
)

// PoolMemberInfo information about what a pool is composed from.
type PoolMemberInfo struct {
	Raid   RaidType
	Member MemberType
	ID     []string
}

// SupportedRaidCapability is types and stripe sizes RAID storage supports.
type SupportedRaidCapability struct {
	Types       []RaidType
	StripeSizes []uint32
}

// WriteCachePolicy represents write cache policy type
type WriteCachePolicy uint32

const (
	// WriteCachePolicyUnknown ...
	WriteCachePolicyUnknown WriteCachePolicy = 1 + iota

	// WriteCachePolicyWriteBack ...
	WriteCachePolicyWriteBack

	// WriteCachePolicyAuto ...
	WriteCachePolicyAuto

	// WriteCachePolicyWriteThrough ...
	WriteCachePolicyWriteThrough
)

// WriteCacheStatus represents write cache status type
type WriteCacheStatus uint32

const (
	// WriteCacheStatusUnknown ...
	WriteCacheStatusUnknown WriteCacheStatus = 1 + iota

	// WriteCacheStatusWriteBack ...
	WriteCacheStatusWriteBack

	// WriteCacheStatusWriteThrough ...
	WriteCacheStatusWriteThrough
)

// ReadCachePolicy represents read cache policy type
type ReadCachePolicy uint32

const (

	// ReadCachePolicyUnknown ...
	ReadCachePolicyUnknown ReadCachePolicy = 1 + iota

	// ReadCachePolicyEnabled ...
	ReadCachePolicyEnabled

	// ReadCachePolicyDisabled ...
	ReadCachePolicyDisabled
)

// ReadCacheStatus represents read cache status type
type ReadCacheStatus uint32

const (
	// ReadCacheStatusUnknown ...
	ReadCacheStatusUnknown ReadCacheStatus = 1 + iota

	// ReadCacheStatusEnabled ...
	ReadCacheStatusEnabled

	// ReadCacheStatusDisabled ...
	ReadCacheStatusDisabled
)

// PhysicalDiskCache represents physical disk caching type
type PhysicalDiskCache uint32

const (
	// PhysicalDiskCacheUnknown ...
	PhysicalDiskCacheUnknown PhysicalDiskCache = 1 + iota

	// PhysicalDiskCacheEnabled ...
	PhysicalDiskCacheEnabled

	// PhysicalDiskCacheDisabled ...
	PhysicalDiskCacheDisabled

	// PhysicalDiskCacheUseDiskSetting ...
	PhysicalDiskCacheUseDiskSetting
)

// VolumeCacheInfo contains information about volume caching values
type VolumeCacheInfo struct {
	WriteSetting       WriteCachePolicy
	WriteStatus        WriteCacheStatus
	ReadSetting        ReadCachePolicy
	ReadStatus         ReadCacheStatus
	PhysicalDiskStatus PhysicalDiskCache
}

// DiskHealthStatus health status of physical disk
type DiskHealthStatus int

const (

	// DiskHealthStatusUnknown represents unknown health status
	DiskHealthStatusUnknown DiskHealthStatus = -1 + iota

	// DiskHealthStatusFail represents fail health status
	DiskHealthStatusFail

	// DiskHealthStatusWarn represents health warning status
	DiskHealthStatusWarn

	// DiskHealthStatusGood represent good health status
	DiskHealthStatusGood
)

// DiskLedStatusBitField Bit field for disk LED status indicators
type DiskLedStatusBitField uint32

const (
	// DiskLedStatusUnknown unknown
	DiskLedStatusUnknown DiskLedStatusBitField = 1 << iota

	// DiskLedStatusIdentOn ident LED is on
	DiskLedStatusIdentOn

	// DiskLedStatusIdentOff ident LED is off
	DiskLedStatusIdentOff

	// DiskLedStatusIdentUnknown ident is unknown
	DiskLedStatusIdentUnknown

	// DiskLedStatusFaultOn  fault LED is on
	DiskLedStatusFaultOn

	// DiskLedStatusFaultOff fault LED is off
	DiskLedStatusFaultOff

	// DiskLedStatusFaultUnknown fault LED is unknown
	DiskLedStatusFaultUnknown
)
