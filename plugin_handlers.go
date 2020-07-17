// SPDX-License-Identifier: 0BSD

package libstoragemgmt

import (
	"encoding/json"
	"fmt"
	"reflect"

	errors "github.com/libstorage/libstoragemgmt-golang/errors"
)

func invalidArgs(msg string, e error) error {
	return &errors.LsmError{
		Code:    errors.TransPortInvalidArg,
		Message: fmt.Sprintf("%s: invalid arguments(s) %s\n", msg, e)}
}

func handleRegister(p *Plugin, msg *requestMsg) (interface{}, error) {

	var register PluginRegister
	if uE := json.Unmarshal(msg.Params, &register); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return nil, p.cb.Mgmt.PluginRegister(&register)
}

func handleUnRegister(p *Plugin, msg *requestMsg) (interface{}, error) {
	return nil, p.cb.Mgmt.PluginUnregister()
}

type tmoSetArgs struct {
	MS    uint32 `json:"ms"`
	Flags uint64 `json:"flags"`
}

func handlePluginInfo(p *Plugin, msg *requestMsg) (interface{}, error) {
	return []string{p.desc, p.ver}, nil
}

func handleTmoSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	var timeout tmoSetArgs
	if uE := json.Unmarshal(msg.Params, &timeout); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return nil, p.cb.Mgmt.TimeOutSet(timeout.MS)
}

func handleTmoGet(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.Mgmt.TimeOutGet(), nil
}

func handleSystems(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.Mgmt.Systems()
}

func handleDisks(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.San.Disks()
}

type search struct {
	Key   string `json:"search_key"`
	Value string `json:"search_value"`
	Flags uint64 `json:"flags"`
}

func handlePools(p *Plugin, msg *requestMsg) (interface{}, error) {
	var s search
	if uE := json.Unmarshal(msg.Params, &s); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	if len(s.Key) > 0 {
		return p.cb.Mgmt.Pools(s.Key, s.Value)
	}

	return p.cb.Mgmt.Pools()
}

func handleVolumes(p *Plugin, msg *requestMsg) (interface{}, error) {
	var s search
	if uE := json.Unmarshal(msg.Params, &s); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	if len(s.Key) > 0 {
		return p.cb.San.Volumes(s.Key, s.Value)
	}

	return p.cb.San.Volumes()
}

type capArgs struct {
	Sys System `json:"system"`
}

func handleCapabilities(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args capArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return p.cb.Mgmt.Capabilities(&args.Sys)
}

type jobArgs struct {
	ID string `json:"job_id"`
}

func handleJobStatus(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args jobArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	job, err := p.cb.Mgmt.JobStatus(args.ID)
	if err != nil {
		return nil, err
	}

	var result [3]interface{}
	result[0] = job.Status
	result[1] = job.Percent
	result[2] = job.Item

	return result, nil
}

func handleJobFree(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args jobArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Mgmt.JobFree(args.ID)
}

func exclusiveOr(item interface{}, job *string, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}

	var result [2]interface{}

	if job != nil {
		result[0] = job
		result[1] = nil
	} else {
		result[0] = nil
		result[1] = item
	}
	return result, nil
}

func handleVolumeCreate(p *Plugin, msg *requestMsg) (interface{}, error) {

	type volumeCreateArgs struct {
		Pool         *Pool               `json:"pool"`
		Name         string              `json:"volume_name"`
		SizeBytes    uint64              `json:"size_bytes"`
		Provisioning VolumeProvisionType `json:"provisioning"`
		Flags        uint64              `json:"flags"`
	}

	var args volumeCreateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	volume, jobID, error := p.cb.San.VolumeCreate(args.Pool, args.Name, args.SizeBytes, args.Provisioning)
	return exclusiveOr(volume, jobID, error)
}

