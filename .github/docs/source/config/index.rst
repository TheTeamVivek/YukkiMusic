YukkiMusic Configuration
========================

Environment Setup
-----------------
  
Here's an Example Sample.env
^^^^^^^^^^^^^^^^^^^^^^^^^^^^
.. code-block:: bash
  
   API_ID = 
   API_HASH = 
   BOT_TOKEN = 
   MONGO_DB_URI = 
   LOG_GROUP_ID = 
   OWNER_ID = 
   STRING_SESSIONS = 
  
- Copy the `example.env` file to `.env`
   .. code-block:: bash

      cp sample.env .env


- **Fill in the necessary variables in your .env :**
  - **API_ID**: The API ID obtained from `Telegram <https://my.telegram.org/auth>`_.
  - **API_HASH**: The API hash obtained from `Telegram <https://my.telegram.org/auth>`_.
  - **BOT_TOKEN**: The bot token you get from `BotFather <https://core.telegram.org/bots#botfather>`_.
  - **MONGO_DB_URI**: The MongoDB connection string for your database.
  - **LOG_GROUP_ID**: The ID of the group to which logs should be sent (set to `0` if no logs are needed).
  - **OWNER_ID**: The ID of the bot owner.
  - **STRING_SESSIONS**: Your session string (usually generated via a method like Pyrogram) or You can generate from `Telegram Tools <https://telegram.tools/session-string-generator#pyrogram>`_ And make sure you environment is Production Don't use Test.

Available Vars
--------------

Here is a List of all Available Vars of `YukkiMusic <https://github.com/TheTeamVivek/YukkiMusic>`_.

Config vars are basically the variables which configure or modify bot to function, which are the basic necessities of plugins or code to work. You have to set the proper mandatory vars to make it functional and to start the basic feature of bot.

1. API_HASH 

2. API_ID 

3. ASSISTANT_LEAVE_TIME 

4. AUTO_LEAVING_ASSISTANT 

5. BOT_TOKEN 

6. CLEANMODE_MINS 

7. DURATION_LIMIT 

8. GIT_TOKEN 

9. GITHUB_REPO 

10. GLOBAL_IMG_URL 

11. HEROKU_API_KEY 

12. HEROKU_APP_NAME 

13. LOG_GROUP_ID 

14. MONGO_DB_URI 

15. OWNER_ID 

16. PING_IMG_URL 

17. PLAYLIST_FETCH_LIMIT 

18. PLAYLIST_IMG_URL 

19. PRIVATE_BOT_MODE 

20. SERVER_PLAYLIST_LIMIT 

21. SONG_DOWNLOAD_DURATION_LIMIT 

22. SOUNCLOUD_IMG_URL 

23. SPOTIFY_ALBUM_IMG_URL 

24. SPOTIFY_ARTIST_IMG_URL 

25. SPOTIFY_CLIENT_ID 

26. SPOTIFY_CLIENT_SECRET 

27. SPOTIFY_PLAYLIST_IMG_URL 

28. START_IMG_URL 

29. STATS_IMG_URL 

30. STREAM_IMG_URL 

31. STRING_SESSIONS

32. SUPPORT_CHANNEL 

33. SUPPORT_GROUP 

34. TELEGRAM_AUDIO_URL 

35. TELEGRAM_EDIT_SLEEP 

36. TELEGRAM_VIDEO_URL 

37. TG_AUDIO_FILESIZE_LIMIT 

38. TG_VIDEO_FILESIZE_LIMIT 

39. UPSTREAM_BRANCH 

40. UPSTREAM_REPO 

41. VIDEO_STREAM_LIMIT 

42. YOUTUBE_EDIT_SLEEP 

42. YOUTUBE_IMG_URL
