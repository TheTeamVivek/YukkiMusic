from pytubefix import YouTube
from pytubefix.cli import on_progress

url = "https://youtu.be/ZQEXRaWSvQ4?si=gC1Kw47OLOADANKa"

yt = YouTube(url, on_progress_callback=on_progress)
print(yt.title)