# 🎵 **YukkiMusic** 🎶

[**YukkiMusic**](https://github.com/TheTeamVivek/YukkiMusic) is a powerful, enhanced version of the original [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot), designed for seamless, high-quality music streaming in Telegram voice chats. Built with **Python** and **Pyrogram**, it offers a robust and user-friendly experience for music lovers and bot developers alike. 🚀


## ⚙️ Configuration

Need help setting up? Check out our detailed configuration guide: [**Configuration Instructions**](https://github.com/TheTeamVivek/YukkiMusic/blob/master/config/README.md).

> [!TIP]
> **Looking to use cookies for authentication?**  
> See: [**Using Cookies for Authentication**](https://github.com/TheTeamVivek/YukkiMusic/blob/master/config/README.md#using-cookies-for-authentication)

## Quick Deployment Options

## Deploy on Heroku
Get started quickly by deploying to Heroku with just one click:

<a href="https://dashboard.heroku.com/new?template=https://github.com/TheTeamVivek/YukkiMusic">
  <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-red?style=for-the-badge&logo=heroku" width="200"/>
</a>

### 🖥️ VPS Deployment Guide

- **Update System and Install Dependencies**:  
  ```bash
  sudo apt update && sudo apt upgrade -y && sudo apt install -y ffmpeg git python3-pip tmux nano
  ```

- **Install uv for Efficient Dependency Management**:
  ```bash
  pip install --upgrade uv
  ```


- **Clone the Repository:**  
  ```bash
  git clone https://github.com/TheTeamVivek/YukkiMusic && cd YukkiMusic
  ```
  

- **Create and Activate a Virtual Environment:**
  - You can create and activate the virtual Environment before cloning the repo.
  ```bash
  uv venv .venv && source .venv/bin/activate
  ```

- Install Python Requirements:  
  ```bash
  uv pip install -e .
  ```

- Copy and Edit Environment Variables:  
  ```bash
  cp sample.env .env && nano .env
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

## 🤝 Get Support

We're here to help you every step of the way! Reach out through:

- **📝 GitHub Issues**: Report bugs or ask questions by [**opening an issue**](https://github.com/TheTeamVivek/YukkiMusic/issues/new?assignees=&labels=question&title=support).

- **💬 Telegram Support**: Connect with us on [**Telegram**](https://t.me/TheTeamVk).

- **👥 Support Channel**: Join our community at
 [**TheTeamVivek**](https://t.me/TheTeamVivek).


## ⭐ Support the Original
Show your love for the project that started it all! If you're using or forking **YukkiMusic**, please **star** the original repository: [**⭐ YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot)


## ❣️ Show Your Support

Love YukkiMusic? Help us grow the project with these simple actions:

- **⭐ Star the Original:** Give a star to [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot).
  
- **🍴 Fork & Contribute**: Dive into the code and contribute to [**YukkiMusic**](https://github.com/TheTeamVivek/YukkiMusic).

- **📢 Spread the Word**: Share your experience on [**Dev.to**](https://dev.to/), [**Medium**](https://medium.com/), or your personal blog.

Together, we can make **YukkiMusic** and **YukkiMusicBot** even better!

## 🙏 Acknowledgments 

A huge thank you to [**Team Yukki**](https://github.com/TeamYukki) for creating the original [**YukkiMusicBot**](https://github.com/TeamYukki/YukkiMusicBot), the foundation of this project. Though the original is now inactive, its legacy lives on.

Special gratitude to [**Pranav-Saraswat**](https://github.com/Pranav-Saraswat) for reviving the project with [**YukkiMusicFork**](https://github.com/Pranav-Saraswat/YukkiMusicFork) (now deleted), which inspired YukkiMusic.

**YukkiMusic** is an imported and enhanced version of the now-deleted **YukkiMusicFork**, with ongoing improvements to deliver the best music streaming experience.
