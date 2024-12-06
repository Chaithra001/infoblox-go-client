package ibclient

import (
	"encoding/json"
	"fmt"
)

type Monitor struct {
	Name string
	Type string
}

// Updating Servers in DtcServerLink with reference
func updateServerReferences(servers []*DtcServerLink, objMgr *ObjectManager) error {
	for _, link := range servers {
		sf := map[string]string{"name": link.Server}
		queryParams := NewQueryParams(false, sf)
		var serverResult []DtcServer
		err := objMgr.connector.GetObject(&DtcServer{}, "dtc:server", queryParams, &serverResult)
		if err != nil {
			return err
		}
		if len(serverResult) > 0 {
			link.Server = serverResult[0].Ref
		} else {
			return fmt.Errorf("dtc:server with name %s not found", link.Server)
		}
	}
	return nil
}

// Updating the topology name with reference
func updateTopologyReference(LbPreferredTopology *string, objMgr *ObjectManager) (*string, error) {
	if LbPreferredTopology != nil {
		TopologyFields := map[string]string{"name": *LbPreferredTopology}
		queryParams := NewQueryParams(false, TopologyFields)
		var topologies []DtcTopology
		err := objMgr.connector.GetObject(&DtcTopology{}, "dtc:topology", queryParams, &topologies)
		if err != nil {
			return nil, err
		}
		if len(topologies) > 0 {
			return &topologies[0].Ref, nil
		} else {
			return nil, fmt.Errorf("dtc:topology with name %s not found", *LbPreferredTopology)
		}
	}
	return nil, nil
}

// get the monitor reference
func getMonitorReference(monitorName string, monitorType string, objMgr *ObjectManager) (string, error) {
	if monitorType == "" {
		return "", nil
	}
	fields := map[string]string{"name": monitorName}
	queryParams := NewQueryParams(false, fields)
	var monitorResult []DtcMonitorHttp

	monitorTypeKey := fmt.Sprintf("dtc:monitor:%s", monitorType)
	err := objMgr.connector.GetObject(&DtcMonitorHttp{}, monitorTypeKey, queryParams, &monitorResult)
	if err != nil {
		return "", err
	}
	if len(monitorResult) > 0 {
		return monitorResult[0].Ref, nil
	}
	return "", fmt.Errorf("dtc:monitor with name %s not found", monitorName)
}

