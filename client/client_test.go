package client

import (
	"reflect"
	"sync"
	"testing"

	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/stretchr/testify/assert"
)

var (
	aUUID0 = "2f77b348-9768-4866-b761-89d5177ecda0"
	aUUID1 = "2f77b348-9768-4866-b761-89d5177ecda1"
	aUUID2 = "2f77b348-9768-4866-b761-89d5177ecda2"
	aUUID3 = "2f77b348-9768-4866-b761-89d5177ecda3"
)

func testOvsSet(t *testing.T, set interface{}) *ovsdb.OvsSet {
	oSet, err := ovsdb.NewOvsSet(set)
	assert.Nil(t, err)
	return oSet
}

func testOvsMap(t *testing.T, set interface{}) *ovsdb.OvsMap {
	oMap, err := ovsdb.NewOvsMap(set)
	assert.Nil(t, err)
	return oMap
}

func updateBenchmark(bridges []string, b *testing.B) {
	bridgeInsert := make(ovsdb.TableUpdate)
	for _, br := range bridges {
		r := newBridgeRow(br)
		bridgeInsert[br] = &ovsdb.RowUpdate{New: r}
	}
	ovsUpdate := ovsdb.TableUpdate{
		"829f8534-94a8-468e-9176-132738cf260a": &ovsdb.RowUpdate{Old: newOvsRow([]string{}), New: newOvsRow(bridges)},
	}
	tu := ovsdb.TableUpdates{
		"Open_vSwitch": ovsUpdate,
		"Bridge":       bridgeInsert,
	}
	ovs := OvsdbClient{
		handlers:      []ovsdb.NotificationHandler{},
		handlersMutex: &sync.Mutex{},
	}
	for n := 0; n < b.N; n++ {
		params := []interface{}{"v1", tu}
		if len(params) != 2 {
			b.Fatalf("Params not 2")
		}
		var reply []interface{}
		err := ovs.update(params, &reply)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func newBridgeRow(name string) *ovsdb.Row {
	return &ovsdb.Row{
		Fields: map[string]interface{}{
			"connection_mode":       []string{},
			"controller":            []string{},
			"datapath_id":           "blablabla",
			"datapath_type":         "",
			"datapath_version":      "",
			"external_ids":          map[string]string{"foo": "bar"},
			"fail_mode":             []string{},
			"flood_vlans":           []string{},
			"flow_tables":           map[string]string{},
			"ipfix":                 []string{},
			"mcast_snooping_enable": false,
			"mirrors":               []string{},
			"name":                  name,
			"netflow":               []string{},
			"other_config":          map[string]string{"baz": "quux"},
			"ports":                 []string{},
			"protocols":             []string{},
			"rstp_enable":           false,
			"rstp_status":           map[string]string{},
			"sflow":                 []string{},
			"status":                map[string]string{},
			"stp_enable":            false,
		},
	}
}

func newOvsRow(bridges []string) *ovsdb.Row {
	return &ovsdb.Row{
		Fields: map[string]interface{}{
			"bridges":          bridges,
			"cur_cfg":          0,
			"datapath_types":   []string{},
			"datapaths":        map[string]string{},
			"db_version":       "8.2.0",
			"dpdk_initialized": false,
			"dpdk_version":     []string{},
			"external_ids":     map[string]string{"system-id": "829f8534-94a8-468e-9176-132738cf260a"},
			"iface_types":      []string{},
			"manager_options":  "6e4cd5fc-f51a-462a-b3d6-a696af6d7a84",
			"next_cfg":         0,
			"other_config":     map[string]string{},
			"ovs_version":      "2.15.90",
			"ssl":              []string{},
			"statistics":       map[string]string{},
			"system_type":      "docker-ovs",
			"system_version":   "0.1",
		},
	}
}

func BenchmarkUpdate1(b *testing.B) {
	updateBenchmark([]string{"foo"}, b)
}

func BenchmarkUpdate2(b *testing.B) {
	updateBenchmark([]string{"foo", "bar"}, b)
}

func BenchmarkUpdate3(b *testing.B) {
	updateBenchmark([]string{"foo", "bar", "baz"}, b)
}

func BenchmarkUpdate5(b *testing.B) {
	updateBenchmark([]string{"foo", "bar", "baz", "quux", "foofoo"}, b)
}

func BenchmarkUpdate8(b *testing.B) {
	updateBenchmark([]string{"foo", "bar", "baz", "quux", "foofoo", "foobar", "foobaz", "fooquux"}, b)
}

func TestEcho(t *testing.T) {
	req := []interface{}{"hi"}
	var reply []interface{}
	ovs := OvsdbClient{
		handlers:      []ovsdb.NotificationHandler{},
		handlersMutex: &sync.Mutex{},
	}
	err := ovs.echo(req, &reply)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(req, reply) {
		t.Error("Expected: ", req, " Got: ", reply)
	}
}

func TestUpdate(t *testing.T) {
	ovs := OvsdbClient{
		handlers:      []ovsdb.NotificationHandler{},
		handlersMutex: &sync.Mutex{},
	}
	var reply []interface{}
	// Update notification should fail for arrays of size < 2
	err := ovs.update([]interface{}{"hello"}, &reply)
	if err == nil {
		t.Error("Expected: error for a dummy request")
	}

	// Update notification should fail if arg[1] is not map[string]map[string]*RowUpdate type
	err = ovs.update([]interface{}{"hello", "gophers"}, &reply)
	if err == nil {
		t.Error("Expected: error for a dummy request")
	}

	// Valid dummy update should pass
	validUpdate := make(map[string]interface{})
	validRowUpdate := make(map[string]ovsdb.RowUpdate)
	validRowUpdate["uuid"] = ovsdb.RowUpdate{}
	validUpdate["table"] = validRowUpdate

	err = ovs.update([]interface{}{"hello", validUpdate}, &reply)
	if err != nil {
		t.Error(err)
	}
}
