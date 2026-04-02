import socket
import http.client
import os
import subprocess

def test_dns(host="google.com"):
    print(f"--- 测试 DNS 解析: {host} ---")
    try:
        addr = socket.gethostbyname(host)
        print(f"成功: {host} -> {addr}")
        return True
    except Exception as e:
        print(f"失败: {e}")
        return False

def test_tcp_connection(host="127.0.0.1", port=7897):
    print(f"--- 测试本地代理端口连接 (Clash): {host}:{port} ---")
    try:
        s = socket.create_connection((host, port), timeout=5)
        print(f"成功: 已连接到 {host}:{port}")
        s.close()
        return True
    except Exception as e:
        print(f"失败: {e}")
        return False

def test_http_request(url="www.google.com"):
    print(f"--- 测试 HTTP 请求 (通过代理): {url} ---")
    proxy_host = "127.0.0.1"
    proxy_port = 7897
    try:
        conn = http.client.HTTPConnection(proxy_host, proxy_port, timeout=10)
        conn.set_tunnel(url)
        conn.request("GET", "/")
        response = conn.getresponse()
        print(f"成功: 状态码 {response.status} {response.reason}")
        return True
    except Exception as e:
        print(f"失败: {e}")
        return False

def test_subprocess_spawn():
    print("--- 测试子进程派生能力 ---")
    try:
        result = subprocess.run(["ls", "-l", "/Applications"], capture_output=True, text=True, timeout=5)
        print(f"成功: 获取到 {len(result.stdout)} 字节输出")
        return True
    except Exception as e:
        print(f"失败: {e}")
        return False

if __name__ == "__main__":
    print("=== Antigravity 环境诊断脚本 ===\n")
    test_dns()
    print()
    test_tcp_connection()
    print()
    test_http_request()
    print()
    test_subprocess_spawn()
    print("\n=== 诊断完成 ===")
