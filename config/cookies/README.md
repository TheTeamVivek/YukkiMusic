<div align="center">

# 🍪 YouTube Cookies

**Authenticate `yt-dlp` with YouTube — bypass age restrictions, region locks & bot detection**

[![batbin](https://img.shields.io/badge/setup-batbin-blue?style=flat-square)](https://batbin.me)
[![refresh](https://img.shields.io/badge/refresh-every%202--4%20days-orange?style=flat-square)]

</div>

---

## 🚀 Method 1 — Remote via Batbin <sub><i>recommended</i></sub>

Keeps cookies out of Git and makes updates easy.

**1 · Export cookies**
Use a browser extension like [Get cookies.txt](https://chrome.google.com/webstore/detail/get-cookiestxt-/) to export your YouTube cookies in **Netscape format**.

**2 · Upload to Batbin**
Paste the cookie text into [batbin.me](https://batbin.me) and create a paste.

**3 · Set the link**
Copy the paste link and add it to your `.env`:

```bash
COOKIES_LINK="https://batbin.me/paste_id"
```

> [!TIP]
> Multiple space-separated URLs are supported for load balancing:
> `COOKIES_LINK="https://batbin.me/id1 https://batbin.me/id2"`

---

## 📁 Method 2 — Local File

Place any Netscape-formatted `.txt` cookie file directly into `config/cookies/`:

```
config/cookies/
├── account1.txt
└── account2.txt
```

The bot automatically picks a **random** `.txt` file each time.

---

## 🔧 Troubleshooting

<details>
<summary><b>"Sign in to confirm you're not a bot"</b></summary>

Your cookies have expired or were exported incorrectly.
<br>→ Re-export fresh cookies from your browser and update the Batbin paste or local file.
</details>

<details>
<summary><b>"No cookie files found"</b></summary>

<br>→ Verify `COOKIES_LINK` is correct, or that your local file has a `.txt` extension.
</details>

---

## 💡 Tips

- 🔒 **Security** — never commit real cookies to Git; use `COOKIES_LINK` or keep them local.
- 🔄 **Refresh** — YouTube cookies expire fast; update every **2–4 days** or when downloads fail.
- 📦 **Redundancy** — use cookies from multiple accounts to reduce rate limits.

---

<div align="center">

← [⚙️ Environment Variables](../README.md)

</div>