func handleVolumeReplicate(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volumeReplicateArgs struct {
		Pool    *Pool               `json:"pool"`
		RepType VolumeReplicateType `json:"rep_type"`
		Flags   uint64              `json:"flags"`
		SrcVol  Volume              `json:"volume_src"`
		Name    string              `json:"name"`
	}

	var args volumeReplicateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	volume, jobID, error := p.cb.San.VolumeReplicate(args.Pool, args.RepType, &args.SrcVol, args.Name)
	return exclusiveOr(volume, jobID, error)
}

func handleVolumeReplicateRange(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volumeReplicateRangeArgs struct {
		RepType VolumeReplicateType `json:"rep_type"`
		Ranges  []BlockRange        `json:"ranges"`
		SrcVol  Volume              `json:"volume_src"`
		DstVol  Volume              `json:"volume_dest"`
		Flags   uint64              `json:"flags"`
	}

	var a volumeReplicateRangeArgs
	if uE := json.Unmarshal(msg.Params, &a); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.VolumeReplicateRange(a.RepType, &a.SrcVol, &a.DstVol, a.Ranges)
}

func handleVolRepRangeBlockSize(p *Plugin, msg *requestMsg) (interface{}, error) {
	type args struct {
		System *System `json:"system"`
		Flags  uint64  `json:"flags"`
	}

	var a args
	if uE := json.Unmarshal(msg.Params, &a); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return p.cb.San.VolumeRepRangeBlkSize(a.System)
}

func handleVolumeResize(p *Plugin, msg *requestMsg) (interface{}, error) {
	type args struct {
		Volume *Volume `json:"volume"`
		Size   uint64  `json:"new_size_bytes"`
		Flags  uint64  `json:"flags"`
	}

	var a args
	if uE := json.Unmarshal(msg.Params, &a); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	fmt.Printf("args = %+v\n", a)

	volume, jobID, error := p.cb.San.VolumeResize(a.Volume, a.Size)
	return exclusiveOr(volume, jobID, error)
}

type volumeArgument struct {
	Volume *Volume `json:"volume"`
	Flags  uint64  `json:"flags"`
}

func handleVolumeEnable(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArgument
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.VolumeEnable(args.Volume)
}

func handleVolumeDisable(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArgument
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.VolumeDisable(args.Volume)
}

func handleVolumeDelete(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volumeDeleteArgs struct {
		Volume *Volume `json:"volume"`
		Flags  uint64  `json:"flags"`
	}

	var args volumeDeleteArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.VolumeDelete(args.Volume)
}

type maskArgs struct {
	Vol   *Volume      `json:"volume"`
	Ag    *AccessGroup `json:"access_group"`
	Flags uint64       `json:"flags"`
}

func handleVolumeMask(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args maskArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.VolumeMask(args.Vol, args.Ag)
}

func handleVolumeUnMask(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args maskArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.VolumeUnMask(args.Vol, args.Ag)
}

func handleVolsMaskedToAg(p *Plugin, msg *requestMsg) (interface{}, error) {
	type argsAg struct {
		Ag    *AccessGroup `json:"access_group"`
		Flags uint64       `json:"flags"`
	}

	var args argsAg
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.VolsMaskedToAg(args.Ag)
}

func handleAccessGroups(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.San.AccessGroups()
}

func handleAccessGroupCreate(p *Plugin, msg *requestMsg) (interface{}, error) {
	type agCreateArgs struct {
		Name     string        `json:"name"`
		InitID   string        `json:"init_id"`
		InitType InitiatorType `json:"init_type"`
		System   *System       `json:"system"`
		Flags    uint64        `json:"flags"`
	}

	var args agCreateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.AccessGroupCreate(args.Name, args.InitID, args.InitType, args.System)
}

func handleAccessGroupDelete(p *Plugin, msg *requestMsg) (interface{}, error) {
	type agDeleteArgs struct {
		Ag    *AccessGroup `json:"access_group"`
		Flags uint64       `json:"flags"`
	}

	var args agDeleteArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.AccessGroupDelete(args.Ag)
}

