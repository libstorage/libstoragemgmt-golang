// SPDX-License-Identifier: 0BSD

package libstoragemgmt

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	lsm "github.com/libstorage/libstoragemgmt-golang"
	errors "github.com/libstorage/libstoragemgmt-golang/errors"
	disks "github.com/libstorage/libstoragemgmt-golang/localdisk"
)

var URI = getEnv("LSM_GO_URI", "sim://")

const PASSWORD = ""
const TMO uint32 = 30000

// Running these tests requires lsmd up and running with the simulator
// plugin available.  In addition a number of tests require things to
// exists which don't exist by default and need to be created.  At the
// moment this needs to be done via lsmcli.  As functionality evolves
// these requirements will be reduced as the unit tests will create
// things as needed.

func rs(pre string, n int) string {
	var l = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		b[i] = l[rand.Intn(len(l))]
	}
	return fmt.Sprintf("%s%s", pre, string(b))
}

func getEnv(variable string, defValue string) string {
	var p = os.Getenv(variable)
	if len(p) > 0 {
		return p
	}
	return defValue
}

func TestConnect(t *testing.T) {
	var c, libError = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, libError)
	assert.Equal(t, nil, c.Close())
}

func TestConnectInvalidUri(t *testing.T) {
	var _, libError = lsm.Client("://", PASSWORD, TMO)
	assert.NotNil(t, libError)
}

func TestMissingPlugin(t *testing.T) {
	var _, libError = lsm.Client("nosuchthing://", PASSWORD, TMO)
	assert.NotNil(t, libError)
}

func TestBadUdsPath(t *testing.T) {
	const KEY = "LSM_UDS_PATH"
	var current = os.Getenv(KEY)

	os.Setenv(KEY, rs("/tmp/", 8))
	var _, libError = lsm.Client(URI, PASSWORD, TMO)
	assert.NotNil(t, libError)

	os.Setenv(KEY, current)
}

func TestPluginInfo(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)

	var pluginInfo, err = c.PluginInfo()
	assert.Nil(t, err)

	t.Logf("%+v", pluginInfo)

	assert.Equal(t, nil, c.Close())
}

func TestAvailablePlugins(t *testing.T) {

	var plugins, err = lsm.AvailablePlugins()
	assert.Nil(t, err)

	t.Logf("%+v", plugins)
}

func TestJobs(t *testing.T) {
	var c, libError = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, libError)

	var pools, pE = c.Pools()
	assert.Nil(t, pE)

	var name = rs("lsm_go_vol_", 12)
	var volume lsm.Volume
	var jobID, vcE = c.VolumeCreate(&pools[0],
		name, 1024*1024*100, lsm.VolumeProvisionTypeDefault, false, &volume)
	assert.Nil(t, vcE)
	assert.NotNil(t, jobID)

	// Supply a bad job id
	var status, percent, err = c.JobStatus("bogus", &volume)
	assert.True(t, (percent >= 0 && percent <= 100))
	assert.NotNil(t, err)
	assert.Equal(t, lsm.JobStatusError, status)

	// Poll for completion using actual jobID
	for true {
		status, percent, err = c.JobStatus(*jobID, &volume)
		assert.True(t, (percent >= 0 && percent <= 100))
		assert.Nil(t, err)
		assert.True(t, status == lsm.JobStatusInprogress || status == lsm.JobStatusComplete)
		if status != lsm.JobStatusInprogress {
			break
		}
	}

	assert.Equal(t, nil, c.Close())
}

func TestAvailablePluginsBadUds(t *testing.T) {
	const KEY = "LSM_UDS_PATH"
	var current = os.Getenv(KEY)
	os.Setenv(KEY, rs("/tmp/", 8))

	var plugins, err = lsm.AvailablePlugins()
	assert.NotNil(t, err)
	assert.NotNil(t, plugins)
	assert.Equal(t, 0, len(plugins))

	t.Logf("%+v", plugins)
	os.Setenv(KEY, current)
}

func TestBadSeach(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)

	var _, sE = c.Volumes("what")
	assert.NotNil(t, sE)

	_, sE = c.NfsExports("what")
	assert.NotNil(t, sE)

	_, sE = c.Pools("what")
	assert.NotNil(t, sE)

	assert.Equal(t, nil, c.Close())
}

func TestGoodSeach(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)

	var volumes, sE = c.Volumes("system_id", "sim-01")
	assert.Nil(t, sE)
	assert.NotNil(t, volumes)
	assert.Greater(t, len(volumes), 0)

	assert.Equal(t, nil, c.Close())
}

func TestSystems(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var systems, sysError = c.Systems()

	assert.Nil(t, sysError)
	assert.Equal(t, 1, len(systems))

	for _, s := range systems {
		t.Logf("%+v", s)
	}
	assert.Equal(t, nil, c.Close())
}

func TestReadCachePercentSet(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var systems, sysError = c.Systems()
	assert.Nil(t, sysError)

	assert.Nil(t, c.SysReadCachePctSet(&systems[0], 0))
	assert.Nil(t, c.SysReadCachePctSet(&systems[0], 100))

	var expectedErr = c.SysReadCachePctSet(&systems[0], 101)
	assert.NotNil(t, expectedErr)
	var e = expectedErr.(*errors.LsmError)
	assert.Equal(t, errors.InvalidArgument, e.Code)

	assert.Equal(t, nil, c.Close())
}

