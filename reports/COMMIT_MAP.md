# Commit hash map (history rewrite of 2026-10-11)

On 2026-10-11 the git history of this repository was rewritten with `git filter-repo` to remove the AI co-author trailer (`Co-Authored-By: Claude …`) from commit messages. Nothing else changed: every rewritten commit has **the same tree (file contents), the same author and committer names, e-mails and dates (to the second, with time zone), the same parents (mapped) and the same commit order** as before; its message is identical except for the removed trailer line. Because a commit hash covers its message and its parents, the trailer commits and every commit descending from them received new hashes. Dates are unchanged, so the order of commits (for example the v2 rules in `docs/PRIORITY_ADMISSION_V2.md` committed before the v2 implementation) can still be checked by date.

- **137 commits were rewritten** (table below, oldest first). The other 55 commits (the early history, the GitHub web edits to README.md on `main`, and branch `frost-integration`) kept their hashes; the GitHub-signed web edits keep their signatures.
- **Results directory names under `benchmark/results/` (and `reports/repro/`, `reports/gates/`, `reports/multihost/`) keep their pre-rewrite short hashes and are NOT renamed.** The same holds for raw data and generated outputs that record a commit: `summary.md`, `env.json`, run logs, deploy manifests, `reports/REPRO.md`, `reports/INDEPENDENCE.md`, `reports/audit/raw/`. Use this table to translate them.
- Commit messages still quote old hashes verbatim (only the trailer was removed).
- In the current tree, references to rewritten commits in NOTES.md, README.md, docs/ and the hand-written reports (reports/*.md, reports/audit/*.md, reports/aws/7c*/ READMEs) were updated to the new hashes in one commit made after the rewrite. Hashes that are part of a file or directory name, quotes of raw file contents and command transcripts in code blocks keep the old value. `docs/instructions/` (verbatim instruction texts) was not changed.
- `reports/HISTORY_PURGE.md` §0 (2026-10-04, "the history is not rewritten") was superseded by this rewrite for the co-author trailers only; its other content is unchanged.
- GitHub may serve the old commits by their SHA until it garbage-collects them; the pre-rewrite history is kept in a local mirror backup by the maintainer.

