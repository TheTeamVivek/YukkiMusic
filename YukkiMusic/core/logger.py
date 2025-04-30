import logging
from logging.handlers import RotatingFileHandler

from config import LOG_FILE_NAME


def setup_logger():
    logging.basicConfig(
        level=logging.INFO,
        format="[%(asctime)s - %(levelname)s] - %(name)s - %(message)s",
        datefmt="%d-%b-%y %H:%M:%S",
        handlers=[
            RotatingFileHandler(LOG_FILE_NAME, maxBytes=5000000, backupCount=10),
            logging.StreamHandler(),
        ],
    )

    logging.getLogger("telethon").setLevel(logging.ERROR)

    logging.getLogger("pyrogram").setLevel(logging.ERROR)
    logging.getLogger("pytgcalls").setLevel(logging.ERROR)
    logging.getLogger("pymongo").setLevel(logging.ERROR)
    logging.getLogger("httpx").setLevel(logging.ERROR)

    logging.getLogger("ntgcalls").setLevel(logging.CRITICAL)
