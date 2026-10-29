<!--
SPDX-FileCopyrightText: 2026 The Crossplane Authors <https://crossplane.io>

SPDX-License-Identifier: CC-BY-4.0
-->

# wildbitca/upjet fork patches on top of upstream v2.5.1

Branch `chore/rebase-v2.5.1` is based on the real upstream tag `v2.5.1` and
carries only the patches upstream does not have. Tag: `v2.5.1-wb.1`.

| # | Patch | Files | Upstream status |
|---|-------|-------|-----------------|
| 1 | Numeric state ids (`NormalizeStateID`, `IDAsExternalName`, `isStateEmpty`) | `pkg/resource/stateid.go`, `pkg/controller/external.go`, `pkg/config/resource.go`, `pkg/terraform/files.go` | Not upstream |
| 2 | Skip the CEL rule for required dynamic parameters | `pkg/types/builder.go` | Not upstream |
| 3 | Skip synthetic TF state when the external ID is empty | `pkg/terraform/files.go` | Not upstream |
| - | `GetImportIDFn` | - | Dropped, upstream has it (v2.4.0, #677) |
| - | DynamicPseudoType envelope unwrap (v2.2.2) | - | Omitted, see below |

## Omitted: dynamic envelope unwrap

The patch targets plugin-framework providers returning `{type, value}`
envelopes (Cloudflare v5). provider-upjet-cloudflare registers its resources
through `WithIncludeList` (Terraform CLI runtime), where the state is JSON and
never contains envelopes, and it canonicalizes zone-setting values itself
(`config/zonesetting/canonicalize.go`). There is no evidence the patch is needed
on v2.5.1, so it is not carried. Re-add it if a plugin-framework consumer shows
the permanent forProvider/atProvider diff.

## Draft upstream issue/PR: numeric state ids (do NOT file yet)

**Title:** CLI runtime fails on resources whose Terraform `id` is a number

Plugin-framework providers may declare `id` as a number. Terraform then writes
a numeric id into the state, while upjet reads it as a string only. The second
reconcile fails with `cannot work with a non-string id: 577654`, and decoding
the observation fails with `ReadString: expects " or n, but found 5`. In
provider-upjet-bunnynet 16 of 24 resources are affected.

Proposal: accept float64 ids in `IDAsExternalName` and `isStateEmpty` (0 counts
as empty), and add `resource.NormalizeStateID` which rewrites a numeric id to
its string form before `SetObservation` in `Observe` and `Import`. Ids above
2^53 or fractional are rejected, since they cannot round-trip through float64.
Only widens accepted input; the string path is untouched. Tests:
`pkg/resource/stateid_test.go`, `pkg/terraform/files_test.go`.

## Draft upstream issue/PR: CEL rule for required dynamic parameters (do NOT file yet)

**Title:** Generated CEL rule for a required dynamic parameter makes the CRD invalid

A required top-level dynamic field is generated as `*apiextensionsv1.JSON`,
rendered as `x-kubernetes-preserve-unknown-fields` without `type`. The CEL type
checker cannot declare an untyped field, so the generated required-parameter
rule fails with `undefined field` and the API server rejects the CRD (seen with
`cloudflare_zone_setting.value`, MRD stuck `Established=False`).

Proposal: do not emit the CEL rule for required dynamic parameters in
`pkg/types/builder.go`; Terraform still enforces the field at apply time. Test
in `pkg/types/builder_test.go`.

## Draft upstream issue/PR: skip synthetic state without external id (do NOT file yet)

With an empty external name, `EnsureTFState` wrote a synthetic state with all
parameters but no id. Plugin-framework providers fail `Read` without an id, so
`Create` was never reached. Proposal: leave the state empty when `tfID == ""`
so Refresh reports the resource as absent and Create proceeds.
