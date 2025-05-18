# 🎵 **YukkiMusic** 🎶

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

- Update and Install Dependencies:  
  ```bash
  sudo apt update && sudo apt upgrade -y && sudo apt install -y ffmpeg git python3-pip tmux nano
  ```

- Install uv:  
  ```bash
  pip install --upgrade uv
  ```


- Clone the Repository:  
  ```bash
  git clone https://github.com/TheTeamVivek/YukkiMusic && cd YukkiMusic
  ```
  

- Create the Virtual Environment:  
  - You can create and activate the virtual Environment before cloning the repo.
  ```bash
  uv venv
  ```

- Activate the Virtual Environment:  
  ```bash
  source .venv/bin/activate
  ```

- Install Python Requirements:  
  ```bash
  uv pip install -e .
  ```

- Copy and Edit Environment Variables:  
  ```bash
  cp sample.env .env
  nano .env
  ```
  After editing, press `Ctrl+X`, then `Y`, and press **Enter** to save the changes.

- Start a tmux Session to Keep the Bot Running:  
  ```bash
  tmux
  ```

- Run the Bot:  
  ```bash
  yukkimusic
  ```

- Detach from the **tmux** Session (Bot keeps running):  
  Press `Ctrl+b`, then `d`


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
