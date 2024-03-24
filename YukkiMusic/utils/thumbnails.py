#
# Copyright (C) 2021-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#
import os
import re
import textwrap
import aiofiles
import aiohttp
import numpy as np
from PIL import Image, ImageChops, ImageDraw, ImageEnhance, ImageFilter, ImageFont
from youtubesearchpython.__future__ import VideosSearch
from config import YOUTUBE_IMG_URL
from YukkiMusic import app

def changeImageSize(maxWidth, maxHeight, image):
    widthRatio = maxWidth / image.size[0]
    heightRatio = maxHeight / image.size[1]
    newWidth = int(widthRatio * image.size[0])
    newHeight = int(heightRatio * image.size[1])
    newImage = image.resize((newWidth, newHeight))
    return newImage


def add_corners(im):
    bigsize = (im.size[0] * 3, im.size[1] * 3)
    mask = Image.new("L", bigsize, 0)
    ImageDraw.Draw(mask).ellipse((0, 0) + bigsize, fill=255)
    mask = mask.resize(im.size, Image.LANCZOS)
    mask = ImageChops.darker(mask, im.split()[-1])
    im.putalpha(mask)


def circle(img):
    h, w = img.size
    a = Image.new('L', [h, w], 0)
    b = ImageDraw.Draw(a)
    b.pieslice([(0, 0), (h, w)], 0, 360, fill=255, outline="white")
    c = np.array(img)
    d = np.array(a)
    e = np.dstack((c, d))
    return Image.fromarray(e)

async def gen_thumb(videoid, user_id):
    if os.path.isfile(f"cache/{videoid}_{user_id}.png"):
        return f"cache/{videoid}_{user_id}.png"

    url = f"https://www.youtube.com/watch?v={videoid}"
    try:
        results = VideosSearch(url, limit=1)
        for result in (await results.next())["result"]:
            try:
                title = result["title"][:30]
                title = re.sub("\W+", " ", title)
                title = title.title()
            except:
                title = "Unsupported Title"
            try:
                duration = result["duration"]
            except:
                duration = "Unknown"
            try:
                views = result["viewCount"]["short"]
            except:
                views = "Unknown"  # Default value if view count is not available
            thumbnail = result["thumbnails"][0]["url"].split("?")[0]
            try:
                result["channel"]["name"]
            except:
                pass

        async with aiohttp.ClientSession() as session:
            async with session.get(thumbnail) as resp:
                if resp.status == 200:
                    f = await aiofiles.open(f"cache/thumb{videoid}.png", mode="wb")
                    await f.write(await resp.read())
                    await f.close()

        try:
            wxy = await app.download_media(
                (await app.get_users(user_id)).photo.big_file_id,
                file_name=f"{user_id}.jpg",
            )
        except:
            wxy = await app.download_media(
                (await app.get_users(app.id)).photo.big_file_id,
                file_name=f"{app.id}.jpg",
            )
        xy = Image.open(wxy)
        a = Image.new("L", [640, 640], 0)
        b = ImageDraw.Draw(a)
        b.pieslice([(0, 0), (640, 640)], 0, 360, fill=255, outline="white")
        c = np.array(xy)
        d = np.array(a)
        e = np.dstack((c, d))
        f = Image.fromarray(e)
        x = f.resize((153, 153))

        youtube = Image.open(f"cache/thumb{videoid}.png")
        zyoutube = Image.open(f"cache/thumb{videoid}.png")
        bg = Image.open(f"assets/ZSTREAM_adobe_express.png")
        image1 = youtube.resize((1280, 720))
        image2 = image1.convert("RGBA")
        background = image2.filter(filter=ImageFilter.BoxBlur(16))
        enhancer = ImageEnhance.Brightness(background)
        background = enhancer.enhance(0.6)
        y = circle(zyoutube).resize((474, 474))
        background.paste(y, (49, 125), mask=y)  # Adjusted placement of YouTube circle image
        image3 = bg.resize((1280, 720))
        image5 = image3.convert("RGBA")
        result_img = Image.alpha_composite(background, image5)
        result_img.paste(x, (353, 471), mask=x)  # Adjusted placement of user photo
        # Adding text to the thumbnail
        draw = ImageDraw.Draw(result_img)
        font = ImageFont.truetype("assets/font2.ttf", 54)
        font2 = ImageFont.truetype("assets/font2.ttf", 89)
        para = textwrap.wrap(title, width=32)
        try:
            text_w, text_h = draw.textsize(f"NOW PLAYING", font=font2)
            draw.text(
                ((1680 - text_w) / 2, 75),
                f"NOW PLAYING",
                fill="white",
                font=font2,
            )
            if para[0]:
                draw.text(
                    (535, 255),
                    f"Title: {para[0]}",
                    fill="white",
                    stroke_width=1,
                    stroke_fill="white",
                    font=font,
                )
            if para[1]:
                draw.text(
                    (535, 255),
                    f"Title: {para[1]}",
                    fill="white",
                    stroke_width=1,
                    stroke_fill="black",
                    font=font,
                )
        except:
            pass
        text_w, text_h = draw.textsize(f"Duration: {duration} Mins", font=font)
        draw.text(
            ((1557 - text_w) / 2, 342),
            f"Duration: {duration} Mins",
            fill="white",
            font=font,
        )
        text_a, text_b = draw.textsize(f"Views: {views}", font=font)
        draw.text(
            ((1530 - text_a) / 2, 432),
            f"Views: {views} ",
            fill="white",
            font=font,
        )
        try:
            os.remove(f"cache/thumb{videoid}.png")
        except:
            pass

        result_img.save(f"cache/{videoid}_{user_id}.png")

        return f"cache/{videoid}_{user_id}.png"

    except Exception as e:
        print(e)
        return YOUTUBE_IMG_URL
    

