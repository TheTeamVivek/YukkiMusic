<div align="center">

# 🌍 Languages

**Multi-language support for YukkiMusic via YAML files**

[![locales](https://img.shields.io/badge/languages-4-green?style=flat-square)](#-available-languages)
[![embed](https://img.shields.io/badge/auto-embedded-blue?style=flat-square)]

</div>

---

## 🌐 Available Languages

| Flag | Language | Code |
| :---: | :--- | :--- |
| 🇺🇸 | English | `en` |
| 🇮🇳 | हिन्दी | `hi` |
| 🇸🇦 | العربية | `ar` |
| 🇹🇷 | Türkçe | `tr` |

---

## ➕ How to Add a Language

**1 · Create a file**
Add `internal/locales/{code}.yml` — e.g. `es.yml` for Spanish.

**2 · Set the display name**
```yaml
name: "🇪🇸 Español"
```

**3 · Translate the keys**
Copy every key from `en.yml` and translate the values.

**4 · Done!**
The file is picked up automatically — **no code changes needed**.

---

## 🔤 Placeholders `{ }`

Any `{key}` inside a string is replaced at runtime with the value passed from code:

```yaml
addauth_success_with_count: "{emoji:✅} Added {user}.\nNow {count}/{limit} authorized users."
```

- `{user}`, `{count}`, `{limit}`, `{error}`, `{status}` etc. — names must match what the code passes.
- Missing translation → the key itself is shown (e.g. `[addauth_success_with_count]`).
- Fallback order: **your language → default language → `en`**.

---

## 🎨 Emoji Tokens `{emoji:...}`

Don't paste raw emojis directly — use `{emoji:X}` tokens instead.

**Case 1 — Predefined emoji**
If the emoji is already listed in the [`emojiTokens`](../locales/emoji.go) map, just put that emoji inside the token:

```yaml
btn: "{emoji:▶️} Resume"
```

**Case 2 — Any other (custom/premium) emoji**
If the emoji is **not** in that map, use its Telegram custom-emoji ID directly:

```yaml
btn: "{emoji:5234567890123456789} Resume"
```

> [!IMPORTANT]
> - Tokens are **language independent** — use the same token in every `.yml`.
> - Predefined tokens auto-expand to premium `<tg-emoji>` where mapped, otherwise plain emoji.

**Example**
```yaml
ping_text: |
  {emoji:⚡} <b>Pong!</b>
  Speed: {speed}s
```

---

## ⚙️ Change Default Language

Set `DEFAULT_LANG` in your `.env` using one of the available codes:

| Code | Language |
| :--- | :--- |
| `en` | 🇺🇸 English *(default)* |
| `hi` | 🇮🇳 हिन्दी |
| `ar` | 🇸🇦 العربية |
| `tr` | 🇹🇷 Türkçe |

```bash
DEFAULT_LANG=hi
```

---

<div align="center">

← [⚙️ Environment Variables](../../config/README.md)

</div>
