from pyrogram import filters
import io
from PIL import Image, ImageDraw, ImageFont
from YukkiMusic import app
from pyrogram.types import ChatMemberUpdated

@app.on_chat_member_updated(filters.group, group=-3)
async def handle_new_member(client, member):
    background = Image.open("assets/welcome.jpg")

    if member.new_chat_member.status == "member":
        photo = member.new_chat_member.user.photo
        if photo:
            photo_down = await client.download_media(photo)
            image = Image.open(io.BytesIO(photo_down))
            image = image.resize((100, 100))
            image = image.convert("RGB")
            circle = Image.new("L", (100, 100), 0)
            draw = ImageDraw.Draw(circle)
            draw.ellipse((0, 0, 100, 100), fill=255)
            image = image.point(
                lambda x: min(x, 255) if circle.getpixel((x, y)) > 128 else 0,
                (0, 0, 100, 100),
            )

            draw = ImageDraw.Draw(background)
            font = ImageFont.truetype("assets/font2.ttf", 15)
            text = f"{member.new_chat_member.user.first_name}\nID: {member.new_chat_member.user.id}\nUSERNAME: @{member.new_chat_member.user.username}"
            text_position = (10, 10)
            draw.text(text_position, text, fill="black", font=font)
            background.paste(image, (50, 50))
            output = io.BytesIO()
            background.save(output, format="JPEG")
            output.seek(0)
            await client.send_photo(member.chat.id, output)