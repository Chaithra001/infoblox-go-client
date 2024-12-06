package ibclient

import "fmt"

func NewEmptyDtcServer() *DtcServer {
	DtcServer := &DtcServer{}
	DtcServer.SetReturnFields(append(DtcServer.ReturnFields(), "extattrs", "auto_create_host_record", "disable", "health", "monitors", "sni_hostname", "use_sni_hostname"))
	return DtcServer
}
func NewDtcServer(comment string,
	name string,
	host string,
	autoCreateHostRecord bool,
	disable bool,
	ea EA,
	monitors []*DtcServerMonitor,
	sniHostname string,
	useSniHostname bool,
) *DtcServer {
	DtcServer := NewEmptyDtcServer()
	DtcServer.Comment = &comment
	DtcServer.Name = &name
	DtcServer.Host = &host
	DtcServer.AutoCreateHostRecord = &autoCreateHostRecord
	DtcServer.Disable = &disable
	DtcServer.Ea = ea
	DtcServer.Monitors = monitors
	DtcServer.SniHostname = &sniHostname
	DtcServer.UseSniHostname = &useSniHostname
	return DtcServer
}

func (objMgr *ObjectManager) CreateDtcServer(
	comment string,
	name string,
	host string,
	autoCreateHostRecord bool,
	disable bool,
	ea EA,
	monitors []map[string]interface{},
	sniHostname string,
	useSniHostname bool,
) (*DtcServer, error) {
	if (useSniHostname && sniHostname == "") || (!useSniHostname && sniHostname != "") {
		return nil, fmt.Errorf("if 'use_sni_hostname' is enabled then 'sni_hostname' must be provided or if 'sni_hostname' is provided then 'use_sni_hostname' must be enabled")
	}
	var serverMonitors []*DtcServerMonitor
	for _, userMonitor := range monitors {
		monitor, okMonitor := userMonitor["monitor"].(Monitor)
		monitorHost, _ := userMonitor["host"].(string)
		if !okMonitor {
			return nil, fmt.Errorf("\"Required field missing: monitor")
		}
		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}

		serverMonitor := &DtcServerMonitor{
			Monitor: monitorRef,
			Host:    monitorHost,
		}

		serverMonitors = append(serverMonitors, serverMonitor)
	}
	dtcServer := NewDtcServer(comment, name, host, autoCreateHostRecord, disable, ea, serverMonitors, sniHostname, useSniHostname)
	ref, err := objMgr.connector.CreateObject(dtcServer)
	if err != nil {
		return nil, err
	}
	dtcServer.Ref = ref
	return dtcServer, nil
}

func (objMgr *ObjectManager) GetDtcServer(queryParams *QueryParams) (*DtcServer, error) {
	var res []DtcServer
	server := NewEmptyDtcServer()
	err := objMgr.connector.GetObject(server, "", queryParams, &res)
	if err != nil {
		return nil, fmt.Errorf("error getting DtcServer object, err: %s", err)
	}
	return &res[0], nil
}
func (objMgr *ObjectManager) UpdateDtcServer(
	ref string,
	comment string,
	name string,
	host string,
	autoCreateHostRecord bool,
	disable bool,
	ea EA,
	monitors []map[string]interface{},
	sniHostname string,
	useSniHostname bool) (*DtcServer, error) {
	if (useSniHostname && sniHostname == "") || (!useSniHostname && sniHostname != "") {
		return nil, fmt.Errorf("If 'use_sni_hostname' is enabled then 'sni_hostname' must be provided or if 'sni_hostname' is provided then 'use_sni_hostname' must be enabled ")
	}
	var serverMonitors []*DtcServerMonitor
	for _, userMonitor := range monitors {
		monitor, okMonitor := userMonitor["monitor"].(Monitor)
		monitorHost, _ := userMonitor["host"].(string)
		if !okMonitor {
			return nil, fmt.Errorf("\"Required field missing: monitor")
		}
		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}

		serverMonitor := &DtcServerMonitor{
			Monitor: monitorRef,
			Host:    monitorHost,
		}

		serverMonitors = append(serverMonitors, serverMonitor)
	}
	DtcServer := NewDtcServer(comment, name, host, autoCreateHostRecord, disable, ea, serverMonitors, sniHostname, useSniHostname)
	DtcServer.Ref = ref
	ref, err := objMgr.connector.UpdateObject(DtcServer, ref)
	if err != nil {
		return nil, err
	}
	DtcServer.Ref = ref
	return DtcServer, nil

}
func (objMgr *ObjectManager) GetDtcServerByRef(ref string) (*DtcServer, error) {
	serverDtc := NewEmptyDtcServer()
	err := objMgr.connector.GetObject(
		serverDtc, ref, NewQueryParams(false, nil), &serverDtc)
	return serverDtc, err
}

func (objMgr *ObjectManager) DeleteDtcServer(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}