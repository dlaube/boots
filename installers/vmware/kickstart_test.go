package vmware

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"testing"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/tinkerbell/boots/conf"
	"github.com/tinkerbell/boots/job"
)

func TestFirstDisk(t *testing.T) {
	tests := []struct {
		slug    string
		version string
		hint    string
		want    string
	}{
		{slug: "", hint: "", want: ""},
		{slug: "", hint: "hint", want: "hint"},
		{slug: "baremetal_5", want: ""},
		{slug: "baremetal_5", hint: "hint", want: "hint"},
		{slug: "c1.small.x86", hint: "hint", want: "hint"},
		{slug: "c1.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "c2.medium.x86", hint: "hint", want: "hint"},
		{slug: "g2.large.x86", hint: "hint", want: "hint"},
		{slug: "m1.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "m1.xlarge.x86:baremetal_2_04", hint: "hint", want: "hint"},
		{slug: "m2.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "n2.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "n2.xlarge.google", hint: "hint", want: "hint"},
		{slug: "s1.large.x86", hint: "hint", want: "hint"},
		{slug: "w1.large.x86", hint: "hint", want: "hint"},
		{slug: "t1.small.x86", hint: "hint", want: "hint"},
		{slug: "x1.small.x86", hint: "hint", want: "hint"},
		{slug: "arbitrary_name", hint: "hint", want: "hint"},
		{slug: "x2.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "n3.xlarge.x86", hint: "KXG60ZNV256G TOSHIBA", want: "KXG60ZNV256G TOS"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q+%q", tc.slug, tc.hint), func(t *testing.T) {
			m := job.NewMock(t, tc.slug, facility)
			m.SetBootDriveHint(tc.hint)
			got := firstDisk(m.Job())
			if got != tc.want {
				t.Errorf("firstDisk(%+v) = %q, want: %q", tc, got, tc.want)
			}
		})
	}
}

func TestScriptKickstart(t *testing.T) {
	manufacturers := []string{"supermicro", "dell"}
	versions := []string{"vmware_esxi_6_0", "vmware_esxi_6_5", "vmware_esxi_6_7", "vmware_esxi_7_0", "vmware_esxi_7_0U2a"}

	diskConfigs := []struct {
		slug    string
		version string
		hint    string
		want    string
	}{
		{slug: "", hint: "", want: ""},
		{slug: "", hint: "hint", want: "hint"},
		{slug: "baremetal_5", want: ""},
		{slug: "baremetal_5", hint: "hint", want: "hint"},
		{slug: "c1.small.x86", hint: "hint", want: "hint"},
		{slug: "c1.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "c2.medium.x86", hint: "hint", want: "hint"},
		{slug: "c3.medium.x86", want: "vmw_ahci,lsi_mr3,lsi_msgpt3"},
		{slug: "g2.large.x86", hint: "hint", want: "hint"},
		{slug: "m1.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "m1.xlarge.x86:baremetal_2_04", hint: "hint", want: "hint"},
		{slug: "m2.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "n2.xlarge.x86", hint: "hint", want: "hint"},
		{slug: "n2.xlarge.google", hint: "hint", want: "hint"},
		{slug: "s1.large.x86", hint: "", want: "vmw_ahci"},
		{slug: "s1.large.x86", hint: "hint", want: "hint"},
		{slug: "s3.xlarge.x86:s3.xlarge.x86.01", hint: "hint", want: "hint"},
		{slug: "s3.xlarge.x86:s3.xlarge.x86.01", hint: "", want: "KXG50ZNV256G_TOSHIBA,vmw_ahci"},
		{slug: "s3.xlarge.x86", hint: "", want: "vmw_ahci,lsi_mr3,lsi_msgpt3"},
		{slug: "w1.large.x86", hint: "", want: ""},
		{slug: "w1.large.x86", hint: "hint", want: "hint"},
		{slug: "arbitrary_name", hint: "hint", want: "hint"},
		{slug: "t1.small.x86", hint: "hint", want: "hint"},
		{slug: "x1.small.x86", hint: "hint", want: "hint"},
		{slug: "x2.xlarge.x86", hint: "hint", want: "hint"},
	}

	conf.PublicIPv4 = net.ParseIP("127.0.0.1")
	conf.PublicFQDN = "boots-test.example.com"

	for _, man := range manufacturers {
		t.Run(man, func(t *testing.T) {
			for _, ver := range versions {
				t.Run(ver, func(t *testing.T) {
					for _, dc := range diskConfigs {
						t.Run(fmt.Sprintf("%q+%q", dc.slug, dc.hint), func(t *testing.T) {
							m := job.NewMock(t, dc.slug, facility)
							m.SetManufacturer(man)
							m.SetOSSlug(ver)
							m.SetIP(net.ParseIP("127.0.0.1"))
							m.SetPassword("password")
							m.SetMAC("00:00:ba:dd:be:ef")
							m.SetBootDriveHint(dc.hint)

							var w strings.Builder
							genKickstart(m.Job(), &w)

							got := w.String()

							bs, err := ioutil.ReadFile(fmt.Sprintf("testdata/ks_%s.txt", dc.want))
							if err != nil {
								t.Fatalf("readfile: %v", err)
							}
							want := string(bs)

							if got != want {
								// Generate a unified diff with friendlier output than cmp.Diff
								edits := myers.ComputeEdits(span.URI("want"), want, got)
								change := gotextdiff.ToUnified("want", "got", want, edits)
								t.Errorf("unexpected diff for expected disk %q:\n%s", dc.want, change)
							}
						})
					}
				})
			}
		})
	}
}

func TestRootpw(t *testing.T) {
	testCases := []struct {
		name       string
		customData interface{}
		want       string
	}{
		{
			"instance password used",
			nil,
			"insecure",
		},
		{
			"CustomData password used",
			map[string]interface{}{"rootpwcrypt": "override"},
			"override",
		},
		{
			"bad CustomData not used",
			[]string{"test"},
			"insecure",
		},
		{
			"CustomData not used if value is not string",
			map[string]interface{}{"rootpwcrypt": 4},
			"insecure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := job.NewMock(t, "some.slug", "test-facility")

			m.SetPassword("insecure")
			m.SetCustomData(tc.customData)

			var w strings.Builder
			genKickstart(m.Job(), &w)

			got := w.String()
			want := fmt.Sprintf("rootpw --iscrypted %s", tc.want)

			if !strings.Contains(got, want) {
				t.Errorf("expected root password not set, expected %s in %s", want, got)
			}
		})
	}
}
