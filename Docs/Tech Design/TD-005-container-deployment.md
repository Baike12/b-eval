# TD-005: Container Deployment to the Cloud Host

| Field | Value |
|---|---|
| **ID** | TD-005 |
| **Status** | Implemented |
| **Implements** | [FEAT-008 Delivery Roadmap and Acceptance Criteria](../Features/platform/delivery-roadmap-and-acceptance-criteria.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the platform is deployed and run on the target host. The *what* and *why* live in the Feature documents it implements.

## 1. Problem Restatement

The platform is a Go binary with an embedded single-page application and a file-backed record store. It runs on a developer machine with `go run ./cmd/b-eval serve`. It has never run anywhere else, so nothing is known about how it behaves when it is packaged, started, restarted, or reached over a network that is not the developer's loopback interface.

The target is a cloud host that already runs four other services. The deployment has to coexist with them, and it has to be reversible.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| The host stays a shared host | The deployment takes one path prefix and touches no existing route |
| No new port is exposed | The platform is reached on the port nginx already serves |
| No toolchain on the server | The binary is cross-compiled on the developer machine and shipped as a file |
| The deployment is reproducible | The image is built from a Dockerfile in the repository, not from an ad-hoc command |
| The data survives a restart | The record store is a bind mount, not a container layer |
| The container is not a shell | The root filesystem is read-only, the container holds no capabilities and runs as an unprivileged user |
| A rollback is one command | Every image is tagged, the three most recent tags are kept, and the container is recreated from a tag |

## 3. Non-Goals

- The deployment does not terminate TLS. The host has no domain and no certificate authority; the platform is reached over the plain HTTP that nginx already serves.
- The deployment does not run the quality gate. `b-eval gate` needs the repository and the Go toolchain, so the gate stays a pipeline step on the developer machine or in CI.
- The deployment does not replace the file-backed store with a database. The store's limits are listed in [TD-004][td-004] and are unchanged by this document.
- The deployment does not introduce a Kubernetes cluster. See §11.

## 4. The Target Host

Every value in this section was measured on the host, not assumed.

| Property | Value | Consequence |
|---|---|---|
| Operating system | Ubuntu 24.04 LTS, x86_64 | The image is built for `linux/amd64` |
| CPU and memory | 4 vCPU, 3.8 GiB RAM | The platform needs one core and under 50 MiB |
| Disk | 40 GB, 25 GB free | The image and the build context fit without pressure |
| Docker | 29.1.3, Compose 2.40.3 | A plain `docker build` and a Compose project are available |
| `registry-1.docker.io` | Unreachable | The base image comes from a registry mirror |
| `docker.m.daocloud.io` | Reachable | The base image is fetched through it |
| `dl-cdn.alpinelinux.org` | Reachable | `apk` resolves during the image build |
| Cloud security group | Opens 80 and 443 only | The platform is served on port 80; no new rule is needed |
| nginx on port 80 | Catch-all for `iteach`; `/blog/` routes to `bblog` | The platform takes a path prefix, not a second catch-all |
| nginx on port 443 | `default_server` for `LexQA` | A new `default_server` on 443 is not free |
| Existing containers | `LexQA-app`, `LexQA-ui`, `LexQA-postgres`, `bblog` | The deployment must not restart them |

## 5. The Image

The runtime image carries the binary and nothing else.

```
alpine:3.20
  └── ca-certificates, tzdata
  └── user beval, UID 10001
  └── /app/b-eval          the cross-compiled static binary
```

**Why the frontend is not in the image.** The Go embed directive puts the built bundle inside the binary, so the image needs no Node runtime, no static-file server and no second process. The image is one file plus a base layer.

**Why the base image is a build argument.** `registry-1.docker.io` is unreachable from this host, so `FROM alpine:3.20` cannot resolve there. The base image is fetched once through `docker.m.daocloud.io` and retagged to the name the Dockerfile expects, which keeps the Dockerfile free of a mirror-specific name. The argument remains so that the build works unchanged on a host that can reach Docker Hub.

**Why `ca-certificates` is in the image.** The HTTP adapter reaches the agent under test over HTTPS. Without the certificate bundle, that request fails with a TLS verification error inside the container while the same request succeeds on the host.

**Why the container runs as UID 10001.** The host's other services use the same unprivileged UID, and the record store is written by the container. A root-owned data directory would let the container write as root and leave files that the next deployment cannot replace.

**Why the root filesystem is read-only.** The platform writes to exactly one path, the data directory, which is a bind mount. Everything else in the image can be read-only.

## 6. The Container

| Setting | Value | Reason |
|---|---|---|
| Image tag | `b-eval:<timestamp>` | A rollback is a re-tag and a re-create |
| Published address | `127.0.0.1:8091` | nginx is the public face; the container has no reason to be reachable itself |
| Data | `/opt/b-eval/data` → `/app/data` | The records outlive the container |
| Restart policy | `unless-stopped` | The platform comes back after a host reboot |
| User | `10001` | The container does not need root |
| Capabilities | All dropped | The platform binds a port above 1024 and needs no capability |
| Root filesystem | Read-only | Only the data mount is writable |
| Health check | `GET /api/healthz` every 30 s | Docker reports the container's real state |
| Log driver | `json-file`, 10 MB × 3 | A runaway log cannot fill the disk |

## 7. The Network Decision

The platform must be reachable, and the host does not have a free place to put it. Three routes were considered.

**A second catch-all on port 80.** Rejected. The catch-all on port 80 belongs to `iteach`, and a second `server_name _` block conflicts with it.

**A new `default_server` on port 443.** Rejected. `LexQA` holds that slot, and taking it would move a service that is live.

**A path prefix on port 80.** Selected. The platform is mounted at `/eval/`, the same way `bblog` is mounted at `/blog/`, and the host's nginx routes it to the container.

The prefix was not free. The single-page application was built for the root path: the router was a `BrowserRouter` without a `basename`, the bundle was emitted with Vite's default `base` of `/`, and the API client called absolute `/api/...` paths. Mounted at `/eval/`, that bundle would have asked the host for `/assets/index-....js`, and the host would have answered with `iteach`'s assets, because `iteach` owns `/assets/`. The API would have been worse: `/api/tasksets` on port 80 is `iteach`'s backend, so the platform would have read another application's data.

The bundle is therefore built with a base path, and three things follow from it:

| What | How | With the base path `/eval/` |
|---|---|---|
| Asset URLs | Vite's `base`, set through `vite build --base` | `/eval/assets/index-....js` |
| API calls | One constant in the API client, `import.meta.env.BASE_URL` + `api` | `/eval/api/tasksets` |
| Client-side routes | The router's `basename` | `/eval/runs` |

The prefix therefore namespaces the assets and the API along with the routes, and the two applications on port 80 no longer overlap on any path.

**Why the base path is a build parameter and not a runtime one.** The bundle is a static asset with no place to read configuration from, and the server would have to know the prefix to serve the assets at it. Baking the prefix into the bundle keeps the server unchanged and the deployment a single build. The consequence is that one image serves one prefix, which is recorded in §12.

**Why the container is published on the loopback interface.** The host's nginx is the public face of the platform. Publishing the container's port on `0.0.0.0` would put a second, unrouted copy of the platform on the network for no benefit.

## 8. The Deployment Flow

```
developer machine                                   host
──────────────────                                   ────
npm run build  (BASE_PATH=/eval/)
  → web/dist
  → internal/web/dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64
  go build
  → dist/b-eval-linux-amd64  (7.2 MiB)
                    │
                    │ scp
                    ▼
                                              /opt/b-eval/
                                                Dockerfile
                                                compose.yaml
                                                nginx/b-eval.conf.template
                                                dist/b-eval-linux-amd64
                                              server_setup.sh
                                                → renders the nginx snippet
                                                → adds the include to iteach.conf
                                                → nginx -t, then reload
                                              remote_run.sh
                                                → docker build → b-eval:<tag>
                                                → docker compose up -d
                                              curl http://127.0.0.1/eval/api/healthz
```

**Why the binary is built on the developer machine.** The server has no Go toolchain and no Node toolchain, and installing them would add hundreds of megabytes to a host that runs four other services. A cross-compiled static binary is 7.2 MiB and needs no toolchain at the far end.

**Why CGO is disabled.** A cross-compiled binary has no C toolchain available, and the platform does not need one: it resolves names through the pure-Go resolver, and it links against nothing else. The result is a static ELF binary that runs on any Linux host with the same architecture.

**Why the image is built on the server rather than shipped.** The build context is one file and the base image is already on the host, so building there costs a few seconds and transfers 7.2 MiB instead of a whole image. The alternative — build locally, `docker save`, `docker load` — is the pattern the other services on this host use, and it remains available by exporting the image and loading it before `remote_run.sh` runs.

**Why the nginx change is guarded.** The snippet is written to a file and `nginx -t` runs before the reload. A configuration that does not parse would take `iteach` and `bblog` down with it, so the original site file is backed up and restored when the check fails.

**Why the tag is a timestamp.** The tag is also the build version, which the binary carries in its `version` variable. A rollback is `docker compose up -d` with an older tag in `.env`.

**Why every ssh step is retried.** The host accepts the connection and then closes a fraction of the ssh sessions it opens, and the client sees `Connection closed by <host> port 22` without an error from the far end. The property belongs to the network rather than to the deployment, and a script that stops at the first dropped session turns a five-second hiccup into a failed release. Every `ssh` and `scp` step of `scripts/deploy.sh` is therefore retried up to six times, five seconds apart, and a step that still fails ends the script with a non-zero status.

## 9. The Data

The record store writes one JSON file per record under the data directory, and the directory is a bind mount. A container that is replaced, rebuilt or pruned keeps its records.

The seeded demonstration data is written with `b-eval seed -data /app/data` **after** the container starts, because the server reads the registered agents from the data directory once, at startup. Seeding a running container therefore leaves the records on disk but not in the server's memory; a restart picks them up. The order matters and is the reason the seed is a separate step from the deployment.

## 10. Why Not a Kubernetes Cluster

The provider's managed container service is one option, and its image registry can serve Docker Hub images through a proxy repository domain. Both were evaluated.

The managed container service is rejected for this workload. The platform is one process with one volume and no scaling requirement; a cluster adds node pools, ingress, a control plane and a per-cluster cost for a service that needs none of them. The host already runs the other services as plain containers, and the deployment follows that pattern.

The registry proxy repository is not rejected on its merits — it is simply not needed. The registry mirror that the host already uses provides the one base image the runtime image needs, so a registry account and a proxy repository would add a dependency without removing a problem.

## 11. The Operator Step That Turned Out To Be Unnecessary

The first deployment published the container on `0.0.0.0:8091` and required an inbound security group rule for the port. The rule was never added, because the platform is now served on the path prefix that nginx already exposes on port 80. **No security group change is required.** The container is bound to the loopback interface and the public route is the host's existing port 80.

## 12. Open Points

- The platform runs as a single replica. The store serialises writes with an in-process lock, which is correct for one process and is not correct for two. The limitation is inherited from [TD-004][td-004] and is not made worse by this deployment.
- The image is built by the classic builder because the host has no `buildx` plugin. The Dockerfile uses no BuildKit-only syntax, so the two builders produce the same image, but the classic builder does not cache layers across builds.
- The `b-eval gate` subcommand cannot run in the runtime image, because the gate needs the repository. The gate stays outside the image.
- The base path is baked into the bundle, so the container serves the application correctly only behind the nginx route that strips the prefix. Reaching the container's port directly returns the application shell without its assets. A second prefix needs a second build.
- The prefix is `/eval`. The host's other applications use `/blog` and the root, so `/eval` is free, but a future application must pick another one.
- The ssh sessions to the host are dropped at a low rate. The deployment retries them, and a deployment that still fails after the last attempt is a host-side problem to look at rather than a script to change.

## 13. Verification

Verified by [TR-004 Container Deployment][tr-004].

## References

- Feature: [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]
- Runtime design: [TD-004 M0 Runtime and Quality Gate System][td-004]
- Test Report: [TR-004 Container Deployment][tr-004]
- Image: [`deploy/Dockerfile`](../../deploy/Dockerfile)
- Container: [`deploy/compose.yaml`](../../deploy/compose.yaml)
- Host preparation: [`deploy/server_setup.sh`](../../deploy/server_setup.sh)
- Container start: [`deploy/remote_run.sh`](../../deploy/remote_run.sh)
- nginx route: [`deploy/nginx/b-eval.conf.template`](../../deploy/nginx/b-eval.conf.template)
- Orchestration: [`scripts/deploy.sh`](../../scripts/deploy.sh)

[feat-008]: <../Features/platform/delivery-roadmap-and-acceptance-criteria.md>
[td-004]: <./TD-004-m0-runtime-and-quality-gate-system.md>
[tr-004]: <../Test Report/TR-004-container-deployment.md>