func TestIscsiChapSet(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)

	var init = "iqn.1994-05.com.domain:01.89bd01"
	var e = c.IscsiChapAuthSet(init, nil, nil, nil, nil)

	var u = rs("user_", 3)
	var p = rs("password_", 3)
	e = c.IscsiChapAuthSet(init, &u, &p, nil, nil)
	assert.Nil(t, e)

	var outU = rs("outuser_", 3)
	var outP = rs("out_password_", 3)
	e = c.IscsiChapAuthSet(init, &u, &p, &outU, &outP)
	assert.Nil(t, e)

	assert.Equal(t, nil, c.Close())
}

func TestVolumes(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.Volumes()

	assert.Nil(t, err)

	for _, i := range items {
		t.Logf("%+v", i)
	}
	assert.Equal(t, nil, c.Close())
}

func TestPools(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.Pools()

	assert.Nil(t, err)
	assert.Greater(t, len(items), 0)

	for _, i := range items {
		t.Logf("%+v", i)
	}
	assert.Equal(t, nil, c.Close())
}

func TestDisks(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.Disks()

	assert.Nil(t, err)
	assert.Greater(t, len(items), 0)

	for _, s := range items {

		assert.Equal(t, "Disk", s.Class)

		if s.DiskType == lsm.DiskTypeSata {
			t.Logf("Got the sata disk!")
		}

		if s.Status&(lsm.DiskStatusOk|lsm.DiskStatusFree) == lsm.DiskStatusOk|lsm.DiskStatusFree {
			t.Logf("Disk OK and FREE %x\n", lsm.DiskStatusOk|lsm.DiskStatusFree)
		}

		t.Logf("%+v", s)
	}
	assert.Equal(t, nil, c.Close())
}

func TestFs(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.FileSystems()

	assert.Nil(t, err)
	assert.Greater(t, len(items), 0)

	for _, i := range items {
		t.Logf("%+v", i)
	}
	assert.Equal(t, nil, c.Close())
}

func TestNfsExports(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.NfsExports()
	assert.Nil(t, err)

	if len(items) == 0 {
		var fs, err = c.FileSystems()
		assert.Nil(t, err)
		var access lsm.NfsAccess

		var exportPath = "/mnt/fubar"
		var auth = "standard"
		access.AnonGID = lsm.AnonUIDGIDNotApplicable
		access.AnonUID = lsm.AnonUIDGIDNotApplicable
		var export lsm.NfsExport

		// Test bad arguments
		access.Ro = make([]string, 0)
		access.Rw = make([]string, 0)
		var e = c.FsExport(&fs[0], &exportPath, &access, &auth, nil, &export)
		assert.NotNil(t, e)

		access.Root = []string{"192.168.1.1"}
		access.Rw = []string{"192.168.1.2"}
		e = c.FsExport(&fs[0], &exportPath, &access, &auth, nil, &export)
		assert.NotNil(t, e)

		access.Root = make([]string, 0)
		access.Rw = []string{"192.168.1.2"}
		access.Ro = []string{"192.168.1.2"}
		e = c.FsExport(&fs[0], &exportPath, &access, &auth, nil, &export)
		assert.NotNil(t, e)

		// This one should be good
		access.Rw = []string{"192.168.1.1"}
		var exportErr = c.FsExport(&fs[0], &exportPath, &access, &auth, nil, &export)
		assert.Nil(t, exportErr)
		assert.Equal(t, exportPath, export.ExportPath)

		var unExportErr = c.FsUnExport(&export)
		assert.Nil(t, unExportErr)
	}

	for _, i := range items {
		t.Logf("%+v", i)

		var unExportErr = c.FsUnExport(&i)
		assert.Nil(t, unExportErr)
	}
	assert.Equal(t, nil, c.Close())
}

func TestNfsAuthTypes(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var authTypes, err = c.NfsExportAuthTypes()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(authTypes))
	assert.Equal(t, "standard", authTypes[0])

	fmt.Printf("%v", authTypes)
	assert.Equal(t, nil, c.Close())
}