type accessGroupInitArgs struct {
	Ag       *AccessGroup  `json:"access_group"`
	ID       string        `json:"init_id"`
	InitType InitiatorType `json:"init_type"`
	Flags    uint64        `json:"flags"`
}

func handleAccessGroupInitAdd(p *Plugin, msg *requestMsg) (interface{}, error) {

	var args accessGroupInitArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.AccessGroupInitAdd(args.Ag, args.ID, args.InitType)
}

func handleAccessGroupInitDelete(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args accessGroupInitArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.AccessGroupInitDelete(args.Ag, args.ID, args.InitType)
}

func handleAgsGrantedToVol(p *Plugin, msg *requestMsg) (interface{}, error) {
	type argsVol struct {
		Vol   *Volume `json:"volume"`
		Flags uint64  `json:"flags"`
	}

	var args argsVol
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.AgsGrantedToVol(args.Vol)
}

func handleIscsiChapAuthSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type argsIscsi struct {
		InitID      string  `json:"init_id"`
		InUser      *string `json:"in_user"`
		InPassword  *string `json:"in_password"`
		OutUser     *string `json:"out_user"`
		OutPassword *string `json:"out_password"`
		Flags       uint64  `json:"flags"`
	}

	var args argsIscsi
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.San.IscsiChapAuthSet(args.InitID, args.InUser, args.InPassword, args.OutUser, args.OutPassword)
}

type volumeArg struct {
	Vol   *Volume `json:"volume"`
	Flags uint64  `json:"flags"`
}

func handleVolHasChildDep(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArg
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.VolHasChildDep(args.Vol)
}

func handleVolChildDepRm(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArg
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.San.VolChildDepRm(args.Vol)
}

func handleTargetPorts(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.San.TargetPorts()
}

func handleVolIdentLedOn(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArg
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return nil, p.cb.San.VolIdentLedOn(args.Vol)
}

func handleVolIdentLedOff(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args volumeArg
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return nil, p.cb.San.VolIdentLedOff(args.Vol)
}

func handleFs(p *Plugin, msg *requestMsg) (interface{}, error) {
	var s search
	if uE := json.Unmarshal(msg.Params, &s); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	if len(s.Key) > 0 {
		return p.cb.File.FileSystems(s.Key, s.Value)
	}
	return p.cb.File.FileSystems()
}

func handleFsCreate(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsCreateArgs struct {
		Pool      *Pool  `json:"pool"`
		Name      string `json:"name"`
		SizeBytes uint64 `json:"size_bytes"`
		Flags     uint64 `json:"flags"`
	}

	var args fsCreateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	fs, jobID, error := p.cb.File.FsCreate(args.Pool, args.Name, args.SizeBytes)
	return exclusiveOr(fs, jobID, error)
}

func handleFsDelete(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsDeleteArgs struct {
		Fs    *FileSystem `json:"fs"`
		Flags uint64      `json:"flags"`
	}

	var args fsDeleteArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}
	return p.cb.File.FsDelete(args.Fs)
}

func handleFsResize(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsResizeArgs struct {
		Fs    *FileSystem `json:"fs"`
		Size  uint64      `json:"new_size_bytes"`
		Flags uint64      `json:"flags"`
	}

	var args fsResizeArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	fs, job, err := p.cb.File.FsResize(args.Fs, args.Size)
	return exclusiveOr(fs, job, err)

}

func handleFsClone(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsCloneArgs struct {
		Fs    *FileSystem         `json:"src_fs"`
		Name  string              `json:"dest_fs_name"`
		Ss    *FileSystemSnapShot `json:"snapshot"`
		Flags uint64              `json:"flags"`
	}

	var args fsCloneArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	fs, job, err := p.cb.File.FsClone(args.Fs, args.Name, args.Ss)
	return exclusiveOr(fs, job, err)

}

