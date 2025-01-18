YukkiMusic Changelog
====================

All notable changes to YukkiMusic will be documented in this file.

v1.3 (2024-11-15)
-----------------
### Added
- Added the ``asyncify`` decorator to convert synchronous functions to asynchronous.
- Added the ``OWNER`` command that will show in the log group or bot PM/DM.

### Removed
- Deprecated the ``oauth2`` method, which is no longer supported. This functionality has been merged into ``yt-dlp``.

### Fixed
- Various bug fixes and performance improvements.

**Full Changelog:** `v1.2...v1.3 <https://github.com/TheTeamVivek/YukkiMusic/compare/v1.2...v1.3>`

v1.2 (2024-11-03)
-----------------
### Added
- Added Multiple Languages Support for commands
- Multiple language support for bot Help Menu (Only for primary plugins, not for external plugins)
- All commands can be used without prefix (Except English commands)
- User can request her data and delete (Except: Chat, Banned Users, Blacklist Chats)
- ``PRIVACY.md`` for YukkiMusic

### Changed
- ``Apple``, ``Carbon``, ``Saavn``, ``Resso``, ``SoundCloud``, ``Spotify``, ``Telegram``, ``YouTube`` are centralized to a class [PlaTForms](https://github.com/TheTeamVivek/YukkiMusic/blob/master/YukkiMusic%2Fplatforms%2F__init__.py)
- Explained Privacy policy in ``/privacy`` command
- Now Assistant will join chat when the chat is private
- Now User Friendly ``README.md``

**Full Changelog:** `v1.1...v1.2 <https://github.com/TheTeamVivek/YukkiMusic/compare/v1.1...v1.2>`

v1.1 (2024-10-14)
-----------------
### Added
- Unlimited assistant support for handling multiple voice chats
- MongoDB Data Export/import Support
- Added JioSaavn Playback support
- Added ``yt-dlp-youtube-oauth2`` to bypass Sign-in Issue
- The currently playing message will be deleted when switching to the next track.

### Changed
- Updated Python Version to 3.12.7-slim
- Improved error handling in music playback
- Enhanced queue management system
- Better formatting for duration display
- Optimized database operations

### Fixed
- Delete files after streams end
- Updated ``langs/en.yml`` standardized to use English letters instead of mini caps.
- Commands are now sourced from ``command.yml``. Any updates to plugin commands will automatically update the help message.

### Removed
- Some unused plugins: ``vars.py``, ``groupass.py``, ``player.py``
- ``Assets`` folder due to lack of use
- Unused dependencies from ``requirements.txt``

**Full Changelog:** `v1.0...v1.1 <https://github.com/TheTeamVivek/YukkiMusic/compare/v1.0...v1.1>`

v1.0 (2024-10-05)
-----------------
- Initial release of YukkiMusic
- Thanks to `Pranav-Saraswat <https://github.com/Pranav-Saraswat>`_ for their YukkiMusicFork for making it work
- Thanks to `TeamYukki <https://github.com/TeamYukki/>`_ for their `YukkiMusicBot <https://github.com/TeamYukki/YukkiMusicBot>`

### Features
- High quality music streaming
- Video streaming capability
- Interactive inline buttons
- Detailed playback statistics
- Group management commands
- Customizable bot settings

### Notes
- Base version established with core functionality
- Compatible with Python 3.9+
- Built with Pyrogram and py-tgcalls
