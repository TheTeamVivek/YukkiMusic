from pyrogram import filters
import io
from PIL import Image, ImageDraw, ImageFont
from YukkiMusic import app
from pyrogram.types import ChatMemberUpdated

@app.on_chat_member_updated(filters.group, group=-3)
async def handle_member_update(client, member):
    if member.new_chat_member and member.new_chat_member.status == "member":
        await send_welcome_message(member.new_chat_member)

@app.on_chat_member_updated(filters.group, group=-3)
async def handle_member_rejoin(client, member):
    if member.new_chat_member and member.new_chat_member.status == "member" and member.new_chat_member.user.id == app.id:
        await send_welcome_message(member.new_chat_member)

async def send_welcome_message(new_member):
    background = Image.open("assets/welcome.jpg")

    if new_member.user.photo:
        photo_down = await app.download_media(new_member.user.photo)
        image = Image.open(io.BytesIO(photo_down))
        image = image.resize((100, 100))
        image = image.convert("RGB")
        circle = Image.new("L", (100, 100), 0)
        draw = ImageDraw.Draw(circle)
        draw.ellipse((0, 0, 100, 100), fill=255)
        image = image.point(
            lambda x: min(x, 255) if circle.getpixel((x, x)) > 128 else 0,
            (0, 0, 100, 100),
        )

        draw = ImageDraw.Draw(background)
        font = ImageFont.truetype("assets/font2.ttf", 15)
        text = f"{new_member.user.first_name}\nID: {new_member.user.id}\nUSERNAME: @{new_member.user.username}"
        text_position = (10, 10)
        draw.text(text_position, text, fill="black", font=font)
        background.paste(image, (50, 50))
        output = io.BytesIO()
        background.save(output, format="JPEG")
        output.seek(0)
        await app.send_photo(new_member.chat.id, output)