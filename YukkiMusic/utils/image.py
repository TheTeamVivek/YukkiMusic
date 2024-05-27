import requests

from config import START_IMG_URL


def gen_image():
    try:
        url = "https://random.imagecdn.app/v1/image?width=1280&height=720&format=json"

        response = requests.get(url)
        data = response.json()
        Z = data.get("url")
        return Z
    except Exception:
        return START_IMG_URL
