// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"context"
	"errors"
	"testing"

	k8sExec "k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"

	"github.com/crossplane/upjet/v2/pkg/resource"
)

type importFakeTR struct {
	resource.Terraformed
}

func (importFakeTR) GetTerraformResourceType() string { return "sentry_team" }
func (importFakeTR) GetName() string                  { return "wildbit" }

func TestWorkspaceImportUsesImportID(t *testing.T) {
	cases := map[string]struct {
		terraformID string
		importID    string
		want        string
	}{
		"ImportIDDiffersFromStateID": {terraformID: "wildbit", importID: "org/wildbit", want: "org/wildbit"},
		"ImportIDEqualsStateID":      {terraformID: "wildbit", importID: "wildbit", want: "wildbit"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got []string
			fe := &testingexec.FakeExec{
				CommandScript: []testingexec.FakeCommandAction{
					func(_ string, args ...string) k8sExec.Cmd {
						got = args
						return &testingexec.FakeCmd{
							CombinedOutputScript: []testingexec.FakeAction{
								func() ([]byte, []byte, error) { return nil, nil, errors.New("stop") },
							},
						}
					},
				},
			}
			w := NewWorkspace(directory, WithExecutor(fe), WithAferoFs(fs), WithFilterFn(filterFn))
			w.terraformID = tc.terraformID
			w.importID = tc.importID
			_, _ = w.Import(context.TODO(), importFakeTR{})
			if len(got) == 0 || got[len(got)-1] != tc.want {
				t.Errorf("terraform import args = %v, want last arg %q", got, tc.want)
			}
		})
	}
}
