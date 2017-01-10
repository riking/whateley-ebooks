from __future__ import print_function
from bs4 import BeautifulSoup
import os
import requests
import argparse
import re
import json
from urlparse import urljoin

parser = argparse.ArgumentParser(description="generate content.json for rssbinge")
parser.add_argument('start_url', metavar='URL', type=str, help='url to start')

args = parser.parse_args()

url = args.start_url

content = []

def process_page(url):
	resp = requests.get(url)
	print("processing", url)
	doc = BeautifulSoup(resp.text, "html.parser")
	for tr in doc.find('table', **{'class': re.compile('category')}).tbody.find_all('tr'):
		tds = tr.find_all('td')
		itemurl = urljoin(url, tds[0].a['href'])
		itemname = tds[0].get_text()
		author = tds[2].get_text()
		print(itemurl, itemname.strip(), author.strip())
		content.append({'url': itemurl, 'title': itemname.strip(), 'description': author.strip()})
	return doc.find('div', **{'class': 'pagination'}).find('a', title="Next")

cur_url = url
while cur_url is not None:
	ret = process_page(cur_url)
	if ret is None:
		cur_url = None
	else:
		cur_url = urljoin(cur_url, ret['href'])

with open('content.json', 'w') as f:
	f.write(json.dumps(list(reversed(content))))