| date (author) | old | new | subject |
|---|---|---|---|
| 2026-09-24 19:08:44 +0530 | `8dac096` | `4f42f09` | spike: Phase 0 threshold RSA (tcrsa) RS256 JWT decision gate |
| 2026-09-24 19:32:29 +0530 | `e0f4bf1` | `4d6c180` | chore(security): Phase 1 secret hygiene |
| 2026-09-24 19:49:44 +0530 | `d75de73` | `fddbf25` | feat(dealer): Phase 2 trusted-dealer threshold RSA key ceremony |
| 2026-09-24 20:01:59 +0530 | `7b08b85` | `b9fb7ea` | feat(signer): Phase 3 threshold RSA signer with independent claims policy |
| 2026-09-24 20:30:25 +0530 | `c1d02e9` | `9678171` | feat(coordinator): Phase 4 threshold coordinator (ExternalJWTSigner v1) |
| 2026-09-24 20:31:14 +0530 | `0885af0` | `01d6424` | chore(legacy): Phase 5 FROST excluded from runtime, document why |
| 2026-09-24 21:00:48 +0530 | `ae51ebd` | `d54ef0d` | docs(notes): highlight N14 JSON parser differential finding |
| 2026-09-24 21:27:13 +0530 | `1f3bbe0` | `c7d78d6` | test: Phase 6 correctness/isolation suite (T1-T12), testmalicious signer, check-images |
| 2026-09-24 21:31:34 +0530 | `8d23d9f` | `32eab39` | test(e2e): kind v1.36.5 e2e harness, Makefile, static-IP compose |
| 2026-09-24 21:50:12 +0530 | `70e182a` | `c4e1525` | fix(check-images): anchor history pattern to build-context paths |
| 2026-09-24 21:50:29 +0530 | `d12ea63` | `ef51a88` | check-images: allow image overrides (negative control) |
| 2026-09-24 21:50:46 +0530 | `2e05943` | `b5450e1` | docs(notes): Phase 6 environment, node image, token-expiration source analysis |
| 2026-09-24 21:55:33 +0530 | `f077169` | `cd9c4ef` | fix(e2e): kube-apiserver patch indices for the overridden-issuer layout |
| 2026-09-24 22:26:23 +0530 | `2f2d7bc` | `4ac126d` | fix(e2e): wait on the restart PID, not all jobs (tee process substitution) |
| 2026-09-24 22:29:05 +0530 | `ab2de56` | `38cc4a1` | fix(e2e): no log prefix for coordinator logs; assert REQ-d strict strategy |
| 2026-09-24 22:44:07 +0530 | `0272318` | `6c9bfba` | fix(e2e,test): wipe generated secrets after e2e; T11 walks the tree for key files |
| 2026-09-24 22:44:23 +0530 | `152941b` | `89d73e8` | test(T11): public-meta.json is public, not key material |
| 2026-09-24 22:53:46 +0530 | `a04c17e` | `d5f71dd` | docs(gate6): Phase 6 evidence logs, e2e run history, T13 negative control |
| 2026-09-25 00:15:28 +0530 | `78384e9` | `c7bfab2` | docs: add CLAUDE_CODE_ADDENDUM_make-it-solid.md (verbatim, from user) |
| 2026-09-25 00:25:06 +0530 | `c8abdf3` | `bc399d5` | feat(security): Phase 6.5 lock down who can call Sign; generic signing errors (N33) |
| 2026-09-25 01:36:47 +0530 | `50985da` | `55e5631` | test(T3): assert generic caller error and detailed coordinator log (N33) |
| 2026-09-25 01:39:46 +0530 | `b2179f0` | `528a62f` | fix(security): unix socket 0600 root:root (nginx umask 077); N2 asserts it |
| 2026-09-25 01:44:01 +0530 | `468cd1c` | `309a031` | fix(security): restrict signer socket via root:root 0700 parent dir |
| 2026-09-25 01:55:27 +0530 | `4d62b42` | `0656d61` | docs(gate6.5): evidence logs and run history |
| 2026-09-25 02:15:31 +0530 | `f39dd68` | `6bc0c7c` | feat(7B): multi-host signer deployment (Level 1, local multi-VM) |
| 2026-09-25 02:26:07 +0530 | `90a01bb` | `1f5163d` | fix(7B e2e): L3 checks read the VM-side exit code; alarms on remote calls |
| 2026-09-25 02:31:17 +0530 | `4432322` | `d29675c` | 7B: teardown script, multihost README, N38/N39 notes |
| 2026-09-25 02:38:05 +0530 | `45dfe68` | `1b1dfaf` | docs(notes): N40 node image rebuilt after my prune |
| 2026-09-25 08:50:50 +0530 | `cc1f6db` | `b712da2` | fix(7B bench): caffeinate, alarms on remote calls, host-sleep detection |
| 2026-09-25 14:59:36 +0530 | `28b8872` | `73d9737` | bench: record host power source and lid state |
| 2026-09-25 15:12:17 +0530 | `4385a48` | `90846a6` | fix(multihost scripts): never give multipass exec /dev/null (client hangs) |
| 2026-09-25 15:13:50 +0530 | `9bb1f16` | `69dca2e` | fix(bench): set -e trap in label assignment at L=0 |
| 2026-09-25 15:19:18 +0530 | `1af5be3` | `7130fb9` | docs(notes): N43 overload at c=50, N44 clock skew after host sleep |
| 2026-09-25 15:33:31 +0530 | `dc9b396` | `28e61b1` | 7B Level 1: benchmark results, INDEPENDENCE.md, multihost e2e evidence |
| 2026-09-25 15:33:57 +0530 | `6ebb013` | `902119a` | chore: remove accidentally committed summarize binary; build bench tools in a temp dir |
| 2026-09-25 15:55:24 +0530 | `b91bf19` | `c8173d8` | fix(N43/N46): signer cancellation checks, admission control, hedged fan-out |
| 2026-09-25 16:01:58 +0530 | `2a75c2f` | `47452bf` | fix(bench): define signer_maxconc before env.json uses it |
| 2026-09-25 16:16:32 +0530 | `713776d` | `89f9456` | 7B post-N43-fix benchmark (PARTIAL): target not met; N47 analysis |
| 2026-09-25 16:46:09 +0530 | `4ff02d3` | `3790e5e` | feat(N48): deadline-aware bounded admission; goodput metrics; N49 target |
| 2026-09-25 17:18:52 +0530 | `aa223ca` | `8131111` | 7B: mandatory clock check (N50); validation deferred to 7A (N51) |
| 2026-09-25 17:24:19 +0530 | `af297c9` | `5db47a2` | docs: Phase 8 README, THREAT_MODEL, CLAIMS_AUDIT, NOTICE |
| 2026-09-25 20:35:11 +0530 | `3d91818` | `dc7d923` | Phase 10: CI (vet, staticcheck, govulncheck, gitleaks, test, kind e2e), make repro, pinned deps |
| 2026-09-25 20:55:24 +0530 | `ea186e4` | `e4c3dc6` | legacy: track cmd/testsign as a file, not a stray gitlink |
| 2026-09-25 21:21:39 +0530 | `a3bdb38` | `ea657e0` | repro: workflow_dispatch job running make repro on a fresh runner; detect existing Docker Engine |
| 2026-09-25 21:22:12 +0530 | `aef95d5` | `fc3a934` | NOTES N52 (GO-2026-6443 acceptance expiry), N53 (repro on fresh runner); repro.yml push trigger |
| 2026-09-25 21:35:04 +0530 | `32c5952` | `20474e7` | Gate 10: REPRO.md from make repro on a fresh GitHub runner (ubuntu-24.04, amd64): PASS |
| 2026-09-25 21:35:04 +0530 | `87b0c23` | `ad9006a` | Phase 11: REVIEW_RESPONSE (R1-R18), PAPER_INPUTS, THREAT_MODEL §7 C1-C7 mapped to existing tests |
| 2026-09-26 22:00:53 +0530 | `27ebe1b` | `f27d2a7` | Phase 7A tooling: AWS provisioning, B1 baseline, single-host driver, 7A summary |
| 2026-09-26 22:12:36 +0530 | `c46daba` | `2cf9ced` | N56: fix racy server-side count in TestHedgedContactsTPlusOneAndRotates; keep EC2 repro attempt 1 (FAIL) |
| 2026-09-26 22:26:40 +0530 | `29abc19` | `c948117` | 7A: pod scale-up benchmark (B0/B1/T strict-all, 0->50/100 pods, 3 reps) + audit-based token latency |
| 2026-09-26 22:27:00 +0530 | `a319464` | `4eb9542` | 7A: move apiserver audit policy out of the ignored audit/ path |
| 2026-09-26 22:32:31 +0530 | `f9c4c58` | `f7a85b5` | repro: build node image from a resumable, sha256-verified official tarball (N55) |
| 2026-09-26 22:45:38 +0530 | `92e2d9b` | `95de434` | 7A: make repro on EC2 m7i-flex.large ap-south-1: PASS (attempt 3, f9c4c58) |
| 2026-09-26 22:46:01 +0530 | `4a918b6` | `6a6f592` | 7A driver: put the pinned Go on PATH (non-login shells) |
| 2026-09-26 22:49:01 +0530 | `8ebb806` | `b1d6de9` | 7A driver: re-pad base64url JWT header in token_kid; report failing line |
| 2026-09-26 23:07:23 +0530 | `efb76a5` | `c022b41` | 7B prep: ssh transport for multihost scripts, Level 2 topology, run-l2.sh, scale-remote.sh |
| 2026-09-26 23:07:32 +0530 | `88473a5` | `0ba3bac` | remove accidentally committed __pycache__; ignore it |
| 2026-09-26 23:08:26 +0530 | `64e9557` | `4d503fe` | 7B prep: Level 2 INDEPENDENCE.md generator (results read from the e2e log) |
| 2026-09-26 23:08:54 +0530 | `6f013dd` | `7c13be4` | NOTES N57 (Level 2 SSH /32 at the SG layer), N58 (ssh transport, TCP-connect RTT) |
| 2026-09-27 00:31:50 +0530 | `c76eee9` | `cc1ba0e` | Phase 7A results: B0/B1/T x 4 T variants x c=1/10/20/50, 3 runs + pod scale-up (m7i-flex.large, ap-south-1) |
| 2026-09-27 00:35:15 +0530 | `c88ca3c` | `8679347` | NOTES N59: 7A kubelet retries, client-side latency beyond the deadline |
| 2026-09-27 00:38:35 +0530 | `b2f63ca` | `5face06` | 7B: strict+optimistic, client vs coordinator latency columns, scale-up with load-based cooldown |
| 2026-09-27 01:12:09 +0530 | `ebabdb1` | `d8ce146` | 7B: Level 2 multihost e2e evidence (L2 invalid: operator network drop), --recheck-l2, chrony-primary clock check |
| 2026-09-27 01:15:02 +0530 | `8be7ca6` | `7820379` | 7B: network-robust run-l2 (detached remote work, INVALID + one re-run on any ssh failure), per-rep scale-up |
| 2026-09-27 01:19:41 +0530 | `0d1aaaf` | `34b785c` | NOTES N60: operator network drop during 7B deploy, L2 recheck PASS, chrony-primary clock check, INVALID rule |
| 2026-09-27 02:48:49 +0530 | `45ff8fb` | `1995319` | Phase 7B results (Level 2, 5 AWS regions) + INDEPENDENCE.md; track evidence logs |
| 2026-09-27 03:14:06 +0530 | `a0d3645` | `230c277` | AWS teardown (zero tagged resources, 5 regions); docs with 7A/7B numbers; NOTES N62 (no 7B re-run, 7C), N63 (revised overload analysis) |
| 2026-09-27 03:14:43 +0530 | `3f345b7` | `ebd02e0` | INDEPENDENCE.md: cite the committed SG rule log (reports/aws/sg-rules.txt) |
| 2026-09-27 03:36:34 +0530 | `bfab4d9` | `5478525` | Overload fix c1, c2, (a): timing instrumentation, verify-until-t, optimistic default with suspicion breaker |
| 2026-09-27 03:44:55 +0530 | `c6af356` | `f08ef84` | Evidence: overload-fix make test (local) and CI kind e2e logs (bfab4d9) |
| 2026-09-27 13:36:10 +0530 | `db30b50` | `1fd5dba` | NOTES N65: breaker bound is per coordinator replica and resets on restart; CI run accepted as kind validation |
| 2026-09-27 14:11:25 +0530 | `1e94f1a` | `75da90a` | NOTES N66: ap-south-1 vCPU quota request d54be0b2... (16 -> 24), decision and timing |
| 2026-09-27 14:27:12 +0530 | `e6758cb` | `1958aa7` | NOTES N66: quota approved (24 vCPU, case 179049863300507); corrected fallback analysis |
| 2026-09-27 14:43:38 +0530 | `d868153` | `927f621` | Phase 7C tooling: kubeadm cluster, system switching, 7C driver and summary (N67) |
| 2026-09-27 14:49:14 +0530 | `ddcfe31` | `559e499` | provision-7c: SG descriptions without '>'; quota plan counts only unlaunched roles |
| 2026-09-27 15:23:27 +0530 | `d5c4ed0` | `349cdfe` | clock-check: resync + sync wait in parallel on every host; one measurement call per host |
| 2026-09-27 15:25:45 +0530 | `87c9ec5` | `7c9584f` | 7C bootstrap: new ssh session after joining the docker group (drop_master) |
| 2026-09-27 15:37:47 +0530 | `02ae373` | `799e8e9` | clock-check: retry the idempotent resync on ssh connection failures (exit 255) |
| 2026-09-27 16:24:15 +0530 | `4fd66c0` | `a0eb1a0` | 7C lib: show the switch output when a system switch fails |
| 2026-09-27 17:18:45 +0530 | `15e8b2b` | `b9137fa` | 7C switch: wait for kube-system to settle without racing the rollout (NotFound) |
| 2026-09-27 17:47:30 +0530 | `a0173b4` | `bcf4405` | 7C bootstrap evidence: all six systems pass mode/kid/TokenReview checks; deploy manifests (T-5, T-same) |
| 2026-09-27 18:01:18 +0530 | `97a60c2` | `cb751c2` | 7C run.sh: missing files give an empty array (pipefail made '\|\| echo []' emit twice) |
| 2026-09-27 18:01:40 +0530 | `7faa09c` | `9b892f3` | 7C run.sh: absolute or relative RES for the summary |
| 2026-09-27 18:03:00 +0530 | `8d7755a` | `e7a1702` | NOTES N68: 7C bring-up findings and fixes |
| 2026-09-27 18:43:34 +0530 | `73be90f` | `a1f2393` | 7C: 25-min ssh alarm for system switches, one retry, switch duration recorded |
| 2026-09-27 20:39:20 +0530 | `0ee7228` | `6d87aaa` | Phase 7C token benchmark results (3 runs x 6 systems x c=1/10/20/50) + fixes |
| 2026-09-27 20:58:28 +0530 | `9a5e7d4` | `f5e2fe8` | 7C: nginx lines from all kubelet log files (rotated + gz); lever and stress phases; signer CPU + audit capture; stress summary |
| 2026-09-27 21:02:45 +0530 | `07bf7d8` | `a16b340` | NOTES N70: Little's law at c=50, signer-compute attribution, nginx capture fix verified, lever/stress design and pre-registered rule |
| 2026-09-27 21:23:10 +0530 | `1e77bee` | `13237ba` | 7C scale-rep: fit check after the previous rep's pods are gone (N71); scale phase restarted |
| 2026-09-28 00:12:29 +0530 | `6b2f6bb` | `2c2736c` | Phase 7C pod scale-up results: 6 systems x 50/100/200 pods x 3 reps, interleaved (0 skipped, 0 invalid) |
| 2026-09-28 00:38:17 +0530 | `bc0ff05` | `c218c73` | Phase 7C lever check: T-same-region optimistic c=50, fan-out all vs hedged, 3 runs, signer CPU |
| 2026-09-28 01:10:31 +0530 | `a436567` | `4e95881` | Phase 7C stress test: CPU-capped signers (CPUQuota=25%, SIGNER_MAX_CONCURRENT=1), c=10..200, 3 runs |
| 2026-09-28 01:23:26 +0530 | `88473b0` | `c0030fc` | Phase 7C docs and teardown: NOTES N72-N74, CLAIMS_AUDIT row 11, REVIEW_RESPONSE R6/R10/R12, PAPER_INPUTS, INDEPENDENCE 7C; zero-resources proof |
| 2026-09-28 01:46:18 +0530 | `48fb8f8` | `9b76cd2` | Design (not implemented): priority-consistent signer admission after DAGOR; NOTES N75 incl. open question on the signer-CPU drop |
| 2026-09-28 03:00:03 +0530 | `8032da3` | `950e920` | N76: quorum-impossible abort (B), priority-consistent signer admission (A), per-replica coordinator certs + fair share, 4 s deadline cap |
| 2026-10-03 14:01:57 +0530 | `b672ab7` | `10a49d9` | N77: stable-identity priority (proposal), DAGOR 4.2.3 histogram controller, storm test for retry amplification |
| 2026-10-03 14:12:27 +0530 | `f745ae2` | `29e2025` | summarize: drop unused type (staticcheck U1000, CI run 37110023361) |
| 2026-10-03 15:42:58 +0530 | `18ac3ac` | `2b27f0c` | N76/N77 AWS session: bootstrap evidence (fresh 7C cluster at f745ae2; T-5 kid 9gu_VGBe..., T-same kid Eu4bZ_v5...; 6 systems smoke-checked) |
| 2026-10-03 21:17:03 +0530 | `4c8823a` | `32474d0` | N78: N76/N77 AWS evaluation results; teardown proof |
| 2026-10-03 22:02:15 +0530 | `293a919` | `fe0d537` | Design (not implemented): priority admission v2, the single follow-up iteration; pre-registered rules R1-R5 and the QUORUM_ABORT default rule (NOTES N79) |
| 2026-10-03 22:41:52 +0530 | `d07a8e7` | `361421d` | N80: priority admission v2 controller (count-and-time window, shed-fraction overload, damping), v1 kept; v2/v2nb variants; summarizer scores R1-R4 and the QUORUM_ABORT rule |
| 2026-10-03 23:17:23 +0530 | `a000c9b` | `bec4906` | v2 AWS session: bootstrap evidence (fresh 7C cluster at d07a8e7; T-5 kid g1TN11q0..., T-same kid XQC677lh...; 6 systems smoke-checked) |
| 2026-10-04 03:17:41 +0530 | `eccc5a8` | `6aff30f` | N81: v2 evaluation results (rules from 293a919): R1-R4 FAIL, R5 PASS; v2 not adopted; QUORUM_ABORT rule -> default off; teardown proof |
| 2026-10-04 03:53:03 +0530 | `37f3c5e` | `6229ee8` | N82: apply the pre-registered decision (QUORUM_ABORT off by default, opt-in); final defaults documented; priority admission evaluated, not recommended; PAPER_INPUTS two-iteration result |
| 2026-10-04 04:02:09 +0530 | `bea27ea` | `68608d5` | N83: code freeze at 37f3c5e (only Phase 12 audit fixes from here) |
| 2026-10-04 04:32:00 +0530 | `0aac79c` | `0785797` | docs: add Phase 12 audit instructions |
| 2026-10-04 13:25:19 +0530 | `2572d08` | `a076db4` | Phase 12 audit C-1: test the coordinator binary's own QUORUM_ABORT default (off) and final defaults |
| 2026-10-04 13:25:19 +0530 | `fc4b8ce` | `e5bfadb` | Phase 12 audit C-2: test that SHARE_FILE names exactly one share file (I4) |
| 2026-10-04 13:26:54 +0530 | `c82f653` | `717069a` | Phase 12 audit E-1: Vault clients never follow a redirect off the configured origin |
| 2026-10-04 13:28:06 +0530 | `c4c267e` | `1b89649` | Phase 12 audit E-2: multihost deploy refuses more than t-1 shares on any host |
| 2026-10-04 13:30:51 +0530 | `9f85a8c` | `9359ce2` | Phase 12 audit C-7: run the spike tcrsa hazard tests in make test and CI |
| 2026-10-04 13:33:38 +0530 | `01272b2` | `da84bfe` | Phase 12 audit G-2, G-4, G-18: per-session AWS evidence restored; scripts never overwrite evidence |
| 2026-10-04 13:34:23 +0530 | `5a77718` | `6c2d65c` | Phase 12 audit G-1: commit the repro step logs; mark the unrecoverable EC2 logs not retained |
| 2026-10-04 13:37:08 +0530 | `d23d0f4` | `189e149` | Phase 12 audit G-3, G-5, G-6, G-9, G-10, G-16, G-17, E-3, I-1, J-3: docs corrections, LICENSE |
| 2026-10-04 13:54:23 +0530 | `5fdf63d` | `6e876dc` | Phase 12 audit: commit the audit reports, defence brief and resolution column |
| 2026-10-04 14:10:04 +0530 | `f4e6a32` | `df1c152` | Phase 12 audit A-1, G-26: REPRO.md is the fresh-runner repro of the fix-loop code; final CI/repro recorded |
| 2026-10-04 14:44:07 +0530 | `29868bd` | `ded335a` | Phase 12 audit G-7, G-8, G-11-G-15, G-19, G-24, G-27, G-28: every claim sentence has evidence |
| 2026-10-04 17:45:20 +0530 | `cba14ca` | `11139f2` | Phase 12 audit C-3: tests for the second-layer checks M5a, M14a, M14b; M3a documented as unreachable |
| 2026-10-04 17:49:37 +0530 | `374a59d` | `af6e093` | Phase 12 audit E-1 (extension): VAULT_ADDR must be https; plain http only with VAULT_DEV_ALLOW_HTTP=1 |
| 2026-10-04 17:57:27 +0530 | `fc69a58` | `be06e77` | Phase 12 audit B-1: define invariants I1-I10 in docs/INVARIANTS.md |
| 2026-10-04 17:58:52 +0530 | `d26a5b5` | `cd699fa` | Phase 12 audit H-3 and history disclosure: grpc acceptance extended with justification; history not rewritten |
| 2026-10-04 18:00:21 +0530 | `7f10c65` | `4a6547d` | Phase 12 audit round 2: resolution column, claims recheck and low-findings proposal |
| 2026-10-04 20:09:44 +0530 | `baf7875` | `124f1f1` | Phase 12 audit round 2: REPRO.md is the repro of the round-2 code; final CI/repro recorded |
| 2026-10-04 22:59:16 +0530 | `046ef34` | `9972e50` | Phase 12 audit C-8, D-1: policy boundary tests; NumericDates must be JSON numbers |
| 2026-10-04 23:00:48 +0530 | `2da5330` | `90b9ab0` | Phase 12 audit C-4, C-5: tests for the coordinator's request_id check and response-size limit |
| 2026-10-04 23:00:48 +0530 | `330113d` | `34db1b8` | Phase 12 audit C-6: bounded waits in TestQueuedRequestCancelledByCallerComputesNoShare |
| 2026-10-04 23:02:48 +0530 | `e1f0bf4` | `4688364` | Phase 12 audit E-4: the dealer refuses a shared-writable output directory and cleans up partial shares |
| 2026-10-04 23:03:16 +0530 | `5517030` | `ca4629e` | Phase 12 audit E-5: ssh control sockets in a private 0700 directory; TOFU accepted with reason |
| 2026-10-04 23:04:16 +0530 | `9c88b02` | `0e6e709` | Phase 12 audit H-1: pin Docker base images by multi-arch index digest |
| 2026-10-04 23:05:12 +0530 | `9867652` | `a422a0a` | Phase 12 audit J-1, G-30: gitleaks false positives fingerprinted; HISTORY_PURGE corrected |
| 2026-10-04 23:06:31 +0530 | `19f678e` | `0a30f22` | Phase 12 audit G-20, G-21, G-22, G-23, G-25, G-29, G-31, G-32, G-35: documentation corrections |
| 2026-10-04 23:59:41 +0530 | `54b1882` | `670e060` | Phase 12 audit C-6: holdSlot never blocks other requests, so missing admission fails fast |
| 2026-10-05 00:09:54 +0530 | `be4904a` | `bb388f5` | Phase 12 audit round 3: resolution column for every low finding; mutation re-runs |
| 2026-10-05 00:25:15 +0530 | `6e3b32b` | `3310103` | NOTES N88: Phase 12 audit complete (final verified commit be4904a) |
| 2026-10-05 14:55:50 +0530 | `e1266b0` | `b0e6ae8` | Review traceability: store the SCORED '26 6A/6B/6C texts verbatim; add R19-R22 (NOTES N89) |
| 2026-10-11 14:49:06 +0530 | `10c0491` | `cff7e44` | Paper support: run-to-run dispersion for 7C main, 7C stress, v2 stress-storm |
| 2026-10-11 15:34:20 +0530 | `02a1306` | `38c13ef` | Merge fix/threshold-rsa into main |
