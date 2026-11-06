"""Gate T4's fixture helper.

    tree.py put <fixture dir> <dest>   copy the fixture to <dest>, @DIR@ replaced by <dest>
    tree.py norm <dir> <out>           copy <dir> to <out>, <dir> replaced by @DIR@ in every
                                       file, and in stub.log the lines of the variables that
                                       differ between two runs (PWD, OLDPWD, SHLVL, _) dropped;
                                       directories are listed in <out>/.dirs
"""
import os
import sys

DROP = ("env PWD=", "env OLDPWD=", "env SHLVL=", "env _=")


def walk(src):
    for root, dirs, files in os.walk(src):
        dirs.sort()
        for f in sorted(files):
            p = os.path.join(root, f)
            yield os.path.relpath(p, src)


def main():
    mode, src, dst = sys.argv[1:4]
    os.makedirs(dst, exist_ok=True)
    if mode == "put":
        for rel in walk(src):
            with open(os.path.join(src, rel), "rb") as f:
                b = f.read().replace(b"@DIR@", dst.encode())
            p = os.path.join(dst, rel)
            os.makedirs(os.path.dirname(p), exist_ok=True)
            with open(p, "wb") as f:
                f.write(b)
    elif mode == "norm":
        dirs = []
        for root, ds, _ in os.walk(src):
            for d in ds:
                dirs.append(os.path.relpath(os.path.join(root, d), src))
        for rel in walk(src):
            with open(os.path.join(src, rel), "rb") as f:
                b = f.read().replace(src.encode(), b"@DIR@")
            if os.path.basename(rel) == "stub.log":
                b = b"".join(l for l in b.splitlines(True) if not l.decode("utf-8", "replace").startswith(DROP))
            p = os.path.join(dst, rel)
            os.makedirs(os.path.dirname(p), exist_ok=True)
            with open(p, "wb") as f:
                f.write(b)
        with open(os.path.join(dst, ".dirs"), "w") as f:
            f.write("".join(d + "\n" for d in sorted(dirs)))
    else:
        sys.exit("tree.py: unknown mode " + mode)


main()