func handleFsFileClone(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsFileCloneArgs struct {
		Fs    *FileSystem         `json:"fs"`
		Src   string              `json:"src_file_name"`
		Dst   string              `json:"dest_file_name"`
		Ss    *FileSystemSnapShot `json:"snapshot"`
		Flags uint64              `json:"flags"`
	}

	var args fsFileCloneArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsFileClone(args.Fs, args.Src, args.Dst, args.Ss)
}

func handleFsSnapShotCreate(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsSnapShotCreateArgs struct {
		Fs    *FileSystem `json:"fs"`
		Name  string      `json:"snapshot_name"`
		Flags uint64      `json:"flags"`
	}

	var args fsSnapShotCreateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	fs, job, err := p.cb.File.FsSnapShotCreate(args.Fs, args.Name)
	return exclusiveOr(fs, job, err)
}

func handleFsSnapShotDelete(p *Plugin, msg *requestMsg) (interface{}, error) {
	type fsSnapShotDeleteArgs struct {
		Fs    *FileSystem         `json:"fs"`
		Ss    *FileSystemSnapShot `json:"snapshot"`
		Flags uint64              `json:"flags"`
	}

	var args fsSnapShotDeleteArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsSnapShotDelete(args.Fs, args.Ss)
}

func handleFsSnapShots(p *Plugin, msg *requestMsg) (interface{}, error) {

	type fsSnapShotArgs struct {
		Fs    *FileSystem `json:"fs"`
		Flags uint64      `json:"flags"`
	}

	var args fsSnapShotArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsSnapShots(args.Fs)
}

func handleFsSnapShotRestore(p *Plugin, msg *requestMsg) (interface{}, error) {

	type fsSnapShotRestoreArgs struct {
		Fs           *FileSystem         `json:"fs"`
		Ss           *FileSystemSnapShot `json:"snapshot"`
		All          bool                `json:"all_files"`
		Files        []string            `json:"files"`
		RestoreFiles []string            `json:"restore_files"`
		Flags        uint64              `json:"flags"`
	}

	var args fsSnapShotRestoreArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsSnapShotRestore(args.Fs, args.Ss, args.All, args.Files, args.RestoreFiles)
}

type fsHasChildDepsArgs struct {
	Fs    *FileSystem `json:"fs"`
	Files []string    `json:"files"`
	Flags uint64      `json:"flags"`
}

func handleFsHasChildDep(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args fsHasChildDepsArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsHasChildDep(args.Fs, args.Files)
}

func handleFsChildDepRm(p *Plugin, msg *requestMsg) (interface{}, error) {
	var args fsHasChildDepsArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.File.FsChildDepRm(args.Fs, args.Files)
}

func handleNfsExports(p *Plugin, msg *requestMsg) (interface{}, error) {
	var s search
	if uE := json.Unmarshal(msg.Params, &s); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	if len(s.Key) > 0 {
		return p.cb.Nfs.Exports(s.Key, s.Value)
	}
	return p.cb.Nfs.Exports()
}

func handleExportFs(p *Plugin, msg *requestMsg) (interface{}, error) {
	type exportArgs struct {
		FsID     *string  `json:"fs_id"`
		Path     *string  `json:"export_path"`
		Root     []string `json:"root_list"`
		Rw       []string `json:"rw_list"`
		Ro       []string `json:"ro_list"`
		AnonUID  int64    `json:"anon_uid"`
		AnonGID  int64    `json:"anon_gid"`
		AuthType *string  `json:"auth_type"`
		Options  *string  `json:"options"`
		Flags    uint64   `json:"flags"`
	}

	var a exportArgs
	if uE := json.Unmarshal(msg.Params, &a); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	// This seems like a blunder in the original API or maybe the preferred way to do it.
	fs, err := p.cb.File.FileSystems("id", *a.FsID)
	if err != nil {
		return nil, err
	}
	if len(fs) != 1 {
		return nil, &errors.LsmError{
			Code:    errors.NotFoundFs,
			Message: fmt.Sprintf("file system with ID=%s not found %d!", *a.FsID, len(fs))}
	}

	access := NfsAccess{Root: a.Root, Rw: a.Rw, Ro: a.Ro, AnonUID: a.AnonUID, AnonGID: a.AnonGID}
	return p.cb.Nfs.FsExport(&fs[0], a.Path, &access, a.AuthType, a.Options)
}

