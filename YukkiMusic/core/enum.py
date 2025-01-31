from enum import Enum, auto


class PlayType(Enum):
    PLAYING = auto()  # Actively playing
    PAUSED = auto()  # Paused
    MUTED = auto()  # Muted
    SPEEDUPED = auto()  # Playing at a faster speed


class SongType(Enum):
    VIDEO = auto()
    AUDIO = auto()


class SourceType(Enum):
    APPLE = auto()
    RESSO = auto()
    SAAVN = auto()
    SOUNDCLOUD = auto()
    SPOTIFY = auto()
    TELEGRAM = auto()
    YOUTUBE = auto()
