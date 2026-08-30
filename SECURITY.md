# Security Policy

## Reporting a Vulnerability

We strongly encourage you to report security vulnerabilities to our private
security mailbox: runtimeradar@outlook.com - first, before disclosing them in
any public forums.

Reports sent to this address are read only by the Runtime Radar maintainers
and are treated as top priority.

To help us triage and fix the issue quickly, please include a description of
the vulnerability and its impact, the affected component (e.g. `auth-center`,
`public-api`, `reverse-proxy`, the Helm chart) and version, steps to reproduce
or a proof of concept, and a suggested fix or mitigation if you have one.

### Scope

Runtime Radar is designed to run inside a Kubernetes cluster and to be
operated by cluster administrators. Please consider the feasibility of an
attack against a correctly secured environment before making a report: issues
that require an already-compromised cluster or administrator-level access to
exploit are generally not considered vulnerabilities in Runtime Radar itself.

In scope are the Runtime Radar services and libraries in this repository, the
Helm chart ([install/helm]), the container images and charts published to
`ghcr.io/runtime-radar`, and the detector SDK and Wasm detector sandbox.

Vulnerabilities in third-party dependencies should be reported to the upstream
project; please let us know if a stable release of Runtime Radar is affected
by such an issue.

#### Issues with Runtime Radar's CI or GitHub workflows

The project does not consider issues affecting Runtime Radar's CI to be in
scope if they only show CI infrastructure being used to build and test
contributor code. CI issues are typically in scope if they can be shown to
lead to the compromise of Runtime Radar release artifacts or release
infrastructure. Some examples of such issues are:

- Issues that lead to the compromise of credentials that can then be used to
  modify release artifacts published to `ghcr.io/runtime-radar`
- Issues that would allow an attacker to bypass required functional or
  security testing, with the aim of introducing unstable or malicious code
  into Runtime Radar releases

### Disclosure

The project aims to acknowledge all contributors for valid reports of security
issues. For reports that affect stable releases of Runtime Radar, the project
will release a GitHub security advisory; reporters will be credited by
name/GitHub handle in the advisory. Disclosure will typically be made at or
shortly after the release of patched versions of Runtime Radar.

The maintainers will decide whether a report meets the requirements for a
GitHub advisory on a case-by-case basis. Some reports may not result in an
associated advisory even if they lead to code changes, for example reports of
issues in pre-release functionality or reports where there is no evidence that
a stable release of Runtime Radar is affected. In such cases, the project aims
to credit reporters with an acknowledgement in the relevant fix commit via a
`Reported-by:` trailer in the commit message.

Runtime Radar is under active pre-1.0 development, and security fixes are
applied to the latest release only. Please make sure you are running the most
recent version and keep vulnerability details private until a fix has been
released.

[install/helm]: https://github.com/Runtime-Radar/runtime-radar/tree/main/install/helm