func TestAccessGroups(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.AccessGroups()
	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	assert.Nil(t, err)

	if len(items) == 0 {
		var systems, sysErr = c.Systems()
		assert.Nil(t, sysErr)

		var ag lsm.AccessGroup
		var agCreateErr = c.AccessGroupCreate(rs("lsm_ag_", 4),
			"iqn.1994-05.com.domain:01.89bd01", lsm.InitiatorTypeIscsiIqn, &systems[0], &ag)
		assert.Nil(t, agCreateErr)

		var maskErr = c.VolumeMask(&volumes[0], &ag)
		assert.Nil(t, maskErr)

		var volsMasked, volsMaskedErr = c.VolsMaskedToAg(&ag)
		assert.Nil(t, volsMaskedErr)
		assert.Equal(t, 1, len(volsMasked))
		assert.Equal(t, volumes[0].Name, volsMasked[0].Name)

		var agsGranted, agsGrantedErr = c.AgsGrantedToVol(&volumes[0])
		assert.Nil(t, agsGrantedErr)
		assert.Equal(t, 1, len(agsGranted))
		assert.Equal(t, ag.Name, agsGranted[0].Name)

		var unmaskErr = c.VolumeUnMask(&volumes[0], &ag)
		assert.Nil(t, unmaskErr)

		volsMasked, volsMaskedErr = c.VolsMaskedToAg(&ag)
		assert.Nil(t, volsMaskedErr)
		assert.Equal(t, 0, len(volsMasked))

		agsGranted, agsGrantedErr = c.AgsGrantedToVol(&volumes[0])
		assert.Nil(t, agsGrantedErr)
		assert.Equal(t, 0, len(agsGranted))

		// Try to add a bad iSCSI iqn
		var agInitAdd lsm.AccessGroup
		var initAddErr = c.AccessGroupInitAdd(&ag, "iqz.1994-05.com.domain:01.89bd02", lsm.InitiatorTypeIscsiIqn, &agInitAdd)
		assert.NotNil(t, initAddErr)

		initAddErr = c.AccessGroupInitAdd(&ag, "not_even_close", lsm.InitiatorTypeWwpn, &agInitAdd)
		assert.NotNil(t, initAddErr)

		initAddErr = c.AccessGroupInitAdd(&ag, "iqn.1994-05.com.domain:01.89bd02", lsm.InitiatorType(100), &agInitAdd)
		assert.NotNil(t, initAddErr)

		initAddErr = c.AccessGroupInitAdd(&ag, "iqn.1994-05.com.domain:01.89bd02", lsm.InitiatorTypeIscsiIqn, &agInitAdd)
		assert.Nil(t, initAddErr)
		assert.NotEqual(t, len(ag.InitIDs), len(agInitAdd.InitIDs))

		initAddErr = c.AccessGroupInitAdd(&ag, "0x002538c571b06a6d", lsm.InitiatorTypeWwpn, &agInitAdd)
		assert.Nil(t, initAddErr)
		assert.NotEqual(t, len(ag.InitIDs), len(agInitAdd.InitIDs))

		var agInitDel lsm.AccessGroup
		var initDelErr = c.AccessGroupInitDelete(&ag, "iqn.1994-05.com.domain:01.89bd02", lsm.InitiatorTypeIscsiIqn, &agInitDel)
		assert.Nil(t, initDelErr)

		items, err = c.AccessGroups()
		assert.Nil(t, err)
	}

	assert.Greater(t, len(items), 0)

	for _, i := range items {
		t.Logf("%+v", i)

		var agDelErr = c.AccessGroupDelete(&i)
		assert.Nil(t, agDelErr)
	}

	items, err = c.AccessGroups()
	assert.Equal(t, 0, len(items))
	assert.Equal(t, nil, c.Close())
}

func TestTargetPorts(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.TargetPorts()

	assert.Nil(t, err)

	for _, i := range items {
		t.Logf("%+v", i)
	}

	assert.Greater(t, len(items), 0)
	assert.Equal(t, nil, c.Close())
}

func TestBatteries(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var items, err = c.Batteries()

	assert.Nil(t, err)

	for _, i := range items {
		t.Logf("%+v", i)
	}

	assert.Greater(t, len(items), 0)
	assert.Equal(t, nil, c.Close())
}

func TestCapabilities(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var systems, sysError = c.Systems()
	assert.Nil(t, sysError)

	var cap, capErr = c.Capabilities(&systems[0])
	assert.Nil(t, capErr)

	assert.True(t, cap.IsSupported(lsm.CapVolumeCreate))
	assert.Equal(t, nil, c.Close())
}

func TestCapabilitiesSet(t *testing.T) {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)
	var systems, sysError = c.Systems()
	assert.Nil(t, sysError)

	var cap, capErr = c.Capabilities(&systems[0])
	assert.Nil(t, capErr)

	var set = []lsm.CapabilityType{lsm.CapVolumeCreate, lsm.CapVolumeCResize}
	assert.True(t, cap.IsSupportedSet(set))
	assert.Equal(t, nil, c.Close())
}

func TestRepBlockSize(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var systems, sysError = c.Systems()
	assert.Nil(t, sysError)

	var repRangeBlkSize, rpbE = c.VolumeRepRangeBlkSize(&systems[0])
	assert.Nil(t, rpbE)
	assert.Equal(t, uint32(512), repRangeBlkSize)
	assert.Equal(t, nil, c.Close())
}

func createVolume(t *testing.T, c *lsm.ClientConnection, name string) *lsm.Volume {
	var pools, poolError = c.Pools()
	assert.Nil(t, poolError)

	var poolToUse = pools[3] // Arbitrary

	var volume lsm.Volume
	var jobID, errVolCreate = c.VolumeCreate(&poolToUse, name, 1024*1024*1, 2, true, &volume)
	assert.Nil(t, errVolCreate)
	assert.Nil(t, jobID)

	return &volume
}

func TestVolumeCreate(t *testing.T) {
	var volumeName = rs("lsm_go_vol_", 8)
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var volume = createVolume(t, c, volumeName)

	assert.Equal(t, volumeName, volume.Name)

	// Try and clean-up
	var volDel, volDelErr = c.VolumeDelete(volume, true)
	assert.Nil(t, volDel)
	assert.Nil(t, volDelErr)
	assert.Equal(t, nil, c.Close())
}

