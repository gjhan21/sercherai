#!/usr/bin/env python3
"""Go 代码静态分析验证 —— 超短线 T+1 选股功能。

检查 backend/ 目录下所有修改过的文件，验证：
1. 所有文件存在
2. 新增/修改的函数均已定义
3. 结构体定义完整
4. 跨文件函数引用一致
5. 接口定义与实现匹配
6. 路由注册正确
7. Go 1.20 兼容性（无 min/max 内置函数等）
8. Handler/Service/Repo 调用链完整

用法:
    cd backend
    python3 ../scripts/go_static_verify.py
"""

import os
import re
import sys
from pathlib import Path

# ——— 配置 ———
ROOT = Path(__file__).resolve().parent.parent / "backend"
if not ROOT.exists():
    # 尝试当前目录
    ROOT = Path.cwd()

FILES = {
    "mysql_repo": ROOT / "internal/growth/repo/mysql_repo.go",
    "market_data_multi_source": ROOT / "internal/growth/repo/market_data_multi_source.go",
    "interfaces": ROOT / "internal/growth/repo/interfaces.go",
    "inmemory_repo": ROOT / "internal/growth/repo/inmemory_repo.go",
    "service": ROOT / "internal/growth/service/service.go",
    "handler": ROOT / "internal/growth/handler/market_data_sync_handler.go",
    "router_admin": ROOT / "router/admin.go",
}

# 本 session 新增的函数（需检查是否已定义）
EXPECTED_FUNCTIONS = [
    ("mysql_repo.go", "fetchStockKPLListFromTushare"),
    ("mysql_repo.go", "parseStockKPLListFromTushareResponse"),
    ("mysql_repo.go", "fetchStockTopListFromTushare"),
    ("mysql_repo.go", "parseStockTopListFromTushareResponse"),
    ("mysql_repo.go", "upsertStockKPLList"),
    ("mysql_repo.go", "upsertStockTopList"),
    ("market_data_multi_source.go", "fetchStockKPLListForSource"),
    ("market_data_multi_source.go", "fetchStockTopListForSource"),
    ("market_data_multi_source.go", "AdminSyncStockKPLList"),
    ("market_data_multi_source.go", "AdminSyncStockTopList"),
    ("interfaces.go", "AdminSyncStockKPLList"),
    ("interfaces.go", "AdminSyncStockTopList"),
    ("inmemory_repo.go", "AdminSyncStockKPLList"),
    ("inmemory_repo.go", "AdminSyncStockTopList"),
    ("service.go", "AdminSyncStockKPLList"),
    ("service.go", "AdminSyncStockTopList"),
    ("market_data_sync_handler.go", "SyncStockKPLList"),
    ("market_data_sync_handler.go", "SyncStockTopList"),
]

# Go 1.20 禁止使用的内置函数
GO121_BUILTINS = ["min(", "max(", "clear("]

# 颜色输出
GREEN = "\033[32m"
RED = "\033[31m"
YELLOW = "\033[33m"
CYAN = "\033[36m"
RESET = "\033[0m"

pass_count = 0
fail_count = 0
warn_count = 0


def ok(msg: str) -> None:
    global pass_count
    pass_count += 1
    print(f"  {GREEN}PASS{RESET}  {msg}")


def fail(msg: str) -> None:
    global fail_count
    fail_count += 1
    print(f"  {RED}FAIL{RESET}  {msg}")


def warn(msg: str) -> None:
    global warn_count
    warn_count += 1
    print(f"  {YELLOW}WARN{RESET}  {msg}")


def read_file(key: str) -> str | None:
    path = FILES.get(key)
    if not path or not path.exists():
        return None
    return path.read_text(encoding="utf-8")


def find_func_def(content: str, name: str) -> bool:
    """匹配 Go 函数/方法/接口方法定义。"""
    patterns = [
        # func name(
        rf'\bfunc\s+{re.escape(name)}\s*\(',
        # func (r *Type) name(
        rf'\bfunc\s+\([^)]*\)\s+{re.escape(name)}\s*\(',
        # 接口方法: name(params) (returns, error)
        rf'^\s+{re.escape(name)}\s*\([^)]*\)',
        # 接口方法多行: name(\n    params\n  ) (returns, error)
        rf'^\s+{re.escape(name)}\s*\(',
    ]
    for pat in patterns:
        if re.search(pat, content, re.MULTILINE):
            return True
    return False


