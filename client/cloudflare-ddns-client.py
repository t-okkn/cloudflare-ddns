# -*- coding: utf-8 -*-

from ipaddress import ip_address
import json
import os
from pprint import pprint
import sys

import requests

GLOBAL_IPFILE = '/path/to/globalip.txt'
URL_BASE = 'https://example.com/cloudflare-ddns/v1'
REQEST_CONFIG = {
    'zone_name': 'example.com',
    'hostnames': ['test']
}

def get_global_ip():
    url = 'https://checkip.amazonaws.com'
    r = requests.get(url)

    if r.status_code != 200:
        return '0.0.0.0'

    i = ip_address(r.text.strip())

    if i.version == 4 and i.is_global:
        return str(i)

    else:
        return '0.0.0.0'

def generate_request_data(global_ip):
    res = {
        'zone_name': REQEST_CONFIG['zone_name'],
        'contents': []
    }

    for hostname in REQEST_CONFIG['hostnames']:
        res['contents'].append({
            'host_name': hostname,
            'content': global_ip
        })

    return res

def main():
    if len(sys.argv) != 2:
        print(f'Usage: python3 {sys.argv[0]} <API_TOKEN>')
        sys.exit(127)

    token = sys.argv[1]
    current_global_ip = ''

    if os.path.isfile(GLOBAL_IPFILE):
        with open(GLOBAL_IPFILE, 'r') as f:
            current_global_ip = f.read().strip()

    now_global_ip = get_global_ip()

    if now_global_ip == '0.0.0.0':
        print('インターネット接続か名前解決に問題が発生しています')
        sys.exit(126)

    if current_global_ip != '' and current_global_ip == now_global_ip:
        print(f'[{now_global_ip}] グローバルIPアドレスに変更はありません')
        sys.exit(0)

    with open(GLOBAL_IPFILE, 'w', encoding='utf-8') as f:
        f.write(now_global_ip)

    url = f'{URL_BASE}/ipv4'
    headers = {
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json'
    }
    payload = json.dumps(generate_request_data(now_global_ip))

    r = requests.put(url, headers=headers, data=payload)
    response = r.json()

    if r.status_code != 200:
        if 'error' in response:
            print(f'[{response["error"]}] {response["message"]}')

        else:
            pprint(response)

        sys.exit(1)

    else:
        if 'results' in response:
            count = 0

            for result in response['results']:
                if result['succeeded']:
                    count += 1
                    print(f'[{result["name"]}] グローバルIPアドレスの更新成功')

                else:
                    print(f'[{result["name"]}] {result["error"]}')

            if count == len(response['results']):
                sys.exit(0)

            else:
                sys.exit(2)

        else:
            pprint(response)
            sys.exit(3)

if __name__ == '__main__':
    main()