func TestVolumeEnableDisable(t *testing.T) {
	var volumeName = rs("lsm_go_vol_", 8)
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var volume = createVolume(t, c, volumeName)
	assert.Equal(t, volumeName, volume.Name)

	var disableErr = c.VolumeDisable(volume)
	assert.Nil(t, disableErr)

	var enableErr = c.VolumeEnable(volume)
	assert.Nil(t, enableErr)

	// Try and clean-up
	var volDel, volDelErr = c.VolumeDelete(volume, true)
	assert.Nil(t, volDel)
	assert.Nil(t, volDelErr)
	assert.Equal(t, nil, c.Close())
}

func TestLEDOnOff(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var volumeName = rs("lsm_go_vol_", 8)
	var volume = createVolume(t, c, volumeName)
	assert.Equal(t, volumeName, volume.Name)

	var onErr = c.VolIdentLedOn(volume)
	assert.Nil(t, onErr)

	var offErr = c.VolIdentLedOff(volume)
	assert.Nil(t, offErr)

	// Try and clean-up
	var volDel, volDelErr = c.VolumeDelete(volume, true)
	assert.Nil(t, volDel)
	assert.Nil(t, volDelErr)
	assert.Equal(t, nil, c.Close())
}

func TestScale(t *testing.T) {

	var c, err = lsm.Client("simc://", "", 30000)
	assert.Nil(t, err)

	var pools, poolError = c.Pools()
	assert.Nil(t, poolError)

	var poolToUse = pools[3]

	for i := 0; i < 10; i++ {
		var volumeName = rs("lsm_go_vol_", 8)

		var volume lsm.Volume
		var jobID, errVolCreate = c.VolumeCreate(&poolToUse, volumeName, 1024*1024*10, 2, true, &volume)
		if errVolCreate != nil {
			fmt.Printf("Created %d volume before we got error %s\n", i, errVolCreate)
			break
		}
		assert.Nil(t, jobID)
	}

	var volumes, vE = c.Volumes()
	assert.Nil(t, vE)
	assert.Greater(t, len(volumes), 10)

	assert.Equal(t, nil, c.Close())
}

func TestVolumeDelete(t *testing.T) {
	var volumeName = rs("lsm_go_vol_", 8)

	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var volume = createVolume(t, c, volumeName)

	var _, errD = c.VolumeDelete(volume, true)
	assert.Nil(t, errD)
	assert.Equal(t, nil, c.Close())
}

func TestJobWait(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var pools, poolError = c.Pools()
	assert.Nil(t, poolError)

	var poolToUse = pools[2] // Arbitrary

	var volumeName = rs("lsm_go_vol_async_", 8)

	var volume lsm.Volume
	var jobID, errCreate = c.VolumeCreate(&poolToUse, volumeName, 1024*1024*100, 2, false, &volume)
	assert.Nil(t, errCreate)
	assert.NotNil(t, jobID)

	var waitForIt = c.JobWait(*jobID, &volume)
	assert.Nil(t, waitForIt)

	assert.Equal(t, volumeName, volume.Name)

	c.VolumeDelete(&volume, true)
	assert.Equal(t, nil, c.Close())
}

func TestTmo(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)

	var tmo uint32 = 32121

	assert.Nil(t, c.TimeOutSet(tmo))

	assert.Equal(t, tmo, c.TimeOutGet())
	assert.Equal(t, nil, c.Close())
}

func TestVolumeResize(t *testing.T) {
	var volumeName = rs("lsm_go_vol_", 8)
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volume = createVolume(t, c, volumeName)
	var resized lsm.Volume
	var _, resizeErr = c.VolumeResize(volume, (volume.BlockSize*volume.NumOfBlocks)*2, true, &resized)
	assert.Nil(t, resizeErr)
	assert.NotEqual(t, volume.NumOfBlocks, resized.NumOfBlocks)

	c.VolumeDelete(&resized, true)
	assert.Equal(t, nil, c.Close())
}

func TestVolumeRaidType(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	var raidInfo, raidInfoErr = c.VolRaidInfo(&volumes[0])
	assert.Nil(t, raidInfoErr)
	assert.NotNil(t, raidInfo)

	assert.Equal(t, nil, c.Close())
}

func TestVolumeCacheInfo(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	var cacheInfo, raidInfoErr = c.VolCacheInfo(&volumes[0])
	assert.Nil(t, raidInfoErr)
	assert.NotNil(t, cacheInfo)

	t.Logf("%+v", cacheInfo)

	assert.Equal(t, nil, c.Close())
}

func TestVolPhyDiskCacheSet(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	var cacheSetErr = c.VolPhyDiskCacheSet(&volumes[0], lsm.PhysicalDiskCacheEnabled)
	assert.Nil(t, cacheSetErr)

	assert.Equal(t, nil, c.Close())
}

func TestVolWriteCacheSet(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	var cacheSetErr = c.VolWriteCacheSet(&volumes[0], lsm.WriteCachePolicyAuto)
	assert.Nil(t, cacheSetErr)

	assert.Equal(t, nil, c.Close())
}

func TestVolReadCacheSet(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumes, volErr = c.Volumes()
	assert.Nil(t, volErr)

	var cacheSetErr = c.VolReadCacheSet(&volumes[0], lsm.ReadCachePolicyEnabled)
	assert.Nil(t, cacheSetErr)

	assert.Equal(t, nil, c.Close())
}

