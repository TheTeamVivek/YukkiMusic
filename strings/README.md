# Yukki Multi-Language Support

- The following languages are currently supported in YukkiMusic. You can edit or change all strings available.

| Code | Language | Contributor |
|------|----------|-------------|
| en   | English  | Thanks to [TeamYukki](https://t.me/TeamYukki) |
| ku   | Kurdish  | Thanks to [Mohammed](https://t.me/IQ7amo) |
| hi   | Hindi    | Thanks to [TeamYukki](https://t.me/TeamYukki) |
| ar   | Arabic   | Thanks to [Mustafa](https://t.me/tr_4z) |
| tr   | Turkish  | Thanks to [TeamYukki](https://t.me/TeamYukki) |
| as   | Assamese | Thanks to [Mungnem Chiranji](https://t.me/ChiranjibKoch) |

---

### Help Us Add More Languages to YukkiMusic! How to Contribute?

1. Translate using the enhanced web-based translator  
   Use [**TranslateIt**](https://vivekkumar-in.github.io/translateit), an enhanced web-based translator created with the help of [**lovable AI ❤️**](https://loveable.dev) — to easily and quickly translate YukkiMusic without editing files manually.

---

If you prefer manual translation, follow the steps below:

1. **Edit the language file manually**  
   Start by editing the [`en.yml`](https://github.com/TheTeamVivek/YukkiMusic/blob/master/strings%2Flangs%2Fen.yml) file, which contains the English language strings. Translate it into your language.

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

For example, to add Hindi commands for `START_COMMAND`:

```yaml
START_COMMAND:
  hi: ["शुरू", "चालू"]
```

Save your changes in the same `commands.yml` file and include it when you submit your translation.

---

By contributing to YukkiMusic translations, you help make it accessible to more users around the world.  
**Thank you for your support! ❤️**
