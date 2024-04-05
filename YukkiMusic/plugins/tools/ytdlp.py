import os
from pyrogram import filters
from youtubesearchpython import VideosSearch
from pytube import YouTube, Stream
from YukkiMusic import app

# Function to handle /yt command
@app.on_message(filters.command("yt"))
async def youtube_dl(client, message):
    # Check if the command has any arguments
    if len(message.command) == 1:
        await message.reply_text("Please specify a video name or link to download.")
        return
    
    # Get the text after /yt command
    query = " ".join(message.command[1:])
    
    # Check if the provided query is a YouTube shorts link
    if 'shorts' in query:
        stream = get_shorts_video(query)
        if stream:
            await send_video(client, message, stream)
        else:
            await message.reply_text("Sorry, couldn't download the video.")
    else:
        # Search for videos using youtube-search-python
        videosSearch = VideosSearch(query, limit=1)
        result = videosSearch.result()["result"][0]
        
        # Get the video link
        video_url = result['link']
        
        # Download the video with 360p quality if available, else download with 480p quality
        stream = get_regular_video(video_url)
        if stream:
            await send_video(client, message, stream)
        else:
            await message.reply_text("Sorry, couldn't download the video.")

def get_shorts_video(url: str) -> Stream:
    try:
        yt = YouTube(url)
        stream = yt.streams.get_by_itag('22')  # Medium quality for shorts
        return stream
    except:
        return None

def get_regular_video(url: str) -> Stream:
    try:
        yt = YouTube(url)
        stream_360p = yt.streams.filter(res='360p').first()
        if stream_360p:
            return stream_360p
        else:
            stream_480p = yt.streams.filter(res='480p').first()
            return stream_480p
    except:
        return None

async def send_video(client, message, stream):
    # Save the thumbnail locally
    thumbnail_filename = f"thumbnail_{stream.video_id}.jpg"
    with open(thumbnail_filename, 'wb') as thumbnail_file:
        thumbnail_file.write(stream.thumbnail_url)
    
    # Download the video
    stream.download(filename='video')
    
    # Send the downloaded video with thumbnail
    await message.reply_video(
        video='video.mp4',
        thumb=thumbnail_filename,
        caption=f"{stream.title}\nDuration: {stream.length} seconds"
    )
    
    # Remove the downloaded files
    os.remove('video.mp4')
    os.remove(thumbnail_filename)