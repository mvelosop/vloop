"""Gate T1's JSON judge: parses vloop's metrics/v4 documents and asserts on the
parse. Modes:
  rows <doc> <want.json>        by_session equals the expected rows, key set and values
  sums <doc> <n>                n rows; cost_usd sums to cost.total_usd within 1e-9,
                                permission_denials to the report's
  b36 <doc>                     13 rows and 8 permission denials in all
  drop <doc>                    print the document without by_session
  bad <doc>                     print the document with by_session a string
  export <jsonl>                no brief record carries by_session
  workspace <doc> <want.json>   an array of documents, each with a by_session list;
                                fx1's equals the expected rows
"""
import json
import sys

KEYS = {"folder", "seq", "schema", "phase", "iteration", "task", "agent", "model", "effort",
        "time_ms", "cost_usd", "cost_source", "credits", "turns", "permission_denials",
        "denials", "is_error"}


def die(msg):
    sys.stderr.write("GATE FAIL: " + msg + "\n")
    sys.exit(1)


def load(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def same(got, want, where):
    if not isinstance(got, list):
        die(f"{where}: by_session is {type(got).__name__}, want a list")
    if len(got) != len(want):
        die(f"{where}: {len(got)} rows, want {len(want)}: {json.dumps(got)}")
    for i, (g, w) in enumerate(zip(got, want)):
        if not isinstance(g, dict):
            die(f"{where}: row {i + 1} is not an object")
        if set(g) != KEYS:
            die(f"{where}: row {i + 1} keys {sorted(g)}, want {sorted(KEYS)}")
        for k in sorted(KEYS):
            gv, wv = g[k], w[k]
            if type(gv) is bool or type(wv) is bool or gv is None or wv is None:
                okv = gv is wv or (type(gv) is bool and gv == wv and type(wv) is bool)
            elif isinstance(wv, (int, float)):
                okv = isinstance(gv, (int, float)) and gv == wv
            else:
                okv = gv == wv
            if not okv:
                die(f"{where}: row {i + 1} ({w['folder']} #{w['seq']}) {k} = {json.dumps(gv)}, want {json.dumps(wv)}")


def sums(doc, n):
    rows = doc.get("by_session")
    if not isinstance(rows, list) or len(rows) != n:
        die(f"by_session has {len(rows) if isinstance(rows, list) else rows!r} rows, want {n}")
    total = sum(r["cost_usd"] for r in rows)
    if abs(total - doc["cost"]["total_usd"]) > 1e-9:
        die(f"rows' cost_usd sum to {total!r}, cost.total_usd is {doc['cost']['total_usd']!r}")
    den = sum(r["permission_denials"] for r in rows)
    if den != doc["permission_denials"]:
        die(f"rows' permission_denials sum to {den}, the report's is {doc['permission_denials']}")
    for r in rows:
        if len(r["denials"]) != r["permission_denials"]:
            die(f"row {r['folder']} #{r['seq']}: {len(r['denials'])} denials listed, permission_denials {r['permission_denials']}")


def main():
    mode, args = sys.argv[1], sys.argv[2:]
    if mode == "rows":
        same(load(args[0]).get("by_session"), load(args[1]), "fx1")
    elif mode == "sums":
        sums(load(args[0]), int(args[1]))
    elif mode == "b36":
        doc = load(args[0])
        if len(doc.get("by_session") or []) != 13 or doc.get("permission_denials") != 8:
            die(f"{len(doc.get('by_session') or [])} rows, {doc.get('permission_denials')} denials; want 13 and 8")
        if abs(doc["cost"]["total_usd"] - 5.6702516) > 1e-6:
            die(f"cost.total_usd {doc['cost']['total_usd']!r}, want 5.6702516…")
    elif mode == "drop":
        doc = load(args[0])
        doc.pop("by_session", None)
        print(json.dumps(doc))
    elif mode == "bad":
        doc = load(args[0])
        doc["by_session"] = "x"
        print(json.dumps(doc))
    elif mode == "export":
        n = 0
        with open(args[0], encoding="utf-8") as f:
            for line in f:
                if not line.strip():
                    continue
                rec = json.loads(line)
                if rec.get("type") == "brief":
                    n += 1
                    if "by_session" in rec or "by_task" in rec:
                        die(f"export brief record {rec.get('brief')} carries by_session or by_task")
        if n == 0:
            die("export printed no brief record")
    elif mode == "workspace":
        docs = load(args[0])
        if not isinstance(docs, list) or not docs:
            die("--workspace --json printed no documents")
        found = False
        for d in docs:
            if not isinstance(d.get("by_session"), list):
                die(f"document {d.get('run_id')} has no by_session list")
            if d.get("run_id") == "B20260101-0000-fx1":
                found = True
                same(d["by_session"], load(args[1]), "workspace fx1")
        if not found:
            die("--workspace --json has no fx1 document")
    else:
        die("unknown mode " + mode)


main()
