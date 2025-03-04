package customipxe

import (
	"context"
	"os"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/tinkerbell/boots/ipxe"
	"github.com/tinkerbell/boots/job"
)

var facility = func() string {
	fac := os.Getenv("FACILITY_CODE")
	if fac == "" {
		fac = "ewr1"
	}

	return fac
}()

func TestScript(t *testing.T) {
	for typ, script := range type2Script {
		t.Run(typ, func(t *testing.T) {
			extraIPXEVars := make([][]string, 2)
			extraIPXEVars[0] = []string{"dynamic_var1", "dynamic_val1"}
			extraIPXEVars[1] = []string{"dynamic_var2", "dynamic_val2"}

			m := job.NewMock(t, typ, facility)
			m.SetIPXEScriptURL("http://127.0.0.1/fake_ipxe_url")

			s := ipxe.NewScript()
			s.Set("iface", "eth0")
			s.Or("shell")
			s.Set("tinkerbell", "http://127.0.0.1")
			s.Set("syslog_host", "127.0.0.1")
			s.Set("ipxe_cloud_config", "packet")

			Installer(extraIPXEVars).BootScript("")(context.Background(), m.Job(), s)
			got := string(s.Bytes())
			if script != got {
				t.Fatalf("%s bad iPXE script:\n%v", typ, diff.LineDiff(script, got))
			}
		})
	}
}

var type2Script = map[string]string{
	"c3.small.x86": `#!ipxe

echo Tinkerbell Boots iPXE
set iface eth0 || shell
set tinkerbell http://127.0.0.1
set syslog_host 127.0.0.1
set ipxe_cloud_config packet
set dynamic_var1 dynamic_val1
set dynamic_var2 dynamic_val2

params
param body Device connected to DHCP system
param type provisioning.104.01
imgfetch ${tinkerbell}/phone-home##params
imgfree

set packet_facility ` + facility + `
set packet_plan c3.small.x86
chain --autofree http://127.0.0.1/fake_ipxe_url
`,
}
