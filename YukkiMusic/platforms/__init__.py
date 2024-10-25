#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from .Apple import Apple
from .Carbon import Carbon
from .JioSavan import Saavn
from .Resso import Resso
from .Soundcloud import SoundCloud
from .Spotify import Spotify
from .Telegram import Telegram
from .Youtube import YouTube


class PlaTForms:
    def __init__(self):
        self.apple = Apple()
        self.carbon = Carbon()
        self.saavn = Saavn()
        self.resso = Resso()
        self.soundcloud = SoundCloud()
        self.spotify = Spotify()
        self.telegram = Telegram()
        self.youtube = YouTube()
