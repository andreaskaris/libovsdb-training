// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package sbdb

// MulticastGroup defines an object in Multicast_Group table
type MulticastGroup struct {
	UUID      string   `ovsdb:"_uuid"`
	Datapath  string   `ovsdb:"datapath"`
	Name      string   `ovsdb:"name"`
	Ports     []string `ovsdb:"ports"`
	TunnelKey int      `ovsdb:"tunnel_key"`
}
