Configuration
=============

Environment Setup
-----------------
  
Here's an Example of ``sample.env``

.. code-block:: bash
  
   API_ID = 
   API_HASH = 
   BOT_TOKEN = 
   MONGO_DB_URI = 
   LOG_GROUP_ID = 
   OWNER_ID = 
   STRING_SESSIONS = 
  
- Copy the ``example.env`` file to ``.env``
   .. code-block:: bash

      cp sample.env .env


- **Fill in the necessary variables in your ``.env`` :**

   - **API_ID**: The API ID obtained from `**Telegram** <https://my.telegram.org/auth>`_.
   - **API_HASH**: The API hash obtained from `Telegram <https://my.telegram.org/auth>`_.
   - **BOT_TOKEN**: The bot token you get from `BotFather <https://core.telegram.org/bots#botfather>`_.
   - **MONGO_DB_URI**: The MongoDB connection string for your database.
   - **LOG_GROUP_ID**: The ID of the group to which logs should be sent (set to `0` if no logs are needed).
   - **OWNER_ID**: The  User ID That will treated as the Owner of the bot, Multiple Id can be separated with a space.
   - **STRING_SESSIONS**: Your session string (usually generated via a method like Pyrogram) or You can generate from `Telegram Tools <https://telegram.tools/session-string-generator#pyrogram>`_ And make sure you environment is **Production** don't use **Test**.

You also can fill these belows Available Vars 

Available Vars
--------------

Here is a List of all Available Vars of `YukkiMusic <https://github.com/TheTeamVivek/YukkiMusic>`_.

Config vars are basically the variables which configure or modify bot to function, which are the basic necessities of plugins or code to work. You have to set the proper mandatory vars to make it functional and to start the basic feature of bot.

Mandatory Vars
^^^^^^^^^^^^^^

- API_ID & API_HASH
   - Go to `my.telegram.org <https://my.telegram.org/auth>`_ then Enter your Phone Number with your country code.

   - After, you are logged in click on API Development Tools.

   - Enter Anything as App name and App short name, Enter `my.telegram.org` in url section

   - That’s it, You'll get your API_ID and API_HASH.

- BOT_TOKEN
   Get token of the bot from the `@Botfather <https://t.me/Botfather>`_ in Telegram

- LOG_GROUP_ID
   .. important::

      You'll need a **Group** for this. 

      Remember to add your Music Bot , Assistant Accounts and Logger Id in Group and Promote them Admin with Full Rights.
   - Add `@MissRose_Bot <https://t.me/MissRose_Bot>`_ in your Group from Add Member > Search ``@MissRose_Bot`` and then Add.

   - After added, Just type ``/id`` in the chat.

   - You'll get the ID of your group.

- OWNER_ID
   .. note::

      Value must be an integer like 0123456789

      You can add multiple userid seperated with a Space

      .. code-block:: bash
         :caption: Example of Multiple Userid

          OWNER_ID = 1234567890 2345617890 8790654321 6578012347

   Your user id (not username) Get it by using command /id on the Group in the reply to your message where Rose Bot was added.

- STRING_SESSIONS
   A list of Pyrogram String Session seperated with comma "," of a Telegram Account which will be joining Group Calls for streaming.

   Your session string (usually generated via a method like Pyrogram) or You can generate from `Telegram Tools <https://telegram.tools/session-string-generator#pyrogram>`_ And make sure you environment is **Production** donn't use **Test**.

   .. code-block:: bash
         :caption: Example of Multiple String Sessions

          STRING_SESSIONS = string1,  string2, string3,  string4

   Like this as your mood you can add multiple String sessions of Your assistant for multiple Assistsant.

- MONGO_DB_URI
       Not a mandatory var, but yes kind off.
   .. note::

      Yukki no longer requires MONGO DB as mandatory. Leave it blank and bot will use Yukki’s database for your bot. Seperate database and Easy to use.

      To maintain bot’s privacy you wont be able to manage sudoers.  Bot will create an separate collection for you and no other bot's database will clash with it.

- COOKIE_LINK
   .. important::

      This is **not a mandatory** variable, but it is necessary for the bot to play songs perfectly due to YouTube verification.  

      Without cookies, the bot may be unable to download songs.  

   **How to obtain COOKIE_LINK:**  

   1. Get your YouTube cookies.  

   2. Go to `batbin.me <https://batbin.me>`_.

   3. Paste your cookies and tap on **Save**.  

   4. Copy the generated URL.  

   5. Set ``COOKIE_LINK`` with this URL.  

   **Alternative Method:**  

   If you don't want to use a URL, you can manually add cookies:  

   1. Navigate to `config/cookies/ <https://github.com/TheTeamVivek/YukkiMusic/tree/dev/config/cookies>`_.  

   2. Create a ``.txt`` file.  

   3. Paste your cookies inside the file.  

   This ensures smooth song playback without verification issues.  

   .. seealso::
  
      Don't know how to get cookies? See :doc:`cookies` This properly

.. toctree::
   :maxdepth: 2
   :hidden:

   cookies