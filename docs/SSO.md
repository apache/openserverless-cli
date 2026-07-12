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

# Legacy embedded SSO configuration ownership

The public `ops config sso` command is supplied by the task repository. The
embedded `ops -config sso` form remains temporarily available for compatibility
with already published scripts and installations. It manages the following
configuration. Resource names can be changed with `--configmap`, `--secret`,
`--statefulset`, and `--container`.

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

`ops config sso disable` removes only the exact `envFrom` references described
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