func handleFsUnexport(p *Plugin, msg *requestMsg) (interface{}, error) {
	type unexportArgs struct {
		Export *NfsExport `json:"export"`
		Flags  uint64     `json:"flags"`
	}

	var args unexportArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Nfs.FsUnExport(args.Export)
}

func handleExportAuthTypes(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.Nfs.ExportAuthTypes()
}

func handleVolRaidCreate(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volRaidCreateArgs struct {
		Name      string   `json:"name"`
		Type      RaidType `json:"raid_type"`
		Disks     []Disk   `json:"disks"`
		StripSize uint32   `json:"strip_size"`
		Flags     uint64   `json:"flags"`
	}

	var args volRaidCreateArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return p.cb.Hba.VolRaidCreate(args.Name, args.Type, args.Disks, args.StripSize)
}

func handleVolRaidCreateCapGet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volRaidCreateCapGetArgs struct {
		Sys   *System `json:"system"`
		Flags uint64  `json:"flags"`
	}

	var args volRaidCreateCapGetArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	result, err := p.cb.Hba.VolRaidCreateCapGet(args.Sys)
	if err != nil {
		return nil, err
	}

	var rc [2]interface{}
	rc[0] = result.Types
	rc[1] = result.StripeSizes
	return rc, nil
}

func handlePoolMemberInfo(p *Plugin, msg *requestMsg) (interface{}, error) {
	type poolMemberInfoArgs struct {
		Pool  *Pool  `json:"pool"`
		Flags uint64 `json:"flags"`
	}

	var args poolMemberInfoArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	result, err := p.cb.Hba.PoolMemberInfo(args.Pool)
	if err != nil {
		return nil, err
	}

	var rc [3]interface{}
	rc[0] = result.Raid
	rc[1] = result.Member
	rc[2] = result.ID
	return rc, nil
}

func handleVolRaidInfo(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volRaidInfoArgs struct {
		Volume *Volume `json:"volume"`
		Flags  uint64  `json:"flags"`
	}

	var args volRaidInfoArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	result, err := p.cb.Hba.VolRaidInfo(args.Volume)
	if err != nil {
		return nil, err
	}

	var rc [5]int32
	rc[0] = int32(result.Type)
	rc[1] = int32(result.StripSize)
	rc[2] = int32(result.DiskCount)
	rc[3] = int32(result.MinIOSize)
	rc[4] = int32(result.OptIOSize)
	return rc, nil
}

func handleBatteries(p *Plugin, msg *requestMsg) (interface{}, error) {
	return p.cb.Hba.Batteries()
}

func handleSystemReadCachePctSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type sysReadCachePctArgs struct {
		System  *System `json:"system"`
		Percent uint32  `json:"read_pct"`
		Flags   uint64  `json:"flags"`
	}

	var args sysReadCachePctArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Cache.SysReadCachePctSet(args.System, args.Percent)
}

func handleVolCacheInfo(p *Plugin, msg *requestMsg) (interface{}, error) {

	type volCacheInfoArgs struct {
		Volume *Volume `json:"volume"`
		Flags  uint64  `json:"flags"`
	}

	var args volCacheInfoArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	info, err := p.cb.Cache.VolCacheInfo(args.Volume)
	if err != nil {
		return nil, err
	}

	var ret [5]uint32
	ret[0] = uint32(info.WriteSetting)
	ret[1] = uint32(info.WriteStatus)
	ret[2] = uint32(info.ReadSetting)
	ret[3] = uint32(info.ReadStatus)
	ret[4] = uint32(info.PhysicalDiskStatus)

	return ret, nil
}

func handleVolPhyDiskCacheSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volPhyDiskCacheSetArgs struct {
		Volume *Volume           `json:"volume"`
		Pdc    PhysicalDiskCache `json:"pdc"`
		Flags  uint64            `json:"flags"`
	}

	var args volPhyDiskCacheSetArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Cache.VolPhyDiskCacheSet(args.Volume, args.Pdc)
}

func handleVolWriteCacheSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volWriteCacheSetArgs struct {
		Volume *Volume          `json:"volume"`
		Wcp    WriteCachePolicy `json:"wcp"`
		Flags  uint64           `json:"flags"`
	}

	var args volWriteCacheSetArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Cache.VolWriteCacheSet(args.Volume, args.Wcp)
}

func handleVolReadCacheSet(p *Plugin, msg *requestMsg) (interface{}, error) {
	type volReadCacheSetArgs struct {
		Volume *Volume         `json:"volume"`
		Rcp    ReadCachePolicy `json:"rcp"`
		Flags  uint64          `json:"flags"`
	}

	var args volReadCacheSetArgs
	if uE := json.Unmarshal(msg.Params, &args); uE != nil {
		return nil, invalidArgs(msg.Method, uE)
	}

	return nil, p.cb.Cache.VolReadCacheSet(args.Volume, args.Rcp)
}

func nilAssign(present interface{}, cb handler) handler {

	// This seems like an epic fail of golang as I got burned by doing present == nil
	// ref. https://groups.google.com/forum/#!topic/golang-nuts/wnH302gBa4I/discussion
	if present == nil || reflect.ValueOf(present).IsNil() {
		return nil
	}
	return cb
}

