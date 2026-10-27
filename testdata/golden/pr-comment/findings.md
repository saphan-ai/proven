**Proven: blocked** — 4 findings: 4 error, 0 warning, 0 info

| Severity | Finding | Location |
| --- | --- | --- |
| Error | Dependency leftpad v1.2.3 is on the known-exploited list | `go.mod:5` |
| Error | CVE-2026-0001 for leftpad v1.2.3 is on the KEV list | `go.mod:5` |
| Error | Dependency evil.example/leftpad v1.2.3 is in an OSV affected range | `go.mod:5` |
| Error | Discount is applied twice | `billing/discount.go:5` |

<details><summary>Error — Dependency leftpad v1.2.3 is on the known-exploited list — go.mod:5</summary>

The change requires evil.example/leftpad v1.2.3, and the sample advisory table has a known-exploited row for that version.

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'
```

| id | aliases | ecosystem | package | introduced | fixed | last_affected | severity | published |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GHSA-px-sample-0001 | CVE-2026-0001 | Go | evil.example/leftpad | 1\.0.0 | 1\.2.4 |  | HIGH | 2026-01-15 |
| GHSA-px-sample-0002 |  | Go | evil.example/leftpad | 2\.0.0 | 2\.1.0 |  | LOW | 2026-03-01 |

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
```

| cve_id | epss | percentile | score_date |
| --- | --- | --- | --- |
| CVE-2026-0001 | 0\.50000 | 0\.90000 | 2026-10-20 |

</details>

<details><summary>Error — CVE-2026-0001 for leftpad v1.2.3 is on the KEV list — go.mod:5</summary>

The change requires evil.example/leftpad v1.2.3. The sample KEV view has CVE-2026-0001 for that advisory.

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
```

| cve_id | vendor_project | product | date_added | known_ransomware |
| --- | --- | --- | --- | --- |
| CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known |

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
```

| cve_id | epss | percentile | score_date |
| --- | --- | --- | --- |
| CVE-2026-0001 | 0\.50000 | 0\.90000 | 2026-10-20 |

</details>

<details><summary>Error — Dependency evil.example/leftpad v1.2.3 is in an OSV affected range — go.mod:5</summary>

evil.example/leftpad v1.2.3 matches GHSA-px-sample-0001: introduced 1.0.0 &lt;= v1.2.3 &lt; fixed 1.2.4. GHSA-px-sample-0002 does not match: v1.2.3 &lt; introduced 2.0.0. alias CVE-2026-0001 is on the KEV list Reachability: not measured.

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'
```

| id | aliases | ecosystem | package | introduced | fixed | last_affected | severity | published |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GHSA-px-sample-0001 | CVE-2026-0001 | Go | evil.example/leftpad | 1\.0.0 | 1\.2.4 |  | HIGH | 2026-01-15 |
| GHSA-px-sample-0002 |  | Go | evil.example/leftpad | 2\.0.0 | 2\.1.0 |  | LOW | 2026-03-01 |

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
```

| cve_id | vendor_project | product | date_added | known_ransomware |
| --- | --- | --- | --- | --- |
| CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known |

**Query** (local-sample, snapshot sample-2026-10-22, form sql)

```sql
SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
```

| cve_id | epss | percentile | score_date |
| --- | --- | --- | --- |
| CVE-2026-0001 | 0\.50000 | 0\.90000 | 2026-10-20 |

**Fix** (named from the advisory table, no test ran): raise `evil.example/leftpad` `v1.2.3` to `1.2.4`. 1\.2.4 is the fixed version of advisory GHSA-px-sample-0001, and it is the smallest fixed version of the rows returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.

</details>

<details><summary>Error — Discount is applied twice — billing/discount.go:5</summary>

Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.

**Failing test** `TestDiscounted` in `billing/discount_test.go` — fails on the change: yes, passes on the base: yes

Command:

```sh
go test -run '^TestDiscounted$' ./billing
```

Output:

```text
base: PASS
change: FAIL got 60 want 80
```

</details>

- Author model: unknown, the change does not declare it
- Reviewer: `NVIDIA` (NVIDIA)
- Second reviewer: not configured

<sub>reviewed by `NVIDIA` · guardrail checked every call</sub>
