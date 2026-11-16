# Customer-built CLI models

These files are recipes and small text adapters. **Proven publishes no image containing vendor CLIs.** You install the CLIs you are entitled to use, build your own image and authenticate it with your own account. No credentials enter the build.

- [Subscription CLI guide](../../docs/CLI_MODELS.md): install, login, local use and limits.
- [Two reviewers](../../docs/TWO_MODELS.md): two APIs, API + CLI, or Claude Code + Copilot.
- [Customer image guide](../../docs/CLI_IMAGES.md): build configuration, private profiles and security checks.

Run the contract checks without accounts: `python3 examples/cli/test_wrappers.py`. They use fake executables and prove adapter behavior, not vendor login or model quality. Vendor options were checked against Claude Code 2.1.319 locally and Copilot CLI 1.0.112 help; recipe versions are overridable.

The `dual` recipe was also built locally on Linux arm64 and its non-root user, Proven executable, Claude Code 2.1.325, Copilot 1.0.112 and `timeout` availability were checked. No account was authenticated, and that customer image was not published. Other architectures and live subscription sessions need their own checks.

The final Alpine variant additionally verified Git 2.56.1, a writable private home and npm's absence from the runtime. On 16 November 2026, Docker Scout indexed 95 packages from its single-platform archive and reported 0 CRITICAL / 0 HIGH (also no MEDIUM/LOW). This is a point-in-time scan of the local arm64 build, not a permanent result for every rebuild.
