import subprocess
import os

def run_command(cmd):
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=10)
        return result.stdout.strip(), result.stderr.strip()
    except Exception as e:
        return "", str(e)

def check_quarantine(path="/Applications/Antigravity.app"):
    print(f"--- 检查苹果隔离属性 (Quarantine): {path} ---")
    stdout, stderr = run_command(f"xattr {path}")
    if "com.apple.quarantine" in stdout:
        print(f"警告: 发现隔离属性。应用可能被 macOS 限制了子进程启动能力。")
    else:
        print("成功: 未发现隔离属性。")

def check_sandbox_entitlements(path="/Applications/Antigravity.app"):
    print(f"--- 检查沙盒权利 (Entitlements): {path} ---")
    stdout, stderr = run_command(f"codesign -d --entitlements - {path}")
    if "com.apple.security.app-sandbox" in stdout:
        print("警告: 应用在沙盒中运行。")
        if "com.apple.security.network.client" not in stdout:
            print("错误: 沙盒中【没有】声明网络客户端权限！这就是无法解析域名和连接 127.0.0.1 的原因。")
        else:
            print("成功: 沙盒已声明网络权限，但可能被内核拦截。")
    else:
        print("成功: 应用似乎不在常规沙盒中，可能受限于路径或系统安全策略。")

def check_socket_filters():
    print("--- 检查内核网络过滤器 (Socket Filters) ---")
    stdout, stderr = run_command("grep -i 'filter' /var/log/system.log | tail -n 5")
    if stdout:
        print(f"最近日志: {stdout}")
    else:
        print("无明显拦截日志记录。")

def test_raw_socket():
    print("--- 测试原始 Socket 创建能力 (检测 Errno 1) ---")
    import socket
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(2)
        s.connect(("127.0.0.1", 7897))
    except Exception as e:
        print(f"捕捉到错误: {e}")
        if "Operation not permitted" in str(e):
            print("关键结论: 这是内核级别的强制拦截。通常由防火墙 (Firewall) 或 EDR 安全软件 (如 Carbon Black, Falcon, Little Snitch) 引起。")

if __name__ == "__main__":
    print("=== Antigravity 深度安全与驱动测试 ===\n")
    check_quarantine()
    print()
    check_sandbox_entitlements()
    print()
    test_raw_socket()
    print("\n=== 诊断完成 ===")
