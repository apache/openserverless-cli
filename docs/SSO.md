<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one
  ~ or more contributor license agreements.  See the NOTICE file
  ~ distributed with this work for additional information
  ~ regarding copyright ownership.  The ASF licenses this file
  ~ to you under the Apache License, Version 2.0 (the
  ~ "License"); you may not use this file except in compliance
  ~ with the License.  You may obtain a copy of the License at
  ~
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing,
  ~ software distributed under the License is distributed on an
  ~ "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  ~ KIND, either express or implied.  See the License for the
  ~ specific language governing permissions and limitations
  ~ under the License.
  -->

# SSO command ownership and 0.9.0 prerequisites

The public `ops config sso` command requires SSO tasks from
`apache/openserverless-task` and an OIDC-capable `apache/openserverless-admin-api`.
At the branch snapshots below, those components are missing from `0.9.0`, even
though the CLI already contains the SSO configuration and login code from
`0.9.1`. Updating the CLI alone does not enable SSO on a `0.9.0` installation.

## Command ownership

| Command | Implementation | Responsibility |
| --- | --- | --- |
| `ops config sso ...` | Task repository: `config/sso/docopts.md`, `config/sso/opsfile.yml`, `config/sso/sso.ts` | Public configuration command, executed with Bun and kubectl. |
| `ops -config sso ...` | CLI: [config/config_tool.go](../config/config_tool.go), [config/sso_tool.go](../config/sso_tool.go) | Legacy embedded configuration tool. |
| `ops -login ...` | CLI: [auth/login.go](../auth/login.go) | SSO device and password login through admin-api, alongside legacy login. |
| `ops ide login ...` | Task repository: `ide/opsfile.yml` | Workspace setup and delegation to the embedded login tool. |

The CLI routes `ops config sso` through its normal task lookup. It does not
fall back to the embedded configuration tool when the SSO task is missing.
The embedded form remains temporarily available for compatibility with
already published scripts and installations; it still requires an
OIDC-capable admin-api for SSO login to work.

## Branch comparison

Checked on 2026-09-10. The links pin the reviewed commits because branch heads
can change independently across repositories.

