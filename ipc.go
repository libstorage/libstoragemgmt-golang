// SPDX-License-Identifier: 0BSD

package libstoragemgmt

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	errors "github.com/libstorage/libstoragemgmt-golang/errors"
)

const (
	udsPathVarName = "LSM_UDS_PATH"
	udsPathDefault = "/var/run/lsm/ipc"
	headerLen      = 10
)

type transPort struct {
	uds   net.Conn
	debug bool
}

func newTransport(pluginUdsPath string, checkErrors bool) (*transPort, error) {
	var c, cError = net.Dial("unix", pluginUdsPath)
	if cError != nil {

		// checkDaemonExists calls newTransport, to prevent unbounded recursion we
		// don't want to check while we are checking :-)
		if checkErrors {
			if checkDaemonExists() {
				return nil, &errors.LsmError{
					Code:    errors.PluginNotExist,
					Message: fmt.Sprintf("plug-in %s not found!", pluginUdsPath)}
			}

			return nil, &errors.LsmError{
				Code:    errors.DameonNotRunning,
				Message: "The libStorageMgmt daemon is not running (process name lsmd)"}
		}

		return nil, cError
	}

	debug := len(os.Getenv("LSM_GO_DEBUG")) > 0
	return &transPort{uds: c, debug: debug}, nil
}

func (t transPort) close() {
	t.uds.Close()
}

type responseMsg struct {
	ID     int              `json:"id"`
	Error  *errors.LsmError `json:"error"`
	Result json.RawMessage  `json:"result"`
}

type requestMsg struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (r *requestMsg) String() string {
	return fmt.Sprintf("ID: %d, Method: %s, Parms: %s", r.ID, r.Method, string(r.Params))
}

func (t *transPort) invoke(cmd string, args map[string]interface{}, result interface{}) error {

	args["flags"] = 0
	msg := map[string]interface{}{
		"method": cmd,
		"id":     100,
		"flags":  0,
		"params": args,
	}

	var msgSerialized, serialError = json.Marshal(msg)
	if serialError != nil {
		return &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Errors serializing parameters %w\n", serialError)}
	}

	if sendError := t.send(string(msgSerialized)); sendError != nil {
		return &errors.LsmError{
			Code:    errors.TransPortComunication,
			Message: fmt.Sprintf("Error writing to unix domain socket %w\n", sendError)}
	}

	var reply, replyError = t.recv()
	if replyError != nil {
		return &errors.LsmError{
			Code:    errors.TransPortComunication,
			Message: fmt.Sprintf("Error reading from unix domain socket %w\n", replyError)}
	}

	var what responseMsg
	if replyUnmarsal := json.Unmarshal(reply, &what); replyUnmarsal != nil {
		return &errors.LsmError{
			Code:    errors.PluginBug,
			Message: fmt.Sprintf("Unparsable response from plugin %w\n", replyUnmarsal)}
	}

	if what.Error != nil {
		return what.Error
	}

	if what.Result != nil {
		// We have a result, parse and return it.
		var unmarshalResult = json.Unmarshal(what.Result, &result)
		if unmarshalResult != nil {
			return &errors.LsmError{
				Code: errors.PluginBug,
				Message: fmt.Sprintf("Plugin returned unexpected response form for (%s) data (%s)",
					cmd, string(what.Result))}
		}

		return nil
	}

	return &errors.LsmError{
		Code:    errors.PluginBug,
		Message: fmt.Sprintf("Unexpected response from plugin %s\n", reply)}

}

func (t *transPort) readRequest() (*requestMsg, error) {
	request, requestError := t.recv()
	if requestError != nil {
		return nil, &errors.LsmError{
			Code:    errors.TransPortComunication,
			Message: fmt.Sprintf("Error reading from unix domain socket %w\n", requestError)}
	}

	var what requestMsg
	if requestUnmarsal := json.Unmarshal(request, &what); requestUnmarsal != nil {
		return nil, &errors.LsmError{
			Code:    errors.TransPortInvalidArg,
			Message: fmt.Sprintf("Unparsable request from client %w\n", requestUnmarsal)}
	}
	return &what, nil
}

func (t *transPort) send(msg string) error {

	var toSend = fmt.Sprintf("%010d%s", len(msg), msg)
	if t.debug {
		fmt.Printf("go-send: %s\n", msg)
	}
	return writeExact(t.uds, []byte(toSend))
}

func (t *transPort) recv() ([]byte, error) {
	hdrLenBuf := make([]byte, headerLen)

	if readError := readExact(t.uds, hdrLenBuf); readError != nil {
		return make([]byte, 0), readError
	}

	msgLen, parseError := strconv.ParseUint(string(hdrLenBuf), 10, 32)
	if parseError != nil {
		return make([]byte, 0), parseError
	}

	msgBuffer := make([]byte, msgLen)
	readError := readExact(t.uds, msgBuffer)

	if t.debug {
		fmt.Printf("go-recv: %s\n", string(msgBuffer))
	}

	return msgBuffer, readError
}

func readExact(c net.Conn, buf []byte) error {
	const tmpBufSize = 1024
	requested := len(buf)
	tmpBuffer := make([]byte, tmpBufSize)
	var current int

	for current < requested {
		remain := requested - current
		if remain > tmpBufSize {
			remain = tmpBufSize
		}

		num, readError := c.Read(tmpBuffer[:remain])
		if readError != nil {
			return readError
		}

		copy(buf[current:], tmpBuffer[:num])
		current += num
	}
	return nil
}

func writeExact(c net.Conn, buf []byte) error {
	wanted := len(buf)
	var written int

	for written < wanted {
		num, writeError := c.Write(buf[written:])
		if writeError != nil {
			return writeError
		}
		written += num
	}

	return nil
}
