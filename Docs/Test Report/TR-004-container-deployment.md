# TR-004: Container Deployment — Verification Report

| Field | Value |
|---|---|
| **ID** | TR-004 |
| **Verifies** | [FEAT-008 Delivery Roadmap and Acceptance Criteria](../Features/platform/delivery-roadmap-and-acceptance-criteria.md) |
| **Technical Design** | [TD-005 Container Deployment to the Cloud Host](<../Tech Design/TD-005-container-deployment.md>) |
| **Environment** | Developer: macOS (darwin/arm64), Go 1.24.3, Node 26.5.0, Chromium via BrowserOS. Host: Ubuntu 24.04 LTS (x86_64), Docker 29.1.3, Compose 2.40.3, 4 vCPU, 3.8 GiB RAM, nginx |
| **Executed** | 2026-10-22 |
| **Result** | **Pass** — the platform builds, ships, starts, serves, evaluates and persists on the target host, and is reachable from outside it |

## 1. Scope

Verifies that the platform can be packaged as a container image, shipped to the cloud host, started there, reached from outside the host, and used for a complete evaluation, without disturbing the four services the host already runs.

Verifies the decisions taken in [TD-005][td-005]: the cross-compiled static binary, the image built on the server, the bind-mounted data directory, the path prefix served by the host's nginx, and the container hardening.

Out of scope: TLS termination, the `b-eval gate` subcommand, which needs the repository and does not run inside the image, and a multi-replica deployment.

## 2. Test Setup

| Item | Value |
|---|---|
| Target host | `root@<host>`, public address `<host>` |
| Image tag | `b-eval:20261022-eval3` |
| Base image | `alpine:3.20`, retagged from `docker.m.daocloud.io/library/alpine:3.20` |
| Binary | `dist/b-eval-linux-amd64`, built with `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` |
| Base path | `/eval/`, baked into the bundle at build time |
| Application directory | `/opt/b-eval` |
| Data directory | `/opt/b-eval/data`, owned by UID 10001 |
| Published address | `127.0.0.1:8091` on the host, `8091` in the container |
| Public route | `http://<host>/eval/`, through the host's nginx on port 80 |
| Task set | `demo` v1, 4 static-QA tasks, seeded into the mounted data directory |
| Agents under test | 2 mock agents: `demo-good` and `demo-weak` |
| Repetitions | 2 per task, so 8 trials per run |

## 3. Build and Ship

| # | Criterion | Method | Result |
|---|---|---|---|
| 1 | The binary is produced for the target architecture without a C toolchain | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` | **Pass** — 7,594,168 bytes, statically linked ELF 64-bit x86-64 executable |
| 2 | The binary carries the frontend bundle | Inspected the served `/eval/` response | **Pass** — the response is the built `index.html` with the hashed asset references, not the placeholder page |
| 3 | The base path reaches the bundle | `BASE_PATH=/eval/ npm run build` | **Pass** — `dist/index.html` references `/eval/assets/index-DFpNWdlj.js` and `/eval/assets/index-CxKdxYcq.css` |
| 4 | The deployment transfers the binary, not an image | `scp` of `dist/b-eval-linux-amd64` | **Pass** — build context 7.606 MB, image 22.5 MB on the host |
| 5 | The server needs no Go toolchain and no Node toolchain | `command -v go`, `command -v node` on the host | **Pass** — neither is installed; the deployment succeeded without them |
| 6 | The base image is obtainable on a host that cannot reach Docker Hub | `docker pull docker.m.daocloud.io/library/alpine:3.20` | **Pass** — pulled and retagged to `alpine:3.20` by `server_setup.sh` |
| 7 | The image builds from the repository's Dockerfile | `docker build --build-arg BASE_IMAGE=alpine:3.20 -t b-eval:20261022-eval3 /opt/b-eval` | **Pass** — 4 layers: minirootfs 8.47 MB, `apk` layer 2.82 MB, binary 7.61 MB, workdir 8.19 kB |

## 4. Container Behaviour

| # | Criterion | Method | Result |
|---|---|---|---|
| 8 | The container starts and reports healthy | `docker ps` after 30 s | **Pass** — `Up (healthy)` |
| 9 | The health endpoint answers | `GET /api/healthz` on the container's port | **Pass** — `{"status":"ok","task_sets":0,"time":"2026-10-22T04:28:35Z"}` |
| 10 | The health check in the image is the one that decides the state | `docker inspect` | **Pass** — `HEALTHCHECK wget -qO- http://127.0.0.1:8091/api/healthz`, status `healthy` |
| 11 | The single-page application is served from the binary | `GET /` on the container's port | **Pass** — HTTP 200, the built `index.html` |
| 12 | The hashed assets are served | `GET /assets/index-ltHlCRr0.js`, `GET /assets/index-CxKdxYcq.css` | **Pass** — HTTP 200, 289,331 bytes and 4,645 bytes, `text/javascript` and the stylesheet |
| 13 | A client-side route falls back to the application | `GET /runs` on the container's port | **Pass** — HTTP 200 |
| 14 | The API answers | `GET /api/tasksets`, `GET /api/metrics` | **Pass** — HTTP 200 with a JSON body |
| 15 | The container restarts and the platform comes back | `docker restart b-eval` | **Pass** — healthy again, agents restored from the data directory |
| 16 | The container comes back after a host reboot | `docker inspect` restart policy | **Pass** — `unless-stopped` |

