# TODO NECESSARY

## 1.

```python

                    if popped.get("mystic"):
                        try:
                            await popped.get("mystic").delete()
                        except Exception:
                            pass
```

Merge this in Cleanmode That delete the message with **CLEANMODE_DELETE_MINS**

## 2.

Implement a env var or using db that help we can disable the Cleanmode and enable it


## 3.
  on M3u8 mode Track return a medium quality audio + video url add support for seperate high qulaity audio and video

## 4.
 
  On livestream every time returns audio url thats cause video playback not work add support for videp by returing or processing the audio and video url
  
## 5. Add this yml files and amd thier translations

```yml
admin_21 : "❌ Failed to reload admincache make sure bot is an admin in your chat"
admin_36: "⏭️ Music is skipped by {}!"
admin_37: "🔁 Music replayed by {0} !"
admin_38: "Bot is unable to seek because duration exceeds.\n\nCurrently played:** {0}** minutes out of **{1}** minutes."

cplay_7: "Channel Play Disabled"
play_20: "Not a live stream"
```

## 6.
  If Possible so remove the /gstats,  it uses userdb and chattopdb that consumes too Database and still useless


## 7.

  Now we can say that YukkiMusic fully support multiples langauge with command, enable, disable so we need to refactor all languages
   


# FEAT CAN BE ADDED [ OPTIONAL ]

### Use **defaultdict** as a replacement of pythons dict

# Dev [contextlib]
Note

Both redirect_stdout() and redirect_stderr() modify global state by replacing objects in the sys module, and should be used with care. The functions are not thread-safe, and may interfere with other operations that expect the standard output streams to be attached to terminal devices