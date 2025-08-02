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
        for k, v in dict(*args, **kwargs).items():
            super().__setitem__(k, self._wrap(v))

    def __getattr__(self, item):
        try:
            return self[item]
        except KeyError:
            raise AttributeError(f"{self.__class__.__name__!r} object has no attribute {item!r}")

    def __setattr__(self, key, value):
        self[key] = self._wrap(value)

    def __setitem__(self, key, value):
        super().__setitem__(key, self._wrap(value))

    def update(self, *args, **kwargs):
        for k, v in dict(*args, **kwargs).items():
            self[k] = self._wrap(v)

    def setdefault(self, key, default=None):
        return super().setdefault(key, self._wrap(default))

    def __dir__(self):
        return list(self.keys()) + super().__dir__()

    def copy(self):
        return AttrDict(self)

    def __repr__(self):
        return f"{self.__class__.__name__}({super().__repr__()})"

    @classmethod
    def _wrap(cls, value):
        if isinstance(value, dict):
            return cls(value)
        if isinstance(value, list):
            return [cls._wrap(v) for v in value]
        if isinstance(value, tuple):
            return tuple(cls._wrap(v) for v in value)
        return value