## 5. Data and Isolation

| # | Criterion | Method | Result |
|---|---|---|---|
| 17 | The records are written to the host, not to the container layer | `ls /opt/b-eval/data` after seeding | **Pass** — `tasksets/demo__v1.json` 5,859 bytes, `agents/demo-good.json` 512 bytes, `agents/demo-weak.json` 462 bytes, all owned by UID 10001 |
| 18 | The registered agents are restored on start | `GET /api/agents` after the restart | **Pass** — both `demo-good` and `demo-weak` are returned with their answers |
| 19 | The container cannot write outside the data mount | `docker exec b-eval touch /app/canary` | **Pass** — `touch: /app/canary: Read-only file system` |
| 20 | The container holds no capabilities | `docker inspect` | **Pass** — `CapDrop=[ALL]`, `ReadonlyRootfs=true` |
| 21 | The container does not run as root | `docker exec b-eval id` | **Pass** — `uid=10001(beval) gid=10001(beval)` |
| 22 | The container is not published beyond the host | `docker ps` | **Pass** — `127.0.0.1:8091->8091/tcp` |
| 23 | The four existing containers keep running | `docker ps` | **Pass** — `LexQA-app` healthy, `LexQA-ui` up, `LexQA-postgres` healthy, `bblog` healthy, all with their original uptimes of 4 to 5 weeks |
| 24 | The deployment does not put the host under load | `uptime`, `docker stats` | **Pass** — load average 0.00, container CPU 0.00%, memory 4.98 MiB of 3.825 GiB |

## 6. A Complete Evaluation on the Host

| # | Criterion | Method | Result |
|---|---|---|---|
| 25 | The demonstration data is seeded into the mounted volume | `docker exec b-eval /app/b-eval seed -data /app/data` | **Pass** — `seeded task set "demo" (4 tasks) and 2 agents into /app/data` |
| 26 | An agent passes the self-check | `POST /api/agents/demo-good/selfcheck` | **Pass** — `{"status":"passed","tasks":4,"action_format_error_rate":0,"contract_violations":0,"usage_reporting_ok":true}` |
| 27 | A run executes and completes | `POST /api/runs` with 2 repetitions over the 4-task set | **Pass** — `run-plan-demo-demo-good-3408113d`, `status: completed`, 8 trials |
| 28 | The report is produced with its confidence | `GET /api/runs/{id}/report` | **Pass** — `main_score.total_score` 1, `valid_sample_size` 8, `planned_sample_size` 8, `coverage` 1, `insufficient_confidence` false, `pass_at_k` 1 |
| 29 | The run is written to the host | `ls /opt/b-eval/data/runs` | **Pass** — `run-plan-demo-demo-good-3408113d.json`, 25,676 bytes |

## 7. The Public Route

