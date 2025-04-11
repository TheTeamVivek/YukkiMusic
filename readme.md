# 🎵 **YukkiMusicFlex** 🎶

[**YukkiMusic**](https://github.com/TheTeamVivek/YukkiMusic) is an enhanced version of the original [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot), optimized for high-quality music streaming in Telegram voice chats. Built with Python and Pyrogram. 🚀

## ⭐ Support the Original
If you're using or forking this project, please show your support by starring the original repo:
[**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot)


## 🚀 Quick Deployment Options

### Deploy on Heroku
Get started quickly by deploying to Heroku with just one click:

<a href="https://dashboard.heroku.com/new?template=https://github.com/TheTeamVivek/YukkiMusic">
  <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-red?style=for-the-badge&logo=heroku" width="200"/>
</a>

### 🖥️ VPS Deployment Guide
  - Update and Install Dependencies: `sudo apt update && sudo apt upgrade -y && sudo apt install -y ffmpeg git python3-pip python3-venv tmux nano`
  - wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
  - sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list'
  - sudo apt update
  - sudo apt install -y google-chrome-stable
  - sudo apt-get install xvfb -y
  - Xvfb :1 -screen 0 1024x768x24 &
  - google-chrome-stable --version # too see if google chrome installed

  - Create the Virtual Environment: `python3 -m venv /root/myenv`

  - Activate Virtual Env: `source /root/myenv/bin/activate`

  - Clone the Repository: `git clone https://github.com/FLEX-GHOST/YukkiMusicFlex && cd YukkiMusicFlex`

  - Install Python Requirements: `pip install -r requirements.txt`

  - Copy and Edit Environment Variables:

    Copy the sample environment file: `cp sample.env .env`

    Edit the variables in the .env file: `nano .env`

  After editing, press `Ctrl+X`, then `Y`, and press **Enter** to save the changes.

  - to run boot and Never Stop create screen with ( screen -S FlexMusic ) And when you wanna back to the screen ( screen -R FlexMusic )  ( source /root/myenv/bin/activate )  ( Run the Bot: `bash start` )
  - and to run ( cookies_manger ) create screen for it with ( screen -S cookies ) And when you wanna back to the screen ( screen -R cookies ) run it ( source /root/myenv/bin/activate ) ( python3 cookie_manager.py ) 

  -  Run the Bot: `bash start`

  - Keep the Bot Running with tmux: `tmux`

To exit the **tmux session** without stopping the bot, press `Ctrl+a`, then `d`.



## ⚙️ Configuration

Need help setting up? Check out our detailed configuration guide: [**Configuration Instructions**](https://github.com/TheTeamVivek/YukkiMusic/blob/master/config/README.md).


## 🤝 Need Help?

We're here to support you through multiple channels:

- [**📝 Open a GitHub Issue**](https://github.com/TheTeamVivek/YukkiMusic/issues/new?assignees=&labels=question&title=support%3A+&body=%23+Support+Question)

- [**💬 Contact Us**](https://t.me/TheTeamVk)

- [**👥 Join Support Group**](https://t.me/TheTeamVk)


## ❣️ Show Your Support

Love YukkiMusic? Here's how you can help:

- ⭐ [**Star the YukkiMusicBot Project**](https://github.com/TeamYukki/YukkiMusicBot).

- 🍴 [**Fork and and contribute to the this Repository**](https://github.com/TheTeamVivek/YukkiMusic)

- 📢 Share your experience on [**Dev.to**](https://dev.to/), [**Medium**](https://medium.com/), or your **personal blog.**

Together, we can make [**YukkiMusic**](https://github.com/TheTeamVivek/YukkiMusic) and [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot) even better!

## 🙏 Special Thanks

A heartfelt thanks to [**Team Yukki**](https://github.com/TeamYukki) for creating the original [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot), which, although now inactive, served as the foundation for this project.  

A special thanks to [**Pranav-Saraswat**](https://github.com/Pranav-Saraswat) for forking and reviving it as [**YukkiMusicFork**](https://github.com/Pranav-Saraswat/YukkiMusicFork), making the bot functional again. However, **YukkiMusicFork** has since been deleted by Pranav.  

Our current project, [**YukkiMusic**](https://github.com/TheTeamVivek/YukkiMusic), is an imported and further improved version of the now-deleted **YukkiMusicFork**.
