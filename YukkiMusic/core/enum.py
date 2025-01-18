from enum import Enum, auto


class PlaybackState(Enum):
    PLAYING = auto()  # Actively playing
    PAUSED = auto()  # Paused
    MUTED = auto()  # Muted
    UNMUTED = auto()  # Unmuted
    SPEEDUPED = auto()  # Playing at a faster speed


class SongType(Enum):
    TRACK = auto()
    PLAYLIST = auto()


class SourceType(Enum):
    APPLE = auto()
    RESSO = auto()
    SAAVN = auto()
    SOUNDCLOUD = auto()
    SPOTIFY = auto()
    TELEGRAM = auto()
    YOUTUBE = auto()
