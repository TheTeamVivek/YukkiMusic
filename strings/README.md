# yukki Multi-Language Support

- The following languages are currently supported in yukkimusic. You can edit or change all strings available.

| Code | Language | Contributor |
|------|----------|-------------|
| en   | English  | Thanks to [Teamyukki](https://t.me/Teamyukki) |
| ku   | Kurdish  | Thanks to [Mohammed](https://t.me/IQ7amo) |
| hi   | Hindi    | Thanks to [Teamyukki](https://t.me/Teamyukki) |
| ar   | Arabic   | Thanks to [Mustafa](https://t.me/tr_4z) |
| tr   | Turkish  | Thanks to [Teamyukki](https://t.me/Teamyukki) |
| as   | Assamese | Thanks to [Mungnem Chiranji](https://t.me/ChiranjibKoch) |

---

### Help Us Add More Languages to yukkimusic! How to Contribute?

1. Translate using the enhanced web-based translator  
   Use [**TranslateIt**](https://vivekkumar-in.github.io/translateit), an enhanced web-based translator created with the help of [**lovable AI ❤️**](https://loveable.dev) — to easily and quickly translate yukkimusic without editing files manually.

---

If you prefer manual translation, follow the steps below:

1. **Edit the language file manually**  
   Start by editing the [`en.yml`](https://github.com/TheTeamVivek/yukkimusic/blob/master/strings%2Flangs%2Fen.yml) file, which contains the English language strings. Translate it into your language.

2. **Submit your translation**  
   Once you've completed your translation, send the edited file to us at [@TheTeamVk](https://t.me/TheTeamVk) or open a pull request in our GitHub repository.

---

### Points to Remember While Editing:

- **Do Not Modify Placeholders:**  
  Placeholders like `{0}`, `{1}`, etc., should remain unchanged, as they are used for dynamic text rendering.

- **Maintain Key Consistency:**  
  Keys such as `"general_1"` or others in the file should not be renamed or modified.

---

### Translating Bot Commands (Optional)

If you want to localize the bot commands in your language, you can do so by editing the `commands.yml` file.

Unlike `en.yml`, this file maps each command to language codes directly. Here’s how it looks:

#### Before:

```yaml
START_COMMAND:
  ar: ["تفعيل", "ميوزك"]
  ku: ["دەستپێکردن", "چالاککردن"]
  en: ["start"]
  tr: ["baslat"]
```

To add new translations, simply follow this format:

```yaml
COMMAND_NAME:
  isolangkey: ["translated_command_1", "translated_command_2"]
```

For example, to add **Hindi** translations to `START_COMMAND`:

```yaml
START_COMMAND:
  ar: ["تفعيل", "ميوزك"]
  ku: ["دەستپێکردن", "چالاککردن"]
  en: ["start"]
  tr: ["baslat"]
  hi: ["शुरू", "चालू"]
```

Or to add French translations to a STOP command:

```yaml
STOP_COMMAND:
  fr: ["arrêter", "stop"]
```

> [!NOTE]
> Make sure you **append** new language keys to existing command blocks instead of creating duplicate `COMMAND_NAME` entries.

Save your changes in the same `commands.yml` file and include it when you submit your translation PR.

---

By contributing to yukkimusic translations, you help make it accessible to more users around the world.  
**Thank you for your support! ❤️**