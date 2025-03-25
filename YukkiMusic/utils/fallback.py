# Radhe Radhe

from .decorators.asyncify import asyncify


@asyncify
def fallback_download():
    pass


async def fallback_track(name):
    pass