func TestPoolMemberInfo(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var pools, volErr = c.Pools()
	assert.Nil(t, volErr)

	var poolInfo, poolInfoErr = c.PoolMemberInfo(&pools[0])
	assert.Nil(t, poolInfoErr)
	assert.NotNil(t, poolInfo)

	t.Logf("%+v", poolInfo)

	assert.Equal(t, nil, c.Close())
}

func TestRaidCreateCapGet(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var systems, sysErr = c.Systems()
	assert.Nil(t, sysErr)

	var cap, errCapGet = c.VolRaidCreateCapGet(&systems[0])
	assert.NotNil(t, cap)
	assert.Nil(t, errCapGet)
	t.Logf("%+v", cap)

	assert.Equal(t, nil, c.Close())
}

func TestVolRaidCreate(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var name = rs("lsm_go_vol_", 4)

	var disks, diskErr = c.Disks()
	assert.Nil(t, diskErr)
	var freeDisks []lsm.Disk

	// Find disks that are OK and FREE to use
	for _, d := range disks {
		if d.Status&(lsm.DiskStatusOk|lsm.DiskStatusFree) == lsm.DiskStatusOk|lsm.DiskStatusFree {
			freeDisks = append(freeDisks, d)
		}
	}

	var volume lsm.Volume
	var volumeErr = c.VolRaidCreate(name, lsm.Raid5, freeDisks, 0, &volume)
	assert.Nil(t, volumeErr)

	if volumeErr == nil {
		assert.Equal(t, name, volume.Name)
		var jobID, volDelErr = c.VolumeDelete(&volume, true)
		assert.Nil(t, jobID)
		assert.Nil(t, volDelErr)
	}

	assert.Equal(t, nil, c.Close())
}

func TestVolumeReplicate(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumeName = rs("lsm_go_vol_", 8)
	var repName = rs("lsm_go_rep_", 4)

	var repVol lsm.Volume
	var srcVol = createVolume(t, c, volumeName)
	var jobID, errRep = c.VolumeReplicate(nil, lsm.VolumeReplicateTypeCopy, srcVol, repName, true, &repVol)
	assert.Nil(t, jobID)
	assert.Nil(t, errRep)

	assert.Equal(t, repName, repVol.Name)

	c.VolumeDelete(&repVol, true)

	var pools, poolError = c.Pools()
	assert.Nil(t, poolError)

	jobID, errRep = c.VolumeReplicate(&pools[3], lsm.VolumeReplicateTypeCopy, srcVol, repName, true, &repVol)
	assert.Nil(t, jobID)
	assert.Nil(t, errRep)

	c.VolumeDelete(&repVol, true)

	c.VolumeDelete(srcVol, true)
	assert.Equal(t, nil, c.Close())

}

func TestVolumeReplicateRange(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var volumeName = rs("lsm_go_vol_", 8)
	var volume = createVolume(t, c, volumeName)

	var ranges []lsm.BlockRange
	ranges = append(ranges, lsm.BlockRange{BlkCount: 100, SrcBlkAddr: 10, DstBlkAddr: 400})

	var jobID, repErr = c.VolumeReplicateRange(lsm.VolumeReplicateTypeCopy, volume, volume, ranges, true)
	assert.Nil(t, jobID)
	assert.Nil(t, repErr)

	c.VolumeDelete(volume, true)
	assert.Equal(t, nil, c.Close())
}

func TestFsCreateResizeCloneDelete(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var pools, pE = c.Pools()
	assert.Nil(t, pE)

	var newFs lsm.FileSystem
	var fsCreateJob, fsCreateErr = c.FsCreate(&pools[2], rs("lsm_go_pool_", 4), 1024*1024*100, true, &newFs)
	assert.Nil(t, fsCreateJob)
	assert.Nil(t, fsCreateErr)

	var resizedFs lsm.FileSystem
	var resizedJob, resizedErr = c.FsResize(&newFs, newFs.TotalSpace*2, true, &resizedFs)
	assert.Nil(t, resizedJob)
	assert.Nil(t, resizedErr)
	assert.NotEqual(t, newFs.TotalSpace, resizedFs)

	var snapShot lsm.FileSystemSnapShot
	var _, ssE = c.FsSnapShotCreate(&resizedFs, rs("lsm_go_ss_", 8), true, &snapShot)
	assert.Nil(t, ssE)

	var cloned lsm.FileSystem
	var cloneFsJob, cloneErr = c.FsClone(&resizedFs, "lsm_go_cloned_fs", nil, true, &cloned)
	assert.Nil(t, cloneFsJob)
	assert.Nil(t, cloneErr)

	var cloned2 lsm.FileSystem
	cloneFsJob, cloneErr = c.FsClone(&resizedFs, "lsm_go_cloned_fs_from_ss", &snapShot, true, &cloned2)
	assert.Nil(t, cloneFsJob)
	assert.Nil(t, cloneErr)

	assert.Equal(t, "lsm_go_cloned_fs", cloned.Name)
	assert.Equal(t, resizedFs.TotalSpace, cloned.TotalSpace)

	var delcloneFsJob, delCloneFsErr = c.FsDelete(&cloned, true)
	assert.Nil(t, delcloneFsJob)
	assert.Nil(t, delCloneFsErr)

	delcloneFsJob, delCloneFsErr = c.FsDelete(&cloned2, true)
	assert.Nil(t, delcloneFsJob)
	assert.Nil(t, delCloneFsErr)

	var delFsJob, delFsErr = c.FsDelete(&resizedFs, true)
	assert.Nil(t, delFsJob)
	assert.Nil(t, delFsErr)

	assert.Equal(t, nil, c.Close())
}

