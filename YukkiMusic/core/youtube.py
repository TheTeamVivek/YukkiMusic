from dataclasses import dataclass, field
from YukkiMusic.utils.formatters import seconds_to_min, time_to_seconds

@dataclass
class Track:
    title: str
    vidid: str
    link: str
    thumb: str
    duration_min: int | None = field(default=None)
    duration_sec: int | None = field(default=None)

    def __post_init__(self):
        if self.duration_min is not None and self.duration_sec is None:
            self.duration_sec = time_to_seconds(self.duration_min)
        elif self.duration_sec is not None and self.duration_min is None:
            self.duration_min = seconds_to_min(self.duration_sec)

class YouTube:
    pass
