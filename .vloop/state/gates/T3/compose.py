"""Gate T3's oracle: the expected output of `kit report <kind> <repo> <vloop> <rc>`,
built from the starting point's report.py output on the same repository by
replacing its first table with P2's Sessions block(s).

    compose.py full <python out> <repo> <vloop> <min rows|-> > want
    compose.py blocks <repo> <vloop> > want-blocks

<min rows> is the number of rows each brief's `vloop metrics <id> --by session`
must print (at least); `-` skips the check (the fixtures where metrics fails).
"""
import json
import os
import re
import subprocess
import sys

PHASES = {"plan", "work", "review", "second-review", "gate-review", "verify"}
SESSION = re.compile(r"^(\d+)-([a-z-]+)\.json$")


def die(msg):
    sys.stderr.write("GATE FAIL: " + msg + "\n")
    sys.exit(1)


def run_ids(repo):
    base = os.path.join(repo, ".vloop/state/runs")
    if not os.path.isdir(base):
        return []
    out = []
    for n in sorted(os.listdir(base)):
        if os.path.isdir(os.path.join(base, n)) and os.path.isfile(
                os.path.join(repo, "docs/briefs", n + ".loop-brief.md")):
            out.append(n)
    return out


def denial_text(d):
    if not isinstance(d, dict):
        return ""
    ti = d.get("tool_input") if isinstance(d.get("tool_input"), dict) else {}
    for v in (d.get("message"), d.get("command"), d.get("file_path"), ti.get("command"), ti.get("file_path")):
        if isinstance(v, str) and v != "":
            return v
    return ""


def block(repo, vloop, rid, min_rows):
    p = subprocess.run([vloop, "metrics", rid, "--by", "session"], cwd=repo, capture_output=True)
    stdout, stderr = p.stdout.decode("utf-8"), p.stderr.decode("utf-8")
    if min_rows != "-":
        if p.returncode != 0:
            die(f"vloop metrics {rid} --by session exited {p.returncode} on a fixture where it must succeed: {stderr.strip()}")
        rows = [l for l in stdout.splitlines() if l.strip()][1:]
        if len(rows) < int(min_rows):
            die(f"vloop metrics {rid} --by session printed {len(rows)} rows, want at least {min_rows}:\n{stdout}")
    if p.returncode == 0:
        fenced = stdout if stdout.endswith("\n") else stdout + "\n"
    else:
        fenced = f"vloop metrics exited {p.returncode}: " + stderr.strip().replace("\n", " ") + "\n"
    lines = [f"Sessions (`vloop metrics {rid} --by session`):", "", "```"]
    lines += fenced[:-1].split("\n")
    lines += ["```", ""]

    n = k = 0
    counts, noschema, recs, dens = {}, 0, [], []
    rdir = os.path.join(repo, ".vloop/state/runs", rid)
    for folder in sorted(x for x in os.listdir(rdir) if os.path.isdir(os.path.join(rdir, x))):
        sdir = os.path.join(rdir, folder, "sessions")
        if not os.path.isdir(sdir):
            continue
        files = []
        for f in os.listdir(sdir):
            m = SESSION.match(f)
            if m:
                files.append((int(m.group(1)), f, m.group(2)))
        for seq, f, fphase in sorted(files):
            n += 1
            rel = f".vloop/state/runs/{rid}/{folder}/sessions/{f}"
            try:
                with open(os.path.join(sdir, f), encoding="utf-8") as fh:
                    r = json.load(fh)
            except ValueError:
                recs.append(f"- {rel}: unreadable")
                continue
            sch = r.get("schema")
            if not isinstance(sch, str) or sch == "":
                noschema += 1
                recs.append(f"- {rel}: no schema")
            else:
                counts[sch] = counts.get(sch, 0) + 1
                ok = subprocess.run([vloop, "schema", "validate", sch, rel], cwd=repo,
                                    capture_output=True).returncode == 0
                if ok:
                    k += 1
                else:
                    recs.append(f"- {rel}: {sch} INVALID")
            phase = r.get("phase") or fphase
            if phase in PHASES:
                for d in r.get("permission_denials") or []:
                    tn = d.get("tool_name") if isinstance(d, dict) and isinstance(d.get("tool_name"), str) else ""
                    dens.append(f"- {folder}/{seq:03d} {phase} {tn}: " + denial_text(d).replace("\n", " "))
    if n == 0:
        lines.append("Session records: 0.")
    else:
        parts = [f"{s} {counts[s]}" for s in sorted(counts)]
        if noschema:
            parts.append(f"no schema {noschema}")
        paren = f" ({', '.join(parts)})" if parts else ""
        lines.append(f"Session records: {n}, schema-valid {k}{paren}.")
    lines += recs
    if dens:
        lines += ["", "Permission denials:", ""] + dens
    lines.append("")
    return lines


def blocks(repo, vloop, min_rows):
    ids = run_ids(repo)
    if not ids:
        return ["Sessions: no brief with runs.", ""]
    out = []
    for rid in ids:
        out += block(repo, vloop, rid, min_rows)
    return out


def main():
    mode = sys.argv[1]
    if mode == "full":
        pyout, repo, vloop, min_rows = sys.argv[2:6]
        with open(pyout, encoding="utf-8") as f:
            lines = f.read().split("\n")
        i = next((n for n, l in enumerate(lines) if l.startswith("| record |") or l.startswith("| round | record |")), None)
        if i is None:
            die("the reference output has no session table")
        j = next((n for n in range(i, len(lines)) if lines[n] == ""), None)
        if j is None:
            die("the reference output's session table has no blank line after it")
        lines[i:j + 1] = blocks(repo, vloop, min_rows)
        sys.stdout.write("\n".join(lines))
    elif mode == "blocks":
        repo, vloop = sys.argv[2:4]
        sys.stdout.write("\n".join(blocks(repo, vloop, "-")))
    else:
        die("unknown mode " + mode)


main()