func TestFsFileClone(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var pools, pE = c.Pools()
	assert.Nil(t, pE)

	var newFs lsm.FileSystem
	var fsCreateJob, fsCreateErr = c.FsCreate(&pools[2], rs("lsm_go_pool_", 4), 1024*1024*100, true, &newFs)
	assert.Nil(t, fsCreateJob)
	assert.Nil(t, fsCreateErr)

	var fsFileCloneJob, fsFcErr = c.FsFileClone(&newFs, "some_file", "some_other_file", nil, true)
	assert.Nil(t, fsFileCloneJob)
	assert.Nil(t, fsFcErr)

	var delFsJob, delFsErr = c.FsDelete(&newFs, true)
	assert.Nil(t, delFsJob)
	assert.Nil(t, delFsErr)

	assert.Equal(t, nil, c.Close())
}

func TestFsSnapShots(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var pools, pE = c.Pools()
	assert.Nil(t, pE)

	var newFs lsm.FileSystem
	var fsCreateJob, fsCreateErr = c.FsCreate(&pools[2], rs("lsm_go_pool_", 4), 1024*1024*100, true, &newFs)
	assert.Nil(t, fsCreateJob)
	assert.Nil(t, fsCreateErr)

	var ss lsm.FileSystemSnapShot
	var ssJob, ssE = c.FsSnapShotCreate(&newFs, "lsm_go_ss", true, &ss)

	assert.Nil(t, ssJob)
	assert.Nil(t, ssE)
	assert.Equal(t, "lsm_go_ss", ss.Name)

	var hasDep, depErr = c.FsHasChildDep(&newFs, make([]string, 0))
	assert.Nil(t, depErr)
	assert.True(t, hasDep)

	// TODO fix simulated FsChildDepRm as its deleting the snapshot instead of removing dependency.
	var jobRm, depRmErr = c.FsChildDepRm(&newFs, make([]string, 0), true)
	assert.Nil(t, depRmErr)
	assert.True(t, hasDep)
	assert.Nil(t, jobRm)

	ssJob, ssE = c.FsSnapShotCreate(&newFs, "lsm_go_ss", true, &ss)

	var snaps, snapsErr = c.FsSnapShots(&newFs)
	assert.Nil(t, snapsErr)

	assert.Equal(t, 1, len(snaps))

	for _, i := range snaps {
		t.Logf("%+v", i)
	}

	var ssDelJob, ssDelErr = c.FsSnapShotDelete(&newFs, &ss, true)
	assert.Nil(t, ssDelJob)
	assert.Nil(t, ssDelErr)

	c.FsDelete(&newFs, true)
	assert.Equal(t, nil, c.Close())
}

func TestFsSnapShotRestore(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	var pools, pE = c.Pools()
	assert.Nil(t, pE)

	var newFs lsm.FileSystem
	var fsCreateJob, fsCreateErr = c.FsCreate(&pools[2], rs("lsm_go_pool_", 4), 1024*1024*100, true, &newFs)
	assert.Nil(t, fsCreateJob)
	assert.Nil(t, fsCreateErr)

	var ss lsm.FileSystemSnapShot
	var ssName = rs("lsm_go_ss_", 4)
	var ssJob, ssE = c.FsSnapShotCreate(&newFs, ssName, true, &ss)

	assert.Nil(t, ssJob)
	assert.Nil(t, ssE)
	assert.Equal(t, ssName, ss.Name)

	var ssRestoreJob, ssRestoreErr = c.FsSnapShotRestore(
		&newFs, &ss, false, make([]string, 0), make([]string, 0), true)
	assert.NotNil(t, ssRestoreErr)

	var files = []string{"/tmp/bar", "/tmp/other"}
	var restoreFiles = make([]string, 0)
	assert.NotEqual(t, len(files), len(restoreFiles))

	ssRestoreJob, ssRestoreErr = c.FsSnapShotRestore(
		&newFs, &ss, false, files, restoreFiles, true)
	assert.NotNil(t, ssRestoreErr)

	ssRestoreJob, ssRestoreErr = c.FsSnapShotRestore(
		&newFs, &ss, true, make([]string, 0), make([]string, 0), true)

	assert.Nil(t, ssRestoreJob)
	assert.Nil(t, ssRestoreErr)

	var org = []string{"/tmp/bar"}
	var rst = []string{"/tmp/fubar"}

	var ssRestoreJobF, ssRestoreErrF = c.FsSnapShotRestore(
		&newFs, &ss, false, org, rst, true)

	assert.Nil(t, ssRestoreJobF)
	assert.Nil(t, ssRestoreErrF)

	var ssDelJob, ssDelErr = c.FsSnapShotDelete(&newFs, &ss, true)
	assert.Nil(t, ssDelJob)
	assert.Nil(t, ssDelErr)

	c.FsDelete(&newFs, true)

	c.FsDelete(&newFs, true)
	assert.Equal(t, nil, c.Close())
}

func TestTemplate(t *testing.T) {
	var c, err = lsm.Client(URI, PASSWORD, TMO)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, nil, c.Close())
}

func contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}

func TestLocalDisk(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var sn, err = disks.SerialNumGet(d)
		var vpd, vpdE = disks.Vpd83Get(d)

		if err == nil {
			assert.True(t, len(sn) > 0)
		} else {
			checkError(t, err)
		}

		if vpdE == nil {
			assert.True(t, len(vpd) > 0)

			var search, searchErr = disks.Vpd83Seach(vpd)
			assert.Nil(t, searchErr)
			assert.True(t, len(search) > 0)
			t.Logf("vpd search result = %v %s\n", search, d)
			assert.True(t, contains(search, d))
		} else {
			checkError(t, vpdE)
		}
	}
}

func TestVpdMissingSearch(t *testing.T) {
	var paths, err = disks.Vpd83Seach(rs("", 16))
	assert.Nil(t, err)
	assert.True(t, len(paths) == 0)
}

func TestRpm(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var rpm, err = disks.RpmGet(d)
		if err != nil {
			checkError(t, err)
		} else {
			t.Logf("rpm = %d\n", rpm)
			assert.True(t, rpm >= -2 && rpm <= 20000)
		}
	}
}

func TestHealthStatus(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var status, err = disks.HealthStatusGet(d)
		if err != nil {
			checkError(t, err)
		} else {
			assert.Equal(t, lsm.DiskHealthStatusGood, status)
		}
	}
}

func checkError(t *testing.T, err error) {
	var e = err.(*errors.LsmError)
	if os.Getuid() == 0 {
		assert.Equal(t, errors.NoSupport, e.Code)
	} else {
		assert.True(t, e.Code == errors.PermissionDenied || e.Code == errors.NoSupport)
	}
}

func TestLinkType(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var _, err = disks.LinkTypeGet(d)
		if err != nil {
			checkError(t, err)
			t.Logf("LinkTypeGet: failed, reason %v for %s\n", err, d)
		}
	}
}

func TestIdentLed(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var err = disks.IndentLedOn(d)
		var offErr = disks.IndentLedOff(d)

		if err != nil {
			checkError(t, err)
			t.Logf("IndentLedOn: failed, reason %v for %s\n", err, d)
		} else {
			t.Logf("IndentLedOn SUCCESS: %s\n", d)
		}

		if offErr != nil {
			checkError(t, err)
			t.Logf("IndentLedOff: failed, reason %v for %s\n", err, d)
		} else {
			t.Logf("IndentLedOff SUCCESS: %s\n", d)
		}
	}
}

func TestFaultLed(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var err = disks.FaultLedOn(d)
		var offErr = disks.FaultLedOff(d)

		if err != nil {
			checkError(t, err)
			t.Logf("FaultLedOn: failed, reason %v for %s\n", err, d)
		} else {
			t.Logf("FaultLedOn SUCCESS: %s\n", d)
		}

		if offErr != nil {
			checkError(t, err)
			t.Logf("FaultLedOff: failed, reason %v for %s\n", err, d)
		} else {
			t.Logf("FaultLedOff SUCCESS: %s\n", d)
		}
	}
}

func TestLedStatusGet(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var status, err = disks.LedStatusGet(d)

		if err != nil {
			checkError(t, err)
		} else {
			t.Logf("status %v\n", status)
		}
	}
}

func TestLocalDiskLinkSpeed(t *testing.T) {
	var diskList, err = disks.List()

	assert.Nil(t, err)
	if len(diskList) == 0 {
		t.Skip("No local disks to test!")
	}

	for _, d := range diskList {
		var linkSpeed, err = disks.LinkSpeedGet(d)

		if err == nil {
			assert.Greater(t, linkSpeed, uint32(0))
		} else {
			checkError(t, err)
			t.Logf("link error %v\n", err)
		}
	}
}

func TestSystemReadCachePct(t *testing.T) {
	assert.Equal(t, lsm.SystemReadCachePctNoSupport, int8(-2))
	assert.Equal(t, lsm.SystemReadCachePctUnknown, int8(-1))
}

func TestSystemStatusType(t *testing.T) {
	assert.Equal(t, lsm.SystemStatusType(1<<0), lsm.SystemStatusUnknown)
	assert.Equal(t, lsm.SystemStatusType(1<<5), lsm.SystemStatusOther)
}

func TestSystemModeType(t *testing.T) {
	assert.Equal(t, lsm.SystemModeType(-2), lsm.SystemModeUnknown)
	assert.Equal(t, lsm.SystemModeType(1), lsm.SystemModeHba)
}

func TestJobStatusType(t *testing.T) {
	assert.Equal(t, lsm.JobStatusType(1), lsm.JobStatusInprogress)
	assert.Equal(t, lsm.JobStatusType(3), lsm.JobStatusError)
}

func TestVolumeReplicateType(t *testing.T) {
	//VolumeReplicateType
	assert.Equal(t, lsm.VolumeReplicateType(-1), lsm.VolumeReplicateTypeUnknown)
	assert.Equal(t, lsm.VolumeReplicateType(2), lsm.VolumeReplicateTypeClone)
	assert.Equal(t, lsm.VolumeReplicateType(5), lsm.VolumeReplicateTypeMirrorAsync)
}