func (d *DtcPool) MarshalJSON() ([]byte, error) {
	type Alias DtcPool
	aux := &struct {
		Monitors []string `json:"monitors,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	// Convert Monitors to a slice of strings
	for _, zone := range d.Monitors {
		if zone != nil {
			aux.Monitors = append(aux.Monitors, zone.Ref)
		}
	}
	return json.Marshal(aux)
}

func (d *DtcPool) UnmarshalJSON(data []byte) error {
	type Alias DtcPool
	aux := &struct {
		Monitors []string `json:"monitors,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Convert Monitors from []string to []*DtcMonitorHttp
	for _, ref := range aux.Monitors {
		d.Monitors = append(d.Monitors, &DtcMonitorHttp{Ref: ref})
	}

	return nil
}

func NewEmptyDtcPool() *DtcPool {
	PoolDtc := &DtcPool{}
	PoolDtc.SetReturnFields(append(PoolDtc.ReturnFields(), "lb_preferred_method", "servers", "lb_dynamic_ratio_preferred", "monitors", "auto_consolidated_monitors", "consolidated_monitors", "disable",
		"extattrs", "health", "lb_alternate_method", "lb_alternate_topology", "lb_dynamic_ratio_alternate", "lb_preferred_topology", "quorum", "ttl", "use_ttl", "availability"))

	return PoolDtc
}
func NewDtcPool(comment string,
	name string,
	LbPreferredMethod string,
	LbDynamicRatioPreferred *SettingDynamicratio,
	servers []*DtcServerLink,
	monitors []*DtcMonitorHttp,
	LbPreferredTopology *string,
	LbAlternateMethod string,
	LbAlternateTopology *string,
	LbDynamicRatioAlternate *SettingDynamicratio,
	eas EA,
	AutoConsolidatedMonitors bool,
	Availability string,
	ConsolidatedMonitors []*DtcPoolConsolidatedMonitorHealth,
	ttl uint32,
	useTTL bool,
	disable bool,
	Quorum uint32,
) *DtcPool {
	DtcPool := NewEmptyDtcPool()
	DtcPool.Comment = &comment
	DtcPool.Name = &name
	DtcPool.LbPreferredMethod = LbPreferredMethod
	DtcPool.Servers = servers
	DtcPool.LbDynamicRatioPreferred = LbDynamicRatioPreferred
	DtcPool.Monitors = monitors
	DtcPool.LbPreferredTopology = LbPreferredTopology
	DtcPool.LbAlternateMethod = LbAlternateMethod
	DtcPool.LbAlternateTopology = LbAlternateTopology
	DtcPool.LbDynamicRatioAlternate = LbDynamicRatioAlternate
	DtcPool.Ea = eas
	DtcPool.AutoConsolidatedMonitors = &AutoConsolidatedMonitors
	DtcPool.Availability = Availability
	DtcPool.ConsolidatedMonitors = ConsolidatedMonitors
	DtcPool.Ttl = &ttl
	DtcPool.UseTtl = &useTTL
	DtcPool.Disable = &disable
	DtcPool.Quorum = &Quorum
	return DtcPool
}

func (objMgr *ObjectManager) CreateDtcPool(
	comment string,
	name string,
	LbPreferredMethod string,
	LbDynamicRatioPreferred map[string]interface{},
	servers []*DtcServerLink,
	monitors []Monitor,
	LbPreferredTopology *string,
	LbAlternateMethod string,
	LbAlternateTopology *string,
	LbDynamicRatioAlternate map[string]interface{},
	eas EA,
	AutoConsolidatedMonitors bool,
	Availability string,
	ttl uint32,
	useTTL bool,
	disable bool,
	Quorum uint32,
) (*DtcPool, error) {
	if name == "" || LbPreferredMethod == "" {
		return nil, fmt.Errorf("name and LbPreferredMethod must be provided to create a pool")
	}
	if LbPreferredMethod == "DYNAMIC_RATIO" && LbDynamicRatioPreferred == nil {
		return nil, fmt.Errorf("LbDynamicRatioPreferred cannot be nil when LbPreferredMethod is set to DYNAMIC_RATIO")
	}
	if LbPreferredMethod == "TOPOLOGY" && LbPreferredTopology == nil {
		return nil, fmt.Errorf("LbPreferredTopology cannot be nil when LbPreferredMethod is set to TOPOLOGY")
	}
	//update servers with server references
	err := updateServerReferences(servers, objMgr)
	if err != nil {
		return nil, err
	}
	// update the monitor in LbDynamicRatioPreferred with reference
	var LbDynamicRatioPreferredMethod *SettingDynamicratio
	if LbDynamicRatioPreferred != nil {
		monitor, _ := LbDynamicRatioPreferred["monitor"].(Monitor)
		method, _ := LbDynamicRatioPreferred["method"].(string)
		monitorMetric, _ := LbDynamicRatioPreferred["monitor_metric"].(string)
		monitorWeighing, _ := LbDynamicRatioPreferred["monitor_weighing"].(string)
		InvertMonitorMetric, _ := LbDynamicRatioPreferred["monitor_invert_monitor"].(bool)

		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}
		LbDynamicRatioPreferredMethod = &SettingDynamicratio{
			Method:              method,
			Monitor:             monitorRef,
			MonitorMetric:       monitorMetric,
			MonitorWeighing:     monitorWeighing,
			InvertMonitorMetric: InvertMonitorMetric,
		}
	} else {
		LbDynamicRatioPreferredMethod = nil
	}

	// Convert monitor names to monitor references
	var monitorResults []*DtcMonitorHttp
	for _, monitor := range monitors {
		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}
		monitorResults = append(monitorResults, &DtcMonitorHttp{Ref: monitorRef})
	}
	//Update the topology name with the topology reference
	LbPreferredTopology, err = updateTopologyReference(LbPreferredTopology, objMgr)
	if err != nil {
		return nil, err
	}
	//Update the topology name with the topology reference
	LbAlternateTopology, err = updateTopologyReference(LbAlternateTopology, objMgr)
	if err != nil {
		return nil, err
	}
	//update the monitor in LbDynamicRatioPreferred with reference
	var LbDynamicRatioAlternateMethod *SettingDynamicratio
	if LbDynamicRatioAlternate != nil {
		monitorAlternate, _ := LbDynamicRatioAlternate["monitor"].(Monitor)
		methodAlternate, _ := LbDynamicRatioAlternate["method"].(string)
		monitorMetricAlternate, _ := LbDynamicRatioAlternate["monitor_metric"].(string)
		monitorWeighingAlternate, _ := LbDynamicRatioAlternate["monitor_weighing"].(string)
		InterferometricAlternate, _ := LbDynamicRatioAlternate["monitor_invert_monitor"].(bool)

		monitorRefAlternate, err := getMonitorReference(monitorAlternate.Name, monitorAlternate.Type, objMgr)
		if err != nil {
			return nil, err
		}
		LbDynamicRatioAlternateMethod = &SettingDynamicratio{
			Method:              methodAlternate,
			Monitor:             monitorRefAlternate,
			MonitorMetric:       monitorMetricAlternate,
			MonitorWeighing:     monitorWeighingAlternate,
			InvertMonitorMetric: InterferometricAlternate,
		}
	} else {
		LbDynamicRatioAlternateMethod = nil
	}

	// Create the DtcPool
	PoolDtc := NewDtcPool(comment, name, LbPreferredMethod, LbDynamicRatioPreferredMethod, servers, monitorResults, LbPreferredTopology, LbAlternateMethod, LbAlternateTopology, LbDynamicRatioAlternateMethod, eas, AutoConsolidatedMonitors, Availability, nil, ttl, useTTL, disable, Quorum)
	ref, err := objMgr.connector.CreateObject(PoolDtc)
	if err != nil {
		return nil, err
	}
	PoolDtc.Ref = ref
	return PoolDtc, nil
}

