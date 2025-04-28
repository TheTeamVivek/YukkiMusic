from abc import ABC, abstractmethod

from ..core.track import Track


class PlatformBase(ABC):
    @abstractmethod
    async def valid(self, link: str) -> bool:
        """
        Checks if the given link is a valid URL for this platform.

        Args:
            link (str): The URL to validate.

        Returns:
            bool: True if the URL is valid for this platform, False otherwise.
        """

    @abstractmethod
    async def track(self, url: str) -> Track:
        """
        Retrieves a track instance from the given URL.
        if the url of the playform but of playlist other kind of track, album etc, so must return a list of elements where first element must be a `Track` instance

        Args:
            url (str): The URL of the track.

        Returns:
            Track: An instance of the Track class representing the track.
        """
