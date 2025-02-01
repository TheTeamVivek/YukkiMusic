Configuration
=============

Environment Setup
-----------------
  
Here's an Example Sample.env

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

You also can fill these belows Available Vars 

Available Vars
--------------

Here is a List of all Available Vars of `YukkiMusic <https://github.com/TheTeamVivek/YukkiMusic>`_.

Config vars are basically the variables which configure or modify bot to function, which are the basic necessities of plugins or code to work. You have to set the proper mandatory vars to make it functional and to start the basic feature of bot.

Mandatory Vars
^^^^^^^^^^^^^^

- API_ID & API_HASH

   - Go to my.telegram.org then Enter your Phone Number with your country code.

   - After, you are logged in click on API Development Tools.

   - Enter Anything as App name and App short name, Enter my.telegram.org in url section

   - That’s it, You'll get your API_ID and API_HASH.

- BOT_TOKEN

   Get it from @Botfather in Telegram

- LOG_GROUP_ID
   .. note::

      You'll need a Group for this. 

      Remember to add your Music Bot , Assistant Accounts and Logger Id in Group and Promote them Admin with Full Rights.
   - Add @MissRose_Bot in your Group from Add Member > Search "@MissRose_Bot" and then Add.

   - After added, Just type "/id" in the chat.

   - You'll get the ID of your group.

- OWNER_ID

   .. note::

      Value must be an integer like 0123456789

      You can add multiple userid seperated with a Space

      .. code-block:: bash
         :caption: Example of Multiple Userid

          OWNER_ID = 1234567890 2345617890 8790654321 6578012347

Your user id (not username) Get it by using command /id on the Group in the reply to your message where Rose Bot was added.