func buildTable(c *PluginCallBacks) map[string]handler {
	return map[string]handler{
		"plugin_info":       handlePluginInfo,
		"plugin_register":   nilAssign(c.Mgmt.PluginRegister, handleRegister),
		"plugin_unregister": nilAssign(c.Mgmt.PluginUnregister, handleUnRegister),
		"systems":           nilAssign(c.Mgmt.Systems, handleSystems),
		"capabilities":      nilAssign(c.Mgmt.Capabilities, handleCapabilities),
		"time_out_set":      nilAssign(c.Mgmt.TimeOutSet, handleTmoSet),
		"time_out_get":      nilAssign(c.Mgmt.TimeOutGet, handleTmoGet),
		"pools":             nilAssign(c.Mgmt.Pools, handlePools),
		"job_status":        nilAssign(c.Mgmt.JobStatus, handleJobStatus),
		"job_free":          nilAssign(c.Mgmt.JobFree, handleJobFree),

		"volume_create":                      nilAssign(c.San.VolumeCreate, handleVolumeCreate),
		"volume_delete":                      nilAssign(c.San.VolumeDelete, handleVolumeDelete),
		"volumes":                            nilAssign(c.San.Volumes, handleVolumes),
		"disks":                              nilAssign(c.San.Disks, handleDisks),
		"volume_replicate":                   nilAssign(c.San.VolumeReplicate, handleVolumeReplicate),
		"volume_replicate_range":             nilAssign(c.San.VolumeReplicateRange, handleVolumeReplicateRange),
		"volume_replicate_range_block_size":  nilAssign(c.San.VolumeRepRangeBlkSize, handleVolRepRangeBlockSize),
		"volume_resize":                      nilAssign(c.San.VolumeResize, handleVolumeResize),
		"volume_enable":                      nilAssign(c.San.VolumeEnable, handleVolumeEnable),
		"volume_disable":                     nilAssign(c.San.VolumeDisable, handleVolumeDisable),
		"volume_mask":                        nilAssign(c.San.VolumeMask, handleVolumeMask),
		"volume_unmask":                      nilAssign(c.San.VolumeUnMask, handleVolumeUnMask),
		"volume_child_dependency":            nilAssign(c.San.VolHasChildDep, handleVolHasChildDep),
		"volume_child_dependency_rm":         nilAssign(c.San.VolChildDepRm, handleVolChildDepRm),
		"volumes_accessible_by_access_group": nilAssign(c.San.VolsMaskedToAg, handleVolsMaskedToAg),
		"access_groups":                      nilAssign(c.San.AccessGroups, handleAccessGroups),
		"access_group_create":                nilAssign(c.San.AccessGroupCreate, handleAccessGroupCreate),
		"access_group_delete":                nilAssign(c.San.AccessGroupDelete, handleAccessGroupDelete),
		"access_group_initiator_add":         nilAssign(c.San.AccessGroupInitAdd, handleAccessGroupInitAdd),
		"access_group_initiator_delete":      nilAssign(c.San.AccessGroupInitDelete, handleAccessGroupInitDelete),
		"access_groups_granted_to_volume":    nilAssign(c.San.AgsGrantedToVol, handleAgsGrantedToVol),
		"iscsi_chap_auth":                    nilAssign(c.San.IscsiChapAuthSet, handleIscsiChapAuthSet),
		"target_ports":                       nilAssign(c.San.TargetPorts, handleTargetPorts),
		"volume_ident_led_on":                nilAssign(c.San.VolIdentLedOn, handleVolIdentLedOn),
		"volume_ident_led_off":               nilAssign(c.San.VolIdentLedOn, handleVolIdentLedOff),

		"fs":                     nilAssign(c.File.FileSystems, handleFs),
		"fs_create":              nilAssign(c.File.FsCreate, handleFsCreate),
		"fs_delete":              nilAssign(c.File.FsDelete, handleFsDelete),
		"fs_resize":              nilAssign(c.File.FsResize, handleFsResize),
		"fs_clone":               nilAssign(c.File.FsClone, handleFsClone),
		"fs_file_clone":          nilAssign(c.File.FsFileClone, handleFsFileClone),
		"fs_snapshot_create":     nilAssign(c.File.FsSnapShotCreate, handleFsSnapShotCreate),
		"fs_snapshot_delete":     nilAssign(c.File.FsSnapShotDelete, handleFsSnapShotDelete),
		"fs_snapshots":           nilAssign(c.File.FsSnapShots, handleFsSnapShots),
		"fs_snapshot_restore":    nilAssign(c.File.FsSnapShotRestore, handleFsSnapShotRestore),
		"fs_child_dependency":    nilAssign(c.File.FsHasChildDep, handleFsHasChildDep),
		"fs_child_dependency_rm": nilAssign(c.File.FsChildDepRm, handleFsChildDepRm),

		"exports":       nilAssign(c.Nfs.Exports, handleNfsExports),
		"export_fs":     nilAssign(c.Nfs.FsExport, handleExportFs),
		"export_remove": nilAssign(c.Nfs.FsUnExport, handleFsUnexport),
		"export_auth":   nilAssign(c.Nfs.ExportAuthTypes, handleExportAuthTypes),

		"volume_raid_create":         nilAssign(c.Hba.VolRaidCreate, handleVolRaidCreate),
		"volume_raid_create_cap_get": nilAssign(c.Hba.VolRaidCreateCapGet, handleVolRaidCreateCapGet),
		"pool_member_info":           nilAssign(c.Hba.PoolMemberInfo, handlePoolMemberInfo),
		"volume_raid_info":           nilAssign(c.Hba.VolRaidInfo, handleVolRaidInfo),
		"batteries":                  nilAssign(c.Hba.Batteries, handleBatteries),

		"system_read_cache_pct_update":      nilAssign(c.Cache.SysReadCachePctSet, handleSystemReadCachePctSet),
		"volume_cache_info":                 nilAssign(c.Cache.VolCacheInfo, handleVolCacheInfo),
		"volume_physical_disk_cache_update": nilAssign(c.Cache.VolPhyDiskCacheSet, handleVolPhyDiskCacheSet),
		"volume_write_cache_policy_update":  nilAssign(c.Cache.VolWriteCacheSet, handleVolWriteCacheSet),
		"volume_read_cache_policy_update":   nilAssign(c.Cache.VolReadCacheSet, handleVolReadCacheSet),
	}
}
