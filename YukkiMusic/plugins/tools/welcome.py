from pyrogram import filters
import io
from PIL import Image, ImageDraw, ImageFont
from YukkiMusic import app


# Define the on_member_added event
@app.on_member_added()
async def handle_new_member(client, member):
    # Open the background image
    background = Image.open("assets/welcome.jpg")

    # Get the user's profile photo
    photo = member.photo

    # If the user has a profile photo
    if photo:
        # Download the user's profile photo
        photo_down = await client.download_media(photo)

        # Open the user's profile photo using Pillow
        image = Image.open(io.BytesIO(photo_down))

        # Resize the user's profile photo to fit in a circle
        image = image.resize((100, 100))
        image = image.convert("RGB")
        circle = Image.new("L", (100, 100), 0)
        draw = ImageDraw.Draw(circle)
        draw.ellipse((0, 0, 100, 100), fill=255)
        image = image.point(
            lambda x: min(x, 255) if circle.getpixel((x, y)) > 128 else 0,
            (0, 0, 100, 100),
        )

        # Create a drawing context
        draw = ImageDraw.Draw(background)

        # Define the font for the text
        font = ImageFont.truetype("assets/font2.ttf", 15)

        # Define the text to be added
        text = f"{member.name}\nID: {member.id}\nUSERNAME: @{member.username}"

        # Define the position to add the text
        text_position = (10, 10)

        # Add the text to the image
        draw.text(text_position, text, fill="black", font=font)

        # Paste the user's profile photo onto the background image
        background.paste(image, (50, 50))

        # Save the new image
        output = io.BytesIO()
        background.save(output, format="JPEG")
        output.seek(0)

        # Send the new image to the group
        await client.send_photo(member.group.id, output)