| Repository | `0.9.0` snapshot | `0.9.1` snapshot | SSO status in `0.9.0` |
| --- | --- | --- | --- |
| `apache/openserverless-cli` | [83075cbc](https://github.com/apache/openserverless-cli/tree/83075cbc890dd9692782e27340fd2eee6b498a59) | [a0dd56cf](https://github.com/apache/openserverless-cli/tree/a0dd56cf5667acb092cfbac14de531f5de139af0) | `config/` and `auth/`, including their tests, are identical to `0.9.1`. No SSO code backport is needed in the CLI. |
| `apache/openserverless-task` | [eeac57e4](https://github.com/apache/openserverless-task/tree/eeac57e4179381ee18fba290ab79d6cec1dafb09) | [71bd01f6](https://github.com/apache/openserverless-task/tree/71bd01f649d4f6b551240b7a3fa1db47954f4c6b) | Missing `config/sso/`, SSO help discovery, `admin/sso/`, and SSO changes to IDE login and user lifecycle tasks. |
| `apache/openserverless-admin-api` | [161c479f](https://github.com/apache/openserverless-admin-api/tree/161c479f318f932e58eced08acd0830e7e75b288) | [35f9a1bc](https://github.com/apache/openserverless-admin-api/tree/35f9a1bcac5f75aab6dfac07d7efa6ac1bd0ae56) | Missing the OIDC bridge, device/password endpoints, SSO namespace mapping and login-triggered provisioning. |
| `apache/openserverless-testing` | [be7b3338](https://github.com/apache/openserverless-testing/tree/be7b33382e02c6543d3e8813724b0177fae1c18d) | [e017ff5f](https://github.com/apache/openserverless-testing/tree/e017ff5fb5b1147cdb772aa584223d6e0d6c83da) | Missing `tests/11-sso-mock.sh` and `tests/mock-oidc-provider.py`. |

The external components to review for a separate backport are:

- **Configuration tasks:** `config/sso/` (including `sso.test.ts`) and the SSO
  entries in `config/docopts.md`. The task uses Bun and kubectl from the task
  prerequisites.
- **Login and lifecycle tasks:** SSO handling in `ide/opsfile.yml`,
  `ide/docopts.md`, `admin/opsfile.yml`, `admin/docopts.md`, `admin/sso/`, and
  SSO display columns in `setup/kubernetes/crds/whisk-user-crd.yaml`.
- **Admin-api:** `openserverless/common/oidc_validator.py`,
  `openserverless/common/sso_namespace.py`,
  `openserverless/impl/auth/oidc_device_flow_service.py`, the SSO additions in
  `openserverless/impl/auth/auth_service.py` and `openserverless/rest/auth.py`,
  their dependencies, configuration documentation and tests. These supply
  `/auth/oidc`, `/auth/oidc/device/start`,
  `/auth/oidc/device/poll`, and `/auth/oidc/password`, all under
  `/system/api/v1`.
- **Integration tests:** the mock provider and SSO smoke script in the testing
  repository. They complement the CLI unit tests with a deployed admin-api
  and Kubernetes resources.

Setting local `SSO_*` values or creating a ConfigMap cannot add the missing
server endpoints. A coordinated task/admin-api backport and integration
validation are required before the public SSO flow can be considered usable
with `0.9.0`.

With the `0.9.0` task snapshot above, `ops config sso --help` and
`ops config sso show` exit with `no command named sso found`.
`ops -config sso --help` still succeeds because that command is embedded in
the CLI. The same CLI can resolve `ops config sso --help` with the `0.9.1`
tasks; this verifies command discovery, not server-side SSO compatibility.

## Kubernetes compatibility

The `0.9.0` deployment manifests and the SSO defaults use different names:

| Resource | `0.9.0` deployment | Embedded SSO / `0.9.1` deployment |
| --- | --- | --- |
| Namespace | `openserverless` | `openserverless` |
| Admin-api StatefulSet and container | `openserverless-system-api` | `openserverless-system-api` |
| `WhiskUser` API version | `openserverless.org/v1` | `openserverless.org/v1` in the `0.9.1` admin-api/operator |

See the task repository's
[0.9.0 admin-api template](https://github.com/apache/openserverless-task/blob/eeac57e4179381ee18fba290ab79d6cec1dafb09/setup/openserverless/system-api/api-template.yaml)
and the operator's
[0.9.0 WhiskUser CRD](https://github.com/apache/openserverless-operator/blob/5df76feda569b374276f4e2c72a46d3a980a4b77/deploy/openserverless-permissions/whisk-user-crd.yaml).
The corresponding
[0.9.1 CRD](https://github.com/apache/openserverless-operator/blob/fdcbbdf2332ae5aa8e22afba0704b37fdb1697bd/deploy/openserverless-permissions/whisk-user-crd.yaml)
uses `openserverless.org`.

`--namespace`, `--statefulset`, and `--container` select the resources patched
by the configuration command. They do not change the namespace or API group
used internally by the
[0.9.1 admin-api provisioning code](https://github.com/apache/openserverless-admin-api/blob/35f9a1bcac5f75aab6dfac07d7efa6ac1bd0ae56/openserverless/impl/auth/auth_service.py)
and its
[Kubernetes client](https://github.com/apache/openserverless-admin-api/blob/35f9a1bcac5f75aab6dfac07d7efa6ac1bd0ae56/openserverless/common/kube_api_client.py).
A future backport must align those values with the deployed operator and CRD;
copying the `0.9.1` image or tasks without adaptation is insufficient.

The external-IdP flow uses the operator's existing `WhiskUser` reconciliation.
Configuring SSO does not install Keycloak or add a Keycloak reconciler to the
operator. An external Keycloak/IdP must already be available.

## Legacy embedded configuration ownership

`ops -config sso` manages the configuration described below. Resource names
can be changed with `--namespace`, `--configmap`, `--secret`, `--statefulset`,
and `--container`.

## Kubernetes resources

The command creates and owns a dedicated ConfigMap. Its default name is
`openserverless-sso-config`, and it contains exactly these keys:

- `OIDC_ISSUER_URL`
- `OIDC_JWKS_URL`
- `OIDC_AUDIENCE`
- `OIDC_CLIENT_ID`
- `OIDC_REQUIRED_GROUP`
- `OIDC_USERNAME_CLAIM`
- `OIDC_GROUPS_CLAIM`
- `SSO_AUTOPROVISION_ON_LOGIN`
- `SSO_AUTOPROVISION_TIMEOUT_SECONDS`
- `SSO_AUTOPROVISION_POLL_SECONDS`
- `SSO_AUTOPROVISION_DEFAULT_SERVICES`
- `SSO_NAMESPACE_PRESERVE_VALID`
- `SSO_NAMESPACE_HASH_LENGTH`
- `SSO_NAMESPACE_MAX_LENGTH`

When `--client-secret` is supplied, the command also creates and owns a
dedicated Secret. Its default name is `openserverless-sso-secret`, and its only
managed key is `OIDC_CLIENT_SECRET`.

The command adds exact, prefix-free `envFrom` references for the managed
ConfigMap and, when applicable, the managed Secret to the selected admin-api
container. It does not create direct `env` entries, volumes, volume mounts, or
annotations.

## Disable behavior

`ops -config sso disable` removes only the exact `envFrom` references described
above and deletes the two dedicated resources with Kubernetes
`--ignore-not-found`. Other `envFrom` entries, all direct `env` entries,
volumes, volume mounts, and existing annotations remain unchanged.

The command is idempotent. If neither managed reference is present, no patch or
rollout command is issued for the StatefulSet. Missing ConfigMaps and Secrets
are not errors. Removing an `envFrom` entry changes the pod template, so
Kubernetes starts the necessary rollout itself. By default the command waits
for that rollout; `--no-rollout` skips the wait and does not issue an additional
restart.

The local `~/.ops/config.json` cleanup is also limited to the keys written by
the SSO command. Unrecognized `SSO_*` keys are preserved.
