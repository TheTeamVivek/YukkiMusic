#
# Copyright (C) 2024 by AnonymousX888@Github, < https://github.com/AnonymousX888 >.
#
# This file is part of < https://github.com/hakutakaid/YukkiMusicFork > project,
# and is released under the MIT License.
# Please see < https://github.com/hakutakaid/YukkiMusicFork/blob/master/LICENSE >
#
# All rights reserved.
#


class AssistantErr(Exception):
    def __init__(self, errr: str):
        super().__init__(errr)