func TestVolumeProvisionType(t *testing.T) {
	assert.Equal(t, lsm.VolumeProvisionType(-1), lsm.VolumeProvisionTypeUnknown)
	assert.Equal(t, lsm.VolumeProvisionType(1), lsm.VolumeProvisionTypeThin)
	assert.Equal(t, lsm.VolumeProvisionType(3), lsm.VolumeProvisionTypeDefault)
}

func TestPoolElementType(t *testing.T) {
	assert.Equal(t, lsm.PoolElementType(1<<1), lsm.PoolElementPool)
	assert.Equal(t, lsm.PoolElementType(1<<6), lsm.PoolElementTypeVolumeThin)
	assert.Equal(t, lsm.PoolElementType(1<<10), lsm.PoolElementTypeSysReserved)
}

func TestPoolUnsupportedType(t *testing.T) {
	assert.Equal(t, lsm.PoolUnsupportedType(1<<0), lsm.PoolUnsupportedVolumeGrow)
	assert.Equal(t, lsm.PoolUnsupportedType(1<<1), lsm.PoolUnsupportedVolumeShink)
}

func TestPoolStatusType(t *testing.T) {
	assert.Equal(t, lsm.PoolStatusType(1), lsm.PoolStatusUnknown)
	assert.Equal(t, lsm.PoolStatusType(1<<4), lsm.PoolStatusDegraded)
	assert.Equal(t, lsm.PoolStatusType(1<<9), lsm.PoolStatusStopped)
	assert.Equal(t, lsm.PoolStatusType(1<<12), lsm.PoolStatusReconstructing)
	assert.Equal(t, lsm.PoolStatusType(1<<15), lsm.PoolStatusGrowing)
}

func TestDiskType(t *testing.T) {
	assert.Equal(t, lsm.DiskType(0), lsm.DiskTypeUnknown)
	assert.Equal(t, lsm.DiskType(3), lsm.DiskTypeAta)
	assert.Equal(t, lsm.DiskType(51), lsm.DiskTypeNlSas)
	assert.Equal(t, lsm.DiskType(54), lsm.DiskTypeHybrid)
}

func TestDiskLinkType(t *testing.T) {
	assert.Equal(t, lsm.DiskLinkType(-2), lsm.DiskLinkTypeNoSupport)
	assert.Equal(t, lsm.DiskLinkType(-1), lsm.DiskLinkTypeUnknown)
	assert.Equal(t, lsm.DiskLinkType(0), lsm.DiskLinkTypeFc)
	assert.Equal(t, lsm.DiskLinkType(2), lsm.DiskLinkTypeSsa)
	assert.Equal(t, lsm.DiskLinkType(11), lsm.DiskLinkTypePciE)
}

func TestDiskStatusType(t *testing.T) {
	assert.Equal(t, lsm.DiskStatusType(1<<0), lsm.DiskStatusUnknown)
	assert.Equal(t, lsm.DiskStatusType(1<<13), lsm.DiskStatusFree)
}

func TestInitiatorType(t *testing.T) {
	assert.Equal(t, lsm.InitiatorType(0), lsm.InitiatorTypeUnknown)
	assert.Equal(t, lsm.InitiatorType(5), lsm.InitiatorTypeIscsiIqn)
	assert.Equal(t, lsm.InitiatorType(7), lsm.InitiatorTypeMixed)
}

func TestPortType(t *testing.T) {
	assert.Equal(t, lsm.PortType(1), lsm.PortTypeOther)
	assert.Equal(t, lsm.PortType(4), lsm.PortTypeIscsi)
}

func TestBatteryType(t *testing.T) {
	assert.Equal(t, lsm.BatteryType(1), lsm.BatteryTypeUnknown)
	assert.Equal(t, lsm.BatteryType(4), lsm.BatteryTypeCapacitor)
}

func TestBatteryStatus(t *testing.T) {
	assert.Equal(t, lsm.BatteryStatus(1), lsm.BatteryStatusUnknown)
	assert.Equal(t, lsm.BatteryStatus(1<<7), lsm.BatteryStatusError)
}

func setup() {
	var c, _ = lsm.Client(URI, PASSWORD, TMO)

	var pools, _ = c.Pools()
	var volumes, _ = c.Volumes()

	if len(volumes) == 0 {
		var volume lsm.Volume
		var _, _ = c.VolumeCreate(
			&pools[1], rs("lsm_go_vol_", 4),
			1024*1024*100,
			lsm.VolumeProvisionTypeDefault, true, &volume)
	}

	var fs, _ = c.FileSystems()
	if len(fs) == 0 {
		var fileSystem lsm.FileSystem
		var _, _ = c.FsCreate(
			&pools[1], rs("lsm_go_fs_", 4), 1024*1024*1000, true, &fileSystem)
	}
}
func TestMain(m *testing.M) {
	setup()

	// This will allow us to reproduce the same sequence if needed
	// if we encounter an error.
	var seed = os.Getenv("LSM_GO_SEED")
	if len(seed) > 0 {
		var sInt, err = strconv.ParseInt(seed, 10, 64)
		if err != nil {
			os.Exit(1)
		}
		rand.Seed(sInt)
	} else {
		var s = time.Now().UnixNano()
		rand.Seed(s)
		fmt.Printf("export LSM_GO_SEED=%v\n", s)
	}

	code := m.Run()
	//shutdown()
	os.Exit(code)
}
