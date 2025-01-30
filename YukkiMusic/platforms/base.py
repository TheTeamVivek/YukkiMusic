from abc import ABC, abstractmethod

from YukkiMusic.utils.formatters import time_to_seconds

from ..core.enum import SourceType 

class TrackDetails:
    def __init__(self, track_info: dict, type: SourceType):
        self.title = track_info.get("title")
        self.link = track_info.get("link")
        self.vidid = track_info.get("vidid")
        self.duration_min = track_info.get("duration_min")
        self.duration_sec = track_info.get("duration_sec", None)
        self.thumb = track_info.get("thumb")

        try:
            if self.duration_sec is None and self.duration_min is not None:
                self.duration_sec = time_to_seconds(self.duration_min)
        except ValueError:
            self.duration_sec = 0
        self.type = type

class PlatformBase(ABC):

    @abstractmethod
    async def valid(self, link: str) -> bool:
        """
        Validates whether the given URL matches the expected format for this service.

        Args:
            url (str): The URL to validate.

        Returns:
            bool: True if the URL is valid for this service, False otherwise.
        """