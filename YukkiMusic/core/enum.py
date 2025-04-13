from enum import Enum, auto


class PlayType(Enum):
    PLAYING = auto()  # playing
    PAUSED = auto()  # Paused
    MUTED = auto()  # Muted

class SpotifyType(Enum):
    TRACK = auto()
    PLAYLIST = auto()
    ALBUM = auto()
    ARTIST = auto()

class SourceType(Enum):
    APPLE = "Apple"
    RESSO = "Resson"
    SAAVN = "JioSaavn"
    SOUNDCLOUD = "Soundcloud"
    SPOTIFY = SpotifyType
    TELEGRAM = "Telegram"
    YOUTUBE = "YouTube"
    M3U8 = "M3U8 Urls"
