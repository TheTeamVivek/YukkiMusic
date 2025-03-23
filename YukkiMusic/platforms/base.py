from abc import ABC, abstractmethod

from ..core.youtube import Track


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

        Args:
            url (str): The URL of the track.

        Returns:
            Track: An instance of the Track class representing the track.
        """

    async def playlist(self, url: str) -> list[Track, str]:
        """
        Retrieves a playlist from the given URL.

        Args:
            url (str): The URL of the playlist.

        Returns:
            list[Track, str]: A list where:
                - The first element is a `Track` instance representing the main track.
                - The other element depends on the platform:
                    - If the platform is YouTube, it will be the video ID (`vidid`).
                    - For other streaming services, it will be the song name.
        """
