from enum import Enum, auto


class PlayType(Enum):
    PLAYING = auto()  # Actively playing
    PAUSED = auto()  # Paused
    MUTED = auto()  # Muted


class SongType(Enum):
    VIDEO = auto()
    AUDIO = auto()


class SourceType(Enum):
    APPLE = "Apple"
    RESSO = "Resson"
    SAAVN = "JioSaavn"
    SOUNDCLOUD = "Soundcloud"
    SPOTIFY = "Spotify"
    TELEGRAM = "Telegram"
    YOUTUBE = "YouTube"