async def gen_qthumb(videoid, user_id):
    if os.path.isfile(f"cache/{videoid}_{user_id}.png"):
        return f"cache/{videoid}_{user_id}.png"

    url = f"https://www.youtube.com/watch?v={videoid}"
    try:
        results = VideosSearch(url, limit=1)
        for result in (await results.next())["result"]:
            try:
                title = result["title"][:30]
                title = re.sub("\W+", " ", title)
                title = title.title()
            except:
                title = "Unsupported Title"
            try:
                duration = result["duration"]
            except:
                duration = "Unknown"
            try:
                views = result["viewCount"]["short"]
            except:
                views = "Unknown"  # Default value if view count is not available
            thumbnail = result["thumbnails"][0]["url"].split("?")[0]
            try:
                result["channel"]["name"]
            except:
                pass

        async with aiohttp.ClientSession() as session:
            async with session.get(thumbnail) as resp:
                if resp.status == 200:
                    f = await aiofiles.open(f"cache/thumb{videoid}.png", mode="wb")
                    await f.write(await resp.read())
                    await f.close()

        try:
            wxy = await app.download_media(
                (await app.get_users(user_id)).photo.big_file_id,
                file_name=f"{user_id}.jpg",
            )
        except:
            wxy = await app.download_media(
                (await app.get_users(app.id)).photo.big_file_id,
                file_name=f"{app.id}.jpg",
            )
        xy = Image.open(wxy)
        a = Image.new("L", [640, 640], 0)
        b = ImageDraw.Draw(a)
        b.pieslice([(0, 0), (640, 640)], 0, 360, fill=255, outline="white")
        c = np.array(xy)
        d = np.array(a)
        e = np.dstack((c, d))
        f = Image.fromarray(e)
        x = f.resize((153, 153))

        youtube = Image.open(f"cache/thumb{videoid}.png")
        zyoutube = Image.open(f"cache/thumb{videoid}.png")
        bg = Image.open(f"assets/ZSTREAM_adobe_express.png")
        image1 = youtube.resize((1280, 720))
        image2 = image1.convert("RGBA")
        background = image2.filter(filter=ImageFilter.BoxBlur(16))
        enhancer = ImageEnhance.Brightness(background)
        background = enhancer.enhance(0.6)
        y = circle(zyoutube).resize((473, 473))
        background.paste(y, (49, 125), mask=y)  # Adjusted placement of YouTube circle image
        image3 = bg.resize((1280, 720))
        image5 = image3.convert("RGBA")
        result_img = Image.alpha_composite(background, image5)
        result_img.paste(x, (353, 472), mask=x)  # Adjusted placement of user photo
        # Adding text to the thumbnail
        draw = ImageDraw.Draw(result_img)
        font = ImageFont.truetype("assets/font2.ttf", 54)
        font2 = ImageFont.truetype("assets/font2.ttf", 89)
        para = textwrap.wrap(title, width=32)
        try:
            text_w, text_h = draw.textsize(f"ADDED TO QUEUE", font=font2)
            draw.text(
                ((1680 - text_w) / 2, 75),
                f"ADDED TO QUEUE",
                fill="white",
                font=font2,
            )
            if para[0]:
                draw.text(
                    (535, 255),
                    f"Title: {para[0]}",
                    fill="white",
                    stroke_width=1,
                    stroke_fill="white",
                    font=font,
                )
            if para[1]:
                draw.text(
                    (535, 255),
                    f"Title: {para[1]}",
                    fill="white",
                    stroke_width=1,
                    stroke_fill="black",
                    font=font,
                )
        except:
            pass
        text_w, text_h = draw.textsize(f"Duration: {duration} Mins", font=font)
        draw.text(
            ((1557 - text_w) / 2, 342),
            f"Duration: {duration} Mins",
            fill="white",
            font=font,
        )
        text_a, text_b = draw.textsize(f"Views: {views}", font=font)
        draw.text(
            ((1530 - text_a) / 2, 432),
            f"Views: {views} ",
            fill="white",
            font=font,
        )
        try:
            os.remove(f"cache/thumb{videoid}.png")
        except:
            pass

        result_img.save(f"cache/{videoid}_{user_id}.png")

        return f"cache/{videoid}_{user_id}.png"

    except Exception as e:
        print(e)
        return YOUTUBE_IMG_URL
    

    
