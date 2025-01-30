from abc import ABC, abstractmethod

from YukkiMusic.utils.formatters import time_to_seconds

from ..core.enum import SourceType
from ..core.youtube import Track

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

    @abstractmethod
    async def track(self, url: str) -> Track:
        pass