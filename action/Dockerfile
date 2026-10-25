# The Action's image: Proven's published image plus the entry point script.
# GitHub runs a container action as root so it can write the workspace and
# $GITHUB_OUTPUT, which the runner owns; the published image's user 65532 cannot.
# Pin PROVEN_IMAGE to a digest (saphan/proven@sha256:...) once make push has
# published the image: docs/ACTION.md, runbook step 1.
ARG PROVEN_IMAGE=saphan/proven:latest
FROM ${PROVEN_IMAGE}
USER root
COPY entrypoint.sh /action/entrypoint.sh
RUN chmod 755 /action/entrypoint.sh
ENTRYPOINT ["/action/entrypoint.sh"]
