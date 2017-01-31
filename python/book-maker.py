from __future__ import print_function
from bs4 import BeautifulSoup
import os
import requests
import argparse
import re
import json
import yaml
import sys

if sys.version_info[0] == 2:
	from urlparse import urljoin
else:
	from urllib.parse import urljoin

parser = argparse.ArgumentParser(description="generate content.json for rssbinge")
parser.add_argument('start_url', metavar='URL', type=str, help='url to start')

args = parser.parse_args()

url = args.start_url

re_url = re.compile('^http://whateleyacademy.net/index.php/(?:([a-z0-9-]+)/)+(\d+)-(.*)$')

content = []

def process_page(url):
	resp = requests.post(url, data={'filter_order': 'a.publish_up', 'filter_order_Dir': 'asc'})
	print("processing", url)
	doc = BeautifulSoup(resp.text, "html.parser")
	for tr in doc.find('table', **{'class': re.compile('category')}).tbody.find_all('tr'):
		tds = tr.find_all('td')
		itemurl = urljoin(url, tds[0].a['href'])
		itemname = tds[0].get_text()
		m = re_url.match(itemurl)
		content.append({'url': itemurl, 'title': itemname.strip(), 'id': m.group(2)})
	pagination = doc.find('div', **{'class': 'pagination'})
	if pagination:
		return pagination.find('a', title="Next")

cur_url = url
while cur_url is not None:
	ret = process_page(cur_url)
	if ret is None:
		cur_url = None
	else:
		cur_url = urljoin(cur_url, ret['href'])

book_parts = [{'toc': item['title'], 'story': {'id': int(item['id'])}} for item in content]
yml_file = {
	'id': 'gen1',
	'title': 'Whateley - Gen1 Complete',
	'author': 'Whateley Academy Canon Authors',
	'author-sort': 'Whateley Academy Canon Authors',
	'uuid': '0b6b1f75-3483-4657-b7f8-3b6c657e9a44',
	'series': 'Whateley Academy Original Canon',
	'publisher': 'Whateley Press',
		}
yml_file['parts'] = book_parts

with open('output.yml', 'w') as f:
	f.write(yaml.dump(yml_file))
#print(json.dumps(list(book_parts)))
