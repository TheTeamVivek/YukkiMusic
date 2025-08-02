#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.


import logging
from logging.handlers import RotatingFileHandler

formatter = logging.Formatter(
    "{asctime} - {levelname} - {message}", style="{", datefmt="%d-%b-%y %H:%M:%S"
)

file_handler = RotatingFileHandler("logs.txt", maxBytes=5_000_000, backupCount=10)
file_handler.setFormatter(formatter)

stream_handler = logging.StreamHandler()
stream_handler.setFormatter(formatter)

logging.basicConfig(
    level=logging.INFO,
    format="{asctime} - {levelname} - {message}",
    style="{",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[file_handler, stream_handler],
)

noisy_modules = ["httpx", "ntgcalls", "pyrogram", "pytgcalls", "pymongo"]

for name in noisy_modules:
    noisy_logger = logging.getLogger(name)
    noisy_logger.setLevel(logging.WARNING)
    noisy_logger.handlers.clear()
    noisy_logger.addHandler(file_handler)
    noisy_logger.addHandler(stream_handler)
    noisy_logger.propagate = False
