import sys
import os
from os import path
from bs4 import BeautifulSoup

RootURL = "http://whateleyacademy.net/"
StoryURL = RootURL + "index.php/-/%d--"
UserAgent =  "Ebook tool - Downloader (+github.com/riking/whateley-ebooks)"

class WhateleyDownloader:
    def __init__(self):
        self.rootdir = os.getcwd()
        self.target = path.join(self.rootdir, "target")
        self.htmlout = path.join(self.target, "html")
        self.definitions = path.join(self.rootdir, "book-definitions")

    def get_url(self, story_id):
        if story_id == 341:
            return "http://whateleyacademy.net/index.php/original-timeline/341-tennyo-goes-to-hell-part-1?showall=&start=1"
        elif story_id == 342:
            return "http://whateleyacademy.net/index.php/original-timeline/342-tennyo-goes-to-hell-part-2?showall=&start=1"
        return StoryURL % story_id

    def download_story(self, story_id):
        pass

if __name__ == "__main__":
    WhateleyDownloader().run(sys.argv[1:])
