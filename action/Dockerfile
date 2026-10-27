# The Action's image: Proven's published image plus curl and the entry point script.
# GitHub runs a container action as root so it can write the workspace and
# $GITHUB_OUTPUT, which the runner owns; the published image's user 65532 cannot.
# :judging is the tag fixed for the judging window: a push to latest does not change it
# (docs/JUDGING.md, which names its digest). make check-lab-gates refuses another tag here.
ARG PROVEN_IMAGE=saphan/proven:judging
FROM ${PROVEN_IMAGE}
USER root

# curl edits the summary comment: BusyBox wget cannot send PATCH (post-comment.sh).
RUN apk add --no-cache curl

COPY entrypoint.sh post-comment.sh /action/
RUN chmod 755 /action/entrypoint.sh && chmod 644 /action/post-comment.sh
ENTRYPOINT ["/action/entrypoint.sh"]
