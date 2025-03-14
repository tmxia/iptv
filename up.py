import os
import re
import requests
from bs4 import BeautifulSoup
import time
import hmac
import hashlib
import base64
import urllib.parse
import json

def get_latest_zip():
    """获取GitHub仓库最新的ZIP文件"""
    github_url = 'https://github.com/fish2018/ZX'
    response = requests.get(github_url)
    soup = BeautifulSoup(response.text, 'html.parser')
    
    zip_files = []
    table = soup.find('table')
    if table:
        rows = table.find_all('tr')
        for row in rows:
            tds = row.find_all('td')
            if len(tds) >= 4:
                file_name = tds[1].text.strip()
                update_time = tds[3].text.strip()
                if file_name.endswith('.zip'):
                    zip_files.append((file_name, update_time))
    
    if zip_files:
        zip_files.sort(key=lambda x: x[1], reverse=True)
        return zip_files[0][0]
    return None

def extract_version_from_filename(file_name):
    """从文件名中提取8位日期版本号（YYYYMMDD）"""
    match = re.search(r'\d{8}', file_name)
    if match:
        return match.group()
    return None

def generate_signed_url(file_name, secret_key):
    """生成带校验的URL（通过ghproxy代理）"""
    # 构造GitHub原始直链
    encoded_file = urllib.parse.quote(file_name)
    github_url = f'https://raw.githubusercontent.com/fish2018/ZX/main/{encoded_file}'
    
    # 添加ghproxy代理
    base_url = f'https://ghproxy.net/{github_url}'
    
    # 生成校验参数
    current_time = int(time.time())
    expires = current_time + 86400  # 24小时有效期
    
    # 构造签名数据
    sign_data = f'{base_url}&e={expires}'.encode()
    
    # 计算HMAC-MD5并进行URL安全的Base64编码
    signature = hmac.new(secret_key, sign_data, hashlib.md5).digest()
    signature_base64 = base64.urlsafe_b64encode(signature).decode().strip('=')
    
    # 组合完整URL
    return f'{base_url}?st={signature_base64}&e={expires}'

def main():
    SECRET_KEY = b'your_secret_key_here'  # 请替换为你的密钥
    
    latest_file = get_latest_zip()
    if not latest_file:
        print("未找到ZIP文件")
        return
    
    version = extract_version_from_filename(latest_file)
    if not version:
        print("错误：无法从文件名中提取版本号（格式需包含8位数字如20250313）")
        return
    
    signed_url = generate_signed_url(latest_file, SECRET_KEY)
    
    output = {
        "name": "凯速仓库",
        "list": [
            {
                "name": "更新版本",
                "url": signed_url,
                "icon": "",
                "version": version
            },
            {
                "name": "当前版本",
                "url": "",
                "icon": "http://127.0.0.1:9978/file/tvbox/img.png",
                "version": ""
            }
        ]
    }
    
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_path = os.path.join(script_dir, 'market.json')
    
    try:
        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(output, f, indent=2, ensure_ascii=False)
        print(f"成功生成market.json文件：{output_path}")
    except IOError as e:
        print(f"文件写入失败：{str(e)}")

if __name__ == "__main__":
    main()
