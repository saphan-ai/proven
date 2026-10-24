#!/bin/sh
# Runs each version of the tool on two jobs through the wrapper and prints one
# line per version and job: version, job, model asked for, tool exit code,
# model logged, model that really served, reference label.
# The model that really served is read from the answer text, by this check,
# from the stand-in's table; neither the tool nor the wrapper computes it.
# The reference label is a literal here; nothing else reads it.
set -eu

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$case_dir/../../../.." && pwd)
scratch="$repo/.saphan/work/corpus-ho-005"
mkdir -p "$scratch"
work=$(mktemp -d "$scratch/run.XXXXXX")
trap 'rm -rf "$work"' EXIT

export SERVICE_CMD="$case_dir/service.sh"

cat > "$work/jobs.txt" <<'EOF'
job-1 model-b summarise
job-2 model-z summarise
EOF

# really_served ANSWER: which stand-in model wrote this answer text
really_served() {
  for m in model-a model-b; do
    if [ "$1" = "$(sh "$case_dir/service.sh" answer "$m")" ]; then
      echo "$m"
      return
    fi
  done
  echo none
}

for version in base change; do
  : > "$work/$version-log.txt"
  sh "$case_dir/wrapper.sh" "$case_dir/$version/tool.sh" "$work/$version-log.txt" "$work/jobs.txt"
  while read -r job model prompt; do
    code=0
    answer=$(sh "$case_dir/$version/tool.sh" "$model" "$prompt" 2> /dev/null) || code=$?
    logged=$(grep "^$job " "$work/$version-log.txt" | cut -d ' ' -f 2)
    printf '%s %s asked %s exit %s logged %s served %s reference %s\n' \
      "$version" "$job" "$model" "$code" "${logged:--}" "$(really_served "$answer")" \
      "$(case "$job" in (job-1) echo "should serve model-b" ;; (*) echo "should not be served" ;; esac)"
  done < "$work/jobs.txt"
done
