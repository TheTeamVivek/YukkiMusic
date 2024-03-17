import stream
import time
import yt_dlp
from youtube_search import YoutubeSearch
from youtubesearchpython import VideosSearch
from youtubesearchpython import SearchVideos
import json


@app.on_message(
    filters.command(ADDPLAYLIST_COMMAND)
    & ~BANNED_USERS
)
@language
async def nothing_playlist(client, message: Message, _):
    if len(message.command) < 2:
        return await message.reply_text("**➻ ᴘʟᴇᴀsᴇ ᴘʀᴏᴠɪᴅᴇ ᴍᴇ ᴀ sᴏɴɢ ɴᴀᴍᴇ ᴏʀ sᴏɴɢ ʟɪɴᴋ ᴏʀ ʏᴏᴜᴛᴜʙᴇ ᴘʟᴀʏʟɪsᴛ ʟɪɴᴋ ᴀғᴛᴇʀ ᴛʜᴇ ᴄᴏᴍᴍᴀɴᴅ..**\n\n**➥ ᴇxᴀᴍᴘʟᴇs:**\n\n▷ `/addplaylist Blue Eyes` (ᴘᴜᴛ ᴀ sᴘᴇᴄɪғɪᴄ sᴏɴɢ ɴᴀᴍᴇ)\n\n▷ /addplaylist [ʏᴏᴜᴛᴜʙᴇ ᴘʟᴀʏʟɪsᴛ ʟɪɴᴋ] (ᴛᴏ ᴀᴅᴅ ᴀʟʟ sᴏɴɢs ғʀᴏᴍ ᴀ ʏᴏᴜᴛᴜʙᴇ ᴘʟᴀʏʟɪsᴛ ɪɴ ʙᴏᴛ ᴘʟᴀʏʟɪsᴛ.)")

    query = message.command[1]

    # Check if the provided input is a YouTube playlist link
    if "youtube.com/playlist" in query:
        adding = await message.reply_text("**🎧 ᴀᴅᴅɪɴɢ sᴏɴɢs ɪɴ ᴘʟᴀʏʟɪsᴛ ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ..**")
        try:
            from pytube import Playlist
            from pytube import YouTube

            playlist = Playlist(query)
            video_urls = playlist.video_urls

        except Exception as e:
            # Handle exception
            return await message.reply_text(f"Error: {e}")

        if not video_urls:
            return await message.reply_text("**➻ ɴᴏ sᴏɴɢs ғᴏᴜɴᴅ ɪɴ ᴛʜᴇ ᴘʟᴀʏʟɪsᴛ ʟɪɴᴋs.\n\n**➥ ᴛʀʏ ᴏᴛʜᴇʀ ᴘʟᴀʏʟɪsᴛ ʟɪɴᴋ**")

        user_id = message.from_user.id
        for video_url in video_urls:
            video_id = video_url.split("v=")[-1]

            try:
                yt = YouTube(video_url)
                title = yt.title
                duration = yt.length
            except Exception as e:
                return await message.reply_text(f"ᴇʀʀᴏʀ ғᴇᴛᴄʜɪɴɢ ᴠɪᴅᴇᴏ ɪɴғᴏ: {e}")

            plist = {
                "videoid": video_id,
                "title": title,
                "duration": duration,
            }

            await save_playlist(user_id, video_id, plist)
        await adding.delete()
        return await message.reply_text(text="**➻ ᴀʟʟ sᴏɴɢs ʜᴀs ʙᴇᴇɴ ᴀᴅᴅᴇᴅ sᴜᴄᴄᴇssғᴜʟʟʏ ғʀᴏᴍ ʏᴏᴜʀ ʏᴏᴜᴛᴜʙᴇ ᴘʟᴀʏʟɪsᴛ ʟɪɴᴋ✅**\n\n**➥ ɪғ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ ʀᴇᴍᴏᴠᴇ ᴀɴʏ sᴏɴɢ ᴛʜᴇɴ ᴄʟɪᴄᴋ ɢɪᴠᴇɴ ʙᴇʟᴏᴡ ʙᴜᴛᴛᴏɴ.\n\n**▷ ᴄʜᴇᴄᴋ ʙʏ » /playlist**\n\n▷ **ᴘʟᴀʏ ʙʏ » /play**")
        pass
    if "youtube.com/@" in query:
        addin = await message.reply_text("**🎧 ᴀᴅᴅɪɴɢ sᴏɴɢs ɪɴ ᴘʟᴀʏʟɪsᴛ ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ..**")
        try:
            from pytube import YouTube

            channel_username = query
            videos = YouTube_videos(f"{query}/videos")
            video_urls = [video['url'] for video in videos]

        except Exception as e:
            # Handle exception
            return await message.reply_text(f"Error: {e}")

        if not video_urls:
            return await message.reply_text("**➻ ɴᴏ sᴏɴɢs ғᴏᴜɴᴅ ɪɴ ᴛʜᴇ YouTube channel.\n\n**➥ ᴛʀʏ ᴏᴛʜᴇʀ YouTube channel ʟɪɴᴋ**")

        user_id = message.from_user.id
        for video_url in video_urls:
            videosid = query.split("/")[-1].split("?")[0]

            try:
                yt = YouTube(f"https://youtu.be/{videosid}")
                title = yt.title
                duration = yt.length
            except Exception as e:
                return await message.reply_text(f"ᴇʀʀᴏʀ ғᴇᴛᴄʜɪɴɢ ᴠɪᴅᴇᴏ ɪɴғᴏ: {e}")

            plist = {
                "videoid": video_id,
                "title": title,
                "duration": duration,
            }

            await save_playlist(user_id, video_id, plist)
           
        await addin.delete()
        return await message.reply_text(text="**➻ ᴀʟʟ sᴏɴɢs ʜᴀs ʙᴇᴇɴ ᴀᴅᴅᴇᴅ sᴜᴄᴄᴇssғᴜʟʟʏ ғʀᴏᴍ ʏᴏᴜʀ ʏᴏᴜᴛᴜʙᴇ channel ʟɪɴᴋ✅**\n\n**➥ ɪғ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ ʀᴇᴍᴏᴠᴇ ᴀɴʏ sᴏɴɢ ᴛʜᴇɴ ᴄʟɪᴄᴋ ɢɪᴠᴇɴ ʙᴇʟᴏᴡ ʙᴜᴛᴛᴏɴ.\n\n**▷ ᴄʜᴇᴄᴋ ʙʏ » /playlist**\n\n▷ **ᴘʟᴀʏ ʙʏ » /play**")
        pass
    # Check if the provided input is a YouTube video link
    if "https://youtu.be" in query:
        try:
            add = await message.reply_text("**🎧 ᴀᴅᴅɪɴɢ sᴏɴɢs ɪɴ ᴘʟᴀʏʟɪsᴛ ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ..**")
            from pytube import Playlist
            from pytube import YouTube
            # Extract video ID from the YouTube lin
            videoid = query.split("/")[-1].split("?")[0]
            user_id = message.from_user.id
            thumbnail = f"https://img.youtube.com/vi/{videoid}/maxresdefault.jpg"
            _check = await get_playlist(user_id, videoid)
            if _check:
                try:
                    await add.delete()
                    return await message.reply_photo(thumbnail, caption=_["playlist_8"])
                except KeyError:
                    pass

            _count = await get_playlist_names(user_id)
            count = len(_count)
            if count == SERVER_PLAYLIST_LIMIT:
                try:
                    return await message.reply_text(_["playlist_9"].format(SERVER_PLAYLIST_LIMIT))
                except KeyError:
                    pass

            try:
                yt = YouTube(f"https://youtu.be/{videoid}")
                title = yt.title
                duration = yt.length
                thumbnail = f"https://img.youtube.com/vi/{videoid}/maxresdefault.jpg"
                plist = {
                    "videoid": videoid,
                    "title": title,
                    "duration": duration,
                }
                await save_playlist(user_id, videoid, plist)
                await add.delete()
                await message.reply_photo(thumbnail, caption="**➻ ᴀᴅᴅᴇᴅ sᴏɴɢ ɪɴ ʏᴏᴜʀ ʙᴏᴛ ᴘʟᴀʏʟɪsᴛ✅**\n\n**➥ ᴄʜᴇᴄᴋ ʙʏ » /playlist**\n\n**➥ ᴅᴇʟᴇᴛᴇ ʙʏ » /delplaylist**\n\n**➥ ᴀɴᴅ ᴘʟᴀʏ ʙʏ » /play (ɢʀᴏᴜᴘs ᴏɴʟʏ)**")
            except Exception as e:
                print(f"Error: {e}")
                await message.reply_text(str(e))
        except Exception as e:
            return await message.reply_text(str(e))
            pass
    else:
        from YukkiMusic import YouTube
        # Add a specific song by name
        query = " ".join(message.command[1:])
        print(query)

        try:
            results = YoutubeSearch(query, max_results=1).to_dict()
            link = f"https://youtube.com{results[0]['url_suffix']}"
            title = results[0]["title"][:40]
            thumbnail = results[0]["thumbnails"][0]
            thumb_name = f"{title}.jpg"
            thumb = requests.get(thumbnail, allow_redirects=True)
            open(thumb_name, "wb").write(thumb.content)
            duration = results[0]["duration"]
            videoid = results[0]["id"]
            views = results[0]["views"]
            channel_name = results[0]["channel"]

            user_id = message.from_user.id
            _check = await get_playlist(user_id, videoid)
            if _check:
                try:
                    return await message.reply_photo(thumbnail, caption=_["playlist_8"])
                except KeyError:
                    pass

            _count = await get_playlist_names(user_id)
            count = len(_count)
            if count == SERVER_PLAYLIST_LIMIT:
                try:
                    return await message.reply_text(_["playlist_9"].format(SERVER_PLAYLIST_LIMIT))
                except KeyError:
                    pass

            m = await message.reply("**🔄 ᴀᴅᴅɪɴɢ ᴘʟᴇᴀsᴇ ᴡᴀɪᴛ... **")
            title, duration_min, _, _, _ = await YouTube.details(videoid, True)
            title = (title[:50]).title()
            plist = {
                "videoid": videoid,
                "title": title,
                "duration": duration_min,
            }

            await save_playlist(user_id, videoid, plist)
            await m.delete()
            await message.reply_photo(thumbnail, caption="**➻ ᴀᴅᴅᴇᴅ sᴏɴɢ ɪɴ ʏᴏᴜʀ ʙᴏᴛ ᴘʟᴀʏʟɪsᴛ✅**\n\n**➥ ᴄʜᴇᴄᴋ ʙʏ » /playlist**\n\n**➥ ᴅᴇʟᴇᴛᴇ ʙʏ » /delplaylist**\n\n**➥ ᴀɴᴅ ᴘʟᴀʏ ʙʏ » /play (ɢʀᴏᴜᴘs ᴏɴʟʏ)**")

        except KeyError:
            return await message.reply_text("ɪɴᴠᴀʟɪᴅ ᴅᴀᴛᴀ ғᴏʀᴍᴀᴛ ʀᴇᴄᴇɪᴠᴇᴅ.")
        except Exception as e:
            pass