func (objMgr *ObjectManager) GetDtcPool(queryParams *QueryParams) (*DtcPool, error) {
	var res []DtcPool
	pool := NewEmptyDtcPool()
	err := objMgr.connector.GetObject(pool, "", queryParams, &res)
	if err != nil {
		return nil, fmt.Errorf("error getting DtcPool object, err: %s", err)
	}
	return &res[0], nil
}
func (objMgr *ObjectManager) UpdateDtcPool(
	ref string,
	comment string,
	name string,
	LbPreferredMethod string,
	LbDynamicRatioPreferred map[string]interface{},
	servers []*DtcServerLink,
	monitors []Monitor,
	LbPreferredTopology *string,
	LbAlternateMethod string,
	LbAlternateTopology *string,
	LbDynamicRatioAlternate map[string]interface{},
	eas EA,
	AutoConsolidatedMonitors bool,
	Availability string,
	userMonitors []map[string]interface{},
	ttl uint32,
	useTTL bool,
	disable bool,
	Quorum uint32,
) (*DtcPool, error) {
	if LbPreferredMethod == "DYNAMIC_RATIO" && LbDynamicRatioPreferred == nil {
		return nil, fmt.Errorf("LbDynamicRatioPreferred cannot be nil when LbPreferredMethod is set to DYNAMIC_RATIO")
	}
	if LbPreferredMethod == "TOPOLOGY" && LbPreferredTopology == nil {
		return nil, fmt.Errorf("LbPreferredTopology cannot be nil when LbPreferredMethod is set to TOPOLOGY")
	}
	//update servers with server references
	err := updateServerReferences(servers, objMgr)
	if err != nil {
		return nil, err
	}
	// Convert LbDynamicRatioPreferred to use monitor reference
	var LbDynamicRatioPreferredMethod *SettingDynamicratio
	if LbDynamicRatioPreferred != nil {
		monitor, _ := LbDynamicRatioPreferred["monitor"].(Monitor)
		method, _ := LbDynamicRatioPreferred["method"].(string)
		monitorMetric, _ := LbDynamicRatioPreferred["monitor_metric"].(string)
		monitorWeighing, _ := LbDynamicRatioPreferred["monitor_weighing"].(string)
		InvertMonitorMetric, _ := LbDynamicRatioPreferred["monitor_invert_monitor"].(bool)

		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}
		LbDynamicRatioPreferredMethod = &SettingDynamicratio{
			Method:              method,
			Monitor:             monitorRef,
			MonitorMetric:       monitorMetric,
			MonitorWeighing:     monitorWeighing,
			InvertMonitorMetric: InvertMonitorMetric,
		}
	} else {
		LbDynamicRatioPreferredMethod = nil
	}
	// Convert monitor names to monitor references
	var monitorResults []*DtcMonitorHttp
	for _, monitor := range monitors {
		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}
		monitorResults = append(monitorResults, &DtcMonitorHttp{Ref: monitorRef})
	}
	//Update the topology name with the topology reference
	LbPreferredTopology, err = updateTopologyReference(LbPreferredTopology, objMgr)
	if err != nil {
		return nil, err
	}
	//Update the topology name with the topology reference
	LbAlternateTopology, err = updateTopologyReference(LbAlternateTopology, objMgr)
	if err != nil {
		return nil, err
	}
	//Convert LbDynamicRatioAlternate to use monitor reference
	var lbDynamicRatioAlternateMethod *SettingDynamicratio
	if LbDynamicRatioAlternate != nil {
		monitorAlternate, _ := LbDynamicRatioAlternate["monitor"].(Monitor)
		methodAlternate, _ := LbDynamicRatioAlternate["method"].(string)
		monitorMetricAlternate, _ := LbDynamicRatioAlternate["monitor_metric"].(string)
		monitorWeighingAlternate, _ := LbDynamicRatioAlternate["monitor_weighing"].(string)
		InvertMonitorMetricAlternate, _ := LbDynamicRatioAlternate["monitor_invert_monitor"].(bool)

		monitorRefAlternate, err := getMonitorReference(monitorAlternate.Name, monitorAlternate.Type, objMgr)
		if err != nil {
			return nil, err
		}
		lbDynamicRatioAlternateMethod = &SettingDynamicratio{
			Method:              methodAlternate,
			Monitor:             monitorRefAlternate,
			MonitorMetric:       monitorMetricAlternate,
			MonitorWeighing:     monitorWeighingAlternate,
			InvertMonitorMetric: InvertMonitorMetricAlternate,
		}
	} else {
		lbDynamicRatioAlternateMethod = nil
	}
	//processing user input to retrieve monitor references and creating a slice of *DtcPoolConsolidatedMonitorHealth structs with updated monitor references.
	var consolidatedMonitors []*DtcPoolConsolidatedMonitorHealth
	for _, userMonitor := range userMonitors {
		monitor, okMonitor := userMonitor["monitor"].(Monitor)
		availability, okAvail := userMonitor["availability"].(string)
		fullHealthComm, _ := userMonitor["full_health_communication"].(bool)
		members, okMember := userMonitor["members"].([]string)
		if !okMonitor {
			return nil, fmt.Errorf("\"Required field missing: monitor")
		}

		if !okAvail {
			return nil, fmt.Errorf("\"Required field missing: availability")
		}

		if !okMember {
			return nil, fmt.Errorf("\"Required field missing: members\"")
		}
		monitorRef, err := getMonitorReference(monitor.Name, monitor.Type, objMgr)
		if err != nil {
			return nil, err
		}

		consolidatedMonitor := &DtcPoolConsolidatedMonitorHealth{
			Members:                 members,
			Monitor:                 monitorRef,
			Availability:            availability,
			FullHealthCommunication: fullHealthComm,
		}

		consolidatedMonitors = append(consolidatedMonitors, consolidatedMonitor)
	}

	PoolDtc := NewDtcPool(comment, name, LbPreferredMethod, LbDynamicRatioPreferredMethod, servers, monitorResults, LbPreferredTopology, LbAlternateMethod, LbAlternateTopology, lbDynamicRatioAlternateMethod, eas, AutoConsolidatedMonitors, Availability, consolidatedMonitors, ttl, useTTL, disable, Quorum)
	PoolDtc.Ref = ref
	reference, err := objMgr.connector.UpdateObject(PoolDtc, ref)
	if err != nil {
		return nil, err
	}
	PoolDtc.Ref = reference

	PoolDtc, err = objMgr.GetDtcPoolByRef(reference)
	if err != nil {
		return nil, err
	}

	return PoolDtc, nil

}
func (objMgr *ObjectManager) GetDtcPoolByRef(ref string) (*DtcPool, error) {
	PoolDtc := NewEmptyDtcPool()
	err := objMgr.connector.GetObject(
		PoolDtc, ref, NewQueryParams(false, nil), &PoolDtc)
	return PoolDtc, err
}
func (objMgr *ObjectManager) DeleteDtcPool(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}
