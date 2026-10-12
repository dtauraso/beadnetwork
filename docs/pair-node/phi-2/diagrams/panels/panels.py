import os

os.chdir(os.path.dirname(os.path.abspath(__file__)))

from panel_defs import build_panels

out = build_panels()
for k, v in out.items():
    v = v.replace('<svg ', '<svg xmlns="http://www.w3.org/2000/svg" ', 1)
    path = k + ".svg"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "w").write(v)
print(len(out), "panels written")
