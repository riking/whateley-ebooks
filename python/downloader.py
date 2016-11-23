import sys
import os
import urllib2
from os import path
from bs4 import BeautifulSoup

RootURL = "http://whateleyacademy.net/"
UserAgent =  "Ebook tool - Downloader (+github.com/riking/whateley-ebooks)"

Headers = {'User-Agent': UserAgent}


class WhateleyStory:
    def __init__(self, downloader, story_id):
        self._downloader = downloader
        self._id = story_id

    def get_url(self):
        if self._id == 341:
            return "http://whateleyacademy.net/index.php/original-timeline/341-tennyo-to-hell-part-1?showall=&start=1"
        elif self._id == 342:
            return "http://whateleyacademy.net/index.php/original-timeline/342-tennyo-to-hell-part-2?showall=&start=1"
        return RootURL + "index.php/-/%d--" % self._id

    def get_raw_filename(self):
        return path.join(self._downloader.htmlout, "%d-raw.html" % self._id)

    def get_filename(self):
        return path.join(self._downloader.htmlout, "%d-story.html" % self._id)

    def is_complete(self):
        return path.exists(self.get_filename())

    def download_raw(self):
        req = urllib2.Request(self.get_url(), None, Headers)
        resp = urllib2.urlopen(req)
        content = resp.read()
        f = open(self.get_raw_filename(), 'w')
        f.write(content)
        f.close()

    def process_story(self):
        f = open(self.get_raw_filename(), 'r')
        content = f.read()
        soup = BeautifulSoup(content, 'html.parser')



class WhateleyDownloader:
    def __init__(self):
        self.rootdir = os.getcwd()
        self.target = path.join(self.rootdir, "target")
        self.htmlout = path.join(self.target, "html")
        self.definitions = path.join(self.rootdir, "book-definitions")


    def download_page(self, ):

    def download_story(self, story_id):
        pass

    def run(self, story_id_list):
        for story_id in story_id_list:
            self.download_story(story_id)

if __name__ == "__main__":
    WhateleyDownloader().run(sys.argv[1:])