| # | Criterion | Method | Result |
|---|---|---|---|
| 30 | The platform is reachable from outside the host | `GET http://<host>/eval/` from the developer machine | **Pass** — HTTP 200, the application shell, `title` `b-eval — Agent Evaluation Platform` |
| 31 | The assets the page references load from outside the host | `GET http://<host>/eval/assets/index-DFpNWdlj.js` and the CSS | **Pass** — HTTP 200, 289,416 bytes and 4,645 bytes, `text/javascript` and the stylesheet |
| 32 | The API is reachable from outside the host | `GET http://<host>/eval/api/healthz` | **Pass** — HTTP 200, `{"status":"ok","task_sets":1,...}` |
| 33 | The application renders and reads its data | Chromium against the public URL | **Pass** — the Dashboard renders with `API ok · 1 task set(s)`, the `demo` task set with 4 tasks, both agents with their self-check states, and the completed run in *Recent runs*. No console errors. |
| 34 | The client-side routes work under the prefix | Chromium against `/eval/runs`, `/eval/task-sets`, `/eval/metrics` | **Pass** — *Runs*, *Task sets* and *Metric registry* render, and the address bar keeps the prefix |
| 35 | The host's other application keeps its own assets | `GET http://<host>/` | **Pass** — `iteach` serves its own page and its own `/assets/index-CkKcN0mL.js`. The two applications share no path. |
| 36 | The host's other application keeps its own API | `GET http://<host>/api/` | **Pass** — the request reaches `iteach`'s backend, not the platform's |
| 37 | The host's other routes are unchanged | `GET http://<host>/blog/` and `GET https://<host>/` | **Pass** — HTTP 200 on both |
| 38 | No new security group rule is needed | External port probe of 80, 443 and 8091 | **Pass** — the platform is served on port 80, which the security group already opens. Port 8091 remains closed, and the container is bound to the loopback interface, so nothing is exposed. |

## 8. Quality Gates

The repository's gates were run against the deployment artefacts.

| Gate | Result |
|---|---|
| `go-fmt`, `go-vet`, `go-test` | **Pass** |
| `web-typecheck`, `web-test`, `web-build` | **Pass** — 15 frontend tests pass |
| `go-build` | **Pass** |
| `docs-english` | **Pass** — the new shell scripts, the compose file and the documents are English-only |
| `docs-links` | **Pass** — every link in the new documents resolves |
| `docs-index` | **Pass** |
| `e2e-smoke` | **Pass** |

All eleven gates returned `passed` with the classification `ready` and the recommended action `proceed`.

## 9. Defects Found During Verification

Two issues were found. Both were in the deployment machinery, and both were fixed in the same change that found them.

**The nginx include was never installed.** The first run of the host preparation failed with `awk: fatal: cannot use gawk builtin 'include' as variable name`, because the variable that carried the include line was named `include`, which is a gawk builtin. The site file was left untouched, nginx stayed valid, and the port 80 route was simply absent. The variable was renamed to `snippet` and the run was repeated; the include was then installed and `nginx -t` passed. The failed run also left a zero-length `iteach.conf.tmp` next to the site file, which is harmless because the rename never happened.

**The seed does not register the agents in a running server.** The server reads the registered agents from the data directory once, at startup, so seeding a container that is already running leaves the records on disk but not in the server's memory. A run against a freshly seeded container fails with `agent "demo-good" is not registered` until the container is restarted.

The behaviour is correct — the agent registry is restored at startup by design — but the order of the deployment steps is not obvious. [TD-005][td-005] §9 now states that the seed runs after the container starts and that a restart picks the records up.

## 10. Conclusion

The platform deploys to the cloud host as a container and works there. The image is 22.5 MB, the container uses 4.98 MiB of memory and no measurable CPU, the data survives a restart and a rebuild, the container runs unprivileged with a read-only root filesystem, and a full evaluation — four tasks, two repetitions, eight trials, a scored report with a confidence interval — completes through the deployed instance.

It is reachable from outside the host at `http://<host>/eval/`, through the port the host's nginx already serves, and the application shares no path with `iteach`, `bblog` or `LexQA`. The four services the host already runs are untouched, and the deployment is reversible: the container is removed with `docker rm -f b-eval`, the directory with `rm -rf /opt/b-eval`, and the nginx include with one line in `/etc/nginx/sites-available/iteach.conf`.

## References

- Feature: [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]
- Technical Design: [TD-005 Container Deployment to the Cloud Host][td-005]
- Runtime design: [TD-004 M0 Runtime and Quality Gate System][td-004]

[feat-008]: <../Features/platform/delivery-roadmap-and-acceptance-criteria.md>
[td-004]: <../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>
[td-005]: <../Tech Design/TD-005-container-deployment.md>
