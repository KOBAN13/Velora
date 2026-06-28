#!/usr/bin/env bash
set -euo pipefail

MODULE="${KUKURUZKA_ESC_MODULE:-github.com/KOBAN13/kukuruzka-esc}"
DEFAULT_PATH="${KUKURUZKA_ESC_PATH:-../kukuruzka-esc}"
DEFAULT_REF="${KUKURUZKA_ESC_REF:-latest}"
DEFAULT_INTERVAL="${KUKURUZKA_ESC_INTERVAL:-10}"

usage() {
	cat <<USAGE
Usage:
  $0 local [path]
      Use a local checkout of ${MODULE} through go.mod replace.

  $0 update [ref]
      Remove local replace, then update ${MODULE} from Git.
      ref defaults to "${DEFAULT_REF}" and can be latest, main, a tag, or a commit hash.

  $0 watch-update [ref] [seconds]
      Run update in a loop. The interval defaults to "${DEFAULT_INTERVAL}" seconds.

  $0 publish [path] <commit-message> [ref]
      In the dependency repository: git add -A, commit if needed, push.
      Then update this project to the pushed ref. If ref is omitted, current branch is used.

Environment:
  KUKURUZKA_ESC_PATH    default local dependency path, currently "${DEFAULT_PATH}"
  KUKURUZKA_ESC_REF     default update ref, currently "${DEFAULT_REF}"
  KUKURUZKA_ESC_INTERVAL update loop interval, currently "${DEFAULT_INTERVAL}"
  KUKURUZKA_ESC_MODULE  module path, currently "${MODULE}"
USAGE
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

require_go_mod() {
	[[ -f go.mod ]] || die "run this script from the Velora repository root"
}

absolute_path() {
	local path="$1"
	(cd "$path" && pwd)
}

validate_dependency_path() {
	local path="$1"

	[[ -d "$path" ]] || die "dependency path does not exist: $path"
	[[ -f "$path/go.mod" ]] || die "dependency path has no go.mod: $path"
	awk -v module="$MODULE" '$1 == "module" && $2 == module { found = 1 } END { exit !found }' "$path/go.mod" || die "dependency go.mod is not module ${MODULE}"
}

print_module_state() {
	go list -m -f '{{if .Replace}}{{.Path}} => {{.Replace.Path}}{{else}}{{.Path}} {{.Version}}{{end}}' "$MODULE"
}

use_local() {
	local path="${1:-$DEFAULT_PATH}"
	local abs_path

	require_go_mod
	validate_dependency_path "$path"
	abs_path="$(absolute_path "$path")"

	go mod edit -replace="${MODULE}=${abs_path}"
	go mod tidy
	print_module_state
}

update_from_git() {
	local ref="${1:-$DEFAULT_REF}"

	require_go_mod
	go mod edit -dropreplace="$MODULE" 2>/dev/null || true
	go get "${MODULE}@${ref}" || return
	go mod tidy || return
	print_module_state || return
}

watch_update_from_git() {
	local ref="${1:-$DEFAULT_REF}"
	local interval="${2:-$DEFAULT_INTERVAL}"

	[[ "$interval" =~ ^[1-9][0-9]*$ ]] || die "interval must be a positive integer"

	while true; do
		printf '[%s] updating %s@%s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$MODULE" "$ref"
		if update_from_git "$ref"; then
			printf '[%s] update finished\n' "$(date '+%Y-%m-%d %H:%M:%S')"
		else
			printf '[%s] update failed; retrying in %s seconds\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$interval" >&2
		fi
		sleep "$interval"
	done
}

publish_and_update() {
	local path="${1:-}"
	local message="${2:-}"
	local ref="${3:-}"
	local branch

	[[ -n "$path" ]] || die "dependency path is required"
	[[ -n "$message" ]] || die "commit message is required"

	require_go_mod
	validate_dependency_path "$path"

	git -C "$path" add -A
	if ! git -C "$path" diff --cached --quiet; then
		git -C "$path" commit -m "$message"
	fi

	git -C "$path" push

	if [[ -z "$ref" ]]; then
		branch="$(git -C "$path" branch --show-current)"
		[[ -n "$branch" ]] || die "cannot infer ref from detached HEAD; pass tag, branch, or commit hash explicitly"
		ref="$branch"
	fi

	update_from_git "$ref"
}

main() {
	local command="${1:-}"
	shift || true

	case "$command" in
		local)
			use_local "$@"
			;;
		update)
			update_from_git "$@"
			;;
		watch-update)
			watch_update_from_git "$@"
			;;
		publish)
			publish_and_update "$@"
			;;
		-h|--help|help|'')
			usage
			;;
		*)
			usage >&2
			die "unknown command: $command"
			;;
	esac
}

main "$@"