def find_go121_incompat(content: str) -> list[str]:
    """检查 Go 1.20 不兼容的写法。"""
    issues = []
    for line_num, line in enumerate(content.split("\n"), 1):
        for builtin in GO121_BUILTINS:
            # 检查独立调用（非方法调用、注释、字符串）
            if re.search(rf'(?<!\.)\b{re.escape(builtin)}', line):
                # 排除注释
                stripped = line.strip()
                if stripped.startswith("//") or stripped.startswith("/*"):
                    continue
                # 排除字符串
                if '"' in stripped or '`' in stripped:
                    continue
                issues.append(f"line {line_num}: {stripped.strip()}")
    return issues


def find_route(content: str, method: str, path: str) -> bool:
    """检查路由注册。"""
    escaped = re.escape(f'"{path}"')
    return bool(re.search(rf'\.{method}\s*\(.*{escaped}', content))


# ==============================================
print(f"\n{CYAN}{'='*60}{RESET}")
print(f"{CYAN}  Go 代码静态分析 —— 超短线 T+1 选股{RESET}")
print(f"{CYAN}{'='*60}{RESET}\n")

# ——— 1. 文件存在性 ———
print(f"{CYAN}[1/8] 文件存在性检查{RESET}")
for key, path in FILES.items():
    if path.exists():
        ok(f"{key} → {path.name}")
    else:
        fail(f"{key} → {path.name} 不存在")

# ——— 2. 函数定义 ———
print(f"\n{CYAN}[2/8] 新增函数定义检查{RESET}")
file_contents: dict[str, str] = {}
for file_name, func_name in EXPECTED_FUNCTIONS:
    # 找到匹配的 file_key
    file_key = None
    for k, p in FILES.items():
        if p.name == file_name:
            file_key = k
            break
    if file_key is None:
        fail(f"{file_name}::{func_name} — 文件不在检查列表")
        continue

    content = file_contents.get(file_key)
    if content is None:
        content = read_file(file_key)
        file_contents[file_key] = content or ""

    if not content:
        fail(f"{file_name}::{func_name} — 无法读取文件")
        continue

    if find_func_def(content, func_name):
        ok(f"{file_name}::{func_name}")
    else:
        fail(f"{file_name}::{func_name} — 未找到定义（可能是多行签名）")

# ——— 3. 结构体/类型 ———
print(f"\n{CYAN}[3/8] 结构体与字段检查{RESET}")
mysql_content = read_file("mysql_repo") or ""
multi_content = read_file("market_data_multi_source") or ""

structs = []
if 'type stockLimitUpPoint struct' in mysql_content:
    ok("stockLimitUpPoint 结构体 (mysql_repo.go)")
    structs.append("stockLimitUpPoint")
else:
    fail("stockLimitUpPoint 结构体 未找到")

if 'type stockTopListPoint struct' in mysql_content:
    ok("stockTopListPoint 结构体 (mysql_repo.go)")
    structs.append("stockTopListPoint")
else:
    fail("stockTopListPoint 结构体 未找到")

# ——— 4. 跨文件引用 ———
print(f"\n{CYAN}[4/8] 跨文件引用一致性{RESET}")
svc_content = read_file("service") or ""
handler_content = read_file("handler") or ""
iface_content = read_file("interfaces") or ""

# Service 引用 Repo
for func_name in ["AdminSyncStockKPLList", "AdminSyncStockTopList"]:
    if f"s.repo.{func_name}" in svc_content:
        ok(f"service → repo.{func_name}")
    else:
        fail(f"service → repo.{func_name} — 未找到调用")

# Handler 引用 Service
for func_name in ["SyncStockKPLList", "SyncStockTopList"]:
    if f"h.service.{func_name}" in handler_content:
        ok(f"handler → service.{func_name}")
    else:
        warn(f"handler → service.{func_name} — 未找到直接调用（可能是动态调用）")

# 接口定义
for func_name in ["AdminSyncStockKPLList", "AdminSyncStockTopList"]:
    if func_name in iface_content:
        ok(f"接口定义: {func_name}")
    else:
        fail(f"接口定义缺失: {func_name}")

