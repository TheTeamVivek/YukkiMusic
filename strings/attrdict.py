#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved

# pylint: disable=missing-module-docstring


class AttrDict(dict):
    """Dict with attribute-style access."""

    def __init__(self, *args, **kwargs):
        super().__init__()
        data = dict(*args, **kwargs)
        for k, v in data.items():
            self[k] = self._wrap(v)

    def __getattr__(self, item):
        try:
            return self[item]
        except KeyError as e:
            raise AttributeError(f"'AttrDict' has no attribute '{item}'") from e

    def __setattr__(self, key, value):
        self[key] = value

    def __setitem__(self, key, value):
        super().__setitem__(key, self._wrap(value))

    def update(self, *args, **kwargs):
        for k, v in dict(*args, **kwargs).items():
            self[k] = self._wrap(v)

    def setdefault(self, key, default=None):
        return super().setdefault(key, self._wrap(default))

    @staticmethod
    def _wrap(value):
        if isinstance(value, dict):
            return AttrDict(value)
        if isinstance(value, list):
            return [AttrDict(v) if isinstance(v, dict) else v for v in value]
        return value