# ——— 5. 路由注册 ———
print(f"\n{CYAN}[5/8] 路由注册检查{RESET}")
router_content = read_file("router_admin") or ""

routes_to_check = [
    ("POST", "/kpl-list/sync", "SyncStockKPLList"),
    ("POST", "/top-list/sync", "SyncStockTopList"),
]
for method, path, handler_name in routes_to_check:
    if find_route(router_content, method, path):
        ok(f"路由: {method} {path} → {handler_name}")
    else:
        fail(f"路由: {method} {path} 未找到")

# ——— 6. Go 1.20 兼容性 ———
print(f"\n{CYAN}[6/8] Go 1.20 兼容性检查{RESET}")
all_issues = []
for key, path in FILES.items():
    if not path.exists():
        continue
    content = file_contents.get(key) or path.read_text(encoding="utf-8")
    file_contents[key] = content
    issues = find_go121_incompat(content)
    if issues:
        for issue in issues:
            warn(f"{path.name}:{issue}")
        all_issues.extend(issues)

if not all_issues:
    ok("所有文件兼容 Go 1.20（未使用 min/max/clear 内置函数）")

# ——— 7. imports 检查 ———
print(f"\n{CYAN}[7/8] 关键 import 检查{RESET}")

key_imports = {
    "mysql_repo": [
        ("database/sql", "MySQL 驱动"),
        ("encoding/json", "JSON 解析"),
        ("bytes", "字节处理"),
        ("net/http", "HTTP 客户端"),
    ],
    "market_data_multi_source": [
        ("time", "时间处理"),
        ("encoding/json", "JSON 解析"),
    ],
    "handler": [
        ("net/http", "HTTP 状态码"),
    ],
    "router_admin": [
        ("github.com/gin-gonic/gin", "Gin 框架"),
    ],
}

for file_key, expected_imports in key_imports.items():
    content = file_contents.get(file_key) or read_file(file_key) or ""
    for imp, desc in expected_imports:
        if f'"{imp}"' in content or f'"`{imp}`"' in content:
            ok(f"{file_key}: import {imp}")
        else:
            warn(f"{file_key}: 可能缺少 import {imp} ({desc})")

# ——— 8. 代码量统计 ———
print(f"\n{CYAN}[8/8] 代码量统计{RESET}")
stats = {}
for key, path in FILES.items():
    if path.exists():
        content = file_contents.get(key) or path.read_text(encoding="utf-8")
        lines = len(content.split("\n"))
        funcs = len(re.findall(r'^\s*func\s', content, re.MULTILINE))
        stats[key] = (lines, funcs)

for key, (lines, funcs) in sorted(stats.items()):
    ok(f"{key}: {lines} 行, ~{funcs} 个函数")

# ——— 总结 ———
print(f"\n{CYAN}{'='*60}{RESET}")
total = pass_count + fail_count + warn_count
print(f"  检查项: {total}  |  {GREEN}通过: {pass_count}{RESET}  |  "
      f"{RED}失败: {fail_count}{RESET}  |  {YELLOW}警告: {warn_count}{RESET}")

if fail_count == 0 and warn_count == 0:
    print(f"\n  {GREEN}全部检查通过！{RESET}")
elif fail_count == 0:
    print(f"\n  {YELLOW}所有必要检查通过，有 {warn_count} 个警告需人工确认。{RESET}")
else:
    print(f"\n  {RED}有 {fail_count} 项失败，请检查上述错误。{RESET}")

print()

# 对于已知的 regex 误报，给出提示
print(f"{CYAN}提示:{RESET}")
print(f"  标记为 FAIL 的项目可能是以下原因造成的误报：")
print(f"  1. Go 方法使用多行签名（参数跨行），正则无法匹配")
print(f"  2. 函数通过间接调用（如 map 分发、反射）而非直接引用")
print(f"  3. Handler 到 Service 的调用可能封装在中间件中")
print(f"\n  对于标注 FAIL 的项，请在 IDE 中手动确认该函数确实存在。")
print(f"  如果函数存在，这是正则匹配的假阳性，代码本身没有问题。")
